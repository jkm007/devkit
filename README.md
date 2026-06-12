# DevKit - 后台管理系统

DevKit 是一个基于 **Go + Vue 3** 的全栈后台管理系统，采用前后端分离架构，包含 Go 后端服务、Vue 3 管理后台，以及一份只读的 Vue Vben Admin 上游源码参考。

## 项目结构

```text
devkit/
├── backend-server/     # Go 后端服务（Gin + GORM + MySQL + Redis）
├── frontend-admin/     # Vue 3 前端管理后台（基于 Vue Vben Admin，Ant Design Vue 版本）
├── vue-vben-admin/     # Vue Vben Admin 上游源码参考，只读，不建议直接修改
├── docs/               # 项目文档、表设计、系统管理说明
└── tmp/                # 临时文件
```

## 技术栈

### 后端：`backend-server/`

| 组件 | 选型 |
| --- | --- |
| 语言 | Go 1.25+ |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0 |
| 缓存 / 队列 | Redis 7.x |
| 日志 | Zap + Lumberjack |
| 配置 | Viper + YAML |
| 认证 | JWT Access Token + Refresh Token |
| 权限 | RBAC：User / Role / Permission / Group |
| WebSocket | Gorilla WebSocket |
| 文件存储 | 本地 / MinIO / OSS / COS |
| 验证码 | go-captcha + 自定义实现，答案 AES-GCM 加密 |
| 邮件 | SMTP，支持 HTML 验证码邮件模板 |
| 限流 | 令牌桶 + 数据库动态限流规则 |
| API 文档 | Swagger / swaggo |

### 前端：`frontend-admin/`

| 组件 | 选型 |
| --- | --- |
| 框架 | Vue 3 |
| 构建 | Vite 8 |
| 语言 | TypeScript 6 |
| UI | Ant Design Vue 4 |
| 状态管理 | Pinia |
| 样式 | Tailwind CSS 4 |
| 包管理 | pnpm 11 |
| 构建编排 | Turborepo |
| 国际化 | vue-i18n |

## 快速启动

### 1. 启动后端

```bash
cd backend-server

# 启动依赖服务：MySQL、Redis、MinIO
docker-compose up -d mysql redis minio

# 方式一：使用 Makefile
make run

# 方式二：没有 make 时直接运行
go run ./cmd/server/main.go
```

后端默认地址：

```text
http://localhost:8080
```

健康检查：

```text
http://localhost:8080/health
```

Swagger 文档：

```text
http://localhost:8080/swagger/index.html
```

> 注意：Swagger 路由仅在 `server.mode = debug` 时开放。配置文件为 `backend-server/config/config.yaml`。

常用后端命令：

```bash
cd backend-server
make run          # 运行服务
make build        # 构建二进制
make test         # 运行测试
make fmt          # gofmt 格式化
make swagger      # 重新生成 Swagger 文档
make migrate      # 执行数据库迁移
make docker-up    # 启动依赖服务
make docker-down  # 停止 Docker 服务
```

如果当前环境没有 `make`，可使用等价命令：

```bash
cd backend-server
go run ./cmd/server/main.go
go test ./...
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g ./cmd/server/main.go -o ./docs
```

### 2. 启动前端

```bash
cd frontend-admin
pnpm install
pnpm dev
```

前端默认地址：

```text
http://localhost:5666
```

常用前端命令：

```bash
cd frontend-admin
pnpm dev                         # 启动 web-antd 开发服务
pnpm --filter @vben/web-antd typecheck
pnpm lint
pnpm build
pnpm test:unit
```

## 核心功能

### 认证与会话

- 用户名密码登录
- 邮箱验证码登录
- 手机号验证码登录
- OAuth / 微信登录入口
- JWT Access Token + Refresh Token 轮换
- Refresh Token 使用 HttpOnly Cookie 承载，前端不持久化长期凭证
- 登出时将 Access Token 加入黑名单，并清理服务端 Refresh Token
- 登录设备管理与踢出
- 密码历史、防重复密码、修改密码

### CSRF 与 CORS

- 全局 CORS 中间件，配置位于 `backend-server/config/config.yaml`
- 全局 CSRF 中间件，采用 Double Submit Cookie 模式
- 非安全方法（POST / PUT / DELETE / PATCH）需要携带：

```http
X-CSRF-Token: <csrf_token>
```

- 登录、注册、发送验证码、刷新 Token、微信登录、公开分享等接口按白名单豁免
- OAuth 解绑等登录后状态变更接口不豁免 CSRF

### 验证码系统

支持以下验证码类型：

| 类型 | 说明 |
| --- | --- |
| slider | 滑块验证 |
| puzzle | 拼图验证 |
| rotation | 旋转验证 |
| point | 点选验证 |
| numeric | 数字验证码 |

安全机制：

- 验证码答案 AES-GCM 加密后存 Redis
- 一次性消费，验证后立即销毁
- 支持最短操作时间检测，降低机器脚本通过率
- 发送邮箱 / 短信验证码前需要先完成图形验证码
- 邮箱 / 短信验证码登录失败纳入登录失败计数

### 风险评分系统

对受保护的 API 路径进行多维度风险评估：

| 维度 | 说明 |
| --- | --- |
| 请求频率 | 短时间请求次数超阈值 |
| Referer | 缺失或异常 Referer |
| Accept-Language | 缺失语言头 |
| User-Agent | 包含可疑关键词，如 bot / spider / crawler |
| 请求间隔 | 请求间隔过短 |

规则和阈值存储在数据库，支持运行时修改。

### RBAC 权限模型

```text
User ──(直接关联)──> Role ──> Permissions
  │
  └── Group ──(递归向上)──> GroupRole ──> Role
```

支持：

- 用户管理
- 角色管理
- 菜单管理
- 分组管理
- 权限码缓存
- 权限版本检测

### 文件与存储

- 文件上传与分片上传
- 文件夹管理
- 回收站
- 文件分享
- 文件标签
- 媒体预览 / 下载 / 流式播放
- 存储桶配置
- 存储连接配置
- 文件类型规则
- 标签路由规则

### 系统管理

- 系统设置
- 安全日志
- 登录设备
- 实名认证
- 角色申请审批
- 限流规则
- 定时任务
- 风险评分查看与清理

## 文档入口

- [后端详细文档](./backend-server/README.md)
- [项目文档](./docs/README.md)
- [系统管理文档](./docs/系统管理/README.md)
- [表设计文档](./docs/_common/表设计/README.md)

> `frontend-admin/` 当前没有根级 README，可参考 `CLAUDE.md` 中的前端架构说明，以及 `frontend-admin/apps/web-antd/` 和 `frontend-admin/packages/` 下的源码结构。

## 开发指南

### Git 提交规范

推荐使用：

```text
feat: 新功能
fix: 修复问题
docs: 文档变更
style: 代码格式
refactor: 重构
test: 测试
chore: 构建 / 工具
```

### 分支管理

- `main`：主分支，稳定版本
- `dev`：开发分支
- `feature/*`：功能分支
- `hotfix/*`：紧急修复分支

### 修改约定

- `vue-vben-admin/` 是上游参考源码，默认只读，不直接修改
- 后端 Swagger 注释修改后，需要重新生成：

```bash
cd backend-server
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g ./cmd/server/main.go -o ./docs
```

- 后端提交前建议运行：

```bash
cd backend-server
go test ./...
```

- 前端提交前建议运行：

```bash
cd frontend-admin
pnpm --filter @vben/web-antd typecheck
```

## License

MIT
