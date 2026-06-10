#!/bin/bash
# ============================================
# 一键部署脚本（备份 + Nginx + 后端 + 前端）
# 每次部署：备份数据 → 部署 Nginx → 部署后端 → 部署前端
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/../env.sh"

echo "============================================"
echo "  DevKit 本地环境部署"
echo "============================================"
echo ""
echo "应用服务器: ${APP_HOST}"
echo "数据库服务器: ${DB_HOST}"
echo "部署目录: ${APP_DIR}"
echo ""

read -p "确认部署? (y/n): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_warn "部署已取消"
    exit 0
fi

# 步骤 1: 备份数据库
log_info "步骤 1/4: 备份数据库"
bash "$SCRIPT_DIR/backup-db.sh"

# 步骤 2: 部署 Nginx
log_info "步骤 2/4: 部署 Nginx"
bash "$SCRIPT_DIR/deploy-nginx.sh"

# 步骤 3: 部署后端
log_info "步骤 3/4: 部署后端"
bash "$SCRIPT_DIR/deploy-backend.sh"

# 步骤 4: 部署前端
log_info "步骤 4/4: 部署前端"
bash "$SCRIPT_DIR/deploy-frontend.sh"

echo ""
echo "============================================"
echo "  ✅ 部署完成！"
echo "============================================"
echo ""
echo "访问地址: http://${APP_HOST}"
echo "API 地址: http://${APP_HOST}/api/"
echo "健康检查: http://${APP_HOST}/api/health"
echo ""
