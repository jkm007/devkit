#!/bin/bash
# ============================================
# 数据库恢复脚本
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/../env.sh"

BACKUP_FILE="$1"

if [ -z "$BACKUP_FILE" ]; then
    echo "用法: $0 <备份文件>"
    echo ""
    echo "可用备份:"
    ls -lh "$SCRIPT_DIR/../../../backups"/backend_db_*.sql.gz 2>/dev/null || echo "  无备份文件"
    exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
    log_error "备份文件不存在: $BACKUP_FILE"
    exit 1
fi

echo "============================================"
echo "  数据库恢复"
echo "============================================"
echo ""
echo "目标数据库: ${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo "备份文件: ${BACKUP_FILE}"
echo ""

read -p "⚠️  确认恢复? 这将覆盖现有数据! (y/n): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_warn "恢复已取消"
    exit 0
fi

log_info "开始恢复数据库..."

# 如果是 gz 文件，先解压
if [[ "$BACKUP_FILE" == *.gz ]]; then
    log_info "解压备份文件..."
    gunzip -c "$BACKUP_FILE" | mysql -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p'${DB_PASSWORD}' ${DB_NAME}
else
    mysql -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p'${DB_PASSWORD}' ${DB_NAME} < "$BACKUP_FILE"
fi

log_info "✅ 数据库恢复完成！"
