# Phase 4 开发计划 — 文件上传与事件机制

## 背景

Phase 1-3 已完成：分层骨架、用户认证与 RBAC、商品管理 API。本阶段实现文件上传功能，并建立事件机制作为未来微服务预留点（当前用进程内事件总线，未来可替换为 NATS/RabbitMQ）。同时预留智能小助手（assistant）微服务接口占位。

## 技术决策

| 决策 | 选择 | 理由 |
|---|---|---|
| 上传端点 | 统一 `POST /api/v1/upload` + `type` form 参数 | 减少路由重复，新增文件类型只需更新校验白名单 |
| 文件存储 URL | 相对路径 `/uploads/<uuid>.<ext>` | Gin Static 文件服务直接可用，切换 S3 后变为完整 URL |
| 事件投递 | 异步 goroutine 分发 | 模拟未来消息队列的 fire-and-forget 语义 |
| FileRepository | 独立 port 接口 | upload use case 需要 Create/GetByID，fileprocessing 需要 UpdateStatus |
| Assistant 接口 | `domain/port/` → `usecase/assistant/` | 遵循端口适配器模式 |

## 文件清单

### 新建文件（21 个）

| 层 | 文件 | 说明 |
|---|---|---|
| **Domain Entity** | `internal/domain/entity/file.go` | File 实体 + FileStatus/FileType 枚举 |
| **Domain Port** | `internal/domain/port/storage.go` | Storage 接口（Upload/Delete） |
| | `internal/domain/port/event_publisher.go` | EventPublisher 接口（Publish/Subscribe） |
| | `internal/domain/port/file_repository.go` | FileRepository 接口（CRUD + UpdateStatus） |
| | `internal/domain/port/assistant_port.go` | AssistantPort 接口（GenerateProductDescription） |
| **Infra Storage** | `internal/infrastructure/storage/local/local.go` | 本地磁盘存储实现 |
| **Infra Event** | `internal/infrastructure/event/bus.go` | 进程内事件总线（Go channel） |
| | `internal/infrastructure/event/bus_test.go` | 事件总线测试 |
| **Infra GORM** | `internal/infrastructure/persistence/gorm/file_model.go` | File GORM model（含 gorm.DeletedAt） |
| | `internal/infrastructure/persistence/gorm/file_repo.go` | FileRepository GORM 实现 |
| **Use Case Upload** | `internal/usecase/upload/upload.go` | UploadUseCase + DTO + 文件类型配置 |
| | `internal/usecase/upload/handle_upload.go` | 上传处理方法（校验→存储→持久化→发布） |
| | `internal/usecase/upload/helpers_test.go` | Mock 仓库/存储/事件总线 |
| | `internal/usecase/upload/upload_test.go` | 上传用例测试（6+ 用例） |
| **Use Case File Processing** | `internal/usecase/fileprocessing/handler.go` | 事件订阅者（日志+状态回写） |
| | `internal/usecase/fileprocessing/handler_test.go` | 订阅者测试 |
| **Use Case Assistant** | `internal/usecase/assistant/assistant.go` | AssistantPort mock 实现 |
| **HTTP Handler** | `internal/interface/http/handler/upload.go` | 上传 HTTP handler |
| | `internal/interface/http/handler/upload_test.go` | Handler 测试 |
| **迁移** | `migrations/000003_create_files_table.up.sql` | files 建表 |
| | `migrations/000003_create_files_table.down.sql` | files 删表 |
| **文档** | `docs/api/upload-api.md` | 上传 API 文档 |

### 修改文件（5 个）

| 文件 | 修改内容 |
|---|---|
| `pkg/errors/errors.go` | 新增 4 个文件相关错误码（6001-6004） |
| `internal/infrastructure/config/config.go` | StorageConfig 新增 Local 字段 |
| `internal/container/container.go` | 新增 5 个字段 + 初始化逻辑 |
| `cmd/server/main.go` | 注册上传路由 + 静态文件服务 |
| `configs/config.example.yaml` / `config.local.yaml` | 添加 `storage.local.base_path` |

## 实施顺序

### 任务 1：Domain 层

**entity/file.go**
- `FileStatusPending = "pending"`, `FileStatusProcessed = "processed"`
- `FileTypeImage = "image"`, `FileTypeDocument = "document"`, `FileTypeVideo = "video"`
- `File` struct: ID, OwnerID(uint), Type(string), OriginalName(string), URL(string), Size(int64), Status(FileStatus), CreatedAt(time.Time)

**port/storage.go** — Storage 接口
**port/event_publisher.go** — EventPublisher 接口
**port/file_repository.go** — FileRepository 接口
**port/assistant_port.go** — AssistantPort 接口

### 任务 2：错误码 + 配置

**修改 `pkg/errors/errors.go`**
```go
CodeFileTooLarge = 6001     → ErrFileTooLarge (413)
CodeInvalidFileType = 6002  → ErrInvalidFileType (400)
CodeFileNotFound = 6003     → ErrFileNotFound (404)
CodeFileUploadFailed = 6004 → ErrFileUploadFailed (500)
```

**修改 `config.go`** — `LocalStorageConfig{BasePath string}`
**修改 YAML** — 添加 `local.base_path: ./uploads`

### 任务 3：基础设施层

**storage/local/local.go**
- 按 `BasePath` 写入文件，自动创建目录
- URL 格式 `/uploads/<uuid><ext>`
- Delete 检查文件存在后删除

**event/bus.go**
- `subscribers map[string][]handler` + `sync.RWMutex`
- Publish 在 goroutine 中分发，不阻塞调用方
- Close 用 `sync.WaitGroup` 等待处理完成
- 注释说明"未来替换为消息队列"

**gorm/file_model.go** — model 含 `gorm.DeletedAt` + `toEntity()`/`toModel()`
**gorm/file_repo.go** — 实现 FileRepository 接口

### 任务 4：Use Case 层

**upload/upload.go**
- 文件类型配置（type → MaxSize + Extensions 白名单）
- `UploadResponse`: ID, Type, OriginalName, URL, Size, Status, CreatedAt
- `FileUploadedEvent`: FileID, Type, URL, OwnerID

| 类型 | 最大 | 扩展名 |
|---|---|---|
| image | 10MB | .jpg .jpeg .png .gif .webp .svg |
| document | 50MB | .pdf .doc .docx .xls .xlsx .txt |
| video | 500MB | .mp4 .mov .avi .mkv |

**upload/handle_upload.go**
1. 校验 type 参数是否合法
2. 校验扩展名是否在对应白名单中
3. 校验文件大小是否超限
4. 生成 UUID 文件名 → Storage.Upload
5. FileRepo.Create 保存元数据
6. EventBus.Publish 发布 `file.uploaded` 事件（失败只打日志）

**fileprocessing/handler.go**
- 订阅 `file.uploaded`
- 若 type==image → 打日志"此处未来会调用 OCR 微服务" → 更新 status 为 processed
- 非 image → 无操作

**assistant/assistant.go**
- 实现 `AssistantPort`
- `GenerateProductDescription`: 50ms 延迟后返回固定文案
- 监听 `ctx.Done()` 模拟超时

### 任务 5：HTTP Handler 层

**handler/upload.go**
- `UploadUseCase` 接口（HandleUpload 方法）
- POST 处理：`c.FormFile("file")` + `c.PostForm("type")`
- 读取文件内容 → 调用 use case → 返回 `UploadResponse`
- 错误处理同 product 的 `handleProductError` 模式

### 任务 6：DI 容器 + 路由

**container.go** — 初始化顺序：EventBus → FileRepo → Storage → UploadUseCase → UploadHandler

**main.go**
```go
r.Protected("POST", "/api/v1/upload", c.UploadHandler.Upload)  // AnyRole
r.Engine().Static("/uploads", cfg.Storage.Local.BasePath)
```

### 任务 7：迁移 + 文档

**migrations/000003_create_files_table.up.sql** — files 表 + 2 个索引
**docs/api/upload-api.md** — 上传接口文档（请求/响应/错误码/curl）

## 验证

```bash
# 架构约束（必须为空）
grep -r "gorm.io\|go-redis\|aws-sdk" internal/domain internal/usecase

# 全量测试
go test ./... -count=1

# 编译检查
go build ./... && go vet ./...

# 端到端
# 注册 → 登录 → 上传图片 → 验证事件链路
```
