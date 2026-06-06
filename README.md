# DevKit - 后台管理系统

基于 Go + Vue3 的全栈后台管理系统，包含后端服务和前端管理界面。

## 项目结构

```
devkit/
├── backend-server/     # Go 后端服务
├── frontend-admin/     # Vue3 前端管理界面
├── vue-vben-admin/     # Vue Vben Admin 框架源码（参考）
├── docs/               # 项目文档
└── tmp/                # 临时文件
```

## 技术栈

### 后端 (backend-server)

| 组件 | 选型 |
|------|------|
| 语言 | Go 1.21+ |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0 |
| 缓存/队列 | Redis 7.x |
| 日志 | Zap + Lumberjack |
| 配置 | Viper |
| 认证 | JWT (Access + Refresh Token) |
| WebSocket | Gorilla WebSocket |
| 文件存储 | 本地 / MinIO / OSS / COS |
| 验证码 | go-captcha + 自定义实现（10 种类型，AES-GCM 加密） |
| 邮件 | SMTP（SSL / STARTTLS），HTML 模板 |
| 限流 | 令牌桶（golang.org/x/time/rate） |

### 前端 (frontend-admin)

| 组件 | 选型 |
|------|------|
| 框架 | Vue 3 |
| 构建 | Vite 8 |
| 语言 | TypeScript |
| UI | Ant Design Vue 4 |
| 状态 | Pinia |
| 样式 | Tailwind CSS 4 |
| 国际化 | vue-i18n |

## 快速启动

### 后端

```bash
cd backend-server

# 1. 启动依赖服务
docker-compose up -d mysql redis

# 2. 修改配置
cp config/config.yaml config/config.local.yaml

# 3. 运行
make run
```

后端服务地址：`http://localhost:8080`
Swagger 文档：`http://localhost:8080/swagger/index.html`

### 前端

```bash
cd frontend-admin

# 1. 安装依赖
pnpm install

# 2. 开发模式
pnpm dev
```

前端服务地址：`http://localhost:5666`

## 核心功能

### 验证码系统

支持 10 种验证码类型，分 go-captcha 库实现和自定义实现两大类：

| 类型 | 说明 |
|------|------|
| slider | 滑块验证 |
| puzzle | 拼图验证 |
| rotation | 旋转验证 |
| point | 点选验证 |
| numeric | 数字验证码 |

- AES-GCM 加密答案，防抓包破解
- 一次性消费（验证后立即销毁）
- 最短操作时间检测（防机器人）
- 配置存储在数据库，修改实时生效
- 登录失败达阈值自动触发

### 风险评分系统

对受保护的 API 路径进行多维度风险评估：

| 维度 | 说明 |
|------|------|
| 请求频率 | 60 秒内请求次数超阈值 |
| Referer | 无 Referer 头 |
| Accept-Language | 无语言头 |
| User-Agent | 包含可疑关键词（bot/spider/crawler 等） |
| 请求间隔 | 请求间隔过短（<500ms） |

- 累计 ≥ 50 分：触发验证码
- 累计 ≥ 100 分：直接拦截
- 风险分自动衰减（每 10 分钟衰减 30%）
- 所有规则和阈值存储在数据库，支持运行时修改

### 邮箱验证码系统

用于注册、重置密码等场景：

```
用户点击"发送验证码"
  → 弹出图形验证码（防机器人）
  → 完成图形验证
  → 生成 6 位随机码 → 存 Redis (5min TTL) → SMTP 发送 HTML 邮件
  → 用户输入验证码 → 提交业务请求
```

**安全机制：** 图形验证码前置 + IP 限流 (1/s) + 发送冷却 (60s) + 失败锁定 (5 次 / 15 分钟)

### 敏感操作保护

修改密码、删除用户/角色/分组/菜单等操作需要完成图形验证码：

- **前端**：操作前弹出图形验证码弹窗
- **后端**：CaptchaGuard 中间件提供额外的风险评分保护

### RBAC 权限模型

```
User ──(直接关联)──> Role ──> Permissions
  │
  └── Group ──(递归向上)──> GroupRole ──> Role
```

### Token 机制

- Access Token：短期有效，JWT 签名
- Refresh Token：长期有效，Redis 存储
- Token 轮换：刷新时自动生成新 Token

## 子项目文档

- [后端详细文档](./backend-server/README.md)
- [前端详细文档](./frontend-admin/README.md)

## 开发指南

### Git 提交规范

```
feat: 新功能
fix: 修复问题
docs: 文档变更
style: 代码格式
refactor: 重构
test: 测试
chore: 构建/工具
```

### 分支管理

- `main`: 主分支，稳定版本
- `develop`: 开发分支
- `feature/*`: 功能分支
- `hotfix/*`: 紧急修复分支

## License

MIT
