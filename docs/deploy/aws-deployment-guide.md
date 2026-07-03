# AWS 部署指南

本文档提供将本系统部署到 AWS 所需的基础设施清单和配置对照。

## 基础设施清单

| 服务 | AWS 产品 | 用途 | 规格建议 |
|---|---|---|---|
| 计算 | ECS Fargate / EKS | 运行应用容器 | 0.25-0.5 vCPU, 512MB 起 |
| 数据库 | RDS PostgreSQL | 主数据库 | db.t3.medium (2vCPU, 4GB) 起 |
| 缓存 | ElastiCache Redis | Token 黑名单 | cache.t3.micro 起 |
| 对象存储 | S3 | 文件存储 | Standard 即可 |
| 容器镜像 | ECR | 存储 Docker 镜像 | — |
| 负载均衡 | ALB | 流量分发 | — |
| 证书 | ACM | HTTPS 证书 | — |
| 域名 | Route53 | DNS 解析 | — |
| 监控 | CloudWatch | 日志/告警 | — |

## 配置对照表

### 应用配置（环境变量）

| 变量 | AWS 对应值 | 说明 |
|---|---|---|
| `APP_ENV` | `production` | 运行环境 |
| `APP_SERVER_PORT` | `8080` | 容器端口（ALB 目标组端口） |
| `APP_DATABASE_HOST` | RDS 端点（如 `xxx.ap-northeast-1.rds.amazonaws.com`） | 数据库地址 |
| `APP_DATABASE_PORT` | `5432` | PostgreSQL 端口 |
| `APP_DATABASE_USER` | `productapp` | 数据库用户 |
| `APP_DATABASE_PASSWORD` | Secrets Manager 或 SSM Parameter Store | 数据库密码 |
| `APP_DATABASE_DBNAME` | `productsystem` | 数据库名称 |
| `APP_DATABASE_SSLMODE` | `require` | RDS 强制 SSL |
| `APP_REDIS_URL` | ElastiCache 主端点（如 `redis://xxx.xxxxx.ng.0001.apne1.cache.amazonaws.com:6379`） | Redis 地址 |
| `APP_JWT_SECRET` | Secrets Manager 管理 | JWT 签名密钥（生产环境必须修改） |
| `APP_STORAGE_DRIVER` | `s3` | 生产环境使用 S3 |
| `APP_STORAGE_S3_ENDPOINT` | `https://s3.ap-northeast-1.amazonaws.com` | AWS S3 端点 |
| `APP_STORAGE_S3_BUCKET` | 创建的 S3 存储桶名称 | 文件存储桶 |
| `APP_STORAGE_S3_REGION` | `ap-northeast-1` | 对应 AWS 区域 |
| `APP_STORAGE_S3_ACCESS_KEY` | IAM 用户 Access Key | S3 访问密钥 |
| `APP_STORAGE_S3_SECRET_KEY` | IAM 用户 Secret Key | S3 秘密密钥 |

### IAM 权限

应用需要以下 S3 权限：

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:DeleteObject",
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::your-app-bucket",
                "arn:aws:s3:::your-app-bucket/*"
            ]
        }
    ]
}
```

## 部署步骤

### 方案一：ECS Fargate（推荐）

1. **创建 Docker 镜像并推送 ECR**
```bash
aws ecr create-repository --repository-name ecommerce-app
aws ecr get-login-password | docker login --username AWS --password-stdin <account>.dkr.ecr.<region>.amazonaws.com
docker build -f deployments/Dockerfile -t ecommerce-app .
docker tag ecommerce-app:latest <account>.dkr.ecr.<region>.amazonaws.com/ecommerce-app:latest
docker push <account>.dkr.ecr.<region>.amazonaws.com/ecommerce-app:latest
```

2. **创建 RDS PostgreSQL**
   - 引擎版本：PostgreSQL 16
   - 安全组：允许来自 ECS 任务的 5432 端口

3. **创建 ElastiCache Redis**
   - 引擎：Redis 7+
   - 集群模式：禁用（单节点即可）

4. **创建 S3 存储桶**
```bash
aws s3 mb s3://your-app-bucket
```

5. **创建 ECS 任务定义**（使用 `deployments/k8s/deployment.yaml` 中的 env 映射）

6. **创建 ALB + 目标组**（健康检查路径 `/health`）

7. **配置环境变量**（将 Secrets Manager 中的密码/密钥注入容器）

### 方案二：EKS（Kubernetes）

参见 `deployments/k8s/README.md`。

## 安全注意事项

1. **JWT 密钥**：生产环境必须使用强随机密钥，通过 Secrets Manager 管理
2. **数据库密码**：使用 RDS 自动生成的密码，或通过 Secrets Manager 轮转
3. **S3 访问密钥**：优先使用 IAM 角色（ECS/EKS 角色），次选 IAM 用户密钥
4. **SSL/TLS**：开启数据库 `sslmode=require`，前端通过 ALB + ACM 使用 HTTPS
5. **备份**：启用 RDS 自动备份 + S3 版本控制
6. **网络**：RDS/ElastiCache 部署在私有子网，仅 ECS 安全组可访问
