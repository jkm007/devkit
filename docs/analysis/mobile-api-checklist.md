# 移动端接口清单

**分析日期**: 2026-06-15
**统计**: 共 45 个接口调用

---

## 一、接口总览

| 类别 | 数量 | 状态 |
|-----|------|------|
| 认证接口 | 14 | ✅ 全部存在 |
| 用户接口 | 7 | ✅ 全部存在 |
| 学习接口 | 12 | ✅ 全部存在 |
| 通知接口 | 4 | ✅ 全部存在 |
| 轮播图 | 1 | ✅ 存在 |
| 文件接口 | 4 | ⚠️ 部分路径不匹配 |
| 其他接口 | 3 | ⚠️ 需要检查 |

---

## 二、详细接口清单

### 2.1 认证接口 (14个) ✅ 全部存在

| # | 接口路径 | 方法 | 说明 | 后端状态 |
|---|---------|------|------|---------|
| 1 | `/auth/login` | POST | 用户名密码登录 | ✅ |
| 2 | `/auth/login-by-email` | POST | 邮箱验证码登录 | ✅ |
| 3 | `/auth/login-by-phone` | POST | 手机验证码登录 | ✅ |
| 4 | `/auth/logout` | POST | 登出 | ✅ |
| 5 | `/auth/refresh` | POST | 刷新Token | ✅ |
| 6 | `/auth/register` | POST | 用户注册 | ✅ |
| 7 | `/auth/captcha` | GET | 获取验证码图片 | ✅ |
| 8 | `/auth/send-code` | POST | 发送邮箱验证码 | ✅ |
| 9 | `/auth/verify-code` | POST | 验证邮箱验证码 | ✅ |
| 10 | `/auth/reset-password` | POST | 重置密码 | ✅ |
| 11 | `/auth/change-password` | PUT | 修改密码 | ✅ |
| 12 | `/auth/devices` | GET | 获取登录设备列表 | ✅ |
| 13 | `/auth/devices/kick-all` | DELETE | 踢出所有其他设备 | ✅ |
| 14 | `/auth/security-logs` | GET | 获取安全日志 | ✅ |

### 2.2 用户接口 (7个) ✅ 全部存在

| # | 接口路径 | 方法 | 说明 | 后端状态 |
|---|---------|------|------|---------|
| 1 | `/user/info` | GET | 获取用户信息 | ✅ |
| 2 | `/user/info` | PUT | 更新用户信息 | ✅ |
| 3 | `/user/home` | GET | 获取首页数据 | ✅ |
| 4 | `/user/privacy` | GET | 获取隐私设置 | ✅ |
| 5 | `/user/privacy` | PUT | 更新隐私设置 | ✅ |
| 6 | `/user/category-bindings` | GET | 获取分类绑定 | ✅ |
| 7 | `/user/category-bindings` | POST | 绑定分类 | ✅ |

### 2.3 学习接口 (12个) ✅ 全部存在

| # | 接口路径 | 方法 | 说明 | 后端状态 |
|---|---------|------|------|---------|
| 1 | `/study/questions` | GET | 获取题目列表 | ✅ |
| 2 | `/study/questions/:id` | GET | 获取题目详情 | ✅ |
| 3 | `/study/questions/:id/favorite` | POST | 收藏题目 | ✅ |
| 4 | `/study/questions/:id/favorite` | DELETE | 取消收藏 | ✅ |
| 5 | `/user/favorites` | GET | 获取收藏列表 | ✅ |
| 6 | `/user/notes` | GET | 获取笔记列表 | ✅ |
| 7 | `/user/notes` | POST | 创建笔记 | ✅ |
| 8 | `/user/notes/:id` | PUT | 更新笔记 | ✅ |
| 9 | `/user/notes/:id` | DELETE | 删除笔记 | ✅ |
| 10 | `/study/practice/questions` | POST | 获取练习题目 | ✅ |
| 11 | `/study/practice/submit` | POST | 提交练习结果 | ✅ |
| 12 | `/study/practice/history` | GET | 获取练习历史 | ✅ |

### 2.4 错题本接口 (6个) ✅ 全部存在

| # | 接口路径 | 方法 | 说明 | 后端状态 |
|---|---------|------|------|---------|
| 1 | `/study/wrong` | GET | 获取错题列表 | ✅ |
| 2 | `/study/wrong/stats` | GET | 获取错题统计 | ✅ |
| 3 | `/study/wrong/random` | GET | 获取随机错题 | ✅ |
| 4 | `/study/wrong/:questionId/mastered` | PUT | 标记已掌握 | ✅ |
| 5 | `/study/wrong/batch-mastered` | POST | 批量标记已掌握 | ✅ |
| 6 | `/study/wrong/:questionId` | DELETE | 移除错题 | ✅ |

### 2.5 智能练习接口 (3个) ✅ 全部存在

| # | 接口路径 | 方法 | 说明 | 后端状态 |
|---|---------|------|------|---------|
| 1 | `/study/practice/smart` | POST | 智能练习 | ✅ |
| 2 | `/study/practice/analysis` | GET | 练习分析 | ✅ |
| 3 | `/study/feedback` | POST | 提交纠错反馈 | ✅ |

### 2.6 通知接口 (4个) ✅ 全部存在

| # | 接口路径 | 方法 | 说明 | 后端状态 |
|---|---------|------|------|---------|
| 1 | `/notifications` | GET | 获取通知列表 | ✅ |
| 2 | `/notifications/unread-count` | GET | 获取未读数量 | ✅ |
| 3 | `/notifications/:id/read` | PUT | 标记已读 | ✅ |
| 4 | `/notifications/read-all` | PUT | 全部已读 | ✅ |

### 2.7 轮播图接口 (1个) ✅ 存在

| # | 接口路径 | 方法 | 说明 | 后端状态 |
|---|---------|------|------|---------|
| 1 | `/banners` | GET | 获取轮播图列表 | ✅ |

### 2.8 文件接口 (4个) ⚠️ 部分路径不匹配

| # | 接口路径 | 方法 | 说明 | 后端状态 |
|---|---------|------|------|---------|
| 1 | `/files/list` | GET | 获取文件列表 | ✅ |
| 2 | `/files/upload/chunk/init` | POST | 初始化分片上传 | ⚠️ 后端是 `/files/upload/init` |
| 3 | `/files/upload/chunk` | POST | 上传分片 | ⚠️ 后端是 `/files/upload/part` |
| 4 | `/files/batch/delete` | POST | 批量删除文件 | ⚠️ 后端是 `/files/batch-delete` |

### 2.9 其他接口 (3个) ⚠️ 需要检查

| # | 接口路径 | 方法 | 说明 | 后端状态 |
|---|---------|------|------|---------|
| 1 | `/auth/devices/:id` | DELETE | 踢出指定设备 | ✅ |
| 2 | `/exam-categories/all` | GET | 获取考试分类列表 | ❌ 需要添加 |
| 3 | `/questions/search` | GET | 搜索题目 | ⚠️ 需要检查 |

---

## 三、接口缺失/不匹配清单

### 3.1 需要修改的接口（路径不匹配）

| 移动端路径 | 后端路径 | 修改建议 |
|-----------|---------|---------|
| `/files/upload/chunk/init` | `/files/upload/init` | 修改前端路径 |
| `/files/upload/chunk` | `/files/upload/part` | 修改前端路径 |
| `/files/batch/delete` | `/files/batch-delete` | 修改前端路径 |

### 3.2 需要添加的接口

| 接口路径 | 方法 | 说明 |
|---------|------|------|
| `/exam-categories/all` | GET | 获取考试分类列表 |
| `/questions/search` | GET | 搜索题目 |

---

## 四、接口调用频率分析

### 高频接口（首页/每次访问）

| 接口 | 调用时机 | 建议 |
|-----|---------|------|
| `/banners` | 首页加载 | ✅ 已添加Redis缓存 |
| `/user/home` | 首页加载 | 建议添加缓存 |
| `/user/info` | 登录后 | 建议添加本地缓存 |
| `/notifications/unread-count` | 首页加载 | 建议WebSocket推送 |

### 中频接口（页面切换）

| 接口 | 调用时机 | 建议 |
|-----|---------|------|
| `/study/questions` | 进入题库 | 建议分页缓存 |
| `/study/wrong/stats` | 进入错题本 | 建议缓存5分钟 |
| `/user/category-bindings` | 分类页面 | 建议缓存 |

---

## 五、后端路由注册状态

### 已注册的路由组

```
/api/v1/auth/*          - 认证接口 ✅
/api/v1/user/*          - 用户接口 ✅
/api/v1/study/*         - 学习接口 ✅
/api/v1/notifications/* - 通知接口 ✅
/api/v1/banners         - 轮播图 ✅
/api/v1/files/*         - 文件接口 ✅
```

### 需要添加的路由

```go
// router.go 中添加
authorized.GET("/exam-categories/all", examCategoryHandler.GetAll)
authorized.GET("/questions/search", questionHandler.Search)
```

---

## 六、总结

### 接口可用性统计

| 状态 | 数量 | 占比 |
|-----|------|------|
| ✅ 完全匹配 | 38 | 84% |
| ⚠️ 路径不匹配 | 3 | 7% |
| ❌ 缺失 | 2 | 4% |
| ⚠️ 待检查 | 2 | 5% |

### 立即需要做的

1. **修改文件接口路径**（3个）
   - `/files/upload/chunk/init` → `/files/upload/init`
   - `/files/upload/chunk` → `/files/upload/part`
   - `/files/batch/delete` → `/files/batch-delete`

2. **添加缺失接口**（2个）
   - `/exam-categories/all`
   - `/questions/search`

### 建议优化

1. 高频接口添加缓存
2. 文件接口统一命名规范
3. 添加WebSocket推送通知
