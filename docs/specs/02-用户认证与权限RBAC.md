# 阶段 2 / 7 — 用户认证与权限 RBAC

## 上文衔接

阶段 1 已完成：Clean Architecture 四层骨架（domain/usecase/interface/infrastructure）、配置加载、日志、数据库和 Redis 连接、依赖注入容器、`/health` 接口、docker-compose 本地环境都已跑通。本阶段在此基础上实现用户注册登录和权限体系，**请在阶段 1 已有的目录结构里继续开发，不要重新创建项目骨架**。

## 全局架构约束（必须遵守）

domain 层和 usecase 层禁止 import 任何具体技术实现。所有外部依赖通过接口注入。路由要明确区分 `Public`（公开接口）和 `Protected`（需要鉴权，且要声明所需角色）两类。

## 本阶段目标

实现完整的用户注册、登录、登出，JWT 鉴权机制，以及基于角色的访问控制（RBAC）中间件，让"接口可配置为公开或需要校验"这个核心需求落地。

## 具体任务清单

1. `domain/entity` 下定义 `User` 实体：包含 id、email 或 phone、password_hash、role（枚举：`admin`/`member`/`customer`）、created_at、updated_at、deleted_at
2. `domain/port` 下定义 `UserRepository` 接口（CreateUser、GetUserByEmail、GetUserByID 等），**只定义接口，不写实现**
3. `infrastructure/persistence/gorm` 下实现 `UserRepository` 的 GORM 版本，并补上对应的数据库迁移脚本（`migrations/` 目录，建议用 golang-migrate 或类似工具管理迁移版本）
4. `usecase/auth` 下实现：
   - 注册用例：密码用 bcrypt 哈希后存储，邮箱/手机号唯一性校验
   - 登录用例：校验密码，签发 access token（建议有效期 15 分钟）和 refresh token（建议有效期 7 天，refresh token 存入 Redis 便于主动吊销）
   - 登出用例：把对应 refresh token 加入 Redis 黑名单
   - 刷新 token 用例：用 refresh token 换新的 access token
5. JWT 的 payload 里必须包含用户 ID 和角色（role），这是后续 RBAC 中间件判断权限的依据
6. `interface/http/middleware` 下实现两个中间件：
   - `AuthMiddleware`：校验 access token 有效性，解析出用户信息存入请求上下文
   - `RBACMiddleware`：接收一个允许的角色列表作为参数，校验当前用户角色是否在列表内，不在则返回 403
7. 完善阶段 1 留的路由注册机制，明确区分：
   ```go
   router.Public("POST", "/api/v1/auth/register", authHandler.Register)
   router.Public("POST", "/api/v1/auth/login", authHandler.Login)
   router.Protected("POST", "/api/v1/auth/logout", authHandler.Logout, AnyRole)
   router.Protected("POST", "/api/v1/auth/refresh", authHandler.Refresh, AnyRole)
   ```
   `Protected` 内部应该自动叠加 `AuthMiddleware` + `RBACMiddleware`，调用方只需要声明允许哪些角色，不需要手动拼装中间件链
8. 实现一个 `GET /api/v1/auth/me` 受保护接口，返回当前登录用户信息，专门用来作为本阶段验证鉴权是否生效的测试入口
9. 补充基础的请求参数校验（邮箱格式、密码强度等），返回统一格式的错误响应

