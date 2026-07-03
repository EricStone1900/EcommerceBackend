# Kubernetes Deployment Guide

## Prerequisites

- A Kubernetes cluster (EKS, AKS, GKE, or self-managed)
- `kubectl` configured with cluster access
- Docker image pushed to a container registry (ECR, Docker Hub, etc.)

## Deploy Steps

1. **Update ConfigMap** with your infrastructure endpoints:
```bash
# Edit configmap.yaml and replace placeholder values
vim deployments/k8s/configmap.yaml
```

2. **Update Secret** with base64-encoded credentials:
```bash
# Encode your secrets
echo -n "your_db_password" | base64
echo -n "your_jwt_secret" | base64
echo -n "your_s3_key" | base64
echo -n "your_s3_secret" | base64

# Paste the encoded values into deployments/k8s/secret.yaml
```

3. **Update Deployment image**:
```bash
# Edit deployment.yaml to set your container image
# Line: image: ${ECR_REPO}:${IMAGE_TAG:-latest}
# Replace with: image: your-registry/ecommerce-app:latest
```

4. **Apply manifests**:
```bash
kubectl apply -f deployments/k8s/configmap.yaml
kubectl apply -f deployments/k8s/secret.yaml
kubectl apply -f deployments/k8s/deployment.yaml
kubectl apply -f deployments/k8s/service.yaml
```

5. **Verify deployment**:
```bash
# Check pods
kubectl get pods -l app=ecommerce-app

# Check logs
kubectl logs -l app=ecommerce-app --tail=50

# Check health
kubectl port-forward svc/ecommerce-app 8080:8080
curl http://localhost:8080/health
```

## Scaling

```bash
# Scale to 5 replicas
kubectl scale deployment ecommerce-app --replicas=5

# Auto-scaling based on CPU
kubectl autoscale deployment ecommerce-app --min=3 --max=10 --cpu-percent=70
```

## Rolling Update

```bash
# Update image version
kubectl set image deployment/ecommerce-app app=your-registry/ecommerce-app:v2.0.0

# Check rollback status
kubectl rollout status deployment/ecommerce-app

# Rollback if needed
kubectl rollout undo deployment/ecommerce-app
```

## Infrastructure Requirements

| Resource | AWS | GCP | Azure |
|---|---|---|---|
| PostgreSQL | RDS | Cloud SQL | Azure Database for PostgreSQL |
| Redis | ElastiCache | Memorystore | Azure Cache for Redis |
| Object Storage | S3 | Cloud Storage | Blob Storage |
| Container Registry | ECR | Artifact Registry | ACR |
| Orchestration | EKS | GKE | AKS |
