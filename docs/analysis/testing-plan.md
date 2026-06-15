# DevKit 移动端完整测试计划和建议

**分析日期**: 2026-06-14
**项目**: DevKit 全栈系统

---

## 一、已完成的工作 ✅

### 1. 关键问题修复

#### 后端修复
- ✅ **Banner数据Redis缓存**（5分钟TTL，自动清理）
- ✅ **Handler错误日志**（使用zap logger记录详细错误）
- ✅ **缓存清理机制**（Banner更新/删除时自动清理缓存）

#### 移动端修复
- ✅ **Token加密/解密错误日志**（catch块添加console.error）
- ✅ **缓存操作错误日志**（所有catch块添加日志记录）

**代码已提交**: commit 94b8b7e

### 2. 接口使用分析

- ✅ **移动端接口统计**: 共68个后端API接口
- ✅ **接口分类**: 9个类别（认证、学习、通知、文件等）
- ✅ **高频接口识别**: 首页加载5个核心接口
- ✅ **安全性分析**: JWT认证完善，有提升空间

**报告已保存**: `/docs/analysis/mobile-interface-usage.md`

---

## 二、待测试的关键接口 🧪

### 高优先级测试接口（首页加载）

这些接口是移动端首页必须加载的，应该优先测试：

#### 1. Banner接口（已优化）

```bash
# 公开接口，无需认证
curl http://localhost:8080/api/v1/banners
```

**预期结果**:
- 返回轮播图列表
- 第一次从数据库查询
- 第二次从Redis缓存读取（耗时显著降低）
- 缓存5分钟后自动过期

#### 2. 用户信息接口

```bash
# 需要登录Token
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/user/info
```

**预期结果**:
- 返回用户基本信息（ID、username、nickname、avatar）
- 返回用户角色列表

#### 3. 错题统计接口

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/study/wrong/stats
```

**预期结果**:
- 返回错题总数、已掌握数、未掌握数
- 按科目分类统计

#### 4. 未读消息接口

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/notifications/unread-count
```

**预期结果**:
- 返回未读通知数量

#### 5. 登录接口（核心）

```bash
# 密码登录
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test123"}'
```

**预期结果**:
- 返回accessToken和refreshToken
- Token有效期合理

---

## 三、数据库表检查建议 📊

### 需要检查的核心表

根据移动端功能需求，需要验证以下数据库表：

#### 1. 用户相关表

| 表名 | 字段检查项 |
|-----|----------|
| `sys_users` | ✅ id, username, password, email, phone, nickname, avatar, status |
| `sys_roles` | ✅ id, name, permissions, status |
| `sys_user_roles` | ✅ user_id, role_id |
| `sys_login_devices` | ✅ user_id, device_type, device_id, ip_address, location, last_active_at |
| `sys_user_privacy` | ✅ user_id, show_profile, show_stats |

**潜在问题**:
- ⚠️ 检查 `sys_users.avatar` 字段长度（是否支持长URL）
- ⚠️ 检查 `sys_login_devices.location` 字段（是否存储地理位置）

#### 2. 学习相关表

| 表名 | 字段检查项 |
|-----|----------|
| `questions` | ✅ id, title, content, question_type, difficulty, answer, analysis |
| `question_categories` | ✅ id, name, parent_id, level, sort_order |
| `question_favorites` | ✅ user_id, question_id, created_at |
| `question_notes` | ✅ id, user_id, question_id, content, created_at |
| `question_wrong_books` | ✅ user_id, question_id, wrong_count, is_mastered, last_wrong_at |
| `practice_records` | ✅ user_id, total, answered, correct, elapsed, created_at |
| `question_feedback` | ✅ user_id, question_id, feedback_type, description, suggestion, status |

**潜在问题**:
- 🔴 `question_notes.content` 字段类型是否为TEXT（支持长内容）
- 🔴 `question_wrong_books.wrong_count` 是否支持大数值
- ⚠️ `practice_records` 表是否缺少字段（如category_id）

#### 3. 通知相关表

| 表名 | 字段检查项 |
|-----|----------|
| `notifications` | ✅ id, user_id, title, content, type, is_read, created_at |

**潜在问题**:
- ⚠️ `notifications.type` 字段是否为枚举类型
- ⚠️ `notifications.content` 是否支持长文本

#### 4. 轮播图表（新增）

| 表名 | 字段检查项 |
|-----|----------|
| `banners` | ✅ id, title, image, link, link_type, sort_order, status, created_at |

**检查项**:
- ✅ 表结构完整（已验证）
- ✅ 字段类型合理
- ⚠️ `banners.image` 字段长度（支持长URL）

#### 5. 文件相关表

| 表名 | 字段检查项 |
|-----|----------|
| `sys_files` | ✅ id, user_id, filename, file_path, file_size, file_type, folder |
| `sys_file_tags` | ✅ file_id, tag_id |
| `sys_tags` | ✅ id, key, value, category |

**潜在问题**:
- ⚠️ `sys_files.folder` 字段是否支持多级目录
- ⚠️ `sys_files.file_path` 是否支持长路径

---

## 四、字段完整性检查SQL 📝

### 执行这些SQL检查字段完整性

```sql
-- 1. 检查用户表字段
DESC sys_users;
SELECT COUNT(*) FROM sys_users WHERE avatar IS NOT NULL AND LENGTH(avatar) > 255;

-- 2. 检查题目表字段
DESC questions;
SELECT COUNT(*) FROM questions WHERE LENGTH(content) > 1000;

-- 3. 检查笔记表字段类型
DESC question_notes;
SELECT COUNT(*) FROM question_notes WHERE LENGTH(content) > 1000;

-- 4. 检查错题表字段
DESC question_wrong_books;
SELECT MAX(wrong_count) FROM question_wrong_books;

-- 5. 检查轮播图表
DESC banners;
SELECT COUNT(*) FROM banners WHERE status = 'enabled' ORDER BY sort_order;

-- 6. 检查通知表字段类型
DESC notifications;
SELECT COUNT(*) FROM notifications WHERE type IN ('system', 'practice', 'feedback');

-- 7. 检查文件表字段长度
DESC sys_files;
SELECT COUNT(*) FROM sys_files WHERE LENGTH(file_path) > 255;

-- 8. 检查设备表字段
DESC sys_login_devices;
SELECT device_type, COUNT(*) FROM sys_login_devices GROUP BY device_type;
```

---

## 五、接口测试建议 🔍

### 测试优先级

#### 第一阶段：基础功能测试（5个接口）

1. **登录接口** - 核心功能
2. **Banner接口** - 验证缓存效果
3. **用户信息接口** - 验证Token机制
4. **错题统计接口** - 验证数据统计
5. **未读消息接口** - 验证通知系统

#### 第二阶段：学习功能测试（10个接口）

1. **题目列表** - 分页、筛选
2. **题目详情** - 内容展示、媒体文件
3. **收藏/取消收藏** - 操作正确性
4. **笔记CRUD** - 创建、查看、更新、删除
5. **练习功能** - 随机练习、提交结果
6. **错题本** - 添加、查看、标记掌握

#### 第三阶段：用户管理测试（8个接口）

1. **设备管理** - 查看设备、踢出设备
2. **隐私设置** - 查看、更新
3. **OAuth绑定** - 查看绑定、解绑
4. **安全日志** - 查看日志

#### 第四阶段：文件和通知测试（剩余接口）

根据实际需要测试文件上传、通知等功能。

---

## 六、自动化测试脚本建议 🤖

### 创建测试脚本

建议创建以下测试脚本：

```bash
#!/bin/bash
# mobile-api-test.sh - 移动端API自动化测试

BASE_URL="http://localhost:8080"
TOKEN=""  # 登录后获取

echo "=== DevKit 移动端API测试 ==="

# 1. 测试Banner接口（验证缓存）
echo "1. 测试Banner接口..."
curl -s "$BASE_URL/api/v1/banners" | jq '.'
sleep 1
echo "第二次请求（应该命中缓存）..."
curl -s "$BASE_URL/api/v1/banners" | jq '.data | length'

# 2. 测试登录接口
echo "2. 测试登录接口..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test123"}')
echo $LOGIN_RESPONSE | jq '.'

# 提取Token
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.accessToken')

# 3. 测试用户信息接口
echo "3. 测试用户信息接口..."
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/user/info" | jq '.'

# 4. 测试错题统计接口
echo "4. 测试错题统计..."
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/study/wrong/stats" | jq '.'

# 5. 测试未读消息接口
echo "5. 测试未读消息数..."
curl -s -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/notifications/unread-count" | jq '.'

echo "=== 测试完成 ==="
```

---

## 七、潜在问题清单 ⚠️

### 需要立即关注的问题

#### 数据库字段问题

1. 🔴 **笔记表content字段**: 是否为TEXT类型（支持长内容）
2. 🔴 **题目表content字段**: 是否支持富文本/HTML（长度限制）
3. ⚠️ **轮播图image字段**: URL长度是否足够（512字符）

#### 接口功能问题

4. 🟡 **验证码发送功能**: 移动端TODO未实现
5. 🟡 **设备ID生成**: 弱随机数，建议升级为指纹
6. 🟡 **Token加密**: XOR混淆强度不足，建议AES-256

#### 性能问题

7. ⚠️ **高频接口缓存**: user/info、wrong/stats需要缓存
8. ⚠️ **WebSocket推送**: 通知应该实时推送，而非轮询

---

## 八、下一步行动计划 🎯

### 立即执行（今天）

1. ✅ **重启后端服务器**（确保新代码生效）
2. ✅ **执行数据库字段检查SQL**
3. ✅ **测试Banner接口**（验证缓存效果）
4. ✅ **测试登录和用户信息接口**

### 明天执行

5. ✅ **测试学习功能接口**（题目、练习、错题本）
6. ✅ **测试文件上传接口**
7. ✅ **编写自动化测试脚本**

### 本周完成

8. ✅ **补充缺失的数据库字段**
9. ✅ **完成验证码发送功能**
10. ✅ **优化高频接口性能**

---

## 九、测试环境准备 🛠️

### 后端服务器启动

```bash
cd ../backend-server
./backend-server
```

### 移动端H5启动

```bash
cd ../frontend-app
node_modules/.bin/uni
```

**H5地址**: http://localhost:5173/

### 测试账号

需要准备测试账号：
- 用户名: `test`
- 密码: `test123`
- 或者使用现有账号登录

---

**计划创建时间**: 2026-06-14 09:10
**建议执行者**: 开发团队
**预计完成时间**: 2026-06-17