# 阶段 3 / 7 — 商品管理 API

## 上文衔接

阶段 1（项目骨架）和阶段 2（用户认证与 RBAC）已完成。当前系统已经有：分层架构、用户注册登录、JWT 鉴权、`Public`/`Protected` 路由机制、RBAC 中间件（可以声明接口允许哪些角色访问）。本阶段在已有基础上实现商品管理功能，**继续在已有目录结构里开发**。

## 全局架构约束（必须遵守）

domain 层和 usecase 层禁止 import 任何具体技术实现。商品相关的数据库访问必须通过 Repository 接口，不能在 handler 或 usecase 里直接写 GORM 代码。

## 本阶段目标

实现商品的增删改查接口，区分不同角色的操作权限，商品列表/详情对应不同的访问级别（详情公开可查看，增删改仅 admin）。

## 具体任务清单

1. `domain/entity` 下定义 `Product` 实体：id、name、description、price、stock、status（如 `on_sale`/`off_sale`）、created_by（关联用户）、created_at、updated_at、deleted_at
2. `domain/port` 下定义 `ProductRepository` 接口（Create、Update、Delete、GetByID、List 带分页筛选排序参数等）
3. `infrastructure/persistence/gorm` 下实现 `ProductRepository`，补充对应数据库迁移脚本
4. `usecase/product` 下实现业务用例：创建、编辑、删除（软删除）、查询单个、分页列表（支持按名称模糊搜索、按状态筛选、按创建时间或价格排序）
5. 路由设计：
   ```go
   router.Public("GET", "/api/v1/products/:id", productHandler.GetDetail)
   router.Protected("GET", "/api/v1/products", productHandler.List, AnyRole)
   router.Protected("POST", "/api/v1/products", productHandler.Create, "admin")
   router.Protected("PUT", "/api/v1/products/:id", productHandler.Update, "admin")
   router.Protected("DELETE", "/api/v1/products/:id", productHandler.Delete, "admin")
   ```
   商品详情公开可访问（未登录用户也能看），列表要求登录但任何角色都可以看，增删改只允许 admin。如果你认为这个权限划分不够合理（比如希望 member 也能看到一些 customer 看不到的字段），可以提出更合理的方案并说明理由。
6. 实现统一的分页响应格式（包含总数、当前页、每页大小、数据列表），后续所有列表类接口都应该复用这个格式
7. 实现统一的错误响应格式（如果阶段 2 已经定义过，本阶段复用，不要重新发明一套）
8. 删除商品要做软删除（用 `deleted_at` 字段），不要物理删除数据

