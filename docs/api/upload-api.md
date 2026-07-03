# 文件上传 API 文档

基础 URL：`http://localhost:8080`

---

## 目录

1. [上传文件](#1-上传文件)

---

## 通用说明

### 响应格式

成功：
```json
{ "code": 0, "message": "success", "data": { ... } }
```

### 错误码（新增）

| code | message | HTTP 状态 | 说明 |
|---|---|---|---|
| `6001` | file size exceeds limit | 413 | 文件大小超过限制 |
| `6002` | invalid file type or extension | 400 | 不支持的文件类型或扩展名 |
| `6003` | file not found | 404 | 文件不存在或已删除 |
| `6004` | file upload failed | 500 | 文件上传存储失败 |

其他通用错误码见 [auth-api.md](auth-api.md)

### 鉴权方式

受保护接口需要在请求头中携带 JWT access token：
```
Authorization: Bearer <access_token>
```

### 权限说明

| 端点 | 访问级别 | 说明 |
|---|---|---|
| `POST /api/v1/upload` | 受保护（AnyRole） | 登录即可，任何角色 |

---

## 1. 上传文件

受保护接口，需要登录（任意角色）。使用 `multipart/form-data` 上传。

### 请求

```
POST /api/v1/upload
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `file` | file | 是 | 要上传的文件 |
| `type` | string | 是 | 文件类型：`image` / `document` / `video` |

### 文件类型限制

| 类型 | 最大大小 | 允许扩展名 |
|---|---|---|
| image | 10MB | .jpg .jpeg .png .gif .webp .svg |
| document | 50MB | .pdf .doc .docx .xls .xlsx .txt .md |
| video | 500MB | .mp4 .mov .avi .mkv |

### 响应示例

```json
// 201 Created
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "type": "image",
    "original_name": "photo.jpg",
    "url": "/uploads/a1b2c3d4.jpg",
    "size": 102400,
    "status": "pending",
    "created_at": "2026-07-03T10:00:00+08:00"
  }
}
```

### 错误示例

```json
// 400 — 缺少文件
{ "code": 1001, "message": "invalid request" }

// 400 — 文件类型无效
{ "code": 6002, "message": "invalid file type or extension" }

// 400 — 缺少 type 参数
{ "code": 6002, "message": "invalid file type or extension" }

// 413 — 文件过大（image > 10MB）
{ "code": 6001, "message": "file size exceeds limit" }

// 500 — 文件上传存储失败
{ "code": 6004, "message": "file upload failed" }
```

### curl 示例

```bash
# 上传图片
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer <token>" \
  -F "file=@photo.jpg" \
  -F "type=image"

# 上传文档
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer <token>" \
  -F "file=@document.pdf" \
  -F "type=document"

# 上传视频
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer <token>" \
  -F "file=@video.mp4" \
  -F "type=video"
```

### 事件机制

上传成功后，系统会自动发布 `file.uploaded` 事件。当前对 image 类型文件，事件订阅者会模拟 OCR 处理（更新文件状态为 `processed`），为未来接入真实 OCR 微服务预留扩展点。

---

## 文件状态说明

| 状态 | 含义 |
|---|---|
| `pending` | 已上传，待处理 |
| `processed` | 已处理完成 |

- 上传成功后初始状态为 `pending`
- 对于 image 类型文件，事件处理完成后状态更新为 `processed`
- 非 image 文件不上传后状态仍保持 `pending`
