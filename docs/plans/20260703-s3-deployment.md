# Phase 6 开发计划 — 云存储实现与部署文件

## 背景

Phase 1-5 已完成：分层骨架、用户认证、商品管理、文件上传（本地存储）、推送服务。本阶段将本地存储补全为 S3 兼容的云存储实现（同时适配 AWS S3 和阿里云 OSS），并完善部署文件（Dockerfile、K8s manifest、README）。

## 架构约束

- **不修改 `domain/port/storage.go` 接口签名** — 只新增一个 S3 实现
- **不修改 usecase 层代码** — 接口设计已验证
- 通过配置 `storage.driver` 切换 local/s3，不改代码

## 文件清单

### 新建文件（7 个）

| 层 | 文件 | 说明 |
|---|---|---|
| **Infra S3** | `internal/infrastructure/storage/s3/s3.go` | S3 兼容 Storage 实现（minio-go v7） |
| **K8s** | `deployments/k8s/deployment.yaml` | Deployment manifest |
| | `deployments/k8s/service.yaml` | Service manifest |
| | `deployments/k8s/configmap.yaml` | ConfigMap 模板 |
| | `deployments/k8s/secret.yaml` | Secret 模板 |
| | `deployments/k8s/README.md` | K8s 部署说明 |
| **文档** | `docs/deploy/aws-deployment-guide.md` | AWS 部署对照表 |

### 修改文件（3 个）

| 文件 | 修改内容 |
|---|---|
| `internal/container/container.go` | 替换 S3 fallthrough 为真实 S3 实现 |
| `internal/infrastructure/config/config.go` | 可选：添加 S3 配置校验 |
| `README.md` | 补充云存储切换说明、环境变量表、构建命令 |

## 实施顺序

### 任务 1：S3 Storage 实现

**`s3.go`** — 使用 `github.com/minio/minio-go/v7` 库实现 `port.Storage`
- 兼容 AWS S3 和阿里云 OSS（MinIO 库天然支持 S3 协议）
- Upload: PutObject → 返回可公开访问 URL
- Delete: RemoveObject
- URL 格式：`https://<bucket>.s3.<region>.amazonaws.com/<filename>`
- 配置通过 `S3Config` 已有字段注入（endpoint/key/secret/bucket/region）
- 实现后更新 `container.go` switch 分支，移除 fallthrough

```go
func (s *Storage) Upload(ctx context.Context, data []byte, filename string) (string, error) {
    _, err := s.client.PutObject(ctx, s.bucket, filename, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
        ContentType: detectContentType(filename),
    })
    if err != nil {
        return "", fmt.Errorf("s3 upload failed: %w", err)
    }
    return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, filename), nil
}
```

本地开发时 MinIO 已通过 docker-compose 运行在 `localhost:9000`，可以直接切换 `storage.driver=s3` 测试。

### 任务 2：Dockerfile 优化审查

当前 Dockerfile 已足够精简（多阶段构建、alpine、CGO_ENABLED=0），确认无需改动。

### 任务 3：Kubernetes 部署文件

**`deployments/k8s/deployment.yaml`**
- 3 副本，滚动更新
- 存活探针 + 就绪探针（`/health`）
- 资源请求/限制
- ConfigMap 挂载配置文件
- Secret 挂载敏感字段

**`deployments/k8s/service.yaml`** — ClusterIP Service，端口 8080

**`deployments/k8s/configmap.yaml`** — 非敏感配置（env、数据库地址）

**`deployments/k8s/secret.yaml`** — 敏感配置模板（JWT 密钥、数据库密码、S3 密钥）

**`deployments/k8s/README.md`** — K8s 部署步骤

### 任务 4：README + 部署文档

**`README.md`** — 补充：
- 「云存储驱动切换」一节：local ↔ s3 切换说明
- 完整环境变量表（含 S3、Push 等新增变量）
- Docker 构建命令更新

**`docs/deploy/aws-deployment-guide.md`**
- 基础设施清单（ECR、RDS、ElastiCache、S3、ECS）
- 配置对照表
- 环境变量设置说明

## 依赖安装

```bash
go get github.com/minio/minio-go/v7
go mod tidy
```

## 验证

```bash
# 编译检查
go build ./... && go vet ./...

# 架构约束（S3 只在 infrastructure 层出现）
grep -r "gorm.io\|aws-sdk" internal/domain internal/usecase

# 全量测试
go test ./... -count=1

# Docker 构建
docker build -f deployments/Dockerfile -t ecommerce-app:latest .
docker images | grep ecommerce-app

# MinIO 本地 S3 验证
# docker-compose 已包含 minio
# 修改 config.local.yaml: storage.driver=s3
# go run cmd/server/main.go serve → 上传 → 验证文件存入 MinIO
```
