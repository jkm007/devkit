#!/bin/bash
# ============================================================
# DevKit 快速部署脚本 (跳过前端编译，使用已有产物)
# ============================================================

set -e

APP_HOST="123.57.201.44"
APP_USER="root"
APP_DIR="/opt/devkit"
BRANCH="main"

BACKEND_DIR="$APP_DIR/backend"
FRONTEND_ADMIN_DIR="$APP_DIR/frontend-admin"
FRONTEND_MOBILE_DIR="$APP_DIR/frontend-mobile"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# ==================== Step 1: 编译后端 ====================
log_info "========== Step 1: 编译后端 =========="
git checkout main
git pull origin main

cd backend-server
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o server ./cmd/server/main.go
log_info "后端编译完成: $(ls -lh server | awk '{print $5}')"
cd ..

# ==================== Step 2: 上传后端 ====================
log_info "========== Step 2: 上传后端 =========="
scp backend-server/server ${APP_USER}@${APP_HOST}:${BACKEND_DIR}/server
scp backend-server/config/config.yaml ${APP_USER}@${APP_HOST}:${BACKEND_DIR}/config.yaml
log_info "后端上传完成"

# ==================== Step 3: 编译管理后台 ====================
log_info "========== Step 3: 编译管理后台 =========="
cd frontend-admin

# 更新生产环境 API 地址
cat > apps/web-antd/.env.production << 'EOF'
VITE_GLOB_API_URL=/api
VITE_GLOB_WS_URL=ws://123.57.201.44/api/ws
VITE_GLOB_APP_TITLE=DevKit Admin
EOF

log_info "编译管理后台 (首次较慢，请耐心等待)..."
pnpm build
log_info "管理后台编译完成: $(du -sh apps/web-antd/dist | awk '{print $1}')"
cd ..

# ==================== Step 4: 编译移动端 ====================
log_info "========== Step 4: 编译移动端 H5 =========="
cd frontend-app

cat > .env.production << 'EOF'
VITE_API_BASE_URL=/api/v1
VITE_WS_URL=ws://123.57.201.44/api/v1/ws
VITE_ENV=production
EOF

log_info "编译移动端 H5..."
./node_modules/.bin/uni build
log_info "移动端编译完成: $(du -sh dist/build/h5 | awk '{print $1}')"
cd ..

# ==================== Step 5: 上传前端 ====================
log_info "========== Step 5: 上传前端文件 =========="
log_info "清空远程旧文件..."
ssh ${APP_USER}@${APP_HOST} "rm -rf ${FRONTEND_ADMIN_DIR}/* ${FRONTEND_MOBILE_DIR}/*"

log_info "上传管理后台..."
scp -r frontend-admin/apps/web-antd/dist/* ${APP_USER}@${APP_HOST}:${FRONTEND_ADMIN_DIR}/

log_info "上传移动端 H5..."
scp -r frontend-app/dist/build/h5/* ${APP_USER}@${APP_HOST}:${FRONTEND_MOBILE_DIR}/
log_info "前端上传完成"

# ==================== Step 6: 配置 Nginx ====================
log_info "========== Step 6: 配置 Nginx =========="

ssh ${APP_USER}@${APP_HOST} 'cat > /etc/nginx/conf.d/devkit.conf << '"'"'NGINX_EOF'"'"'
# ============================================================
# DevKit Nginx 配置
# 80   → 移动端 H5
# 8081 → 管理后台
# ============================================================

# --- 移动端 H5 (80 端口) ---
server {
    listen 80;
    listen [::]:80;
    server_name _;
    root /opt/devkit/frontend-mobile;

    client_max_body_size 200m;

    proxy_buffering on;
    proxy_buffer_size 256k;
    proxy_buffers 8 512k;
    proxy_busy_buffers_size 512k;

    location ^~ /api/v1/ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 600;
        proxy_send_timeout 600;
        proxy_read_timeout 600;
    }

    location ^~ /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 600;
        proxy_send_timeout 600;
        proxy_read_timeout 600;
    }

    location ^~ /minio/ {
        proxy_pass http://10.0.50.108:9000/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        index index.html;
        try_files $uri $uri/ /index.html;
    }
}

# --- 管理后台 (8081 端口) ---
server {
    listen 8081;
    listen [::]:8081;
    server_name _;
    root /opt/devkit/frontend-admin;

    client_max_body_size 200m;

    proxy_buffering on;
    proxy_buffer_size 256k;
    proxy_buffers 8 512k;
    proxy_busy_buffers_size 512k;

    location ^~ /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 600;
        proxy_send_timeout 600;
        proxy_read_timeout 600;
    }

    location ^~ /api/ws/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 600;
        proxy_send_timeout 600;
        proxy_read_timeout 600;
    }

    location ^~ /minio/ {
        proxy_pass http://10.0.50.108:9000/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        index index.html;
        try_files $uri $uri/ /index.html;
    }
}
NGINX_EOF'

log_info "测试 Nginx 配置..."
ssh ${APP_USER}@${APP_HOST} "nginx -t"

log_info "重载 Nginx..."
ssh ${APP_USER}@${APP_HOST} "systemctl reload nginx"

# ==================== Step 7: 重启后端 ====================
log_info "========== Step 7: 重启后端 =========="

ssh ${APP_USER}@${APP_HOST} << 'SSH_EOF'
pkill -f "./server" 2>/dev/null || true
sleep 2

cd /opt/devkit/backend
nohup ./server > /opt/devkit/backend/app.log 2>&1 &
sleep 3

if pgrep -f "./server" > /dev/null; then
    echo "✅ 后端启动成功 (PID: $(pgrep -f './server'))"
    curl -s http://127.0.0.1:8080/api/v1/health 2>/dev/null || echo "健康检查: 端口未响应"
else
    echo "❌ 后端启动失败"
    tail -30 /opt/devkit/backend/app.log
fi
SSH_EOF

# ==================== 完成 ====================
log_info "============================================================"
log_info "🎉 部署完成！"
log_info "============================================================"
log_info "移动端 H5:    http://${APP_HOST}"
log_info "管理后台:     http://${APP_HOST}:8081"
log_info "后端 API:     http://${APP_HOST}/api/v1"
log_info "============================================================"
