#!/bin/bash
# ============================================
# 从阿里云同步数据库脚本
# 支持增量同步和全量同步
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"
source "$DEPLOY_DIR/env.sh"

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

# 阿里云配置
ALIYUN_DB_HOST="114.215.190.52"
ALIYUN_DB_PORT="3306"
ALIYUN_DB_USER="root"
ALIYUN_DB_PASSWORD="root123456"

# 显示帮助
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --full        全量同步（默认）"
    echo "  --incremental 增量同步（只同步新表）"
    echo "  --tables      只同步指定表（用逗号分隔）"
    echo "  --dry-run     只显示要同步的表，不执行"
    echo "  --help        显示帮助"
    echo ""
    echo "示例:"
    echo "  $0                          # 全量同步"
    echo "  $0 --incremental            # 增量同步"
    echo "  $0 --tables sys_storage_bucket,sys_tag  # 只同步指定表"
}

# 解析参数
SYNC_MODE="full"
TABLES=""
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --full)
            SYNC_MODE="full"
            shift
            ;;
        --incremental)
            SYNC_MODE="incremental"
            shift
            ;;
        --tables)
            TABLES="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

log_info "=========================================="
log_info "  从阿里云同步数据库"
log_info "=========================================="
log_info "源数据库: ${ALIYUN_DB_HOST}:${ALIYUN_DB_PORT}"
log_info "目标数据库: ${DB_HOST}:${DB_PORT}"
log_info "同步模式: ${SYNC_MODE}"
echo ""

# 确认同步
if [ "$DRY_RUN" = false ]; then
    read -p "确认同步? 这将覆盖本地数据! (y/n): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_warn "同步已取消"
        exit 0
    fi
fi

# 获取要同步的表
get_tables() {
    if [ -n "$TABLES" ]; then
        echo "$TABLES"
        return
    fi

    if [ "$SYNC_MODE" = "incremental" ]; then
        # 获取阿里云有但本地没有的表
        ALIYUN_TABLES=$(mysql -h ${ALIYUN_DB_HOST} -P ${ALIYUN_DB_PORT} -u ${ALIYUN_DB_USER} -p"${ALIYUN_DB_PASSWORD}" ${DB_NAME} -e "SHOW TABLES;" 2>/dev/null | tail -n +2 | sort)
        LOCAL_TABLES=$(mysql -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p"${DB_PASSWORD}" ${DB_NAME} -e "SHOW TABLES;" 2>/dev/null | tail -n +2 | sort)

        # 找出差异表
        NEW_TABLES=$(comm -23 <(echo "$ALIYUN_TABLES") <(echo "$LOCAL_TABLES"))
        echo "$NEW_TABLES" | tr '\n' ',' | sed 's/,$//'
    else
        # 全量同步：获取所有表
        mysql -h ${ALIYUN_DB_HOST} -P ${ALIYUN_DB_PORT} -u ${ALIYUN_DB_USER} -p"${ALIYUN_DB_PASSWORD}" ${DB_NAME} -e "SHOW TABLES;" 2>/dev/null | tail -n +2 | tr '\n' ',' | sed 's/,$//'
    fi
}

TABLE_LIST=$(get_tables)

if [ -z "$TABLE_LIST" ]; then
    log_info "没有需要同步的表"
    exit 0
fi

log_info "要同步的表: ${TABLE_LIST}"

if [ "$DRY_RUN" = true ]; then
    log_info "DRY RUN 模式，不执行同步"
    exit 0
fi

# 执行同步
log_step "开始同步..."

IFS=',' read -ra TABLE_ARRAY <<< "$TABLE_LIST"
for table in "${TABLE_ARRAY[@]}"; do
    table=$(echo "$table" | xargs)  # trim whitespace
    if [ -z "$table" ]; then
        continue
    fi

    log_info "同步表: ${table}"

    # 导出表结构和数据
    DUMP_FILE="/tmp/${table}_sync_$(date +%Y%m%d_%H%M%S).sql"
    mysqldump -h ${ALIYUN_DB_HOST} -P ${ALIYUN_DB_PORT} -u ${ALIYUN_DB_USER} -p"${ALIYUN_DB_PASSWORD}" \
        --single-transaction --routines --triggers --events \
        ${DB_NAME} "${table}" > "$DUMP_FILE" 2>/dev/null

    # 导入到本地
    mysql -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p"${DB_PASSWORD}" ${DB_NAME} < "$DUMP_FILE" 2>/dev/null

    # 清理
    rm -f "$DUMP_FILE"

    log_info "✅ ${table} 同步完成"
done

# 同步后执行基础数据初始化
log_step "初始化基础数据..."
bash "$DEPLOY_DIR/../scripts/db-migrate.sh" local seed

log_info "=========================================="
log_info "  数据库同步完成！"
log_info "=========================================="
