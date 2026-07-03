# 推送服务 API 文档

基础 URL：`http://localhost:8080`

---

## 目录

1. [注册设备 Token](#1-注册设备-token)
2. [删除设备 Token](#2-删除设备-token)
3. [发送测试推送](#3-发送测试推送)

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
| `7001` | push token not found | 404 | 用户未注册设备 Token |
| `7002` | push send failed | 500 | 推送发送失败 |
| `7003` | invalid platform, only 'ios' is supported | 400 | 不支持的平台类型 |

其他通用错误码见 [auth-api.md](auth-api.md)

### 鉴权方式

受保护接口需要在请求头中携带 JWT access token：
```
Authorization: Bearer <access_token>
```

### 权限说明

| 端点 | 访问级别 | 说明 |
|---|---|---|
| `POST /api/v1/push/token` | 受保护（AnyRole） | 注册设备 token |
| `DELETE /api/v1/push/token` | 受保护（AnyRole） | 删除设备 token |
| `POST /api/v1/push/test` | 受保护（AnyRole） | 发送测试推送 |

---

## 1. 注册设备 Token

受保护接口，需要登录（任意角色）。注册 iOS 设备的推送 token，重复注册同一 token 会自动更新（幂等）。

### 请求

```
POST /api/v1/push/token
Authorization: Bearer <access_token>
Content-Type: application/json
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `device_token` | string | 是 | iOS 设备推送 token |
| `platform` | string | 是 | 平台类型，当前仅支持 `ios` |

### 响应示例

```json
// 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "platform": "ios",
    "created_at": "2026-07-03T10:00:00+08:00",
    "updated_at": "2026-07-03T10:00:00+08:00"
  }
}
```

### 错误示例

```json
// 400 — device_token 为空
{ "code": 1002, "message": "validation failed: device_token cannot be empty" }

// 400 — 不支持的平台
{ "code": 7003, "message": "invalid platform, only 'ios' is supported" }
```

### curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/push/token \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"device_token":"your-device-token-here","platform":"ios"}'
```

---

## 2. 删除设备 Token

受保护接口，需要登录（任意角色）。删除指定设备的推送 token（用于用户登出或卸载 App 时）。

### 请求

```
DELETE /api/v1/push/token
Authorization: Bearer <access_token>
Content-Type: application/json
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `device_token` | string | 是 | 要删除的设备推送 token |

### 响应示例

```json
// 200 OK
{ "code": 0, "message": "success" }
```

### curl 示例

```bash
curl -X DELETE http://localhost:8080/api/v1/push/token \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"device_token":"your-device-token-here"}'
```

---

## 3. 发送测试推送

受保护接口，需要登录（任意角色）。向当前用户的所有已注册设备发送一条测试推送通知。

### 请求

```
POST /api/v1/push/test
Authorization: Bearer <access_token>
```

### 响应示例

```json
// 200 OK — 成功发送到 1 个设备
{
  "code": 0,
  "message": "success",
  "data": {
    "sent": 1
  }
}
```

```json
// 200 OK — 成功发送到 2 个设备（用户有多个设备）
{
  "code": 0,
  "message": "success",
  "data": {
    "sent": 2
  }
}
```

### 错误示例

```json
// 404 — 用户未注册任何设备
{ "code": 7001, "message": "push token not found" }
```

### curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/push/test \
  -H "Authorization: Bearer <token>"
```

---

## 推送架构说明

当前推送系统使用 **Stub 实现**（仅打印日志），为未来接入 APNs 预留完整扩展点：

```
客户端注册 Token → PushTokenRepo (DB 持久化)
                 → 返回注册成功

发送测试推送 → PushTokenRepo.GetByUserID → Notifier.SendPush (Stub)
                 → 打日志（包含 user_id, device_token, title, body）
                 → 返回 sent 计数
```

### 未来接入 APNs

接入真实 APNs 时，只需：
1. 在 `internal/infrastructure/push/apns/` 下实现 `port.Notifier` 接口
2. 在 `container.go` 中将 `pushstub.NewNotifier` 替换为 APNs 实现
3. `port.Notifier` 接口不需要任何修改

详细接入指南见 `internal/infrastructure/push/apns/doc.go`
