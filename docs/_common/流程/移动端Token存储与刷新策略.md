# 移动端 Token 存储与刷新策略

本文档说明 H5、App、小程序、App WebView 等正式用户端接入 DevKit 登录认证时的 Token 存储、刷新、设备标识和 CSRF 边界策略。

## 1. Token 类型

| Token | 默认有效期 | 用途 | 后端存储 |
|-------|------------|------|----------|
| AccessToken | 7 天 | 调用认证接口，放入 `Authorization: Bearer <token>` | 不落库；登出/改密时 SHA-256 写入 Redis 黑名单 |
| RefreshToken | 30 天 | 轮换刷新 AccessToken 和 RefreshToken | SHA-256 哈希存入 Redis `refresh_token:{userID}` |

登录成功后，后端同时：

1. 在响应体返回 `accessToken`、`refreshToken`。
2. 写入 HttpOnly Cookie：`access_token`、`refresh_token`。

移动端推荐优先使用响应体中的 Token，避免依赖 Cookie。

---

## 2. 各端推荐存储策略

### Web 管理端

| 项目 | 建议 |
|------|------|
| AccessToken | Pinia 持久化或内存 +持久化，当前管理端已持久化 |
| RefreshToken | 当前管理端已持久化，刷新时也兼容 Cookie |
| 请求认证 | `Authorization: Bearer <accessToken>` |
| 刷新方式 | Cookie `refresh_token` + Header/Body `refreshToken` 兼容 |
| CSRF | 保留 Double Submit Cookie 模式 |

### H5 浏览器

| 项目 | 建议 |
|------|------|
| AccessToken | 内存优先；如需保持登录可存 localStorage/sessionStorage，但需做好 XSS 防护 |
| RefreshToken | 优先存安全封装层；纯浏览器 H5 可使用 Cookie 或由业务决定是否持久化 |
| 请求认证 | `Authorization: Bearer <accessToken>` |
| 刷新方式 | 推荐 `X-Refresh-Token` 或 body `refreshToken`，不要强依赖第三方 Cookie |
| CSRF | 纯 Token 模式可使用 `X-Mobile-Token-Mode: true` 跳过 CSRF；Cookie 模式仍需 CSRF |

### App WebView

| 项目 | 建议 |
|------|------|
| AccessToken | 原生容器安全存储后注入 WebView，或 WebView 内存保存 |
| RefreshToken | 原生安全存储，避免长时间暴露在 JS 环境 |
| 请求认证 | `Authorization: Bearer <accessToken>` |
| 刷新方式 | `X-Refresh-Token` 或 body `refreshToken` |
| CSRF | 使用纯 Token 模式，不依赖 Cookie |

### 原生 App

| 项目 | 建议 |
|------|------|
| AccessToken | Keychain / Keystore / 系统安全存储 |
| RefreshToken | Keychain / Keystore / 系统安全存储，严禁明文日志输出 |
| 请求认证 | `Authorization: Bearer <accessToken>` |
| 刷新方式 | `X-Refresh-Token` 或 body `refreshToken` |
| CSRF | 不需要浏览器 CSRF；高风险接口可使用验证码/风控二次校验 |

### 小程序

| 项目 | 建议 |
|------|------|
| AccessToken | 小程序安全存储接口，必要时内存保存 |
| RefreshToken | 小程序安全存储接口，定期清理 |
| 请求认证 | `Authorization: Bearer <accessToken>` |
| 刷新方式 | `X-Refresh-Token` 或 body `refreshToken` |
| CSRF | 不使用 Cookie 认证时可走纯 Token 模式 |

---

## 3. 登录请求 Header 规范

所有正式用户端建议统一上报设备信息。

### 通用 Header

```http
X-Client-Type: h5
X-Device-ID: device-uuid
X-Platform: ios
```

`X-Client-Type` 可选值：

```text
web / h5 / app / miniapp
```

`X-Platform` 可选值：

```text
web / h5 / ios / android / miniapp / windows / macos / linux
```

### 移动端扩展 Header

```http
X-App-Version: 1.2.3
X-System-Version: iOS 17.5
X-Device-Model: iPhone 15 Pro
X-Channel: appstore
```

这些字段会记录到 `sys_login_devices`：

| Header | 数据库字段 |
|--------|------------|
| `X-Client-Type` | `device_type` |
| `X-App-Version` | `app_version` |
| `X-System-Version` | `system_version` |
| `X-Device-Model` | `device_model` |
| `X-Platform` | `platform` |
| `X-Channel` | `channel` |

---

## 4. Token 刷新流程

### 推荐请求

```http
POST /api/v1/auth/refresh
Authorization: Bearer <expired-or-current-access-token>
X-Client-Type: app
X-Mobile-Token-Mode: true
X-Refresh-Token: <refreshToken>
```

也可以使用请求体：

```json
{
  "refreshToken": "<refreshToken>"
}
```

### 后端读取优先级

```text
Cookie refresh_token → Header X-Refresh-Token → Body refreshToken
```

### 后端刷新逻辑

1. 解析 RefreshToken，校验签名算法、签名和过期时间。
2. 查询用户和角色。
3. 生成新的 AccessToken 和 RefreshToken。
4. 计算旧 RefreshToken SHA-256 哈希。
5. Redis Lua 脚本原子校验旧哈希并替换为新哈希。
6. 返回新 Token，并同步写 Cookie。

并发刷新时，只有第一个请求能成功，其他请求会因为 RefreshToken 哈希不匹配失败。

---

## 5. CSRF 边界策略

### Web Cookie 模式

后台 Web 管理端或浏览器 Cookie 认证模式需要 CSRF：

```http
Cookie: csrf_token=<token>
X-CSRF-Token: <token>
```

后端使用 Double Submit Cookie 模式校验 Cookie 和 Header 是否一致。

### 移动端纯 Token 模式

H5/App/小程序如果不依赖 Cookie 作为认证凭据，而是使用：

```http
Authorization: Bearer <accessToken>
```

可以显式声明纯 Token 模式：

```http
X-Mobile-Token-Mode: true
X-Client-Type: app
Authorization: Bearer <accessToken>
```

后端只有同时满足以下条件才跳过 CSRF：

1. `X-Mobile-Token-Mode: true`。
2. `X-Client-Type` 是 `h5` / `app` / `miniapp`。
3. 存在 `Authorization: Bearer <token>`。

这样避免仅伪造 `X-Client-Type` 就绕过 Web Cookie 模式的 CSRF 防护。

### 高风险接口建议

即使移动端跳过 CSRF，高风险接口仍建议使用：

- 图形验证码；
- 风险评分 `CaptchaGuard`；
- 短信/邮箱二次验证；
- 操作日志和安全通知。

---

## 6. 登出和失效策略

### 登出

调用：

```http
POST /api/v1/auth/logout
Authorization: Bearer <accessToken>
```

后端会：

1. 将 AccessToken SHA-256 写入 Redis 黑名单 `token_blacklist:{sha256(token)}`。
2. 删除 Redis `refresh_token:{userID}`。
3. 清空 Cookie。

移动端本地也必须删除：

- AccessToken；
- RefreshToken；
- 当前用户信息；
- 权限码缓存。

### 修改密码 / 重置密码

修改密码或重置密码后，后端会删除 RefreshToken，使旧会话无法继续刷新。

---

## 7. 多端登录和踢设备

登录成功后，后端根据 `X-Device-ID` 记录设备。移动端应确保同一设备长期使用稳定的设备 ID。

设备管理接口：

```http
GET /api/v1/auth/devices?deviceType=app
DELETE /api/v1/auth/devices/{id}
DELETE /api/v1/auth/devices/kick-all
```

踢出设备时，后端写入 Redis：

```text
kicked_device:{userID}:{deviceID}
```

后续该设备请求会被 `JWTAuth` 拒绝。

---

## 8. 客户端接入伪代码

```ts
async function loginByPhone(phone: string, code: string) {
  const result = await request.post('/auth/login-by-phone', { phone, code }, {
    headers: {
      'X-Client-Type': 'app',
      'X-Device-ID': getDeviceId(),
      'X-App-Version': getAppVersion(),
      'X-System-Version': getSystemVersion(),
      'X-Device-Model': getDeviceModel(),
      'X-Platform': getPlatform(),
      'X-Channel': getChannel(),
    },
  });
  saveAccessToken(result.accessToken);
  saveRefreshToken(result.refreshToken);
}

async function refresh() {
  const result = await request.post('/auth/refresh', {
    refreshToken: getRefreshToken(),
  }, {
    headers: {
      Authorization: `Bearer ${getAccessToken()}`,
      'X-Client-Type': 'app',
      'X-Mobile-Token-Mode': 'true',
      'X-Refresh-Token': getRefreshToken(),
    },
  });
  saveAccessToken(result.accessToken);
  saveRefreshToken(result.refreshToken);
}
```

---

## 9. 安全要求

1. RefreshToken 不允许写入日志。
2. 移动端 RefreshToken 应使用系统安全存储。
3. AccessToken 过期后优先刷新，不要频繁重新登录。
4. 刷新失败时必须清理本地 Token 并跳转登录。
5. Web Cookie 模式继续保留 CSRF。
6. 移动端纯 Token 模式必须显式带 `X-Mobile-Token-Mode: true`。
7. CORS 不允许生产环境使用 `*` + credentials。
