#!/bin/bash
# ============================================
# 完整部署脚本
# 支持阿里云和本地环境
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_DIR="$(dirname "$DEPLOY_DIR")"

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
    echo "用法: $0 <环境> [组件]"
    echo ""
    echo "环境:"
    echo "  aliyun    - 阿里云环境"
    echo "  local     - 本地环境"
    echo ""
    echo "组件（可选，默认全部部署）:"
    echo "  backend   - 只部署后端"
    echo "  frontend  - 只部署前端"
    echo "  db        - 只执行数据库迁移"
    echo "  all       - 全部部署（默认）"
    echo ""
    echo "示例:"
    echo "  $0 local"
    echo "  $0 aliyun backend"
    echo "  $0 local db"
}

# 检查参数
if [ -z "$1" ]; then
    show_help
    exit 1
fi

ENV="$1"
COMPONENT="${2:-all}"

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

log_info "=========================================="
log_info "  环境: $ENV"
log_info "  组件: $COMPONENT"
log_info "=========================================="

# 编译后端
build_backend() {
    log_step "编译后端..."
    cd "$PROJECT_DIR/backend-server"

    if [ "$ENV" = "aliyun" ]; then
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o backend-server cmd/server/main.go
    else
        go build -o backend-server cmd/server/main.go
    fi

    log_info "✅ 后端编译完成: $(ls -lh backend-server | awk '{print $5}')"
}

# 编译前端
build_frontend() {
    log_step "编译前端..."
    cd "$PROJECT_DIR/frontend-admin"

    pnpm build

    log_info "✅ 前端编译完成"
}

# 部署后端
deploy_backend() {
    log_step "部署后端..."

    build_backend

    # 创建远程目录
    ssh ${APP_USER}@${APP_HOST} "mkdir -p ${APP_DIR}/backend/config ${APP_DIR}/backend/logs"

    # 上传文件
    scp "$PROJECT_DIR/backend-server/backend-server" ${APP_USER}@${APP_HOST}:${APP_DIR}/backend/
    scp "$PROJECT_DIR/backend-server/config/config.yaml" ${APP_USER}@${APP_HOST}:${APP_DIR}/backend/config/

    # 更新配置
    if [ "$ENV" = "aliyun" ]; then
        ssh ${APP_USER}@${APP_HOST} "cd ${APP_DIR}/backend/config && sed -i 's|host: .*|host: ${DB_HOST}|' config.yaml && sed -i 's|password: .*|password: \"${DB_PASSWORD}\"|' config.yaml"
    fi

    # 创建 systemd 服务
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

    # 重启服务
    ssh ${APP_USER}@${APP_HOST} "systemctl restart devkit-backend"

    # 健康检查
    sleep 3
    HTTP_CODE=$(ssh ${APP_USER}@${APP_HOST} "curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:${BACKEND_PORT}/health")

    if [ "$HTTP_CODE" = "200" ]; then
        log_info "✅ 后端部署成功！"
    else
        log_error "❌ 健康检查失败，HTTP 状态码: $HTTP_CODE"
        ssh ${APP_USER}@${APP_HOST} "journalctl -u devkit-backend --no-pager -n 20"
        exit 1
    fi
}

# 部署前端
deploy_frontend() {
    log_step "部署前端..."

    build_frontend

    # 创建远程目录
    ssh ${APP_USER}@${APP_HOST} "mkdir -p ${APP_DIR}/frontend"

    # 上传文件
    scp -r "$PROJECT_DIR/frontend-admin/apps/web-antd/dist/"* ${APP_USER}@${APP_HOST}:${APP_DIR}/frontend/

    log_info "✅ 前端部署成功！"
}

# 执行数据库迁移
deploy_db() {
    log_step "执行数据库迁移..."

    bash "$SCRIPT_DIR/db-migrate.sh" "$ENV" sync

    log_info "✅ 数据库迁移完成！"
}

# 执行操作
case "$COMPONENT" in
    backend)
        deploy_backend
        ;;
    frontend)
        deploy_frontend
        ;;
    db)
        deploy_db
        ;;
    all)
        deploy_db
        deploy_backend
        deploy_frontend
        ;;
    *)
        log_error "未知组件: $COMPONENT"
        show_help
        exit 1
        ;;
esac

log_info "=========================================="
log_info "  部署完成！"
log_info "=========================================="
