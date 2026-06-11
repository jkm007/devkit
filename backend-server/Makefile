.PHONY: run build clean test lint vuln-check docker-up docker-down

# 变量
APP_NAME = backend-server
MAIN_PATH = ./cmd/server/main.go

# 运行
run:
	go run $(MAIN_PATH)

# 构建
build:
	CGO_ENABLED=0 go build -o bin/$(APP_NAME) $(MAIN_PATH)

# 清理
clean:
	rm -rf bin/

# 测试
test:
	go test -v ./...

# 代码检查
lint:
	golangci-lint run ./...

# 格式化
fmt:
	gofmt -w .

# 依赖漏洞扫描
vuln-check:
	govulncheck ./...

# 启动依赖服务
docker-up:
	docker-compose up -d mysql redis minio

# 停止依赖服务
docker-down:
	docker-compose down

# 构建 Docker 镜像
docker-build:
	docker build -t $(APP_NAME) .

# 运行 Docker 容器
docker-run:
	docker-compose up -d

# 数据库迁移
migrate:
	go run $(MAIN_PATH) -migrate

# 生成 Swagger 文档
swagger:
	swag init -g $(MAIN_PATH) -o ./docs

# 帮助
help:
	@echo "可用命令:"
	@echo "  make run          - 运行服务"
	@echo "  make build        - 构建二进制文件"
	@echo "  make clean        - 清理构建文件"
	@echo "  make test         - 运行测试"
	@echo "  make lint         - 代码检查"
	@echo "  make fmt          - 格式化代码"
	@echo "  make docker-up    - 启动依赖服务"
	@echo "  make docker-down  - 停止依赖服务"
	@echo "  make docker-build - 构建 Docker 镜像"
	@echo "  make docker-run   - 运行 Docker 容器"
	@echo "  make swagger      - 生成 Swagger 文档"
	@echo "  make vuln-check   - 依赖漏洞扫描"
