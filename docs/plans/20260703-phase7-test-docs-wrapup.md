# Phase 7 开发计划 — 测试、压测与文档收尾

## 背景

Phase 1-6 已完成全部系统核心功能：骨架搭建、用户认证与 RBAC、商品 CRUD、文件上传与事件机制、推送服务、S3 云存储与部署文件。本阶段是最终收尾阶段，目标是补齐测试覆盖率、添加集成测试、提供压测脚本、生成 Swagger API 文档、完善最终 README 以及清理遗留调试代码。**本阶段不新增业务功能。**

## 当前状态概览

| 维度 | 状态 |
|---|---|
| 单元测试 | 30 个测试文件，~3,253 行测试代码，usecase 层基本覆盖但 `assistant/` 无测试 |
| 集成测试 | 不存在 |
| 压测脚本 | 无（`scripts/` 目录为空） |
| Swagger 文档 | 不存在（无 swaggo 依赖，handler 无注解） |
| API 文档 | `docs/api/` 有手写 Markdown（auth, product, upload, push） |
| README | 存在但缺失架构图、压测说明、微服务集成指引 |
| 代码清理 | 未执行 |

## 任务清单

### 任务 1：补充 usecase 层单元测试

**现状分析：**
- `auth/` — 16 个测试，覆盖全面 ✅
- `product/` — 20 个测试，覆盖全面 ✅
- `push/` — 15 个测试，覆盖全面 ✅
- `upload/` — 11 个测试，覆盖全面 ✅
- `fileprocessing/` — 4 个测试，覆盖全面 ✅
- `assistant/` — **0 个测试，需要新增** ❌

**`assistant/assistant.go` 需补充测试：**
- `TestGenerateDescription_Success` — 正常生成描述
- `TestGenerateDescription_ContextTimeout` — context 超时场景
- `TestGenerateDescription_Cancellation` — context 取消场景

遵循现有模式：在 `internal/usecase/assistant/` 下创建 `helpers_test.go`（mock 定义）和 `generate_description_test.go`（测试函数）。

### 任务 2：HTTP 集成测试

**新建 `tests/integration/` 目录，包含：**

1. **`tests/integration/suite_test.go`** — 测试套件
   - 使用 GORM SQLite 内存数据库（纯 Go、无外部依赖）
   - 自动运行 migration 建表
   - 在 `TestMain` 中完成初始化

2. **`tests/integration/auth_test.go`** — 核心链路测试
   - 注册 → 登录 → 获取受保护接口（`/api/v1/auth/me`）
   - 注册重复邮箱 → 409
   - 错误密码登录 → 401
   - 未携带 token 访问受保护接口 → 401
   - 无效 token 访问 → 401

3. **`tests/integration/product_test.go`** — 商品链路测试
   - admin 登录 → 创建商品 → 查询详情 → 列表查询 → 删除

### 任务 3：压测脚本

**新建 `scripts/loadtest.js`（k6 脚本）：**

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 50,
  duration: '1m',
  thresholds: {
    http_req_duration: ['p(95)<2000', 'p(99)<5000'],
    http_req_failed: ['rate<0.01'],
  },
};
```

覆盖路径：
1. `POST /api/v1/auth/login` — 登录（先注册一个固定用户）
2. `GET /api/v1/products/:id` — 商品详情（公开接口，无需认证）
3. `GET /api/v1/products` — 商品列表（已登录，带分页参数）

输出：P50/P95/P99 延迟、错误率、吞吐量。

### 任务 4：本地压测执行与记录

- 启动 docker-compose（PostgreSQL + Redis）
- 运行 migration + seed 数据（商品数据）
- 执行 k6 压测
- 将结果记录到 `docs/benchmark/results.md`
- 如发现性能瓶颈（连接池耗尽、缺索引），在本阶段修复

### 任务 5：Swagger/OpenAPI 文档

**步骤：**

1. 添加 swaggo 依赖：
```bash
go get github.com/swaggo/swag
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
go install github.com/swaggo/swag/cmd/swag@latest
```

2. 在 `cmd/server/main.go` 添加 `// @title` 等包级 swagger 注解

3. 为所有 handler 方法添加 swagger 注解（共 5 个文件，15 个端点）：

| 文件 | 端点数 | 路径范围 |
|---|---|---|
| `auth.go` | 5 | `/api/v1/auth/*` |
| `product.go` | 5 | `/api/v1/products/*` |
| `upload.go` | 1 | `/api/v1/upload` |
| `push.go` | 3 | `/api/v1/push/*` |
| `health.go` | 1 | `/health` |

4. 在 `main.go` 路由注册中添加：
```go
import ginSwagger "github.com/swaggo/gin-swagger"
import swaggerFiles "github.com/swaggo/files"

// 开发环境挂载 Swagger UI
if c.Config.Server.Env == "development" {
    r.Engine().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
```

5. 生成文档：`swag init -g cmd/server/main.go -o docs/`
6. 将生成的 `docs/swagger.json`、`docs/swagger.yaml`、`docs/docs.go` 纳入版本管理

### 任务 6：最终版 README

**`README.md`** 补充以下内容：

1. **项目简介与架构图（文字版）** — 层间依赖图 + 数据流方向
2. **本地启动步骤** — 从零开始（假设只有 Docker）
3. **环境变量完整表** — 含 S3、Push 等新增变量
4. **如何运行测试** — 单元测试 + 集成测试 + 覆盖率
5. **如何运行压测** — k6 命令
6. **如何切换本地/云存储** — 配置 + 环境变量两种方式
7. **部署指引** — 引用 `docs/deploy/aws-deployment-guide.md`
8. **微服务接入指引** — 三个预留点：OCR、AI Assistant、APNs
   - 指明要修改的文件、实现的接口、不需要改的接口

### 任务 7：代码清理

- 搜索 `// TODO` — 判断是否保留，保留的写清除计划
- 搜索注释掉的废弃代码块 — 删除
- 搜索 `fmt.Println` / `log.Println` 调试输出 — 替换为结构化日志或删除
- 搜索死代码（没有被引用的函数/类型）— 删除

## 文件清单

### 新建文件

| 文件路径 | 说明 |
|---|---|
| `internal/usecase/assistant/generate_description_test.go` | Assistant 用例测试 |
| `internal/usecase/assistant/helpers_test.go` | Assistant 测试内 mock 定义 |
| `tests/integration/suite_test.go` | 集成测试套件 |
| `tests/integration/auth_test.go` | 认证集成测试 |
| `tests/integration/product_test.go` | 商品集成测试 |
| `scripts/loadtest.js` | k6 压测脚本 |
| `docs/benchmark/results.md` | 压测结果记录（执行后填写） |

### 修改文件

| 文件路径 | 修改内容 |
|---|---|
| `cmd/server/main.go` | 添加 swagger 包级注解 + `/swagger/*any` 路由 |
| `internal/interface/http/handler/auth.go` | 5 个 handler 方法添加 swagger 注解 |
| `internal/interface/http/handler/product.go` | 5 个 handler 方法添加 swagger 注解 |
| `internal/interface/http/handler/upload.go` | 1 个 handler 方法添加 swagger 注解 |
| `internal/interface/http/handler/push.go` | 3 个 handler 方法添加 swagger 注解 |
| `internal/interface/http/handler/health.go` | 1 个 handler 方法添加 swagger 注解 |
| `go.mod` / `go.sum` | 新增 swaggo 依赖 |
| `README.md` | 完整重构——架构图、压测说明、微服务接入指引 |
| `docs/swagger.json` | swag init 生成 |
| `docs/swagger.yaml` | swag init 生成 |
| `docs/docs.go` | swag init 生成 |

## 实施顺序

```
任务 1 (单元测试) ──→ 任务 2 (集成测试) ──→ 任务 5 (Swagger)
                                     ↓
               任务 3 (压测脚本) ──→ 任务 4 (压测执行) ──→ 任务 6 (README)
                                                           ↓
                                                     任务 7 (代码清理)
```

1. 先补充测试（单元 + 集成），确保代码正确性
2. 再实现 Swagger，文档化 API
3. 写压测脚本并执行，验证性能
4. 最后整理 README 和清理代码

## 验证

```bash
# 1. 编译
go build ./... && go vet ./... || echo "FAIL"

# 2. 架构约束
grep -r "gorm.io\|go-redis\|aws-sdk" internal/domain internal/usecase || echo "PASS"

# 3. 全量测试 + 覆盖率
go test ./... -count=1 -cover

# 4. Swagger 生成
swag init -g cmd/server/main.go -o docs/
ls docs/swagger.json docs/swagger.yaml || echo "FAIL"

# 5. Docker 构建
docker build -f deployments/Dockerfile -t ecommerce-app:latest .

# 6. 压测（依赖 docker-compose 基础设施就绪）
k6 run --vus 50 --duration 1m scripts/loadtest.js

# 7. 代码清理检查
grep -rn "TODO\|FIXME\|fmt.Println\|log.Println" . --include="*.go" | grep -v "_test.go" | grep -v vendor/
```

## 依赖安装

```bash
# swaggo
go get github.com/swaggo/swag
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
go install github.com/swaggo/swag/cmd/swag@latest

# SQLite 驱动（集成测试用）
go get gorm.io/driver/sqlite

go mod tidy
```
