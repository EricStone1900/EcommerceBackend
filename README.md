# 商品管理系统 (Product Management System)

一个使用 Go 实现的电商后台服务，采用 **Clean Architecture + Hexagonal（端口适配器）** 双重架构约束。

## 架构概览

```
┌─────────────────────────────────────────────────────┐
│                 interface 层 (Gin)                    │
│   HTTP handler · RBAC 中间件 · router.Public/Protected │
├─────────────────────────────────────────────────────┤
│                  usecase 层                            │
│   auth / product / upload / push / assistant          │
│   核心业务编排，只依赖 domain 层接口                    │
├─────────────────────────────────────────────────────┤
│                  domain 层                             │
│   entity/ · port/ (Repository·Storage·EventPublisher· │
│   Notifier·AssistantPort 等接口)                      │
│   纯业务定义，零外部依赖                                │
├─────────────────────────────────────────────────────┤
│               infrastructure 层                        │
│   GORM · Redis · S3/Local · EventBus · APNs Stub     │
│   domain 接口的具体实现，可替换                         │
└─────────────────────────────────────────────────────┘
```

**数据流方向：** interface → usecase → domain（接口）← infrastructure（实现）

**层间隔离原则：**
- `domain/` 和 `usecase/` **禁止** import 任何第三方包（gorm.io、go-redis、aws-sdk 等）
- 所有外部集成点先定义接口（`domain/port/`），再实现（`infrastructure/`）
- 路由通过 `router.Public()` / `router.Protected()` 注册，禁止 handler 内鉴权

验证命令：`grep -r "gorm.io\|go-redis\|aws-sdk" internal/domain internal/usecase` — 必须为空

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
| 文件存储 | 本地磁盘 / S3（可切换） | 本地开发用 local，云端用 s3 |
| 事件总线 | 进程内 Channel（预留 NATS） | 异步处理（如 OCR 模拟） |
| 推送 | Stub 日志（预留 APNs） | 本地验证推送链路 |
| 测试 | testify + GORM SQLite | 单元测试 + 集成测试 |
| 性能 | k6 | 压测 |
| API 文档 | swaggo / Swagger | OpenAPI 自动生成 |
| 容器 | Docker + docker-compose | 本地开发 / 部署 |

## API 文档

启动服务后访问：http://localhost:8080/swagger/index.html

Swagger 文档基于 handler 注释自动生成（swaggo），更新 handler 后重新生成：

```bash
swag init -g cmd/server/main.go -o docs/
```

## 目录结构

```
.
├── cmd/server/main.go         # 程序入口 + 路由表
├── internal/
│   ├── domain/
│   │   ├── entity/            # 纯业务实体
│   │   └── port/              # Repository / Storage / Notifier 等接口定义
│   ├── usecase/               # 业务用例编排（auth / product / upload / push / assistant）
│   ├── interface/http/
│   │   ├── handler/           # Gin handler（含 swagger 注解）
│   │   ├── middleware/        # JWT 认证 + RBAC 中间件
│   │   └── router/            # 路由注册器
│   ├── infrastructure/
│   │   ├── cache/redis/       # Redis 连接 + token store
│   │   ├── config/            # 配置加载
│   │   ├── event/             # 事件总线
│   │   ├── log/               # 日志初始化
│   │   ├── persistence/gorm/  # GORM 模型 + 仓库实现
│   │   ├── push/              # 推送服务（stub + APNs 预留）
│   │   └── storage/           # 文件存储（local + s3）
│   └── container/             # 依赖注入容器
├── pkg/
│   ├── errors/                # 业务错误码（17 个预定义错误）
│   ├── jwt/                   # JWT 工具（access 15min + refresh 7d）
│   └── response/              # 统一响应格式
├── tests/integration/         # HTTP 集成测试（SQLite 内存数据库）
├── migrations/                # 数据库迁移（4 个版本）
├── configs/                   # 配置文件
├── deployments/               # Dockerfile / docker-compose / k8s
│   └── k8s/                   # Kubernetes 部署清单
├── scripts/
│   └── loadtest.js            # k6 压测脚本
└── docs/
    ├── api/                   # 手写 API 文档
    ├── deploy/                # 部署指南
    ├── plans/                 # 开发计划
    └── specs/                 # 阶段规格文档
```

## 本地开发

### 前置条件

- Go 1.22+
- Docker & docker-compose
- swag CLI（可选，用于 Swagger 文档生成）

### 一键启动

```bash
# 1. 复制配置文件
cp configs/config.example.yaml configs/config.local.yaml

# 2. 编辑配置（按需修改数据库密码等）
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

Swagger UI：访问 http://localhost:8080/swagger/index.html

## 环境变量

所有配置均可通过环境变量覆盖，适用于云上部署：

| 变量 | 说明 | 默认值 |
|---|---|---|
| `APP_ENV` | 运行环境（development/production） | `development` |
| `APP_SERVER_PORT` | 服务端口 | `8080` |
| `APP_DATABASE_HOST` | 数据库地址 | `localhost` |
| `APP_DATABASE_PORT` | 数据库端口 | `5432` |
| `APP_DATABASE_USER` | 数据库用户 | `productapp` |
| `APP_DATABASE_PASSWORD` | 数据库密码 | — |
| `APP_DATABASE_DBNAME` | 数据库名称 | `productsystem` |
| `APP_DATABASE_SSLMODE` | SSL 模式 | `disable` |
| `APP_REDIS_URL` | Redis 连接串 | `redis://localhost:6379` |
| `APP_JWT_SECRET` | JWT 签名密钥 | — |
| `APP_CONFIG_PATH` | 配置文件路径 | `configs/config.local.yaml` |
| `APP_STORAGE_DRIVER` | 存储驱动（local/s3） | `local` |
| `APP_STORAGE_LOCAL_BASE_PATH` | 本地存储路径 | `./uploads` |
| `APP_STORAGE_S3_ENDPOINT` | S3 端点 | `http://localhost:9000` |
| `APP_STORAGE_S3_BUCKET` | S3 存储桶 | `products` |
| `APP_STORAGE_S3_REGION` | S3 区域 | `us-east-1` |
| `APP_STORAGE_S3_ACCESS_KEY` | S3 访问密钥 | — |
| `APP_STORAGE_S3_SECRET_KEY` | S3 秘密密钥 | — |

## 云存储驱动切换

系统支持本地磁盘和 S3 兼容存储两种模式，通过配置或环境变量切换，**无需修改代码**。

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
# configs/config.local.yaml
storage:
  driver: s3
  s3:
    endpoint: https://s3.ap-northeast-1.amazonaws.com
    bucket: your-app-bucket
    region: ap-northeast-1
    access_key: YOUR_ACCESS_KEY
    secret_key: YOUR_SECRET_KEY
```

### 通过环境变量切换（无需改配置文件）

```bash
APP_STORAGE_DRIVER=s3 \
APP_STORAGE_S3_ENDPOINT=https://minio.example.com \
APP_STORAGE_S3_BUCKET=products \
APP_STORAGE_S3_ACCESS_KEY=minioadmin \
APP_STORAGE_S3_SECRET_KEY=minioadmin \
go run cmd/server/main.go serve
```

### 本地以 S3 模式测试（使用 MinIO）

```bash
APP_STORAGE_DRIVER=s3 \
APP_STORAGE_S3_ENDPOINT=http://localhost:9000 \
APP_STORAGE_S3_ACCESS_KEY=minioadmin \
APP_STORAGE_S3_SECRET_KEY=minioadmin \
APP_STORAGE_S3_BUCKET=products \
go run cmd/server/main.go serve
```

## 运行测试

### 单元测试 + 集成测试

```bash
# 全部测试（含集成测试）
go test ./... -v -count=1

# 指定包测试
go test ./internal/usecase/... -v -count=1
go test ./internal/interface/http/handler/... -v -count=1
go test ./tests/integration/... -v -count=1

# 覆盖率
go test ./... -cover -count=1

# 指定包覆盖率（usecase 层目标 ≥ 70%）
go test ./internal/usecase/... -cover -count=1
```

### 架构约束检查

```bash
# 验证 domain/usecase 层没有引入外部依赖
grep -r "gorm.io\|go-redis\|aws-sdk" internal/domain internal/usecase || echo "PASS: 无架构违规"
```

## 负载压测

系统要求支持 50+ 并发。压测脚本使用 [k6](https://k6.io/)（需安装）：

```bash
# macOS 安装 k6
brew install k6

# 启动完整系统（如果尚未启动）
docker-compose -f deployments/docker-compose.yml up -d
go run cmd/server/main.go migrate

# 执行压测（50 并发，持续 1 分钟）
k6 run --vus 50 --duration 1m scripts/loadtest.js
```

压测覆盖以下高频路径：
- `POST /api/v1/auth/login` — 登录
- `GET /api/v1/products?page=1&page_size=20` — 商品列表（需认证）
- `GET /api/v1/products/:id` — 商品详情（公开）

输出：P50/P95/P99 延迟、错误率、吞吐量。

## 构建

### Docker 构建

```bash
docker build -f deployments/Dockerfile -t ecommerce-app:latest .

# 运行容器（需先启动 PostgreSQL 和 Redis）
docker run -p 8080:8080 --env-file .env ecommerce-app:latest
```

构建采用多阶段构建 + alpine 基础镜像，产物小于 15MB，CGO_ENABLED=0。

## 部署

### AWS 部署

完整的 AWS 基础设施清单和配置对照详见 [AWS 部署指南](docs/deploy/aws-deployment-guide.md)：

| 服务 | AWS 产品 | 规格建议 |
|---|---|---|
| 计算 | ECS Fargate | 0.25-0.5 vCPU, 512MB |
| 数据库 | RDS PostgreSQL | db.t3.medium |
| 缓存 | ElastiCache Redis | cache.t3.micro |
| 对象存储 | S3 | Standard |
| 负载均衡 | ALB | — |

### Kubernetes 部署

详细步骤见 [K8s 部署说明](deployments/k8s/README.md)。

```bash
# 应用 K8s 清单
kubectl apply -f deployments/k8s/configmap.yaml
kubectl apply -f deployments/k8s/secret.yaml
kubectl apply -f deployments/k8s/deployment.yaml
kubectl apply -f deployments/k8s/service.yaml
```

## 角色与权限矩阵

| 接口 | 未登录 | customer | member | admin |
|---|---|---|---|---|
| 商品详情 `GET /products/:id` | ✅ | ✅ | ✅ | ✅ |
| 商品列表 `GET /products` | ❌ 401 | ✅ | ✅ | ✅ |
| 创建商品 `POST /products` | ❌ | ❌ 403 | ❌ 403 | ✅ |
| 编辑/删除商品 | ❌ | ❌ 403 | ❌ 403 | ✅ |
| 文件上传 `POST /upload` | ❌ | ✅ | ✅ | ✅ |
| 注册/登录 | ✅（公开） | — | — | — |
| 推送 token 注册 | ❌ | ✅ | ✅ | ✅ |

## API 端点一览

| 方法 | 路径 | 认证 | 角色 | 说明 |
|---|---|---|---|---|
| GET | `/health` | 公开 | — | 健康检查 |
| GET | `/swagger/*any` | 公开 | — | Swagger UI（开发环境） |
| POST | `/api/v1/auth/register` | 公开 | — | 注册 |
| POST | `/api/v1/auth/login` | 公开 | — | 登录 |
| POST | `/api/v1/auth/logout` | 受保护 | 任意 | 登出 |
| POST | `/api/v1/auth/refresh` | 受保护 | 任意 | 刷新令牌 |
| GET | `/api/v1/auth/me` | 受保护 | 任意 | 当前用户 |
| GET | `/api/v1/products` | 受保护 | 任意 | 商品列表 |
| GET | `/api/v1/products/:id` | 公开 | — | 商品详情 |
| POST | `/api/v1/products` | 受保护 | admin | 创建商品 |
| PUT | `/api/v1/products/:id` | 受保护 | admin | 更新商品 |
| DELETE | `/api/v1/products/:id` | 受保护 | admin | 删除商品 |
| POST | `/api/v1/upload` | 受保护 | 任意 | 文件上传 |
| POST | `/api/v1/push/token` | 受保护 | 任意 | 注册推送令牌 |
| DELETE | `/api/v1/push/token` | 受保护 | 任意 | 删除推送令牌 |
| POST | `/api/v1/push/test` | 受保护 | 任意 | 测试推送 |

## 微服务预留接入指引

系统设计了三个微服务预留点。当前阶段使用本地 mock 实现，接入真实服务**无需修改接口签名**。

### 1. OCR 微服务接入

**当前状态：** 文件上传后通过 `EventPublisher.Publish("file.uploaded", payload)` 发布事件，`fileprocessing/handler.go` 订阅事件并模拟处理。

**接入步骤：**
1. ✅ **不需要修改 `domain/port/EventPublisher` 接口** — 事件格式 `{"FileID", "FileName", "FileType", "UploadedBy"}` 已在 `fileprocessing/handler.go` 中定义
2. **实现远程事件消费者**：将 `infrastructure/event/bus.go` 中的内存 channel 替换为 NATS 或 Kafka
3. **部署 OCR 微服务**：订阅 `file.uploaded` 事件，处理后将结果通过 OCR 微服务 API 写回
4. **通知方式**：OCR 完成后可通过 `EventPublisher.Publish("file.processed", payload)` 通知主系统

**需要修改的文件：**
- `internal/domain/port/event_publisher.go` — 接口不动
- `internal/infrastructure/event/bus.go` — 替换为 NATS 客户端
- `internal/usecase/fileprocessing/handler.go` — 添加真实 OCR 调用

### 2. AI 智能助手接入

**当前状态：** `internal/usecase/assistant/assistant.go` 中的 `mockAssistant` 返回固定文案，模拟 50ms 延迟。

**接入步骤：**
1. ✅ **不需要修改 `domain/port/AssistantPort` 接口** — `GenerateProductDescription(ctx, productID uint) (string, error)` 签名不变
2. **创建真实实现**：在 `internal/infrastructure/assistant/` 下创建新文件，实现 `port.AssistantPort`
3. 调用远程 AI 微服务的 HTTP/gRPC API，使用 `context.Context` 控制超时
4. **替换容器注入**：修改 `internal/container/container.go`，将 `port.AssistantPort` 从 `mockAssistant` 替换为真实实现

**需要修改的文件：**
- `internal/domain/port/assistant_port.go` — 接口不动
- `internal/infrastructure/assistant/` — 新建，实现真实 AI 调用
- `internal/container/container.go` — 替换注入

### 3. APNs 推送接入

**当前状态：** `internal/infrastructure/push/stub/notifier.go` 记录日志桩实现，`internal/infrastructure/push/apns/doc.go` 提供接入指引。

**接入步骤：**
1. ✅ **不需要修改 `domain/port/Notifier` 接口** — `SendPush(ctx, userID, deviceToken, title, body) error` 签名不变
2. **创建 APNs 实现**：在 `internal/infrastructure/push/apns/` 下创建 `notifier.go` 实现 `port.Notifier`
3. 使用 `github.com/sideshow/apns2` 或苹果 HTTP/2 API 发送推送
4. **替换容器注入**：修改 `internal/container/container.go`，将 `port.Notifier` 从 `stub.Notifier` 替换为 `apns.Notifier`

**需要修改的文件：**
- `internal/domain/port/notifier.go` — 接口不动
- `internal/infrastructure/push/apns/notifier.go` — 新建，实现真实 APNs 调用
- `internal/container/container.go` — 替换注入

## 开发命令速查

```bash
go run cmd/server/main.go serve    # 启动服务
go run cmd/server/main.go migrate  # 运行迁移
go run cmd/server/main.go migrate --down  # 回滚迁移
go test ./... -v -count=1          # 全量测试
go test ./... -cover -count=1      # 测试覆盖率
go build ./...                     # 编译
go vet ./...                       # 静态检查
swag init -g cmd/server/main.go -o docs/  # 更新 Swagger 文档
docker build -f deployments/Dockerfile -t ecommerce-app:latest .  # 构建镜像
k6 run --vus 50 --duration 1m scripts/loadtest.js  # 压测
```

## 响应格式

```json
// 成功
{ "code": 0, "data": {...}, "message": "success" }

// 错误
{ "code": 40001, "data": null, "message": "用户名或密码错误" }

// 分页列表
{ "code": 0, "data": { "total": 100, "page": 1, "page_size": 20, "list": [...] } }
```

## 相关文档

- [Swagger API 文档](http://localhost:8080/swagger/index.html)（需启动服务）
- [API 参考文档](docs/api/)（手写 Markdown）
- [AWS 部署指南](docs/deploy/aws-deployment-guide.md)
- [K8s 部署说明](deployments/k8s/README.md)
- [开发规格文档](docs/specs/)
