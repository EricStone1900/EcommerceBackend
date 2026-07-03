# 商品管理 API 文档

基础 URL：`http://localhost:8080`

---

## 目录

1. [获取商品详情](#1-获取商品详情)
2. [获取商品列表](#2-获取商品列表)
3. [创建商品](#3-创建商品)
4. [更新商品](#4-更新商品)
5. [删除商品](#5-删除商品)

---

## 通用说明

### 响应格式

成功：
```json
{ "code": 0, "message": "success", "data": { ... } }
```

分页列表响应：
```json
{ "code": 0, "message": "success", "data": { "total": 100, "page": 1, "page_size": 20, "list": [...] } }
```

### 错误码（新增）

| code | message | HTTP 状态 | 说明 |
|---|---|---|---|
| `4004` | product not found | 404 | 商品不存在或已删除 |

其他通用错误码见 [auth-api.md](auth-api.md)

### 鉴权方式

受保护接口需要在请求头中携带 JWT access token：
```
Authorization: Bearer <access_token>
```

### 权限说明

| 端点 | 访问级别 | 说明 |
|---|---|---|
| `GET /api/v1/products/:id` | 公开 | 未登录用户也可访问 |
| `GET /api/v1/products` | 受保护（AnyRole） | 登录即可，任何角色 |
| `POST /api/v1/products` | 受保护（admin） | 仅管理员 |
| `PUT /api/v1/products/:id` | 受保护（admin） | 仅管理员 |
| `DELETE /api/v1/products/:id` | 受保护（admin） | 仅管理员 |

---

## 1. 获取商品详情

公开接口，无需认证。获取单个商品的详细信息。

### 请求

```
GET /api/v1/products/:id
```

| 参数 | 类型 | 说明 |
|---|---|---|
| `id` | path | 商品 ID |

### 响应示例

```json
// 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "商品名称",
    "description": "商品描述",
    "price": 99.99,
    "stock": 50,
    "status": "on_sale",
    "created_by": 1,
    "created_at": "2026-07-03T10:00:00+08:00",
    "updated_at": "2026-07-03T10:00:00+08:00"
  }
}
```

### 错误示例

```json
// 404 — 商品不存在或已删除
{ "code": 4004, "message": "product not found" }
```

### curl 示例

```bash
curl http://localhost:8080/api/v1/products/1
```

---

## 2. 获取商品列表

受保护接口，需要登录（任意角色）。支持分页、搜索、筛选和排序。

### 请求

```
GET /api/v1/products?page=1&page_size=20&name=关键字&status=on_sale&sort_by=created_at&sort_desc=true
Authorization: Bearer <access_token>
```

### 查询参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|---|---|---|---|---|
| `page` | int | 否 | 1 | 页码 |
| `page_size` | int | 否 | 20 | 每页数量（最大 100） |
| `name` | string | 否 | — | 商品名称模糊搜索（大小写不敏感） |
| `status` | string | 否 | — | 按状态筛选：`on_sale` / `off_sale` |
| `sort_by` | string | 否 | `created_at` | 排序字段：`created_at` / `price` |
| `sort_desc` | bool | 否 | `true` | 是否降序排序 |

### 响应示例

```json
// 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 1,
    "page": 1,
    "page_size": 20,
    "list": [
      {
        "id": 1,
        "name": "商品名称",
        "description": "商品描述",
        "price": 99.99,
        "stock": 50,
        "status": "on_sale",
        "created_by": 1,
        "created_at": "2026-07-03T10:00:00+08:00",
        "updated_at": "2026-07-03T10:00:00+08:00"
      }
    ]
  }
}
```

### curl 示例

```bash
# 默认查询（第一页，20条，按创建时间倒序）
curl http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer <token>"

# 按名称搜索，按价格升序排列
curl "http://localhost:8080/api/v1/products?name=手机&sort_by=price&sort_desc=false" \
  -H "Authorization: Bearer <token>"

# 筛选在售商品
curl "http://localhost:8080/api/v1/products?status=on_sale" \
  -H "Authorization: Bearer <token>"
```

---

## 3. 创建商品

受保护接口，仅 admin 角色可访问。

### 请求

```
POST /api/v1/products
Authorization: Bearer <admin_access_token>
Content-Type: application/json
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | string | 是 | 商品名称，不能为空 |
| `description` | string | 否 | 商品描述，默认为空 |
| `price` | number | 是 | 价格，必须大于 0 |
| `stock` | int | 否 | 库存数量，不能为负，默认 0 |
| `status` | string | 否 | 状态，`on_sale` 或 `off_sale`，默认 `on_sale` |

### 响应示例

```json
// 201 Created
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "商品名称",
    "description": "商品描述",
    "price": 99.99,
    "stock": 50,
    "status": "on_sale",
    "created_by": 1,
    "created_at": "2026-07-03T10:00:00+08:00",
    "updated_at": "2026-07-03T10:00:00+08:00"
  }
}
```

### 错误示例

```json
// 400 — 名称不能为空
{ "code": 1002, "message": "validation failed: name cannot be empty" }

// 400 — 价格必须大于 0
{ "code": 1002, "message": "validation failed: price must be greater than 0" }

// 400 — 非法状态值
{ "code": 1002, "message": "validation failed: status must be 'on_sale' or 'off_sale'" }

// 403 — 非 admin 角色
{ "code": 2003, "message": "forbidden" }
```

### curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"手机","description":"最新款智能手机","price":2999.00,"stock":100}'
```

---

## 4. 更新商品

受保护接口，仅 admin 角色可访问。全字段更新。

### 请求

```
PUT /api/v1/products/:id
Authorization: Bearer <admin_access_token>
Content-Type: application/json
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | string | 是 | 商品名称 |
| `description` | string | 否 | 商品描述 |
| `price` | number | 是 | 价格 |
| `stock` | int | 否 | 库存数量 |
| `status` | string | 否 | 状态：`on_sale` / `off_sale` |

### 响应示例

```json
// 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "更新后的名称",
    "description": "更新后的描述",
    "price": 1999.00,
    "stock": 80,
    "status": "off_sale",
    "created_by": 1,
    "created_at": "2026-07-03T10:00:00+08:00",
    "updated_at": "2026-07-03T11:00:00+08:00"
  }
}
```

### 错误示例

```json
// 404 — 商品不存在
{ "code": 4004, "message": "product not found" }
```

### curl 示例

```bash
curl -X PUT http://localhost:8080/api/v1/products/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"更新名称","description":"更新描述","price":1999.00,"stock":80,"status":"off_sale"}'
```

---

## 5. 删除商品

受保护接口，仅 admin 角色可访问。使用软删除（设置 `deleted_at` 字段），不会物理删除数据。

### 请求

```
DELETE /api/v1/products/:id
Authorization: Bearer <admin_access_token>
```

### 响应示例

```json
// 200 OK
{ "code": 0, "message": "success" }
```

### 错误示例

```json
// 404 — 商品不存在或已被删除
{ "code": 4004, "message": "product not found" }
```

### curl 示例

```bash
curl -X DELETE http://localhost:8080/api/v1/products/1 \
  -H "Authorization: Bearer <admin_token>"
```

---

## 商品状态说明

| 状态 | 含义 |
|---|---|
| `on_sale` | 在售 |
| `off_sale` | 下架 |

- 创建商品时，`status` 默认为 `on_sale`
- 下架商品仍可在列表中查询到，但详情页可据此状态做相应展示
- 删除（软删除）后，通过 API 查询返回 404
