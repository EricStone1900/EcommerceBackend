# 商品管理系统 (Product Management System)

一个使用 Go 实现的电商后台服务，采用 Clean Architecture + Hexagonal 架构。

## 技术栈

| 组件 | 选型 | 用途 |
|---|---|---|
| 语言 | Go 1.22+ | — |
| Web 框架 | Gin | HTTP 路由 / 中间件 |
| ORM | GORM | 数据库操作（infrastructure 层） |
| 数据库 | PostgreSQL 16 | 主存储 |
| 缓存 | Redis 7 | token 黑名单 / 限流 |
| 配置 | Viper | 文件 + 环境变量 |
| 日志 | zap | 结构化 JSON 日志 |
| 文件存储 | 本地磁盘 / S3（可切换）| 本地开发用 local，云端用 s3 |
| 事件总线 | 进程内 Channel（预留 NATS）| 异步处理（如 OCR 模拟）|
| 推送 | Stub 日志（预留 APNs）| 本地验证推送链路 |
| 容器 | Docker + docker-compose | 本地开发 / 部署 |

## 目录结构

```
.
├── cmd/server/main.go         # 程序入口
├── internal/
│   ├── domain/
│   │   ├── entity/            # 纯业务实体
│   │   └── port/              # Repository / Storage / Notifier 等接口定义
│   ├── usecase/               # 业务用例编排（auth / product / upload / push / assistant）
│   ├── interface/http/        # Gin handler / middleware / router
│   ├── infrastructure/        # GORM / Redis / 文件存储 / 事件 / 推送实现
│   └── container/             # 依赖注入容器
├── pkg/                       # 跨模块通用工具（JWT、错误码、响应格式）
├── migrations/                # 数据库迁移（4 个版本）
├── configs/                   # 配置文件
├── deployments/               # Dockerfile / docker-compose / k8s
├── scripts/                   # 辅助脚本
└── docs/                      # 文档（API、计划、规格）
```

## 本地开发

### 前置条件

- Go 1.22+
- Docker & docker-compose

### 一键启动

```bash
# 1. 复制配置文件
cp configs/config.example.yaml configs/config.local.yaml

# 2. 编辑配置（按需修改）
# vi configs/config.local.yaml

# 3. 启动基础设施（PostgreSQL + Redis + MinIO）
docker-compose -f deployments/docker-compose.yml up -d

# 4. 运行数据库迁移
go run cmd/server/main.go migrate

# 5. 启动服务
go run cmd/server/main.go serve
```

### 验证

```bash
# Health check
curl http://localhost:8080/health

# 期望输出:
# {"status":"ok","database":"connected","redis":"connected","uptime":"2s"}
```

## 环境变量

所有配置均可通过环境变量覆盖，适用于云上部署：

| 变量 | 说明 | 默认值 |
|---|---|---|
| `APP_ENV` | 运行环境 | `development` |
| `APP_SERVER_PORT` | 服务端口 | `8080` |
| `APP_DATABASE_HOST` | 数据库地址 | `localhost` |
| `APP_DATABASE_PORT` | 数据库端口 | `5432` |
| `APP_DATABASE_USER` | 数据库用户 | `productapp` |
| `APP_DATABASE_PASSWORD` | 数据库密码 | — |
| `APP_DATABASE_DBNAME` | 数据库名称 | `productsystem` |
| `APP_DATABASE_SSLMODE` | SSL 模式 | `disable` |
| `APP_REDIS_URL` | Redis 连接串 | `redis://localhost:6379` |
| `APP_JWT_SECRET` | JWT 密钥 | — |
| `APP_CONFIG_PATH` | 配置文件路径 | `configs/config.local.yaml` |
| `APP_STORAGE_DRIVER` | 存储驱动 | `local` |
| `APP_STORAGE_LOCAL_BASE_PATH` | 本地存储路径 | `./uploads` |
| `APP_STORAGE_S3_ENDPOINT` | S3 端点 | `http://localhost:9000` |
| `APP_STORAGE_S3_BUCKET` | S3 存储桶 | `products` |
| `APP_STORAGE_S3_REGION` | S3 区域 | `us-east-1` |
| `APP_STORAGE_S3_ACCESS_KEY` | S3 访问密钥 | — |
| `APP_STORAGE_S3_SECRET_KEY` | S3 秘密密钥 | — |

## 云存储驱动切换

系统支持本地磁盘和 S3 兼容存储两种模式，通过配置切换，**无需修改代码**。

### 本地开发（默认）

```yaml
# configs/config.local.yaml
storage:
  driver: local
  local:
    base_path: ./uploads
```

### 云端部署（S3）

```yaml
# configs/config.local.yaml（或通过环境变量）
storage:
  driver: s3
  s3:
    endpoint: https://s3.ap-northeast-1.amazonaws.com
    bucket: your-app-bucket
    region: ap-northeast-1
    access_key: YOUR_ACCESS_KEY
    secret_key: YOUR_SECRET_KEY
```

或通过环境变量切换（无需改配置文件）：

```bash
APP_STORAGE_DRIVER=s3 \
APP_STORAGE_S3_ENDPOINT=https://minio.example.com \
APP_STORAGE_S3_BUCKET=products \
APP_STORAGE_S3_ACCESS_KEY=minioadmin \
APP_STORAGE_S3_SECRET_KEY=minioadmin \
go run cmd/server/main.go serve
```

### 本地以 S3 模式测试（使用 MinIO）

docker-compose 已包含 MinIO 服务（`localhost:9000`），可直接切换：

```bash
# 方式一：通过环境变量
APP_STORAGE_DRIVER=s3 \
APP_STORAGE_S3_ENDPOINT=http://localhost:9000 \
APP_STORAGE_S3_ACCESS_KEY=minioadmin \
APP_STORAGE_S3_SECRET_KEY=minioadmin \
APP_STORAGE_S3_BUCKET=products \
go run cmd/server/main.go serve
```

## 架构约束

- `domain/` 和 `usecase/` 禁止 import 任何具体技术实现（gorm.io、go-redis、aws-sdk 等）
- 所有外部集成点必须先在 `domain/port/` 定义接口，再在 `infrastructure/` 实现
- 路由必须通过 `router.Public()` 或 `router.Protected()` 注册，禁止在 handler 内部做鉴权

## 开发命令

```bash
go run cmd/server/main.go serve    # 启动服务
go run cmd/server/main.go migrate  # 运行迁移
go test ./... -v                    # 运行测试
go build ./...                      # 编译
go vet ./...                        # 静态检查
grep -r "gorm.io\|go-redis\|aws-sdk" internal/domain internal/usecase  # 架构约束检查
```

## Docker 构建

```bash
# 构建镜像
docker build -f deployments/Dockerfile -t ecommerce-app:latest .

# 查看镜像
docker images | grep ecommerce-app

# 运行容器（需先启动 PostgreSQL 和 Redis）
docker run -p 8080:8080 --env-file .env ecommerce-app:latest
```

## 部署

详见文档：
- [AWS 部署指南](docs/deploy/aws-deployment-guide.md)
- [K8s 部署说明](deployments/k8s/README.md)
