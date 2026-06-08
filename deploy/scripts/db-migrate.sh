#!/bin/bash
# ============================================
# 数据库迁移脚本
# 用于同步表结构和基础数据
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

# 显示帮助
show_help() {
    echo "用法: $0 <环境> [操作]"
    echo ""
    echo "环境:"
    echo "  aliyun    - 阿里云环境"
    echo "  local     - 本地环境"
    echo ""
    echo "操作:"
    echo "  migrate   - 执行迁移（默认）"
    echo "  seed      - 初始化基础数据"
    echo "  sync      - 同步所有（迁移+基础数据）"
    echo "  verify    - 验证数据库状态"
    echo ""
    echo "示例:"
    echo "  $0 local migrate"
    echo "  $0 aliyun seed"
    echo "  $0 local sync"
}

# 检查参数
if [ -z "$1" ]; then
    show_help
    exit 1
fi

ENV="$1"
ACTION="${2:-migrate}"

# 加载环境配置
if [ "$ENV" = "aliyun" ]; then
    source "$DEPLOY_DIR/aliyun/env.sh"
elif [ "$ENV" = "local" ]; then
    source "$DEPLOY_DIR/local/env.sh"
else
    log_error "未知环境: $ENV"
    show_help
    exit 1
fi

# 统一使用网络连接数据库
DB_CMD="mysql -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p'${DB_PASSWORD}' ${DB_NAME}"

log_info "=========================================="
log_info "  环境: $ENV"
log_info "  操作: $ACTION"
log_info "=========================================="

# 执行迁移
do_migrate() {
    log_step "执行表结构迁移..."

    # 读取迁移文件
    MIGRATION_DIR="$SCRIPT_DIR/../../backend-server/migrations"

    for sql_file in "$MIGRATION_DIR"/*.sql; do
        if [ -f "$sql_file" ]; then
            filename=$(basename "$sql_file")
            log_info "执行迁移: $filename"

            mysql -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p"${DB_PASSWORD}" ${DB_NAME} < "$sql_file" 2>/dev/null || true
        fi
    done

    log_info "✅ 表结构迁移完成"
}

# 初始化基础数据
do_seed() {
    log_step "初始化基础数据..."

    # 1. 存储桶管理 - 同步默认桶
    log_info "同步存储桶配置..."

    # 检查 sys_storage_bucket 表是否存在
    table_exists=$(eval "$DB_CMD -e \"SHOW TABLES LIKE 'sys_storage_bucket';\"" 2>/dev/null | grep -c "sys_storage_bucket" || true)

    if [ "$table_exists" = "0" ]; then
        log_warn "sys_storage_bucket 表不存在，跳过存储桶同步"
    else
        # 同步 MinIO 默认桶（从 sys_system_settings 读取）
        eval "$DB_CMD -e \"
            INSERT INTO sys_storage_bucket (name, driver, endpoint, bucket, access_key, secret_key, use_ssl, purpose, is_default, status, description)
            SELECT
                'MinIO 默认桶',
                'minio',
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'minio_endpoint' AND group_key = 'storage'), ''),
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'minio_bucket' AND group_key = 'storage'), 'devkit'),
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'minio_access_key' AND group_key = 'storage'), ''),
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'minio_secret_key' AND group_key = 'storage'), ''),
                0,
                'file',
                0,
                1,
                '由部署脚本自动创建的 MinIO 默认桶'
            FROM DUAL
            WHERE NOT EXISTS (SELECT 1 FROM sys_storage_bucket WHERE driver = 'minio' AND name = 'MinIO 默认桶');
        \"" 2>/dev/null || log_warn "MinIO 桶同步失败（可能已存在）"

        # 同步本地默认桶
        eval "$DB_CMD -e \"
            INSERT INTO sys_storage_bucket (name, driver, purpose, is_default, status, description)
            SELECT '本地默认存储', 'local', 'file', 1, 1, '系统内置的本地文件存储'
            FROM DUAL
            WHERE NOT EXISTS (SELECT 1 FROM sys_storage_bucket WHERE driver = 'local' AND name = '本地默认存储');
        \"" 2>/dev/null || log_warn "本地桶同步失败（可能已存在）"

        # 同步 OSS 桶（如果配置了）
        eval "$DB_CMD -e \"
            INSERT INTO sys_storage_bucket (name, driver, endpoint, bucket, access_key, secret_key, purpose, is_default, status, description)
            SELECT
                '阿里云 OSS 桶',
                'oss',
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'oss_endpoint' AND group_key = 'storage'), ''),
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'oss_bucket' AND group_key = 'storage'), ''),
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'oss_access_key_id' AND group_key = 'storage'), ''),
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'oss_access_key_secret' AND group_key = 'storage'), ''),
                'file',
                0,
                0,
                '阿里云 OSS 存储桶（未启用）'
            FROM DUAL
            WHERE NOT EXISTS (SELECT 1 FROM sys_storage_bucket WHERE driver = 'oss' AND name = '阿里云 OSS 桶');
        \"" 2>/dev/null || log_warn "OSS 桶同步失败（可能已存在）"

        # 同步 COS 桶（如果配置了）
        eval "$DB_CMD -e \"
            INSERT INTO sys_storage_bucket (name, driver, region, bucket, access_key, secret_key, purpose, is_default, status, description)
            SELECT
                '腾讯云 COS 桶',
                'cos',
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'cos_region' AND group_key = 'storage'), ''),
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'cos_bucket' AND group_key = 'storage'), ''),
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'cos_secret_id' AND group_key = 'storage'), ''),
                COALESCE((SELECT value FROM sys_system_settings WHERE key = 'cos_secret_key' AND group_key = 'storage'), ''),
                'file',
                0,
                0,
                '腾讯云 COS 存储桶（未启用）'
            FROM DUAL
            WHERE NOT EXISTS (SELECT 1 FROM sys_storage_bucket WHERE driver = 'cos' AND name = '腾讯云 COS 桶');
        \"" 2>/dev/null || log_warn "COS 桶同步失败（可能已存在）"

        log_info "✅ 存储桶配置同步完成"
    fi

    # 2. 标签管理 - 同步默认标签
    log_info "同步默认标签..."

    tag_table_exists=$(eval "$DB_CMD -e \"SHOW TABLES LIKE 'sys_tag';\"" 2>/dev/null | grep -c "sys_tag" || true)

    if [ "$tag_table_exists" != "0" ]; then
        eval "$DB_CMD -e \"
            INSERT INTO sys_tag (tag_key, tag_value, tag_name, description, created_at, updated_at)
            VALUES
                ('type', 'image', '图片', '图片类型文件', NOW(), NOW()),
                ('type', 'video', '视频', '视频类型文件', NOW(), NOW()),
                ('type', 'audio', '音频', '音频类型文件', NOW(), NOW()),
                ('type', 'document', '文档', '文档类型文件', NOW(), NOW()),
                ('type', 'archive', '压缩包', '压缩包类型文件', NOW(), NOW()),
                ('source', 'user', '用户上传', '用户手动上传的文件', NOW(), NOW()),
                ('source', 'system', '系统生成', '系统自动生成的文件', NOW(), NOW()),
                ('sensitivity', 'public', '公开', '公开访问的文件', NOW(), NOW()),
                ('sensitivity', 'private', '私有', '私有文件', NOW(), NOW())
            ON DUPLICATE KEY UPDATE tag_name = VALUES(tag_name);
        \"" 2>/dev/null || log_warn "标签同步失败"

        log_info "✅ 默认标签同步完成"
    fi

    # 3. 路由规则 - 同步默认路由
    log_info "同步默认路由规则..."

    routing_table_exists=$(eval "$DB_CMD -e \"SHOW TABLES LIKE 'sys_tag_routing';\"" 2>/dev/null | grep -c "sys_tag_routing" || true)

    if [ "$routing_table_exists" != "0" ]; then
        eval "$DB_CMD -e \"
            INSERT INTO sys_tag_routing (name, conditions, match_type, driver, bucket, path_prefix, priority, is_default, status, description, created_at, updated_at)
            SELECT
                '图片存储到 MinIO',
                '{\\\"tags\\\":[{\\\"key\\\":\\\"type\\\",\\\"value\\\":\\\"image\\\"}]}',
                'all',
                'minio',
                'devkit',
                'images/',
                100,
                0,
                1,
                '图片文件自动存储到 MinIO 的 images 目录',
                NOW(),
                NOW()
            FROM DUAL
            WHERE NOT EXISTS (SELECT 1 FROM sys_tag_routing WHERE name = '图片存储到 MinIO');
        \"" 2>/dev/null || log_warn "路由规则同步失败"

        log_info "✅ 默认路由规则同步完成"
    fi

    log_info "✅ 基础数据初始化完成"
}

# 验证数据库状态
do_verify() {
    log_step "验证数据库状态..."

    echo ""
    echo "=== 表结构 ==="
    eval "$DB_CMD -e 'SHOW TABLES;'" 2>/dev/null

    echo ""
    echo "=== 存储桶配置 ==="
    eval "$DB_CMD -e 'SELECT id, name, driver, is_default, status FROM sys_storage_bucket;'" 2>/dev/null || echo "表不存在"

    echo ""
    echo "=== 存储配置 ==="
    eval "$DB_CMD -e \"SELECT \\\`key\\\`, value FROM sys_system_settings WHERE group_key = 'storage' AND (\\\`key\\\` LIKE '%enabled%' OR \\\`key\\\` = 'storage_driver');\"" 2>/dev/null || echo "无存储配置"

    echo ""
    echo "=== 标签统计 ==="
    eval "$DB_CMD -e 'SELECT COUNT(*) as total FROM sys_tag;'" 2>/dev/null || echo "表不存在"

    echo ""
    echo "=== 路由规则统计 ==="
    eval "$DB_CMD -e 'SELECT COUNT(*) as total FROM sys_tag_routing;'" 2>/dev/null || echo "表不存在"
}

# 执行操作
case "$ACTION" in
    migrate)
        do_migrate
        ;;
    seed)
        do_seed
        ;;
    sync)
        do_migrate
        do_seed
        ;;
    verify)
        do_verify
        ;;
    *)
        log_error "未知操作: $ACTION"
        show_help
        exit 1
        ;;
esac

log_info "=========================================="
log_info "  操作完成！"
log_info "=========================================="
