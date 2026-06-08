#!/bin/bash
# ============================================
# 数据库备份脚本
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/../env.sh"

BACKUP_DIR="$SCRIPT_DIR/../../../backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/backend_db_${TIMESTAMP}.sql"

log_info "========== 开始备份数据库 =========="

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 备份数据库
log_info "备份数据库 ${DB_NAME}..."
mysqldump -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p'${DB_PASSWORD}' \
    --single-transaction --routines --triggers --events \
    ${DB_NAME} > "$BACKUP_FILE"

# 压缩
log_info "压缩备份文件..."
gzip "$BACKUP_FILE"

# 统计
BACKUP_SIZE=$(ls -lh "${BACKUP_FILE}.gz" | awk '{print $5}')
log_info "备份完成: ${BACKUP_FILE}.gz (${BACKUP_SIZE})"

# 清理旧备份（保留最近 7 个）
log_info "清理旧备份..."
cd "$BACKUP_DIR"
ls -t backend_db_*.sql.gz 2>/dev/null | tail -n +8 | xargs rm -f

log_info "✅ 数据库备份完成！"
