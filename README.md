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
| 容器 | Docker + docker-compose | 本地开发 / 部署 |

## 目录结构

```
.
├── cmd/server/main.go         # 程序入口
├── internal/
│   ├── domain/
│   │   ├── entity/            # 纯业务实体
│   │   └── port/              # Repository / Storage 等接口定义
│   ├── usecase/               # 业务用例编排
│   ├── interface/http/        # Gin handler / middleware / router
│   ├── infrastructure/        # GORM / Redis / 文件存储 / 事件 / 推送实现
│   └── container/             # 依赖注入容器
├── pkg/                       # 跨模块通用工具
├── migrations/                # 数据库迁移
├── configs/                   # 配置文件
├── deployments/               # Dockerfile / docker-compose / k8s
├── scripts/                   # 辅助脚本
└── docs/                      # 文档
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

# 3. 启动基础设施
docker-compose -f deployments/docker-compose.yml up -d

# 4. 运行数据库迁移（占位，后续阶段实现）
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

### 环境变量

| 变量 | 说明 | 默认值 |
|---|---|---|
| `APP_ENV` | 运行环境 | `development` |
| `APP_SERVER_PORT` | 服务端口 | `8080` |
| `APP_DATABASE_HOST` | 数据库地址 | `localhost` |
| `APP_DATABASE_PORT` | 数据库端口 | `5432` |
| `APP_DATABASE_USER` | 数据库用户 | `productapp` |
| `APP_DATABASE_PASSWORD` | 数据库密码 | — |
| `APP_DATABASE_DBNAME` | 数据库名称 | `productsystem` |
| `APP_REDIS_URL` | Redis 连接串 | `redis://localhost:6379` |
| `APP_JWT_SECRET` | JWT 密钥 | — |
| `APP_CONFIG_PATH` | 配置文件路径 | `configs/config.local.yaml` |

环境变量优先级高于配置文件，适用于云上部署时覆盖配置。

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
```

## 结束人工测试
```bash
# 启动
docker-compose -f deployments/docker-compose.yml up -d

# 验证
curl http://localhost:8080/health

# 期望输出:
# {"status":"ok","database":"connected","redis":"connected","uptime":"2s"}

# 暂时关闭PostgreSQL 容器
docker-compose -f deployments/docker-compose.yml stop postgres
# 验证
curl http://localhost:8080/health

# 构建镜像
docker build -f deployments/Dockerfile -t ecommerce-app:latest .
# 查看镜像大小
docker images | grep ecommerce-app 
# 删除镜像
docker rmi ecommerce-app:latest

```