#!/bin/bash
# ============================================
# Nginx 配置部署脚本
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/../env.sh"

log_info "========== 部署 Nginx 配置 =========="

# 1. 上传配置文件
log_info "上传 Nginx 配置..."
scp "$SCRIPT_DIR/../configs/nginx.conf" ${APP_USER}@${APP_HOST}:/etc/nginx/nginx.conf
scp "$SCRIPT_DIR/../configs/devkit.conf" ${APP_USER}@${APP_HOST}:/etc/nginx/conf.d/devkit.conf

# 2. 测试配置
log_info "测试 Nginx 配置..."
ssh ${APP_USER}@${APP_HOST} "nginx -t"

# 3. 重载配置
log_info "重载 Nginx..."
ssh ${APP_USER}@${APP_HOST} "systemctl reload nginx"

# 4. 验证
log_info "验证 Nginx 状态..."
ssh ${APP_USER}@${APP_HOST} "systemctl status nginx --no-pager | head -5"

log_info "✅ Nginx 部署成功！"
