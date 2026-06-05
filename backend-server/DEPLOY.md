# 部署指南

## 服务器信息

- 服务器 IP：123.57.201.44
- 数据库地址：114.215.190.52
- 数据库用户：root / root123456
- Redis 地址：114.215.190.52:6379

## 目录结构

```
/opt/devkit/
├── frontend/              # 前端静态文件
│   ├── index.html
│   ├── js/
│   ├── css/
│   └── ...
├── backend/               # 后端程序
│   ├── backend-server     # 可执行文件
│   └── config/
│       └── config.yaml    # 配置文件
```

## 编译部署

### 1. 编译后端（静态编译避免 GLIBC 问题）

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o backend-server cmd/server/main.go
```

### 2. 上传到服务器

```bash
# 上传程序
scp backend-server root@服务器IP:/opt/devkit/backend/

# 上传配置（首次部署）
scp config/config.yaml root@服务器IP:/opt/devkit/backend/config/
```

### 3. 修改配置文件

在服务器上修改 `/opt/devkit/backend/config/config.yaml`：

```yaml
mysql:
  host: 114.215.190.52  # 数据库地址

redis:
  host: 114.215.190.52  # Redis 地址

cors:
  allow_origins:
    - "http://服务器IP"
    - "http://服务器IP:80"

server:
  mode: release  # 生产环境
```

## Systemd 服务

### 服务文件

`/etc/systemd/system/devkit-backend.service`

```ini
[Unit]
Description=DevKit Backend Server
After=network.target

[Service]
Type=simple
ExecStart=/opt/devkit/backend/backend-server
WorkingDirectory=/opt/devkit/backend
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### 服务管理

```bash
# 启用服务
systemctl enable devkit-backend

# 启动服务
systemctl start devkit-backend

# 查看状态
systemctl status devkit-backend

# 重启服务
systemctl restart devkit-backend

# 查看日志
journalctl -u devkit-backend -f
```

## 数据库初始化

### 从 108 数据库导出完整数据

```bash
mysqldump -h 10.0.50.108 -u root -p'root123456' backend_db > backend_db_full.sql
```

### 导入到目标数据库

```bash
# 创建数据库
mysql -h 114.215.190.52 -u root -p'root123456' -e "DROP DATABASE IF EXISTS backend_db; CREATE DATABASE backend_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入数据
mysql -h 114.215.190.52 -u root -p'root123456' backend_db < backend_db_full.sql
```

## 默认账号

| 用户名 | 密码 | 角色 | 说明 |
|--------|------|------|------|
| vben | 123456 | super | 超级管理员，全部权限 |
| admin | 123456 | admin | 管理员，系统管理权限 |
| jack | 123456 | user | 普通用户，无权限 |

## Nginx 配置

`/etc/nginx/nginx.conf`

```nginx
server {
    listen 80;
    server_name _;
    root /opt/devkit/frontend;

    # API 路径
    location ^~ /api/ {
        rewrite ^/api/(.*) /$1 break;
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # 前端静态文件
    location / {
        limit_except GET HEAD {
            deny all;
        }
        index index.html;
        try_files $uri $uri/ /index.html;
    }
}
```

## API 端点

- 前端请求：`/api/auth/login`
- Nginx 代理：去掉 `/api` 前缀
- 后端实际：`/auth/login`

## 访问地址

- 前端：http://服务器IP
- API：http://服务器IP/api/
- Swagger：http://服务器IP/api/swagger/
- 健康检查：http://服务器IP/api/health