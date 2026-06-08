#!/bin/bash
# ============================================
# Nginx 安装和配置脚本 - 本地环境
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/../env.sh"

log_info "========== 安装配置 Nginx =========="

# 1. 检查 Nginx 是否已安装
NGINX_INSTALLED=$(ssh ${APP_USER}@${APP_HOST} "nginx -v 2>&1 || echo 'not_installed'")

if [[ "$NGINX_INSTALLED" == *"not_installed"* ]]; then
    log_info "安装 Nginx..."
    ssh ${APP_USER}@${APP_HOST} << 'INSTALL_SCRIPT'
# 检测包管理器
if command -v dnf &> /dev/null; then
    dnf install -y nginx
elif command -v yum &> /dev/null; then
    yum install -y nginx
elif command -v apt-get &> /dev/null; then
    apt-get update && apt-get install -y nginx
else
    echo "无法安装 Nginx，请手动安装"
    exit 1
fi
INSTALL_SCRIPT
else
    log_info "Nginx 已安装: $NGINX_INSTALLED"
fi

# 2. 上传配置文件
log_info "上传 Nginx 配置..."
scp "$SCRIPT_DIR/../configs/nginx.conf" ${APP_USER}@${APP_HOST}:/etc/nginx/nginx.conf
scp "$SCRIPT_DIR/../configs/devkit.conf" ${APP_USER}@${APP_HOST}:/etc/nginx/conf.d/devkit.conf

# 3. 测试配置
log_info "测试 Nginx 配置..."
ssh ${APP_USER}@${APP_HOST} "nginx -t"

# 4. 启动/重载 Nginx
log_info "启动 Nginx..."
ssh ${APP_USER}@${APP_HOST} << 'NGINX_SCRIPT'
systemctl enable nginx
systemctl restart nginx
systemctl status nginx --no-pager | head -5
NGINX_SCRIPT

log_info "✅ Nginx 部署成功！"
