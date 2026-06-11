# AGENTS.md

## 仓库结构

多项目仓库，三个独立子项目共用一个 git 根目录：

- `backend-server/` — Go 后端（Gin + GORM + MySQL + Redis）
- `frontend-admin/` — Vue 3 前端（pnpm monorepo，Ant Design Vue 4）
- `vue-vben-admin/` — 上游参考，**禁止修改**

## 关键约束

- `vue-vben-admin/` 只读，永远不要编辑其中的文件。
- 前端通过 `preinstall` 脚本强制使用 `pnpm`。Node `^22.18.0 || ^24.0.0`，pnpm `>=11.0.0`。
- 前端 pre-commit 钩子会运行 `pnpm lint` + `pnpm check:type`（通过 lefthook）。
- 提交信息格式：`type(scope): subject`，由 commitlint 强制执行。
- 后端 `config.yaml` 包含真实凭证，本地覆盖请用 `config.local.yaml`（已 gitignore）。

## 后端命令

所有命令在 `backend-server/` 目录下执行：

```
make run              # 启动服务（需要 MySQL + Redis）
make build            # CGO_ENABLED=0 编译到 bin/
make test             # go test -v ./...
make lint             # golangci-lint run ./...
make fmt              # gofmt -w .
make swagger          # swag init → ./docs
make migrate          # 通过 -migrate 标志运行数据库迁移
make docker-up        # 启动 MySQL + Redis + MinIO 容器
make docker-down      # 停止容器
```

启动顺序：config → logger → MySQL → Redis → migrations → 默认用户 → WebSocket hub → routes → HTTP server。

API 分层：`handler/` → `service/` → `repository/` → `model/`，请求自顶向下流动。

入口：`cmd/server/main.go`

## 前端命令

所有命令在 `frontend-admin/` 目录下执行：

```
pnpm install          # 安装依赖（强制 pnpm）
pnpm dev              # 开发服务器，端口 5666，代理 /api/v1 → localhost:8080
pnpm build            # Turborepo 构建所有包
pnpm lint             # ESLint + OxLint（通过 vsh）
pnpm format           # 自动格式化（vsh lint --format）
pnpm check            # 循环依赖 + 依赖检查 + 类型检查
pnpm check:type       # 仅类型检查（turbo run typecheck）
pnpm test:unit        # Vitest run --dom
```

Monorepo 结构：
- `apps/web-antd/` — 主管理端应用，仅保留此 app 变体。
- `packages/` — 共享库（@core, constants, icons, locales, stores, styles, types, utils）
- `internal/` — 构建工具（lint-configs, tailwind-config, tsconfig, vite-config）

API 代理：开发服务器将 `/api/v1` 转发到 `http://localhost:8080`（配置在 `apps/web-antd/vite.config.ts`）。

## 关键配置文件

| 文件 | 用途 |
|------|------|
| `backend-server/config/config.yaml` | 后端运行配置（数据库、Redis、JWT、存储、CORS） |
| `backend-server/config/config.local.yaml` | 本地覆盖（已 gitignore，从 config.yaml 复制创建） |
| `frontend-admin/apps/web-antd/.env.development` | 开发环境（端口 5666、API 地址） |
| `frontend-admin/pnpm-workspace.yaml` | 工作区结构 + pnpm catalog（集中管理依赖版本） |
| `frontend-admin/turbo.json` | Turborepo 任务流水线 |
| `frontend-admin/lefthook.yml` | Git 钩子（pre-commit: lint + typecheck; commit-msg: commitlint） |

## 认证架构

JWT 双 Token（Access + Refresh）。RBAC 模型：User → Role → Permissions，支持 Group 层级继承。Refresh Token 存储在 Redis 中，刷新时自动轮换。
