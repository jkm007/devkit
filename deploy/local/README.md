# 本地环境部署

## 服务器信息

| 服务 | IP | 端口 | 说明 |
|------|-----|------|------|
| 应用服务器 | 10.0.50.103 | 80, 8080 | Nginx + 后端 |
| 数据库 | 10.0.50.108 | 3306 | MySQL 8.0 (Docker) |
| Redis | 10.0.50.108 | 6379 | Redis 7 (Docker) |
| MinIO | 10.0.50.108 | 9000, 9001 | 对象存储 (Docker) |

## 目录结构

```
deploy/local/
├── env.sh                    # 环境变量配置
├── README.md                 # 本文档
├── configs/
│   ├── nginx.conf            # Nginx 主配置
│   ├── devkit.conf           # DevKit 站点配置
│   └── config.yaml           # 后端配置模板
└── scripts/
    ├── deploy-all.sh         # 一键部署（备份+部署）
    ├── deploy-backend.sh     # 部署后端
    ├── deploy-frontend.sh    # 部署前端
    ├── deploy-nginx.sh       # 安装部署 Nginx
    ├── backup-db.sh          # 备份数据库
    └── db-sync.sh            # 从阿里云同步数据库（首次使用）
```

## 快速开始

### 一键部署（推荐）

每次部署自动备份数据库，然后部署 Nginx + 后端 + 前端：

```bash
cd deploy/local
bash scripts/deploy-all.sh
```

### 单独部署

```bash
# 备份数据库
bash scripts/backup-db.sh

# 安装配置 Nginx（首次）
bash scripts/deploy-nginx.sh

# 部署后端
bash scripts/deploy-backend.sh

# 部署前端
bash scripts/deploy-frontend.sh
```

### 首次部署 - 从阿里云同步数据

```bash
bash scripts/db-sync.sh
```

## 部署流程

每次执行 `deploy-all.sh` 会按以下顺序执行：

1. **备份数据库** — 备份到 `/opt/devkit/backups/`，保留最近 10 个备份
2. **部署 Nginx** — 上传配置，重载 Nginx
3. **部署后端** — 编译 Go 程序，上传并重启 systemd 服务
4. **部署前端** — 编译 Vue 应用，上传到 `/opt/devkit/frontend/`

## 访问地址

- 前端：http://10.0.50.103
- API：http://10.0.50.103/api/
- 健康检查：http://10.0.50.103/api/health
- MinIO 控制台：http://10.0.50.108:9001

## 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| vben | 123456 | 超级管理员 |
| admin | 123456 | 管理员 |
| jack | 123456 | 普通用户 |
