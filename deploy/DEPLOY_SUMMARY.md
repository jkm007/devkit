# 部署脚本使用说明

## 快速部署

### 本地环境

```bash
# 完整部署（数据库 + 后端 + 前端）
./scripts/deploy-all.sh local

# 只执行数据库迁移
./scripts/deploy-all.sh local db

# 只部署后端
./scripts/deploy-all.sh local backend

# 只部署前端
./scripts/deploy-all.sh local frontend
```

### 阿里云环境

```bash
# 完整部署
./scripts/deploy-all.sh aliyun

# 只执行数据库迁移
./scripts/deploy-all.sh aliyun db
```

## 数据库管理

### 验证数据库状态

```bash
./scripts/db-migrate.sh local verify
```

输出示例：
```
=== 表结构 ===
sys_storage_bucket
sys_system_settings
sys_tag
sys_tag_routing

=== 存储桶配置 ===
id  name              driver  is_default  status
9   MinIO 默认桶      minio   1           1
15  本地默认存储      local   1           1

=== 存储配置 ===
storage_minio_enabled  true
storage_cos_enabled    false
```

### 执行数据库迁移

```bash
# 同步所有（迁移 + 基础数据）
./scripts/db-migrate.sh local sync

# 只执行表结构迁移
./scripts/db-migrate.sh local migrate

# 只初始化基础数据
./scripts/db-migrate.sh local seed
```

### 从阿里云同步数据库

```bash
# 全量同步
./local/scripts/db-sync.sh

# 增量同步（只同步新表）
./local/scripts/db-sync.sh --incremental

# 只同步指定表
./local/scripts/db-sync.sh --tables sys_storage_bucket,sys_tag

# 只显示要同步的表，不执行
./local/scripts/db-sync.sh --dry-run
```

## 存储桶管理

### 数据库表结构

`sys_storage_bucket` 表：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| name | VARCHAR(100) | 桶名称 |
| driver | VARCHAR(20) | 驱动: local, minio, oss, cos |
| endpoint | VARCHAR(500) | 服务端点 |
| bucket | VARCHAR(200) | 桶名称 |
| access_key | VARCHAR(500) | 访问密钥 |
| secret_key | VARCHAR(500) | 密钥 |
| region | VARCHAR(100) | 区域 |
| purpose | VARCHAR(100) | 用途: file, backup, avatar, temp |
| is_default | TINYINT(1) | 是否默认 |
| status | TINYINT(1) | 状态: 1=启用, 0=禁用 |

### 同步流程

1. **存储配置** → **存储桶管理**：保存存储配置时自动同步
2. **存储桶管理** → **上传流程**：默认桶决定使用哪个驱动

### 手动同步

如果需要手动同步存储桶配置：

```bash
./scripts/db-migrate.sh local seed
```

## 环境配置

### 本地环境 (local/env.sh)

```bash
APP_HOST="10.0.50.103"
DB_HOST="10.0.50.108"
DB_PASSWORD="root123456"
MINIO_HOST="10.0.50.108"
MINIO_BUCKET="devkit"
```

### 阿里云环境 (aliyun/env.sh)

```bash
APP_HOST="123.57.201.44"
DB_HOST="114.215.190.52"
DB_PASSWORD="root123456"
MINIO_HOST="114.215.190.52"
MINIO_BUCKET="test"
```

## 常见问题

### 1. 上传文件失败

检查：
1. 存储桶管理中是否有启用的默认桶
2. 存储配置中对应的驱动是否启用
3. 连接信息是否正确

```bash
# 验证存储桶配置
./scripts/db-migrate.sh local verify
```

### 2. 数据库表不存在

```bash
# 执行迁移
./scripts/db-migrate.sh local migrate
```

### 3. 存储桶配置不同步

```bash
# 手动同步
./scripts/db-migrate.sh local seed
```

### 4. 从阿里云同步数据

```bash
# 全量同步
./local/scripts/db-sync.sh

# 只同步新表
./local/scripts/db-sync.sh --incremental
```

## 部署检查清单

- [ ] 数据库表结构完整
- [ ] 存储桶配置已同步
- [ ] 默认桶已设置
- [ ] 标签和路由规则已初始化
- [ ] 后端服务正常运行
- [ ] 前端可以访问
- [ ] 上传功能正常

## 文件结构

```
deploy/
├── README.md                    # 部署指南
├── DEPLOY_SUMMARY.md            # 部署脚本使用说明（本文件）
├── aliyun/                      # 阿里云环境
│   ├── env.sh                   # 环境变量
│   ├── configs/                 # 配置文件
│   └── scripts/                 # 部署脚本
├── local/                       # 本地环境
│   ├── env.sh                   # 环境变量
│   ├── configs/                 # 配置文件
│   └── scripts/                 # 部署脚本
│       └── db-sync.sh           # 从阿里云同步数据库
└── scripts/                     # 通用脚本
    ├── deploy-all.sh            # 完整部署脚本
    └── db-migrate.sh            # 数据库迁移脚本
```
