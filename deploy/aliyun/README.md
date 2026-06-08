# 阿里云环境部署

## 服务器信息

| 服务 | IP | 端口 | 说明 |
|------|-----|------|------|
| 应用服务器 | 123.57.201.44 | 80, 8080 | Nginx + 后端 |
| 数据库 | 114.215.190.52 | 3306 | MySQL 8.0 |
| Redis | 114.215.190.52 | 6379 | Redis 7 |
| MinIO | 114.215.190.52 | 9000, 9001 | 对象存储 |

## 目录结构

```
deploy/aliyun/
├── env.sh                    # 环境变量配置
├── README.md                 # 本文档
├── configs/
│   ├── nginx.conf            # Nginx 主配置
│   ├── devkit.conf           # DevKit 站点配置
│   └── backend-config.yaml   # 后端配置模板
└── scripts/
    ├── deploy-all.sh         # 一键部署
    ├── deploy-backend.sh     # 部署后端
    ├── deploy-frontend.sh    # 部署前端
    ├── deploy-nginx.sh       # 部署 Nginx
    ├── db-backup.sh          # 数据库备份
    └── db-restore.sh         # 数据库恢复
```

## 快速开始

### 一键部署

```bash
cd deploy/aliyun
bash scripts/deploy-all.sh
```

### 单独部署

```bash
# 部署后端
bash scripts/deploy-backend.sh

# 部署前端
bash scripts/deploy-frontend.sh

# 部署 Nginx
bash scripts/deploy-nginx.sh
```

### 数据库操作

```bash
# 备份数据库
bash scripts/db-backup.sh

# 恢复数据库
bash scripts/db-restore.sh /path/to/backup.sql.gz
```

## 服务器目录结构

```
/opt/devkit/
├── frontend/              # 前端静态文件
│   ├── index.html
│   ├── assets/
│   └── ...
└── backend/
    ├── backend-server     # 后端程序
    ├── config/
    │   └── config.yaml    # 配置文件
    ├── logs/              # 日志目录
    └── uploads/           # 本地上传文件（如果使用 local 存储）
```

## 服务管理

```bash
# 后端服务
systemctl start devkit-backend
systemctl stop devkit-backend
systemctl restart devkit-backend
systemctl status devkit-backend
journalctl -u devkit-backend -f

# Nginx
systemctl start nginx
systemctl stop nginx
systemctl reload nginx
nginx -t
```

## 访问地址

- 前端：http://123.57.201.44
- API：http://123.57.201.44/api/
- 健康检查：http://123.57.201.44/api/health
- MinIO 控制台：http://114.215.190.52:9001

## 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| vben | 123456 | 超级管理员 |
| admin | 123456 | 管理员 |
| jack | 123456 | 普通用户 |
