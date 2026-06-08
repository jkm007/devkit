#!/bin/bash
# ============================================
# 后端部署脚本
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/../env.sh"

log_info "========== 开始部署后端 =========="

# 1. 编译后端
log_info "编译后端..."
cd "$SCRIPT_DIR/../../../backend-server"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o backend-server cmd/server/main.go
log_info "编译完成: $(ls -lh backend-server | awk '{print $5}')"

# 2. 创建远程目录
log_info "创建远程目录..."
ssh ${APP_USER}@${APP_HOST} "mkdir -p ${APP_DIR}/backend/config ${APP_DIR}/backend/logs"

# 3. 上传文件
log_info "上传后端程序..."
scp backend-server ${APP_USER}@${APP_HOST}:${APP_DIR}/backend/

log_info "上传配置文件..."
scp config/config.yaml ${APP_USER}@${APP_HOST}:${APP_DIR}/backend/config/

# 4. 修改远程配置
log_info "更新远程配置..."
ssh ${APP_USER}@${APP_HOST} << 'REMOTE_SCRIPT'
cd /opt/devkit/backend/config

# 使用 sed 更新配置
sed -i "s|host: .*|host: ${DB_HOST}|" config.yaml
sed -i "s|password: .*|password: \"${DB_PASSWORD}\"|" config.yaml
REMOTE_SCRIPT

# 5. 创建/更新 systemd 服务
log_info "配置 systemd 服务..."
ssh ${APP_USER}@${APP_HOST} << 'SERVICE_SCRIPT'
cat > /etc/systemd/system/devkit-backend.service << 'EOF'
[Unit]
Description=DevKit Backend Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/opt/devkit/backend/backend-server
WorkingDirectory=/opt/devkit/backend
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
SERVICE_SCRIPT

# 6. 重启服务
log_info "重启后端服务..."
ssh ${APP_USER}@${APP_HOST} "systemctl restart devkit-backend && sleep 2 && systemctl status devkit-backend --no-pager"

# 7. 健康检查
log_info "健康检查..."
sleep 3
HTTP_CODE=$(ssh ${APP_USER}@${APP_HOST} "curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:${BACKEND_PORT}/health")
if [ "$HTTP_CODE" = "200" ]; then
    log_info "✅ 后端部署成功！"
else
    log_error "❌ 健康检查失败，HTTP 状态码: $HTTP_CODE"
    ssh ${APP_USER}@${APP_HOST} "journalctl -u devkit-backend --no-pager -n 20"
    exit 1
fi
