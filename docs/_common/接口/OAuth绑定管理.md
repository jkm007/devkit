# OAuth 绑定管理接口文档

> 对应表：`sys_oauth_users`

## 接口概览

| 接口 | 方法 | 用途 | 认证 |
|------|------|------|------|
| `/auth/oauth/bindings` | GET | 获取当前用户的第三方绑定列表 | 需要 |
| `/auth/oauth/bind-url` | GET | 获取第三方授权 URL | 需要 |
| `/auth/oauth/callback` | GET | 第三方授权回调 | 不需要 |
| `/auth/oauth/unbind` | POST | 解绑第三方账号 | 需要 |

---

## 1. 获取第三方绑定列表

### 接口信息

- **路径：** `GET /auth/oauth/bindings`
- **用途：** 获取当前用户已绑定的第三方账号列表
- **认证：** 需要 `Authorization: Bearer <token>`

### 返回格式

```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "provider": "wechat",
      "providerUsername": "微信用户",
      "providerAvatar": "https://wx.qlogo.cn/xxx",
      "createdAt": "2026-05-01 10:00:00"
    },
    {
      "id": 2,
      "provider": "github",
      "providerUsername": "vben-dev",
      "providerAvatar": "https://avatars.githubusercontent.com/u/xxx",
      "createdAt": "2026-04-15 08:00:00"
    }
  ],
  "error": null,
  "message": "ok"
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | number | 绑定记录 ID |
| `provider` | string | 提供商：wechat/github/google |
| `providerUsername` | string | 第三方用户名 |
| `providerAvatar` | string | 第三方头像 |
| `createdAt` | string | 绑定时间 |

---

## 2. 获取第三方授权 URL

### 接口信息

- **路径：** `GET /auth/oauth/bind-url`
- **用途：** 获取第三方平台的授权页面 URL，前端跳转到该 URL 进行授权
- **认证：** 需要 `Authorization: Bearer <token>`

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `provider` | string | 是 | 提供商：wechat/github/google |
| `redirectUri` | string | 否 | 授权完成后的回调地址 |

### 返回格式

```json
{
  "code": 0,
  "data": {
    "url": "https://github.com/login/oauth/authorize?client_id=xxx&redirect_uri=xxx&scope=user"
  },
  "error": null,
  "message": "ok"
}
```

---

## 3. 第三方授权回调

### 接口信息

- **路径：** `GET /auth/oauth/callback`
- **用途：** 第三方平台授权完成后的回调接口，由第三方平台重定向到此
- **认证：** 不需要（通过 state 参数防 CSRF）

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `provider` | string | 是 | 提供商 |
| `code` | string | 是 | 授权码 |
| `state` | string | 是 | 状态参数（防 CSRF） |

### 返回格式

**成功 — 绑定到已有账号：**

```json
{
  "code": 0,
  "data": {
    "action": "bind",
    "provider": "github",
    "providerUsername": "vben-dev"
  },
  "error": null,
  "message": "ok"
}
```

**成功 — 第三方登录：**

```json
{
  "code": 0,
  "data": {
    "action": "login",
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "vben",
      "nickname": "Vben"
    }
  },
  "error": null,
  "message": "ok"
}
```

---

## 4. 解绑第三方账号

### 接口信息

- **路径：** `POST /auth/oauth/unbind`
- **用途：** 解绑指定的第三方账号
- **认证：** 需要 `Authorization: Bearer <token>`

### 请求体

```json
{
  "provider": "github"
}
```

### 请求字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `provider` | string | 是 | 提供商：wechat/github/google |

### 返回格式

```json
{
  "code": 0,
  "data": null,
  "error": null,
  "message": "ok"
}
```

### 副作用

- 删除 `sys_oauth_users` 表中对应的绑定记录
- 在 `sys_security_logs` 表记录解绑事件

### 错误响应

```json
// 未绑定该提供商
{
  "code": -1,
  "data": null,
  "error": "This provider is not bound.",
  "message": "This provider is not bound."
}

// 至少保留一个登录方式
{
  "code": -1,
  "data": null,
  "error": "Cannot unbind the only login method.",
  "message": "Cannot unbind the only login method."
}
```

---

## 支持的第三方平台

| provider | 说明 | 授权方式 |
|----------|------|---------|
| `wechat` | 微信（小程序/公众号） | OAuth2.0 |
| `github` | GitHub | OAuth2.0 |
| `google` | Google | OAuth2.0 / OIDC |

---

## 前端调用示例

```typescript
import {
  getOAuthBindings,
  getOAuthBindUrl,
  unbindOAuth,
} from '#/api/auth';

// 获取已绑定列表
const bindings = await getOAuthBindings();

// 获取授权 URL 并跳转
const { url } = await getOAuthBindUrl({ provider: 'github' });
window.location.href = url;

// 解绑
await unbindOAuth({ provider: 'github' });
```
