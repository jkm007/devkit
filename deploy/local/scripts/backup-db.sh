#!/bin/bash
# ============================================
# 数据库备份脚本 - 本地环境
# 每次部署前自动备份，只备份不同步
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/../env.sh"

BACKUP_DIR="/opt/devkit/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql.gz"

log_info "========== 备份数据库 =========="

# 1. 创建备份目录
ssh ${APP_USER}@${APP_HOST} "mkdir -p ${BACKUP_DIR}"

# 2. 备份数据库（通过应用服务器连接数据库服务器）
log_info "备份数据库 ${DB_NAME}..."
ssh ${APP_USER}@${APP_HOST} "mysqldump -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p'${DB_PASSWORD}' \
    --single-transaction --routines --triggers --events \
    ${DB_NAME} 2>/dev/null | gzip > ${BACKUP_FILE}"

BACKUP_SIZE=$(ssh ${APP_USER}@${APP_HOST} "ls -lh ${BACKUP_FILE} | awk '{print \$5}'")
log_info "备份完成: ${BACKUP_FILE} (${BACKUP_SIZE})"

# 3. 清理旧备份（保留最近 10 个）
log_info "清理旧备份..."
ssh ${APP_USER}@${APP_HOST} "cd ${BACKUP_DIR} && ls -t ${DB_NAME}_*.sql.gz 2>/dev/null | tail -n +11 | xargs rm -f"

REMAINING=$(ssh ${APP_USER}@${APP_HOST} "ls ${BACKUP_DIR}/${DB_NAME}_*.sql.gz 2>/dev/null | wc -l")
log_info "保留最近 ${REMAINING} 个备份"

log_info "✅ 数据库备份完成！"
