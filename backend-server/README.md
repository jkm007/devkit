# Backend Server

基于 Go + Gin + GORM + MySQL + Redis 的后台管理服务。

## 目录结构说明

```
backend-server/
├── cmd/server/
│   └── main.go                          # 程序入口，初始化各组件并启动 HTTP 服务
├── config/
│   ├── config.go                        # 配置结构体定义 + Viper 加载逻辑
│   └── config.yaml                      # 配置文件模板（数据库、Redis、JWT、存储等）
├── internal/                            # 内部实现，不允许外部包导入
│   ├── handler/                         # HTTP 处理器（参数绑定、校验、调用 service、返回响应）
│   │   ├── auth_handler.go              # 认证相关（登录、登出、Token 刷新、权限码、用户信息）
│   │   ├── user_handler.go              # 用户管理 CRUD
│   │   ├── role_handler.go              # 角色管理 CRUD
│   │   ├── menu_handler.go              # 菜单管理 CRUD + 用户菜单树
│   │   ├── group_handler.go             # 分组管理 CRUD
│   │   └── ws_handler.go                # WebSocket 连接处理（升级、读写协程）
│   ├── middleware/                       # Gin 中间件
│   │   ├── auth.go                      # JWT 认证中间件（Access Token 校验）
│   │   ├── permission.go                # 权限码校验（调用 Service 层，带 Redis 缓存）
│   │   ├── cors.go                      # CORS 跨域配置
│   │   ├── ratelimit.go                 # 令牌桶限流（带 IP TTL 自动清理）
│   │   ├── rbac.go                      # RBAC 角色权限控制中间件
│   │   └── logger.go                    # 请求日志中间件（Zap 结构化日志）
│   ├── model/                           # GORM 数据模型定义
│   │   ├── base.go                      # BaseModel（ID、CreatedAt、UpdatedAt、DeletedAt）
│   │   ├── user.go                      # 用户模型（sys_users）
│   │   ├── role.go                      # 角色模型（sys_roles），权限以 JSON 数组存储
│   │   ├── menu.go                      # 菜单模型（sys_menus），含 MenuMeta、MenuTree
│   │   ├── group.go                     # 分组模型（sys_groups），含 GroupTree、GroupRole
│   │   └── user_role.go                 # 用户-角色关联表（sys_user_roles）
│   ├── repository/                      # 数据访问层（GORM CRUD）
│   │   ├── user_repo.go                 # 用户仓库（分页查询、角色同步）
│   │   ├── role_repo.go                 # 角色仓库
│   │   ├── menu_repo.go                 # 菜单仓库
│   │   ├── group_repo.go                # 分组仓库（递归删除、递归获取角色）
│   │   └── escape.go                    # SQL LIKE 通配符转义工具
│   ├── service/                         # 业务逻辑层（编排、事务管理）
│   │   ├── auth_service.go              # 认证服务（登录、Token 刷新+轮换、权限码缓存、数据初始化）
│   │   ├── user_service.go              # 用户服务（CRUD、角色同步、权限缓存失效）
│   │   ├── role_service.go              # 角色服务（CRUD、权限 JSON 序列化）
│   │   ├── menu_service.go              # 菜单服务（树构建、i18n 翻译、唯一性校验）
│   │   └── group_service.go             # 分组服务（树构建、递归删除、角色同步）
│   ├── router/
│   │   └── router.go                    # 路由注册、分组、中间件挂载（含 WebSocket 路由）
│   ├── ws/
│   │   └── hub.go                       # WebSocket Hub（按用户分组、广播/点对点消息）
│   ├── task/
│   │   └── scheduler.go                 # 后台任务调度（Go Channel 异步任务、Redis 延迟队列消费）
│   └── validator/
│       └── validator.go                 # 自定义参数校验器（注册到 go-playground/validator）
├── pkg/                                 # 可复用工具包（可被其他项目引用）
│   ├── database/
│   │   ├── mysql.go                     # MySQL 连接初始化（GORM）
│   │   └── redis.go                     # Redis 连接初始化
│   ├── jwt/
│   │   └── jwt.go                       # JWT 生成、解析（Access + Refresh Token，HS256）
│   ├── captcha/
│   │   └── captcha.go                   # 图形验证码（纯 Go 实现，内存存储）
│   ├── crypto/
│   │   └── aes.go                       # AES-256-CBC 对称加密/解密
│   ├── response/
│   │   └── response.go                  # 统一 API 响应封装（Success/BadRequest/Unauthorized/Forbidden/TooManyRequests 等）
│   ├── logger/
│   │   └── zap.go                       # Zap 日志初始化（配置级别、输出格式、lumberjack 切割）
│   ├── cache/
│   │   └── redis.go                     # Redis 缓存工具（String/Hash/ZSet 常用操作封装）
│   ├── mq/
│   │   ├── delay_queue.go               # Redis ZSet 延迟队列
│   │   └── stream.go                    # Redis Stream 消费组消息队列
│   ├── sensitive/
│   │   └── filter.go                    # 敏感词过滤（DFA 算法）
│   ├── storage/                         # 文件存储抽象层（Strategy 模式）
│   │   ├── storage.go                   # Storage 接口定义 + 工厂函数
│   │   ├── local.go                     # 本地文件存储实现（开发环境）
│   │   ├── minio.go                     # MinIO 对象存储实现（私有化部署）
│   │   ├── oss.go                       # 阿里云 OSS 实现（TODO）
│   │   └── cos.go                       # 腾讯云 COS 实现（TODO）
│   └── cdn/
│       └── cdn.go                       # CDN URL 生成、图片处理（缩略图、裁剪）
├── migrations/
│   └── migrate.go                       # 数据库迁移（GORM AutoMigrate）
├── docs/
│   ├── docs.go                          # Swagger 生成文件
│   ├── swagger.go                       # Swagger 入口
│   ├── swagger.json
│   └── swagger.yaml
├── scripts/
│   ├── init.sql                         # 建表 DDL（6 张表，与 GORM 模型一致）
│   ├── init_menu.sql                    # 菜单种子数据（Dashboard + 系统管理 + 按钮权限）
│   └── init_data.sql                    # 默认角色和用户种子数据
├── Dockerfile                           # 多阶段构建镜像
├── docker-compose.yml                   # 本地开发环境编排（MySQL、Redis、MinIO）
├── Makefile                             # 常用命令快捷入口
├── go.mod                               # Go 模块依赖
└── go.sum                               # 依赖校验
```

## 分层架构

```
请求 → router → middleware → handler → service → repository → model / database
                        ↑           ↓
                   permission   pkg (jwt, cache, storage, mq...)
                   (调用 service
                    获取权限码)
```

| 层 | 职责 | 依赖方向 |
|---|------|---------|
| **router** | 路由注册、中间件编排、Handler 实例化 | → handler, middleware |
| **middleware** | JWT 认证、权限校验（调用 Service 层 + Redis 缓存）、限流、CORS、日志 | → service |
| **handler** | 参数绑定、校验、调用 service、返回响应 | → service |
| **service** | 业务逻辑编排、权限码缓存、Token 轮换、事务管理 | → repository, pkg |
| **repository** | 数据库 CRUD，构造函数显式注入 `*gorm.DB` | → model |
| **model** | 数据结构定义、GORM 标签 | 无依赖 |

## 核心设计

### RBAC 权限模型

```
User ──(直接关联)──> Role ──> Permissions (JSON 数组)
  │
  └── Group ──(递归向上)──> GroupRole ──> Role ──> Permissions
```

- 权限码存储在 Role 的 `Permissions` 字段（JSON 数组）
- 用户通过**直接关联角色**和**分组继承角色**（递归向上查找父分组）获得权限
- 权限码在 Service 层获取后**缓存到 Redis**（10 分钟 TTL），用户角色/分组变更时主动失效

### Token 机制

- **Access Token**：短期有效，JWT HS256 签名
- **Refresh Token**：长期有效，以 SHA256 哈希存储在 Redis
- **Token 轮换**：刷新时同时生成新的 Access + Refresh Token，旧 Refresh Token 自动失效

### 限流

- 令牌桶算法，按 IP 限流
- 后台 goroutine 每 5 分钟清理 10 分钟未活跃的 IP 记录，防止内存泄漏

## 快速启动

```bash
# 1. 复制配置文件并修改
cp config/config.yaml config/config.local.yaml

# 2. 启动依赖服务
docker-compose up -d mysql redis

# 3. 生成 Swagger 文档
make swagger

# 4. 运行
go run cmd/server/main.go

# 或使用 Makefile
make run
```

## Swagger 文档

启动服务后访问：

```
http://localhost:8080/swagger/index.html
```

重新生成文档：

```bash
make swagger
```

## API 端点

| 方法 | 路径 | 认证 | 权限码 | 说明 |
|------|------|------|--------|------|
| GET | `/health` | ✗ | — | 健康检查 |
| POST | `/auth/login` | ✗ | — | 用户登录 |
| POST | `/auth/logout` | ✗ | — | 用户登出 |
| POST | `/auth/refresh` | ✗ | — | Token 刷新（轮换） |
| GET | `/user/info` | JWT | — | 获取当前用户信息 |
| GET | `/auth/codes` | JWT | — | 获取权限码列表 |
| GET | `/auth/permission-version` | JWT | — | 权限版本 hash |
| GET | `/ws` | JWT | — | WebSocket 连接 |
| GET | `/menu/all` | JWT | — | 获取用户菜单树 |
| GET | `/system/user/list` | JWT | system:user:view | 用户列表 |
| POST | `/system/user` | JWT | system:user:add | 创建用户 |
| PUT | `/system/user/:id` | JWT | system:user:edit | 更新用户 |
| DELETE | `/system/user/:id` | JWT | system:user:delete | 删除用户 |
| GET | `/system/role/list` | JWT | system:role:view | 角色列表 |
| GET | `/system/role/:id` | JWT | system:role:view | 角色详情 |
| POST | `/system/role` | JWT | system:role:add | 创建角色 |
| PUT | `/system/role/:id` | JWT | system:role:edit | 更新角色 |
| DELETE | `/system/role/:id` | JWT | system:role:delete | 删除角色 |
| GET | `/system/menu/list` | JWT | system:menu:view | 菜单列表 |
| POST | `/system/menu` | JWT | system:menu:add | 创建菜单 |
| PUT | `/system/menu/:id` | JWT | system:menu:edit | 更新菜单 |
| DELETE | `/system/menu/:id` | JWT | system:menu:delete | 删除菜单 |
| GET | `/system/group/list` | JWT | system:group:view | 分组列表 |
| POST | `/system/group` | JWT | system:group:add | 创建分组 |
| PUT | `/system/group/:id` | JWT | system:group:edit | 更新分组 |
| DELETE | `/system/group/:id` | JWT | system:group:delete | 删除分组 |

## 技术栈

| 组件 | 选型 |
|------|------|
| 语言 | Go 1.21+ |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0 |
| 缓存/队列 | Redis 7.x |
| 日志 | Zap + Lumberjack |
| 配置 | Viper |
| 认证 | JWT (Access + Refresh Token, HS256) |
| WebSocket | Gorilla WebSocket (Hub 模式) |
| 文件存储 | 本地 / MinIO / 阿里云 OSS / 腾讯云 COS |
| 限流 | golang.org/x/time/rate (令牌桶) |
