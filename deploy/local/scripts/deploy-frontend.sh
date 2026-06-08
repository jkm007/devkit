#!/bin/bash
# ============================================
# 前端部署脚本 - 本地环境
# ============================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/../env.sh"

log_info "========== 开始部署前端 =========="

# 1. 编译前端
log_info "编译前端..."
cd "$SCRIPT_DIR/../../../frontend-admin"
pnpm build
log_info "编译完成"

# 2. 打包
log_info "打包前端文件..."
cd apps/web-antd/dist
tar czf /tmp/devkit-frontend.tar.gz .
log_info "打包大小: $(ls -lh /tmp/devkit-frontend.tar.gz | awk '{print $5}')"

# 3. 上传
log_info "上传前端文件..."
scp /tmp/devkit-frontend.tar.gz ${APP_USER}@${APP_HOST}:/tmp/

# 4. 解压部署
log_info "解压部署..."
ssh ${APP_USER}@${APP_HOST} << 'DEPLOY_SCRIPT'
# 备份旧版本
if [ -d /opt/devkit/frontend ]; then
    mv /opt/devkit/frontend /opt/devkit/frontend.bak.$(date +%Y%m%d%H%M%S)
fi

# 创建目录并解压
mkdir -p /opt/devkit/frontend
cd /opt/devkit/frontend
tar xzf /tmp/devkit-frontend.tar.gz
rm -f /tmp/devkit-frontend.tar.gz

# 设置权限
chmod -R 755 /opt/devkit/frontend
DEPLOY_SCRIPT

# 5. 清理旧备份（保留最近 3 个）
log_info "清理旧备份..."
ssh ${APP_USER}@${APP_HOST} << 'CLEANUP_SCRIPT'
cd /opt/devkit
ls -dt frontend.bak.* 2>/dev/null | tail -n +4 | xargs rm -rf
CLEANUP_SCRIPT

log_info "✅ 前端部署成功！"
