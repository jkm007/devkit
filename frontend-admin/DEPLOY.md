# 前端部署指南

## 编译

```bash
# 进入前端目录
cd frontend-admin

# 编译生产版本
pnpm build --filter @vben/web-antd
```

编译输出目录：`apps/web-antd/dist/`

## 配置说明

### 生产环境配置

文件：`apps/web-antd/.env.production`

```env
VITE_BASE=/

# API 地址（使用 nginx 代理，相对路径）
VITE_GLOB_API_URL=/api

# 压缩方式
VITE_COMPRESS=none

# 路由模式
VITE_ROUTER_HISTORY=hash

# 应用标题
VITE_APP_TITLE=系统管理
```

### 重要说明

- `VITE_GLOB_API_URL=/api` - 使用相对路径，通过 nginx 代理转发到后端
- 不要使用完整的 URL（如 `https://mock-napi.vben.pro/api`），否则会请求到错误的地址

## 上传部署

```bash
# 上传到服务器
scp -r apps/web-antd/dist/* root@服务器IP:/opt/devkit/frontend/
```

## Nginx 配置

前端需要配合 nginx 代理 API 请求：

```nginx
# API 路径 - 前端请求 /api/*，代理到后端去掉 /api 前缀
location ^~ /api/ {
    rewrite ^/api/(.*) /$1 break;
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}

# 前端静态文件
location / {
    index index.html;
    try_files $uri $uri/ /index.html;
}
```

## 路由说明

- 前端使用 hash 路由模式：`/#/dashboard`
- 所有路由请求都返回 `index.html`，由前端处理

## API 请求流程

```
前端请求: /api/auth/login
    ↓
Nginx 代理: rewrite ^/api/(.*) /$1 break
    ↓
后端接收: /auth/login
```

## 访问地址

- 登录页：http://服务器IP/#/auth/login
- Dashboard：http://服务器IP/#/dashboard

## 默认账号

| 用户名 | 密码   | 角色  |
| ------ | ------ | ----- |
| vben   | 123456 | super |
| admin  | 123456 | admin |
| jack   | 123456 | user  |

## 常见问题

### 405 Not Allowed

原因：nginx 静态文件配置捕获了 POST 请求

解决：使用 `^~` 前缀让 API location 优先匹配，或限制静态 location 只允许 GET/HEAD

### 请求到错误的 API 地址

原因：`.env.production` 中 `VITE_GLOB_API_URL` 配置错误

解决：改为 `/api`（相对路径）

### 页面空白

原因：nginx 没有正确配置前端路由 fallback

解决：确保 `try_files $uri $uri/ /index.html;` 配置正确
