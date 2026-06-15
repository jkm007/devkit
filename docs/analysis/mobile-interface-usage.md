# 移动端接口使用情况分析报告

**分析日期**: 2026-06-14
**项目**: DevKit 移动端应用 (frontend-app)

---

## 一、接口统计总览 📊

移动端应用共使用了 **68个** 后端API接口。

### 接口分类统计

| 类别 | 接口数量 | 说明 |
|-----|---------|------|
| **认证接口** | 14 | 登录、注册、Token管理、验证码 |
| **用户信息** | 8 | 用户资料、设备管理、隐私设置 |
| **学习功能** | 25 | 题目、练习、错题本、收藏、笔记 |
| **通知消息** | 5 | 通知列表、未读计数、标记已读 |
| **题目纠错** | 4 | 提交反馈、查看历史 |
| **轮播图** | 1 | 公开接口，首页展示 |
| **分类绑定** | 4 | 用户绑定科目分类 |
| **文件管理** | 11 | 上传、下载、预览、删除 |
| **用户功能** | 6 | 设备、隐私、OAuth、安全日志 |

---

## 二、详细接口列表 📝

### 2.1 认证接口 (14个)

| 接口路径 | HTTP方法 | 说明 | 是否需要认证 |
|---------|---------|------|------------|
| `/auth/login` | POST | 用户名密码登录 | ❌ |
| `/auth/login-by-email` | POST | 邮箱验证码登录 | ❌ |
| `/auth/login-by-phone` | POST | 手机验证码登录 | ❌ |
| `/auth/register` | POST | 用户注册 | ❌ |
| `/auth/logout` | POST | 登出 | ✅ |
| `/auth/refresh` | POST | 刷新Token | ❌ (需RefreshToken) |
| `/auth/captcha` | GET | 获取验证码图片 | ❌ |
| `/auth/send-code` | POST | 发送邮箱验证码 | ❌ |
| `/auth/verify-code` | POST | 验证邮箱验证码 | ❌ |
| `/auth/reset-password` | POST | 重置密码 | ❌ |
| `/auth/change-password` | PUT | 修改密码 | ✅ |
| `/api/v1/user/info` | GET | 获取用户信息 | ✅ |
| `/api/v1/user/info` | PUT | 更新用户信息 | ✅ |
| `/api/v1/auth/devices` | GET | 获取登录设备列表 | ✅ |

### 2.2 设备管理接口 (4个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/auth/devices` | GET | 获取登录设备列表 |
| `/api/v1/auth/devices/kick-all` | DELETE | 踢出所有其他设备 |
| `/api/v1/auth/devices/:id` | DELETE | 踢出指定设备 |
| `/api/v1/auth/security-logs` | GET | 获取安全日志 |

### 2.3 隐私设置接口 (2个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/user/privacy` | GET | 获取隐私设置 |
| `/api/v1/user/privacy` | PUT | 更新隐私设置 |

### 2.4 OAuth接口 (3个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/auth/oauth/bindings` | GET | 获取OAuth绑定列表 |
| `/api/v1/auth/oauth/bind-url` | GET | 获取OAuth绑定URL |
| `/api/v1/auth/oauth/unbind` | POST | 解绑OAuth |

### 2.5 角色申请接口 (3个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/auth/role-applications/available-roles` | GET | 获取可申请的角色 |
| `/api/v1/auth/role-applications` | POST | 提交角色申请 |
| `/api/v1/auth/role-applications` | GET | 获取我的角色申请列表 |

### 2.6 学习功能接口 (25个)

#### 题目接口 (6个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/study/questions` | GET | 获取题目列表 |
| `/api/v1/study/questions/:id` | GET | 获取题目详情 |
| `/api/v1/study/questions/:id/favorite` | POST | 收藏题目 |
| `/api/v1/study/questions/:id/favorite` | DELETE | 取消收藏 |
| `/api/v1/user/favorites` | GET | 获取收藏列表 |
| `/api/v1/user/category-bindings` | GET | 获取分类绑定 |

#### 笔记接口 (4个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/user/notes` | GET | 获取笔记列表 |
| `/api/v1/user/notes` | POST | 创建笔记 |
| `/api/v1/user/notes/:id` | PUT | 更新笔记 |
| `/api/v1/user/notes/:id` | DELETE | 删除笔记 |

#### 练习接口 (5个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/study/practice/questions` | POST | 获取练习题目 |
| `/api/v1/study/practice/submit` | POST | 提交练习结果 |
| `/api/v1/study/practice/history` | GET | 获取练习历史 |
| `/api/v1/study/practice/smart` | POST | 智能练习 |
| `/api/v1/study/practice/analysis` | GET | 练习分析 |

#### 错题本接口 (6个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/study/wrong` | GET | 获取错题列表 |
| `/api/v1/study/wrong/stats` | GET | 获取错题统计 |
| `/api/v1/study/wrong/random` | GET | 获取随机错题 |
| `/api/v1/study/wrong/:questionId/mastered` | PUT | 标记已掌握 |
| `/api/v1/study/wrong/batch-mastered` | POST | 批量标记已掌握 |
| `/api/v1/study/wrong/:questionId` | DELETE | 移除错题 |

#### 分类绑定接口 (4个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/user/category-bindings` | GET | 获取分类绑定 |
| `/api/v1/user/category-bindings` | POST | 绑定分类 |
| `/api/v1/user/category-bindings/:id` | PUT | 设为主分类 |
| `/api/v1/user/category-bindings/:id` | DELETE | 解绑分类 |

### 2.7 通知接口 (5个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/notifications` | GET | 获取通知列表 |
| `/api/v1/notifications/unread-count` | GET | 获取未读数量 |
| `/api/v1/notifications/:id/read` | PUT | 标记已读 |
| `/api/v1/notifications/read-all` | PUT | 全部已读 |
| `/api/v1/notifications/:id` | DELETE | 删除通知 |

### 2.8 题目纠错接口 (4个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/study/feedback` | POST | 提交纠错反馈 |
| `/api/v1/study/feedback` | GET | 获取纠错列表 |
| `/api/v1/study/feedback/:id` | GET | 获取纠错详情 |
| `/api/v1/study/feedback/:id` | DELETE | 删除纠错反馈 |

### 2.9 轮播图接口 (1个)

| 接口路径 | HTTP方法 | 说明 | 是否需要认证 |
|---------|---------|------|------------|
| `/api/v1/banners` | GET | 获取启用的轮播图列表 | ❌ |

### 2.10 文件管理接口 (11个)

| 接口路径 | HTTP方法 | 说明 |
|---------|---------|------|
| `/api/v1/files/upload` | POST | 上传文件（小文件） |
| `/api/v1/files/upload/chunk/init` | POST | 初始化分片上传 |
| `/api/v1/files/upload/chunk` | POST | 上传分片 |
| `/api/v1/files/upload/chunk/complete/:uploadId` | POST | 完成分片上传 |
| `/api/v1/files/:fileId` | GET | 获取文件信息 |
| `/api/v1/files/:fileId/download` | GET | 获取下载链接 |
| `/api/v1/files/:fileId/preview` | GET | 获取预览链接 |
| `/api/v1/files/:fileId` | DELETE | 删除文件 |
| `/api/v1/files/batch/delete` | POST | 批量删除文件 |
| `/api/v1/files/list` | GET | 获取文件列表 |

---

## 三、接口使用频率分析 🔍

### 高频接口（首页加载）

移动端首页加载时会调用以下接口：

1. **`/api/v1/banners`** - 获取轮播图（公开接口）
2. **`/api/v1/user/info`** - 获取用户信息
3. **`/api/v1/user/home`** - 获取首页数据（未在API文件中找到，推测在user-feature.ts）
4. **`/api/v1/study/wrong/stats`** - 获取错题统计
5. **`/api/v1/notifications/unread-count`** - 获取未读消息数

**建议**: 这些高频接口应该添加缓存机制。

### 核心功能接口

#### 学习模块（使用最多）
- 题目列表、详情、收藏：6个接口
- 笔记管理：4个接口
- 练习功能：5个接口
- 错题本：6个接口

**占比**: 31% (21/68)

#### 用户管理
- 认证：14个接口
- 设备管理：4个接口
- 隐私设置：2个接口
- OAuth：3个接口

**占比**: 34% (23/68)

---

## 四、接口命名规范分析 ✅

### 良好的命名规范

1. **RESTful 设计**: 大部分接口遵循 RESTful 规范
   - GET：查询
   - POST：创建
   - PUT：更新
   - DELETE：删除

2. **资源层级清晰**:
   - `/api/v1/study/questions/:id/favorite` - 题目的收藏操作
   - `/api/v1/user/notes/:id` - 用户笔记的指定ID

3. **版本化管理**: 所有接口使用 `/api/v1` 前缀

### 建议改进的命名

1. **部分接口路径不一致**:
   - `/auth/login` (无/api/v1前缀) vs `/api/v1/user/info` (有前缀)
   - 建议：统一使用 `/api/v1/auth/login`

2. **动作命名混合**:
   - `/api/v1/study/wrong/:questionId/mastered` (动词mastered)
   - 建议：改为 `/api/v1/study/wrong/:questionId/mark-mastered`

---

## 五、接口安全性分析 🔒

### 认证接口（无需登录）

**数量**: 10个

包括：
- 登录（3种方式）
- 注册
- 验证码相关（3个）
- 重置密码
- Token刷新

**安全措施**:
- ✅ 验证码机制（防止暴力破解）
- ✅ 验证码频率限制（60秒冷却）
- ⚠️ 建议：添加 IP 限流

### 需认证接口（需要登录）

**数量**: 58个

**安全措施**:
- ✅ JWT Token认证
- ✅ Token加密存储（XOR+Base64）
- ✅ Token黑名单机制
- ✅ 设备踢出检查
- ✅ Fail-closed策略

**建议改进**:
- 🔴 升级Token加密方案（AES-256）
- 🟡 添加请求日志记录
- 🟡 添加权限验证（部分接口可能缺少）

---

## 六、接口性能优化建议 🚀

### 高频接口缓存

| 接口 | 建议 |
|-----|------|
| `/api/v1/banners` | ✅ 已添加Redis缓存（5分钟） |
| `/api/v1/user/info` | 建议添加本地缓存（1分钟） |
| `/api/v1/study/wrong/stats` | 建议添加缓存（5分钟） |
| `/api/v1/notifications/unread-count` | 建议添加WebSocket实时推送 |

### 分页接口优化

所有列表接口支持分页：
- ✅ 默认pageSize合理（20）
- ✅ 最大pageSize限制（50）
- ⚠️ 建议：添加总数限制，防止超大分页

---

## 七、接口文档完整性 📚

### 有Swagger文档的接口

根据后端代码分析，部分接口有Swagger注释：
- ✅ 登录设备接口（4个）
- ⚠️ 其他接口缺少Swagger注释

### 建议完善Swagger文档

需要补充Swagger注释的接口：
- 认证接口（除login外）
- 学习功能接口（25个）
- 通知接口（5个）
- 文件管理接口（11个）

---

## 八、总结与建议 📝

### 移动端接口使用情况总结

1. **接口数量适中**: 68个接口，覆盖核心功能
2. **命名规范良好**: 大部分遵循RESTful规范
3. **安全性到位**: JWT认证完善，但有提升空间
4. **性能可优化**: 高频接口需要缓存机制

### 建议优化项

#### 高优先级

1. **统一接口路径前缀**: 所有接口使用 `/api/v1` 前缀
2. **高频接口缓存**: user/info、wrong/stats等添加缓存
3. **完善Swagger文档**: 为所有接口添加注释

#### 中等优先级

4. **升级Token加密**: 从XOR升级到AES-256
5. **添加请求日志**: 记录用户行为，便于审计
6. **接口限流细化**: 公开接口添加更严格的限流

#### 低优先级

7. **接口命名统一**: 动词命名改为规范形式
8. **WebSocket推送**: 未读消息改为实时推送

---

**分析完成时间**: 2026-06-14 09:00
**下一步**: 测试接口可用性