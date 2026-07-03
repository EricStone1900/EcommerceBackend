# 阶段 5 / 7 — 推送服务与设备 Token 管理

## 上文衔接

阶段 1-4 已完成：分层骨架、用户认证与 RBAC、商品管理、文件上传与事件机制。本阶段实现推送服务的设备 Token 注册能力，并预留 APNs 真实推送的接入点（本阶段不需要真正接入 APNs）。继续在已有目录结构里开发。

## 全局架构约束（必须遵守）

`domain/port` 下定义 `Notifier` 接口，usecase 层只依赖这个接口，不知道底层是打日志还是真的调用 APNs。

## 本阶段目标

1. 用户可以注册自己的 iOS 设备 token
2. 实现 `Notifier` 接口的桩实现（Stub），当前只打日志或写入数据库记录，预留好未来接 APNs 的位置

## 具体任务清单

1. `domain/entity` 下定义 `PushToken` 实体：id、user_id、device_token、platform（如 `ios`，预留未来扩展 `android`）、created_at
2. `domain/port` 下定义：
   - `PushTokenRepository` 接口（Create、GetByUserID、Delete）
   - `Notifier` 接口：`SendPush(ctx context.Context, userID string, title string, body string) error`
3. `infrastructure/persistence/gorm` 下实现 `PushTokenRepository`，补充数据库迁移脚本
4. `infrastructure/push/stub` 下实现 `Notifier` 的桩版本：根据 userID 查出对应的设备 token，打印一条结构化日志（包含 token、title、body），或者额外写入一张 `push_logs` 表记录"本次本应发送的推送内容"，方便后续真接入 APNs 时可以对照验证逻辑是否正确
5. 预留 `infrastructure/push/apns/` 空目录或者一个带详细注释的文件，写明"未来接入 APNs 时在此实现 Notifier 接口，需要的配置项包括：APNs 证书/密钥、bundle ID、生产/沙箱环境地址"等，让后续真正接入时一目了然要做什么
6. `usecase/push` 下实现：
   - 设备 token 注册用例（同一用户重复注册同一个 token 要做幂等处理，不要插入重复记录）
   - 设备 token 删除用例（用户登出或卸载 App 时调用）
   - 一个简单的"发送测试推送"用例，调用 `Notifier.SendPush`，用于本阶段验证链路
7. 路由设计：
   ```go
   router.Protected("POST", "/api/v1/push/token", pushHandler.RegisterToken, AnyRole)
   router.Protected("DELETE", "/api/v1/push/token", pushHandler.DeleteToken, AnyRole)
   router.Protected("POST", "/api/v1/push/test", pushHandler.SendTest, AnyRole)
   ```

