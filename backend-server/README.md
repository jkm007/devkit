# Backend Server

基于 Go + Gin + GORM + MySQL + Redis 的后台管理服务。

## 目录结构说明

```
backend-server/
├── cmd/server/
│   └── main.go                          # 程序入口：配置 → 日志 → MySQL → Redis → 迁移 → 默认数据 → WebSocket → 路由 → HTTP 服务
├── config/
│   ├── config.go                        # 配置结构体定义 + Viper 加载逻辑
│   └── config.yaml                      # 配置文件模板（数据库、Redis、JWT、存储、邮箱等）
├── internal/                            # 内部实现，不允许外部包导入
│   ├── handler/                         # HTTP 处理器（参数绑定、校验、调用 service、返回响应）
│   │   ├── auth_handler.go              # 认证相关（登录、登出、Token 刷新、权限码、用户信息、注册、重置密码、修改密码）
│   │   ├── captcha_handler.go           # 图形验证码获取接口
│   │   ├── verify_code_handler.go       # 邮箱验证码（发送、校验），发送前校验图形验证码
│   │   ├── user_handler.go              # 用户管理 CRUD
│   │   ├── role_handler.go              # 角色管理 CRUD
│   │   ├── menu_handler.go              # 菜单管理 CRUD + 用户菜单树
│   │   ├── group_handler.go             # 分组管理 CRUD
│   │   ├── system_setting_handler.go    # 系统配置管理（按 group 分组读写）
│   │   ├── login_device_handler.go      # 登录设备管理
│   │   ├── oauth_handler.go             # OAuth 第三方登录
│   │   ├── password_history_handler.go  # 密码历史记录
│   │   ├── role_application_handler.go  # 角色申请审批
│   │   ├── security_log_handler.go      # 安全日志查询
│   │   ├── user_privacy_handler.go      # 用户隐私设置
│   │   ├── user_real_name_handler.go    # 实名认证
│   │   └── ws_handler.go               # WebSocket 连接处理（升级、读写协程）
│   ├── middleware/                       # Gin 中间件
│   │   ├── auth.go                      # JWT 认证中间件（Access Token 校验）
│   │   ├── permission.go                # 权限码校验（调用 Service 层，带 Redis 缓存）
│   │   ├── captcha_guard.go             # 风险评分中间件（多维度评估 → 触发/拦截），数据库可配置规则
│   │   ├── security_log.go              # 安全日志中间件（记录操作行为）
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
│   │   ├── user_role.go                 # 用户-角色关联表（sys_user_roles）
│   │   ├── system_setting.go            # 系统配置模型（sys_system_settings），支持按 group 分组
│   │   ├── security_log.go              # 安全日志模型
│   │   ├── login_device.go              # 登录设备模型
│   │   ├── oauth_user.go                # OAuth 用户关联模型
│   │   ├── password_history.go          # 密码历史模型
│   │   ├── role_application.go          # 角色申请模型
│   │   ├── user_privacy.go              # 用户隐私模型
│   │   └── user_real_name.go            # 实名认证模型
│   ├── repository/                      # 数据访问层（GORM CRUD）
│   │   ├── user_repo.go                 # 用户仓库（分页查询、角色同步、按用户名/邮箱查询）
│   │   ├── role_repo.go                 # 角色仓库（按名称查询）
│   │   ├── menu_repo.go                 # 菜单仓库
│   │   ├── group_repo.go                # 分组仓库（递归删除、递归获取角色）
│   │   ├── system_setting_repo.go       # 系统配置仓库
│   │   ├── security_log_repo.go         # 安全日志仓库
│   │   ├── login_device_repo.go         # 登录设备仓库
│   │   ├── oauth_user_repo.go           # OAuth 用户仓库
│   │   ├── password_history_repo.go     # 密码历史仓库
│   │   ├── role_application_repo.go     # 角色申请仓库
│   │   ├── user_privacy_repo.go         # 用户隐私仓库
│   │   ├── user_real_name_repo.go       # 实名认证仓库
│   │   └── escape.go                    # SQL LIKE 通配符转义工具
│   ├── service/                         # 业务逻辑层（编排、事务管理）
│   │   ├── auth_service.go              # 认证服务（登录、Token 刷新+轮换、注册、重置密码、修改密码）
│   │   ├── verify_code_service.go       # 邮箱验证码服务（生成、发送、校验、频率限制、失败锁定）
│   │   ├── risk_service.go              # 风险评分配置服务（从数据库加载规则、缓存、刷新）
│   │   ├── user_service.go              # 用户服务（CRUD、角色同步、权限缓存失效）
│   │   ├── role_service.go              # 角色服务（CRUD、权限 JSON 序列化）
│   │   ├── menu_service.go              # 菜单服务（树构建、i18n 翻译、唯一性校验）
│   │   ├── group_service.go             # 分组服务（树构建、递归删除、角色同步）
│   │   ├── system_setting_service.go    # 系统配置服务（按 group 读写，修改后自动刷新关联配置）
│   │   ├── security_log_service.go      # 安全日志服务
│   │   ├── login_device_service.go      # 登录设备服务
│   │   ├── oauth_service.go             # OAuth 服务
│   │   ├── password_history_service.go  # 密码历史服务
│   │   ├── role_application_service.go  # 角色申请服务
│   │   ├── user_privacy_service.go      # 用户隐私服务
│   │   ├── user_real_name_service.go    # 实名认证服务
│   │   └── event_types.go              # 事件类型定义
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
│   ├── captcha/                         # 图形验证码（10 种类型，AES-GCM 加密答案）
│   │   ├── captcha.go                   # 核心逻辑：Generate、Verify、Redis 存储、AES 加解密
│   │   ├── gocaptcha.go                 # go-captcha 库集成（滑块、拼图、旋转、点选、数字）
│   │   ├── slider.go                    # 滑块验证码（自定义实现）
│   │   ├── puzzle.go                    # 拼图验证码（自定义实现）
│   │   ├── rotation.go                  # 旋转验证码（自定义实现）
│   │   ├── point.go                     # 点选验证码（自定义实现）
│   │   ├── numeric.go                   # 数字验证码（自定义实现）
│   │   └── background.go               # 验证码背景图生成
│   ├── email/                           # 邮件发送通用包
│   │   ├── email.go                     # SMTP 发送（SSL 465 / STARTTLS 587），配置从数据库读取并缓存
│   │   └── template.go                  # HTML 邮件模板（验证码邮件：渐变头部 + 大字验证码 + 过期提示）
│   ├── crypto/
│   │   └── aes.go                       # AES-256-GCM 对称加密/解密
│   ├── response/
│   │   └── response.go                  # 统一 API 响应封装（Success/BadRequest/Unauthorized/Forbidden/CaptchaRequired 等）
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
│   │   ├── oss.go                       # 阿里云 OSS 实现
│   │   └── cos.go                       # 腾讯云 COS 实现
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
│   ├── init.sql                         # 建表 DDL
│   ├── init_menu.sql                    # 菜单种子数据（Dashboard + 系统管理 + 按钮权限）
│   ├── init_data.sql                    # 默认角色和用户种子数据
│   └── init_system_settings.sql         # 系统配置种子数据（基础配置 + 验证码配置 + 风险评分规则）
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
                   permission   pkg (jwt, cache, captcha, email, storage, mq...)
                   captcha_guard
                   (调用 service
                    获取配置)
```

| 层 | 职责 | 依赖方向 |
|---|------|---------|
| **router** | 路由注册、中间件编排、Handler 实例化 | → handler, middleware |
| **middleware** | JWT 认证、权限校验、风险评分、安全日志、限流、CORS | → service, pkg |
| **handler** | 参数绑定、校验、调用 service、返回响应 | → service, pkg |
| **service** | 业务逻辑编排、验证码生成/校验、Token 轮换、事务管理 | → repository, pkg |
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

### 验证码系统

支持 10 种验证码类型，分两大类：

| 类型 | 库 | 说明 |
|------|-----|------|
| slider | go-captcha / 自定义 | 滑块验证 |
| puzzle | go-captcha / 自定义 | 拼图验证 |
| rotation | go-captcha / 自定义 | 旋转验证 |
| point | go-captcha / 自定义 | 点选验证 |
| numeric | go-captcha / 自定义 | 数字验证码 |

- **AES-GCM 加密**答案，防抓包破解
- **一次性消费**：验证后立即从 Redis 删除（`GetDel`）
- **最短操作时间检测**：防止机器人秒答
- 配置存储在 `sys_system_settings` 表，修改实时生效（无需重启）
- 登录失败次数达阈值自动触发验证码

### 风险评分系统（CaptchaGuard）

对受保护的 API 路径进行多维度风险评估，高风险请求触发验证码，超高风险直接拦截。

**评估维度：**

| 规则 | 说明 | 默认分值 |
|------|------|---------|
| frequency | 60 秒内请求频率超阈值 | 20 |
| no_referer | 无 Referer 头 | 15 |
| no_lang | 无 Accept-Language 头 | 10 |
| ua | User-Agent 包含可疑关键词 | 25 |
| interval | 请求间隔过短（<500ms） | 15 |

**阈值：**
- 触发验证码：累计 ≥ 50 分
- 直接拦截：累计 ≥ 100 分
- 风险分自动衰减（每 10 分钟衰减 30%）

所有规则和阈值存储在数据库 `sys_system_settings` 表的 `risk_score` 分组中，支持运行时修改。

### 邮箱验证码系统

用于注册、重置密码等场景，流程：**图形验证码 → 发送邮箱验证码 → 提交业务请求**。

**安全机制（4 层防护）：**

| 层 | 机制 | 说明 |
|----|------|------|
| 1 | 图形验证码 | 发送前必须完成图形验证，防止接口被滥用 |
| 2 | IP 限流 | `/auth/send-code` 路由级别 1 req/s |
| 3 | 发送冷却 | 同一邮箱+目的 60 秒内不能重复发送 |
| 4 | 失败锁定 | 同一邮箱+目的 5 次验证失败后锁定 15 分钟 |

**Redis Key 设计：**
- `verify_code:{purpose}:{email}` — 验证码存储，TTL 5 分钟
- `verify_code_cooldown:{purpose}:{email}` — 发送冷却，TTL 60 秒
- `verify_code_fail:{purpose}:{email}` — 失败计数，TTL 15 分钟

**邮件发送：** 支持 SSL（端口 465）和 STARTTLS（端口 587），配置从数据库读取并缓存。HTML 模板包含渐变头部、大字验证码、过期提示。

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

### 认证（无需登录）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/login` | 用户登录（失败达阈值触发验证码） |
| POST | `/auth/logout` | 用户登出 |
| POST | `/auth/refresh` | Token 刷新（轮换） |
| GET | `/auth/captcha` | 获取图形验证码（限流 5/s） |
| POST | `/auth/send-code` | 发送邮箱验证码（需图形验证码，限流 1/s） |
| POST | `/auth/verify-code` | 验证邮箱验证码 |
| POST | `/auth/register` | 用户注册（需邮箱验证码） |
| POST | `/auth/reset-password` | 重置密码（需邮箱验证码） |

### 用户相关（需 JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/user/info` | 获取当前用户信息 |
| GET | `/auth/codes` | 获取权限码列表 |
| GET | `/auth/permission-version` | 权限版本 hash |
| PUT | `/auth/change-password` | 修改密码（需图形验证码） |

### 系统管理（需 JWT + 权限码）

| 方法 | 路径 | 权限码 | 说明 |
|------|------|--------|------|
| GET | `/system/user/list` | system:user:view | 用户列表 |
| POST | `/system/user` | system:user:add | 创建用户 |
| PUT | `/system/user/:id` | system:user:edit | 更新用户 |
| DELETE | `/system/user/:id` | system:user:delete | 删除用户（需图形验证码） |
| GET | `/system/role/list` | system:role:view | 角色列表 |
| GET | `/system/role/:id` | system:role:view | 角色详情 |
| POST | `/system/role` | system:role:add | 创建角色 |
| PUT | `/system/role/:id` | system:role:edit | 更新角色 |
| DELETE | `/system/role/:id` | system:role:delete | 删除角色（需图形验证码） |
| GET | `/system/menu/list` | system:menu:view | 菜单列表 |
| POST | `/system/menu` | system:menu:add | 创建菜单 |
| PUT | `/system/menu/:id` | system:menu:edit | 更新菜单 |
| DELETE | `/system/menu/:id` | system:menu:delete | 删除菜单（需图形验证码） |
| GET | `/system/group/list` | system:group:view | 分组列表 |
| POST | `/system/group` | system:group:add | 创建分组 |
| PUT | `/system/group/:id` | system:group:edit | 更新分组 |
| DELETE | `/system/group/:id` | system:group:delete | 删除分组（需图形验证码） |
| GET/PUT | `/system/settings` | system:setting:view | 系统配置（按 group 读写） |

### 其他

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/ws` | WebSocket 连接（需 JWT） |

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
| 验证码 | go-captcha + 自定义实现（AES-GCM 加密） |
| 邮件 | SMTP（SSL / STARTTLS） |
| 限流 | golang.org/x/time/rate (令牌桶) |
