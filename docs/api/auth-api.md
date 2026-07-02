# 用户认证 API 文档

基础 URL：`http://localhost:8080`

---

## 目录

1. [注册](#1-注册)
2. [登录](#2-登录)
3. [登出](#3-登出)
4. [刷新令牌](#4-刷新令牌)
5. [获取当前用户](#5-获取当前用户)

---

## 通用说明

### 响应格式

所有接口统一返回 JSON 格式：

```json
// 成功
{
  "code": 0,
  "message": "success",
  "data": { ... }
}

// 失败
{
  "code": 40001,
  "message": "错误描述",
  "data": null
}
```

### 错误码说明

| code | message | HTTP 状态 | 说明 |
|---|---|---|---|
| `0` | success | 200/201 | 请求成功 |
| `1001` | invalid request | 400 | 请求参数格式错误 |
| `1002` | validation failed | 400 | 参数校验失败 |
| `2001` | unauthorized | 401 | 未提供 token 或 token 无效 |
| `2002` | token expired | 401 | token 已过期 |
| `2003` | forbidden | 403 | 权限不足 |
| `3001` | user not found | 404 | 用户不存在 |
| `3002` | email already registered | 409 | 邮箱已被注册 |
| `3003` | invalid email or password | 401 | 邮箱或密码错误 |
| `5001` | internal server error | 500 | 服务器内部错误 |

### 鉴权方式

受保护接口需要在请求头中携带 JWT access token：

```
Authorization: Bearer <access_token>
```

### 角色说明

| 角色 | 说明 |
|---|---|
| `customer` | 普通用户（注册时默认角色） |
| `member` | 会员 |
| `admin` | 管理员 |

---

## 1. 注册

创建新用户账户并返回令牌。

### 请求

```
POST /api/v1/auth/register
Content-Type: application/json
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `email` | string | 是 | 邮箱地址，需符合邮箱格式，最长 255 字符 |
| `password` | string | 是 | 密码，8-128 字符，须包含至少一个字母和一个数字 |

### 响应示例

```json
// 201 Created
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "role": "customer"
    }
  }
}
```

### 错误示例

```json
// 400 — 邮箱已注册
{
  "code": 3002,
  "message": "email already registered"
}

// 400 — 参数校验失败
{
  "code": 1002,
  "message": "validation failed: email invalid email format"
}

// 400 — 密码太弱
{
  "code": 1002,
  "message": "validation failed: password must be at least 8 characters"
}
```

### curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Password1"}'
```

---

## 2. 登录

使用邮箱和密码登录，返回令牌。

### 请求

```
POST /api/v1/auth/login
Content-Type: application/json
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `email` | string | 是 | 注册邮箱 |
| `password` | string | 是 | 密码 |

### 响应示例

```json
// 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "role": "customer"
    }
  }
}
```

### 错误示例

```json
// 401 — 邮箱或密码错误（不区分是邮箱不存在还是密码错误，防止邮箱枚举）
{
  "code": 3003,
  "message": "invalid email or password"
}
```

### curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Password1"}'
```

---

## 3. 登出

吊销指定的 refresh token（使其无法再用于刷新令牌）。

### 请求

```
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
Content-Type: application/json
```

### 请求头

| 参数 | 说明 |
|---|---|
| `Authorization` | Bearer access token（上一步返回的 `access_token`） |

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `refresh_token` | string | 是 | 要吊销的 refresh token（上一步返回的 `refresh_token`） |

### 响应示例

```json
// 200 OK
{
  "code": 0,
  "message": "success"
}
```

> **幂等性说明**：对已吊销或已过期的 token 再次调用登出仍然返回成功。

### 错误示例

```json
// 401 — 未提供 token 或 token 无效
{
  "code": 2001,
  "message": "unauthorized"
}

// 403 — 尝试吊销不属于当前用户的 token
{
  "code": 2003,
  "message": "forbidden"
}
```

### curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

---

## 4. 刷新令牌

使用 refresh token 换取新的 access token 和 refresh token（轮换机制）。

### 请求

```
POST /api/v1/auth/refresh
Authorization: Bearer <access_token>
Content-Type: application/json
```

### 请求头

| 参数 | 说明 |
|---|---|
| `Authorization` | Bearer access token |

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `refresh_token` | string | 是 | 有效的 refresh token |

### 响应示例

```json
// 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...（新）",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...（新）",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "role": "customer"
    }
  }
}
```

> **重要**：每次刷新都会生成全新的令牌对，旧的 refresh token 会被吊销（轮换机制）。如果使用了已吊销的 refresh token，则返回 401。

### 错误示例

```json
// 401 — token 已过期
{
  "code": 2002,
  "message": "token expired"
}

// 401 — token 已被吊销
{
  "code": 2001,
  "message": "unauthorized"
}

// 400 — 使用了 access token 而非 refresh token
{
  "code": 1001,
  "message": "invalid request"
}
```

### curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

---

## 5. 获取当前用户

返回当前登录用户的个人信息（用于验证鉴权是否生效）。

### 请求

```
GET /api/v1/auth/me
Authorization: Bearer <access_token>
```

### 请求头

| 参数 | 说明 |
|---|---|
| `Authorization` | Bearer access token |

### 响应示例

```json
// 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "role": "customer"
  }
}
```

### 错误示例

```json
// 401 — 未提供 token
{
  "code": 2001,
  "message": "unauthorized"
}

// 401 — token 已过期
{
  "code": 2002,
  "message": "token expired"
}
```

### curl 示例

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

---

## 令牌说明

| 类型 | 有效期 | 存储 | 说明 |
|---|---|---|---|
| access token | 15 分钟 | 无状态（JWT 自带） | 用于 API 鉴权，不存储服务端 |
| refresh token | 7 天 | Redis (`refresh_token:{userID}:{tokenID}`) | 用于刷新 access token，支持主动吊销 |

### JWT Payload 结构

```json
{
  "user_id": 1,
  "role": "customer",
  "type": "access",
  "sub": "1",
  "iss": "ecommerce-backend",
  "iat": 1700000000,
  "exp": 1700000900
}
```

| 字段 | 说明 |
|---|---|
| `user_id` | 用户 ID |
| `role` | 用户角色（admin / member / customer） |
| `type` | 令牌类型（`access` 或 `refresh`） |
| `sub` | 主题，值为用户 ID 的字符串形式 |
| `iss` | 签发者，固定值 `ecommerce-backend` |
| `jti` | 仅 refresh token 有，UUID v4，用于吊销 |
