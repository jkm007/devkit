# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DevKit is a full-stack admin management system (后台管理系统) built with Go + Vue 3. This is a **multi-project repository** with three independent sub-projects sharing one git repo:

- `backend-server/` — Go backend (Gin + GORM + MySQL + Redis)
- `frontend-admin/` — Vue 3 frontend (customized fork of Vue Vben Admin, Ant Design Vue variant)
- `vue-vben-admin/` — Upstream Vue Vben Admin v5.7.0 kept as reference (read-only, do not modify)

## Common Commands

### Backend (`backend-server/`)

```bash
cd backend-server
make run              # Run the server locally
make build            # Build binary
make test             # Run tests
make lint             # Run golangci-lint
make fmt              # Format code
make swagger          # Regenerate Swagger docs
make migrate          # Run database migrations

# Docker
make docker-up        # Start MySQL + Redis + MinIO + backend via docker-compose
make docker-down      # Stop all containers
make docker-build     # Build Docker image
```

### Frontend (`frontend-admin/`)

```bash
cd frontend-admin
pnpm install          # Install dependencies
pnpm dev              # Start dev server (web-antd app, port 5666)
pnpm build            # Build all packages via Turborepo
pnpm test:unit        # Run Vitest unit tests
pnpm lint             # Run ESLint + OxLint
pnpm format           # Format code
pnpm check            # Circular deps + dep check + type check
```

## Architecture

### Backend (`backend-server/`)

Standard Go layered architecture:

- **`cmd/server/main.go`** — Entry point. Boot: config → logger → MySQL → Redis → migrations → default users → WebSocket hub → routes → HTTP server → graceful shutdown
- **`config/`** — Viper YAML config (`config.yaml` + `config.go` struct definitions)
- **`internal/`** — Core business logic:
  - `handler/` → `service/` → `repository/` → `model/` — request flows through these layers in order
  - `middleware/` — Auth/JWT, CORS, RBAC, rate limiter, security log
  - `router/` — Route definitions with permission-based access control
  - `ws/` — WebSocket hub
- **`pkg/`** — Reusable packages: cache, captcha (10 types), crypto (AES-256), database, jwt, logger, mq, response, storage (local/minio/oss/cos)
- **`migrations/`** — SQL migration files orchestrated by `migrate.go`

**Auth**: JWT with Access + Refresh tokens. RBAC model: User → Role → Permissions, with Group hierarchy.

### Frontend (`frontend-admin/`)

pnpm monorepo with Turborepo, based on Vue Vben Admin framework:

- **`apps/web-antd/`** — The main admin application (Ant Design Vue 4)
  - `src/api/` — API modules matching backend routes (system/, account/)
  - `src/views/` — Page components (dashboard, account, system management)
  - `src/router/` — Dynamic route loading with permission-based access control
  - `src/store/` — Pinia stores (auth store handles login/logout/permission version)
- **`packages/`** — Shared libraries (@core, constants, icons, locales, stores, styles, types, utils)
- **`internal/`** — Build tooling (lint-configs, tailwind-config, tsconfig, vite-config)

**API proxy**: Dev server proxies `/api` to the Go backend (configured in `apps/web-antd/vite.config.ts` and `.env.development`).

### Relationship Between Frontend Projects

`frontend-admin` is a customized derivative of `vue-vben-admin`. Only the `web-antd` app variant is retained. Shared packages are duplicated (not linked as submodule). `vue-vben-admin/` should be treated as read-only reference.

## Key Configuration

- `backend-server/config/config.yaml` — Backend runtime config (DB, Redis, JWT, storage, CORS, rate limits)
- `frontend-admin/apps/web-antd/.env.development` — Dev environment (port 5666, API backend URL)
- `frontend-admin/pnpm-workspace.yaml` — Workspace structure + pnpm catalog (centralized dependency versions)
- `frontend-admin/turbo.json` — Turborepo task pipeline
- `frontend-admin/lefthook.yml` — Git hooks (pre-commit: lint + typecheck; commit-msg: commitlint)

## Commit Convention

Frontend uses commitlint with format: `type(scope): subject`. Enforced by Lefthook git hooks.
