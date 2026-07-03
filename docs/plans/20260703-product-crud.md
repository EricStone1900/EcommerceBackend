# Phase 3 开发计划 — 商品管理 API

## 背景

Phase 1（项目骨架）和 Phase 2（用户认证与 RBAC）已完成。当前系统已有：分层架构、JWT 鉴权、`Public`/`Protected` 路由机制、RBAC 中间件、统一响应/错误格式。本阶段实现商品 CRUD 功能，遵循域层/usecase 层零 GORM 引用的架构约束。

## 权限设计（按 spec 要求）

| 端点 | 访问级别 | 说明 |
|---|---|---|
| `GET /api/v1/products/:id` | 公开 | 未登录用户也能访问 |
| `GET /api/v1/products` | 受保护（AnyRole） | 登录即可，任何角色 |
| `POST /api/v1/products` | 受保护（admin 仅） | 创建商品 |
| `PUT /api/v1/products/:id` | 受保护（admin 仅） | 编辑商品 |
| `DELETE /api/v1/products/:id` | 受保护（admin 仅） | 删除商品（软删除） |

## 技术决策

| 决策点 | 选择 | 理由 |
|---|---|---|
| 分页实现 | GORM Scopes + Count 后 Offset/Limit | 与已有 `PaginatedData` 结构配合 |
| 排序白名单 | 只允许 `created_at` / `price` | 防止 SQL 注入式排序 |
| 模糊搜索 | `LOWER(name) LIKE ?` | PostgreSQL 兼容，大小写不敏感 |
| 价格字段 | `DECIMAL(10,2)` DB 类型，Go `float64` | 简单电商场景够用 |
| 库存校验 | `stock >= 0` | 不允许负数库存 |
| 删除策略 | 软删除（`deleted_at`） | 遵循架构铁律，不在 entity 中暴露 |
| 新错误码 | `ErrProductNotFound` (4004) | 与已有错误体系兼容 |

## 需要创建/修改的文件汇总

| 文件 | 操作 | 说明 |
|---|---|---|
| `pkg/errors/errors.go` | **修改** | 新增 `ErrProductNotFound` (4004) 错误码 |
| `internal/domain/entity/product.go` | 创建 | Product 实体 + ProductStatus 枚举 |
| `internal/domain/port/product_repository.go` | 创建 | Repository 接口 + Filter/Result 结构体 |
| `internal/infrastructure/persistence/gorm/product_repo.go` | 创建 | GORM 实现（model 含 DeletedAt） |
| `internal/usecase/product/product.go` | 创建 | 共享 DTO + UseCase 结构体 + 构造函数 |
| `internal/usecase/product/create.go` | 创建 | 创建商品用例 |
| `internal/usecase/product/update.go` | 创建 | 更新商品用例 |
| `internal/usecase/product/delete.go` | 创建 | 软删除商品用例 |
| `internal/usecase/product/get.go` | 创建 | 获取单个商品用例 |
| `internal/usecase/product/list.go` | 创建 | 分页列表用例 |
| `internal/usecase/product/helpers_test.go` | 创建 | Mock 仓库 + 测试辅助函数 |
| `internal/usecase/product/create_test.go` | 创建 | 5+ 测试用例 |
| `internal/usecase/product/update_test.go` | 创建 | 3+ 测试用例 |
| `internal/usecase/product/delete_test.go` | 创建 | 2+ 测试用例 |
| `internal/usecase/product/get_test.go` | 创建 | 3+ 测试用例 |
| `internal/usecase/product/list_test.go` | 创建 | 4+ 测试用例 |
| `internal/interface/http/handler/product.go` | 创建 | HTTP handler + UseCase 接口 |
| `internal/interface/http/handler/product_test.go` | 创建 | Handler 测试（5+ 测试用例） |
| `internal/container/container.go` | **修改** | 新增 ProductRepo/UseCase/Handler 字段和初始化 |
| `cmd/server/main.go` | **修改** | 注册 5 条商品路由 |
| `migrations/000002_create_products_table.up.sql` | 创建 | 建表脚本 |
| `migrations/000002_create_products_table.down.sql` | 创建 | 删表脚本 |
| `docs/plans/20260703-product-crud.md` | 创建 | 本计划文档副本 |
| `docs/api/product-api.md` | 创建 | 商品 API 文档 |

---

## 实施顺序（严格按此顺序，前置依赖先完成）

### 任务 1：新增错误码 + 迁移脚本

**修改 `pkg/errors/errors.go`**：添加 `ErrProductNotFound` 的 code/message/http 映射

**新建迁移脚本**：`migrations/000002_create_products_table.up.sql` + `.down.sql`

### 任务 2：Product 实体 + Repository 接口

**`internal/domain/entity/product.go`**：
```go
type ProductStatus string
const ( ProductStatusOnSale ProductStatus = "on_sale"; ProductStatusOffSale ProductStatus = "off_sale" )
type Product struct { ID, Name, Description, Price, Stock, Status, CreatedBy, CreatedAt, UpdatedAt }
```

**`internal/domain/port/product_repository.go`**：
```go
type ProductRepository interface { Create/Update/Delete/GetByID/List }
type ProductFilter struct { Page, PageSize int; Name string; Status *ProductStatus; SortBy string; SortDesc bool }
type ProductListResult struct { Products []*entity.Product; Total int64 }
```

### 任务 3：GORM ProductRepository 实现

遵循 `user_repo.go` 的 model/entity 分离模式：
- 私有 `productModel` 含 `gorm.DeletedAt`
- `toEntity()` / `toModel()` 转换方法
- `List` 动态构建 WHERE/ORDER BY/OFFSET/LIMIT
- 软删除使用 GORM 的 `Delete`（自动设置 `deleted_at`）

### 任务 4：Use Case 层（含测试）

`internal/usecase/product/` 目录，遵循 auth 包的一文件一操作模式：
- `product.go` — DTO 定义、`productToResponse` 转换、`ProductUseCase` 结构体（依赖 `port.ProductRepository` 接口，零 GORM 引用）
- 5 个操作文件 + 5 个测试文件

### 任务 5：HTTP Handler 层（含测试）

**`internal/interface/http/handler/product.go`**：
- `ProductUseCase` 接口（5 方法）+ `ProductHandler` 结构体
- 5 个 handler 方法对应 5 条路由
- 复用 `handleProductError` 模式（type-assert `*errors.Error`）

**Handler 测试**：mock usecase + gin.TestMode + httptest

### 任务 6：DI 容器与路由注册

**`internal/container/container.go`**：新增 3 个字段，按顺序初始化

**`cmd/server/main.go`**：注册 5 条路由，使用 `entity.RoleAdmin` 作为角色参数

### 任务 7：API 文档

`docs/api/product-api.md`：5 个接口的完整文档（请求/响应/错误码/curl）

---

## 验证方式

### 编译器
```bash
go build ./...
go vet ./...
```

### 测试（全部 70+ 测试用例）
```bash
go test ./... -count=1
```

### 架构约束（输出必须为空）
```bash
grep -r "gorm.io\|go-redis\|aws-sdk" internal/domain internal/usecase
```

### 运行验证
```bash
go run cmd/server/main.go migrate
# 注册 → 登录获取 admin token → 创建商品 → 查询列表 → 查看详情 → 更新 → 删除
```
