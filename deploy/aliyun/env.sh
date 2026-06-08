#!/bin/bash
# ============================================
# 阿里云环境配置
# ============================================

# 应用服务器
APP_HOST="123.57.201.44"
APP_USER="root"
APP_DIR="/opt/devkit"

# 数据库服务器
DB_HOST="114.215.190.52"
DB_PORT="3306"
DB_USER="root"
DB_PASSWORD="root123456"
DB_NAME="backend_db"

# Redis 服务器
REDIS_HOST="114.215.190.52"
REDIS_PORT="6379"
REDIS_PASSWORD="redis123456"

# MinIO 服务器
MINIO_HOST="114.215.190.52"
MINIO_API_PORT="9000"
MINIO_CONSOLE_PORT="9001"
MINIO_USER="minioadmin"
MINIO_PASSWORD="minioadmin123"
MINIO_BUCKET="test"

# 后端配置
BACKEND_PORT="8080"
JWT_SECRET="your-jwt-secret-key-here"
AES_KEY="your-aes-32-byte-key-here!!!!"
CAPTCHA_SECRET="your-captcha-secret-key-here"

# 前端配置
FRONTEND_PORT="80"
VITE_GLOB_API_URL="/api"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
