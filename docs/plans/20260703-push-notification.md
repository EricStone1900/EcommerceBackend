# Phase 5 开发计划 — 推送服务与设备 Token 管理

## 背景

Phase 1-4 已完成：分层骨架、用户认证与 RBAC、商品管理、文件上传与事件机制。本阶段实现推送服务的设备 Token 注册能力，预留 APNs 真实推送接入点（本阶段仅实现 Stub 打日志版本）。

## 技术决策

| 决策 | 选择 | 理由 |
|---|---|---|
| Token 幂等 | PostgreSQL ON CONFLICT DO UPDATE | 消除竞态，单次 DB 往返，一致性好 |
| 删除策略 | user_id + device_token 联合定位 | 支持同一用户多设备注销单个 token |
| 错误码 | 7001-7003（新段） | 不与 5001(InternalError) 冲突 |
| Platform | 实体预留 `android`，当前只接受 `ios` | 避免未来加 Android 支持时改 schema |
| SendTest | 只给自己发（请求者 userID） | 当前无管理员主动推送需求 |

## 文件清单

### 新建文件（17 个）

| 层 | 文件 | 说明 |
|---|---|---|
| **Entity** | `internal/domain/entity/push_token.go` | PushToken 实体 + Platform 枚举 |
| **Port** | `internal/domain/port/push_token_repository.go` | PushTokenRepository 接口 |
| | `internal/domain/port/notifier.go` | Notifier 接口（SendPush） |
| **Infra GORM** | `internal/infrastructure/persistence/gorm/push_token_model.go` | GORM model（含 gorm.DeletedAt） |
| | `internal/infrastructure/persistence/gorm/push_token_repo.go` | GORM 实现（upsert + 唯一索引） |
| **Infra Stub** | `internal/infrastructure/push/stub/notifier.go` | Notifier 桩实现（打日志） |
| **Infra APNs** | `internal/infrastructure/push/apns/doc.go` | 注释文件，说明接入 APNs 所需配置 |
| **Use Case** | `internal/usecase/push/push.go` | PushUseCase + DTOs + 构造函数 |
| | `internal/usecase/push/register_token.go` | 注册设备 Token |
| | `internal/usecase/push/delete_token.go` | 删除设备 Token |
| | `internal/usecase/push/send_test.go` | 发送测试推送 |
| | `internal/usecase/push/helpers_test.go` | Mock 依赖 |
| | `internal/usecase/push/register_token_test.go` | 注册测试 |
| | `internal/usecase/push/delete_token_test.go` | 删除测试 |
| | `internal/usecase/push/send_test_test.go` | 发送测试 |
| **Handler** | `internal/interface/http/handler/push.go` | HTTP handler |
| | `internal/interface/http/handler/push_test.go` | Handler 测试 |
| **迁移** | `migrations/000004_create_push_tokens_table.up.sql` | push_tokens 建表 |
| | `migrations/000004_create_push_tokens_table.down.sql` | push_tokens 删表 |
| **文档** | `docs/api/push-api.md` | 推送 API 文档 |

### 修改文件（3 个）

| 文件 | 修改内容 |
|---|---|
| `pkg/errors/errors.go` | 新增 3 个推送错误码（7001-7003） |
| `internal/container/container.go` | 新增 PushTokenRepo, Notifier, PushUseCase, PushHandler |
| `cmd/server/main.go` | 注册 3 条推送路由 |

## 实施顺序

### 任务 1：Domain 层

**entity/push_token.go**
```go
type Platform string
const ( PlatformIOS Platform = "ios"; PlatformAndroid Platform = "android" )
type PushToken struct { ID, UserID(uint), DeviceToken(string), Platform(Platform), CreatedAt, UpdatedAt }
```

**port/push_token_repository.go**
```go
type PushTokenRepository interface {
    Create(ctx, *entity.PushToken) error
    GetByUserID(ctx, uint) ([]*entity.PushToken, error)
    DeleteByUserAndDevice(ctx, userID uint, deviceToken string) error
}
```

**port/notifier.go**
```go
type Notifier interface {
    SendPush(ctx, userID uint, deviceToken, title, body string) error
}
```

### 任务 2：错误码

**修改 `pkg/errors/errors.go`**
- `CodePushTokenNotFound = 7001` → ErrPushTokenNotFound (404)
- `CodePushSendFailed = 7002` → ErrPushSendFailed (500)
- `CodeInvalidPlatform = 7003` → ErrInvalidPlatform (400)

### 任务 3：基础设施层

**gorm/push_token_model.go** — model 含 `gorm.DeletedAt` + 复合唯一索引 `(user_id, device_token)` + `toEntity()`/`toPushTokenModel()`

**gorm/push_token_repo.go** — Create 使用 `Clauses(clause.OnConflict{...})` 实现 upsert，写入后回读填充 ID。GetByUserID 按 user_id 查询未被软删除的 token。DeleteByUserAndDevice 执行软删除。

**push/stub/notifier.go** — 接收 userID, deviceToken, title, body 打结构化日志。可选写入 push_logs 表（当前暂不实现 logs）。

**push/apns/doc.go** — 包注释文件说明未来接入 APNs 需要的配置项和注意事项。

### 任务 4：Use Case 层

**push/push.go** — DTO 定义、PushUseCase 结构体（依赖 PushTokenRepository + Notifier + Logger）

**push/register_token.go**
- 校验 device_token 非空
- 校验 platform 为 "ios"（预留 "android"）
- 调用 repo.Create（upsert 幂等）

**push/delete_token.go**
- 调用 repo.DeleteByUserAndDevice(idempotent — 不存在也返回 nil)

**push/send_test.go**
- 通过 repo.GetByUserID 查找用户的所有 token
- 若无 token → ErrPushTokenNotFound
- 对每个 token 调用 notifier.SendPush
- 返回 `{"sent": N}`

### 任务 5：HTTP Handler 层

**handler/push.go**
- PushUseCase 接口 + PushHandler 结构体
- RegisterToken: POST, JSON body → 200
- DeleteToken: DELETE, JSON body → 200
- SendTest: POST, no body → 200

### 任务 6：DI 容器 + 路由

**container.go** — 新增 4 个字段，在 Phase 4 后添加初始化

**main.go**
```go
r.Protected("POST", "/api/v1/push/token", c.PushHandler.RegisterToken)
r.Protected("DELETE", "/api/v1/push/token", c.PushHandler.DeleteToken)
r.Protected("POST", "/api/v1/push/test", c.PushHandler.SendTest)
```

### 任务 7：迁移 + 文档

**push_tokens 表** — id, user_id(→users), device_token, platform, created_at, updated_at, deleted_at
唯一索引 `idx_push_tokens_user_device ON (user_id, device_token)`

**docs/api/push-api.md** — 3 个接口文档

## 验证

```bash
# 架构约束
grep -r "gorm.io\|go-redis\|aws-sdk" internal/domain internal/usecase

# 全量测试
go test ./... -count=1

# 编译检查
go build ./... && go vet ./...

# 端到端
# 登录 → 注册设备 token → 发送测试推送 → 删除 token
```
