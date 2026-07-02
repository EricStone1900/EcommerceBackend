# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## 项目简介

这是一个使用 Go 实现的商品管理系统后端，包含用户认证（JWT + RBAC）、商品 CRUD、文件上传、事件驱动的微服务预留点，以及 APNs 推送预留。

**当前状态：Phase 0 — 项目脚手架。** 暂无任何 Go 代码。目录结构、go.mod、基础设施代码尚未创建。后续分阶段实现。

---

## 当前要使用的命令（Phase 0 → Phase 1 启动）

项目 init（首次操作）：
```bash
go mod init github.com/EricStone1900/ecommerce-backend
go mod tidy
```

开发服务器：
```bash
go run cmd/server/main.go serve        # 启动应用
go run cmd/server/main.go migrate      # 运行数据库迁移
```

测试：
```bash
go test ./internal/usecase/... -cover  # 用例层覆盖率
go test ./... -v                       # 全部测试（详细输出）
```

代码质量：
```bash
gofmt -w . && goimports -w .           # 格式化 + import 排序
golangci-lint run ./...                # lint 检查
```

架构约束检查：
```bash
grep -r "gorm.io\|go-redis\|aws-sdk" internal/domain internal/usecase
```

Swagger 文档生成（代码就绪后）：
```bash
swag init -g cmd/server/main.go -o docs/
```

构建镜像：
```bash
docker build -t product-system .
```

压测（50 并发，持续 1 分钟）：
```bash
k6 run --vus 50 --duration 1m scripts/loadtest.js
```

---

## 架构总览

采用 **Clean Architecture + 端口适配器（Hexagonal）** 双重约束：

```
┌─────────────────────────────────────────────┐
│              interface 层                    │  ← HTTP handler / middleware / router
│   Gin handler · RBAC中间件 · 路由表          │
├─────────────────────────────────────────────┤
│              usecase 层                      │  ← 业务编排，只依赖 domain 接口
│   auth / product / upload / push / assistant │
├─────────────────────────────────────────────┤
│              domain 层                       │  ← 纯业务实体 + 接口定义（端口）
│   entity/ · port/（Repository·Storage·      │
│   EventPublisher·Notifier 等接口）           │
├─────────────────────────────────────────────┤
│           infrastructure 层                  │  ← 具体技术实现，可替换
│   GORM · Redis · S3/Local · NATS · APNs桩  │
└─────────────────────────────────────────────┘
```

**数据流方向：** interface → usecase → domain（接口）← infrastructure（实现）

### 层间依赖

| 层 | 可依赖 | 禁依赖 |
|---|---|---|
| domain | 仅 Go 标准库 | 任何第三方包、ORM、云 SDK |
| usecase | domain（接口），标准库 | gorm.io、go-redis、aws-sdk、云 SDK |
| interface | usecase、domain | GORM/Redis/AWS 等具体实现 |
| infrastructure | domain（实现接口）、第三方包 | — |

---

## 阶段实施路线图

| 阶段 | 内容 | 产出物 |
|---|---|---|
| **Phase 1** | 基础骨架 + 用户认证 | go.mod、main.go、Gin 路由、JWT 认证、用户注册/登录、PostgreSQL + Redis 基础设施 |
| **Phase 2** | 商品 CRUD + 文件上传 | 商品管理接口、文件上传（本地存储 + S3接口）、事件发布占位 |
| **Phase 3** | 推送 + 智能助手 | APNs 桩实现、推送任务编排、智能助手 mock 接口 |
| **Phase 4** | 微服务化 | NATS 实际接入、APNs 实现、AI 助手接入、可观测性完善 |

各阶段不阻塞，按需迭代。微服务预留点（EventPublisher·Notifier·AssistantPort）的接口签名一旦确定不要随意修改。

---

## 铁律约束（违反即要求重写）

### 层级隔离
- `domain/` 和 `usecase/` 目录下的任何 `.go` 文件，**禁止 import** 以下包：
  - `gorm.io/gorm`
  - `github.com/redis/go-redis`
  - `github.com/aws/aws-sdk-go`
  - 任何云厂商 SDK
  - 任何具体数据库驱动
- 验证方式：`grep -r "gorm.io\|go-redis\|aws-sdk" internal/domain internal/usecase` 结果必须为空

### 接口优先
- 所有外部集成点必须先在 `domain/port/` 定义接口，再在 `infrastructure/` 实现
- 接口方法签名必须包含 `context.Context`，返回 `error`，体现"可能失败的远程调用"语义
- 示例（正确）：`Upload(ctx context.Context, data []byte, filename string) (url string, err error)`
- 示例（错误）：`Upload(data []byte, filename string) string`

### 路由权限声明
- 所有路由必须通过 `router.Public()` 或 `router.Protected()` 注册，**禁止**在 handler 内部做鉴权判断
- `Protected` 必须显式声明允许的角色列表，不允许隐式放行

### 删除策略
- 所有删除一律软删除（`deleted_at`），禁止物理删除
- 迁移脚本放 `migrations/`，用 golang-migrate 管理版本

---

## 技术栈与版本

| 组件 | 选型 | 说明 |
|---|---|---|
| 语言 | Go 1.22+ | 使用泛型、range-over-int 等现代特性 |
| Web 框架 | Gin | 路由、中间件、参数绑定 |
| ORM | GORM | 只在 infrastructure 层出现 |
| 数据库 | PostgreSQL 16 | 本地用 docker-compose 启动 |
| 缓存 | Redis 7 | token黑名单、限流计数器 |
| 文件存储 | 本地磁盘（dev）/ S3（prod）| 通过 Storage 接口切换 |
| 事件总线 | 进程内 channel（dev）/ NATS（prod）| 通过 EventPublisher 接口切换 |
| 推送 | Stub 日志（dev）/ APNs（后续）| 通过 Notifier 接口切换 |
| 鉴权 | JWT（access 15min + refresh 7d）| access token 无状态，refresh 存 Redis |
| 配置 | Viper | 环境变量优先级高于配置文件 |
| 日志 | zap | JSON 结构化日志，dev 模式用 console 格式 |
| API 文档 | swaggo | 注释自动生成 OpenAPI/Swagger |
| 容器 | Docker 多阶段构建 + docker-compose | 最终镜像基于 alpine |

---

## 角色与权限矩阵

| 接口 | 未登录 | customer | member | admin |
|---|---|---|---|---|
| 商品详情 `GET /products/:id` | ✅ | ✅ | ✅ | ✅ |
| 商品列表 `GET /products` | ❌ 401 | ✅ | ✅ | ✅ |
| 创建商品 `POST /products` | ❌ | ❌ 403 | ❌ 403 | ✅ |
| 编辑/删除商品 | ❌ | ❌ 403 | ❌ 403 | ✅ |
| 文件上传 | ❌ | ✅ | ✅ | ✅ |
| 注册/登录 | ✅（公开）| — | — | — |
| 推送 token 注册 | ❌ | ✅ | ✅ | ✅ |

---

## 本地开发环境

### 依赖服务（docker-compose 包含）

| 服务 | 地址 | 用途 |
|---|---|---|
| PostgreSQL | localhost:5432 | 主数据库 |
| Redis | localhost:6379 | 缓存/token黑名单 |
| MinIO | localhost:9000 | 本地 S3 兼容存储 |

### 本地启动

```bash
cp configs/config.example.yaml configs/config.local.yaml
# 编辑 config.local.yaml 填入本地值（数据库密码等）
docker-compose up -d
go run cmd/server/main.go migrate
go run cmd/server/main.go serve
```

### 关键环境变量

```bash
APP_ENV=development          # development | production
APP_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_NAME=productsystem
DB_USER=productapp
DB_PASSWORD=your_password
REDIS_URL=redis://localhost:6379
JWT_SECRET=your_jwt_secret
STORAGE_DRIVER=local         # local | s3
```

---

## 代码规范

### 响应格式

```json
// 成功
{ "code": 0, "data": {...}, "message": "success" }
// 错误
{ "code": 40001, "data": null, "message": "用户名或密码错误" }
// 分页列表
{ "code": 0, "data": { "total": 100, "page": 1, "page_size": 20, "list": [...] } }
```

### 图片命名规则

由于此项目使用 MinIO 兼容的 S3 接口，上传的图片通过代理 URL 返回。定义图片命名规则以约束图片 URL 在请求方使用：

```
{uuid}.{ext}  示例: a1b2c3d4.jpg
```

### 错误处理
- 使用 `fmt.Errorf("operation failed: %w", err)` 包装错误，保留 stack
- 使用 `errors.Is()` / `errors.As()` 判断错误类型，禁止直接比较 error string
- 自定义业务错误类型放 `pkg/errors/`，包含错误码和 HTTP 状态码映射
- handler 层统一捕获错误，转换为统一的 JSON 响应格式

### 并发
- 数据库连接池：`MaxOpenConns=100, MaxIdleConns=10, ConnMaxLifetime=1h`（可通过配置覆盖）
- 所有外部调用都要传 `context`，设置合理的 timeout
- 不允许在 goroutine 里直接操作共享变量而不加锁

---

## Claude Code 工作指引

### 可用的 Go 技能

本仓库已安装 40+ 个 [cc-skills-golang](https://github.com/samber/cc) 技能。处理以下任务时，Claude 会自动触发对应技能：

| 任务场景 | 自动触发技能 |
|---|---|
| 写 API handler | `golang-gin`, `golang-error-handling` |
| 定义接口/结构体 | `golang-structs-interfaces`, `golang-naming` |
| 写测试 | `golang-testing`, `golang-stretchr-testify` |
| 数据库操作 | `golang-database`, `golang-project-layout`（迁移目录） |
| CLI 命令 | `golang-spf13-cobra`, `golang-cli` |
| 配置加载 | `golang-spf13-viper` |
| DI 容器 | `golang-uber-fx` 或 `golang-google-wire`（按选型） |
| 并发 | `golang-concurrency`, `golang-context` |
| 错误处理 | `golang-error-handling` |
| Swagger 文档 | `golang-swagger` |
| 安全审计 | `golang-security` |
| CI/CD | `golang-continuous-integration` |

### 实施流程（参考全局指令）

1. **复杂功能** → 先用 `planner` agent 制定计划，写入 `docs/plans/<日期>-<名称>.md`
2. **TDD** → 使用 `tdd-guide` agent 强制执行"先写测试 → 实现 → 重构"（要求 usecase 层覆盖 ≥70%）
3. **代码审查** → 实现后立即使用 `code-reviewer` agent
4. **提交** → 遵循 Conventional Commits（feat/fix/refactor/test/chore）
5. **验证架构** → 每次提交前运行 `grep -r "gorm.io" internal/domain internal/usecase` 确保为空

### docs/plans/ 文档命名

```
docs/plans/<YYYYMMDD>-<简短英文描述>.md
示例：docs/plans/20260701-auth-module.md
```

---

## 微服务预留点

这是架构设计的核心价值，每个预留点都必须有真实的接口和事件（不仅仅是注释）：

```
文件上传完成
    └─→ EventPublisher.Publish("file.uploaded", payload)
            └─→ 【当前】进程内订阅者：更新文件状态为 processed，打日志
            └─→ 【未来】OCR 微服务订阅此事件，独立进程消费

商品详情页
    └─→ AssistantPort.GenerateDescription(ctx, productID)
            └─→ 【当前】mock 实现：返回固定文案
            └─→ 【未来】替换为真实 AI 微服务的 HTTP/gRPC 调用

用户行为触发推送
    └─→ Notifier.SendPush(ctx, userID, title, body)
            └─→ 【当前】stub 实现：写 push_logs 表
            └─→ 【未来】替换为 APNs 实现（infrastructure/push/apns/）
```

---

## 验证命令速查

```bash
# 架构层级隔离检查（必须输出为空）
grep -r "gorm.io\|go-redis\|aws-sdk" internal/domain internal/usecase

# 运行全部测试
go test ./... -v

# 测试覆盖率（usecase 层目标 ≥ 70%）
go test ./internal/usecase/... -cover

# 构建镜像
docker build -t product-system .

# 压测（50并发，持续1分钟）
k6 run --vus 50 --duration 1m scripts/loadtest.js

# Swagger 文档生成
swag init -g cmd/server/main.go -o docs/

# 数据库迁移
go run cmd/server/main.go migrate
```
