# 移动端问题总结和解决方案

## 问题1：API路径重复 ✅ 已解决

**问题现象**：
```
GET http://10.0.50.207:8080/api/v1/api/v1/user/home 404
```

**原因**：
- baseURL已包含 `/api/v1`
- API文件和.vue文件也包含 `/api/v1`

**解决方案**：
- ✅ 批量移除API文件前缀（commit 555ccc0）
- ✅ 批量移除.vue文件前缀（commit 2d693fd）

---

## 问题2：后台轮播图管理 ✅ 已创建

**解决方案**：
- ✅ 创建Banner管理页面（commit 2cdaf8a）
- ✅ 创建API接口文件
- ✅ 创建编辑弹窗

**需要配置**：
需要在数据库添加菜单和权限：

```sql
-- 1. 添加菜单
INSERT INTO sys_menus (parent_id, name, path, component, icon, sort_order, status, created_at, updated_at)
VALUES (
  (SELECT id FROM sys_menus WHERE name = '系统管理'),
  '轮播图管理',
  '/system/banner',
  '/system/banner/list',
  'PictureOutlined',
  8,
  1,
  NOW(),
  NOW()
);

-- 2. 添加权限
INSERT INTO sys_permissions (name, code, description, created_at, updated_at)
VALUES
('轮播图查看', 'system:banner:view', '查看轮播图列表', NOW(), NOW()),
('轮播图添加', 'system:banner:add', '添加轮播图', NOW(), NOW()),
('轮播图编辑', 'system:banner:edit', '编辑轮播图', NOW(), NOW()),
('轮播图删除', 'system:banner:delete', '删除轮播图', NOW(), NOW());

-- 3. 给角色分配权限（假设角色ID为2）
INSERT INTO sys_role_permissions (role_id, permission_id, created_at)
SELECT 2, id, NOW() FROM sys_permissions WHERE code LIKE 'system:banner:%';

-- 4. 添加默认Banner数据
INSERT INTO banners (title, image, link, link_type, sort_order, status, created_at, updated_at)
VALUES
('📚 题库更新：新增500道网络协议真题', '/uploads/banner1.jpg', '/pages/question/list', 'internal', 1, 'enabled', NOW(), NOW()),
('🎯 智能练习上线', '/uploads/banner2.jpg', '/pages/practice/smart', 'internal', 2, 'enabled', NOW(), NOW()),
('📕 错题本复习功能', '/uploads/banner3.jpg', '/pages/profile/wrong-book', 'internal', 3, 'enabled', NOW(), NOW());
```

---

## 问题3：移动端配置问题 🔧 需要修复

### 3.1 TabBar图标不显示

**原因**：
- SVG图标在H5环境可能不支持
- uni-app对SVG支持有限

**解决方案**：

**方案1：转换为PNG图标**
```bash
# 需要准备PNG格式的图标文件
# 建议尺寸：48x48像素
# 修改pages.json：
"iconPath": "static/images/tabbar/home.png",
"selectedIconPath": "static/images/tabbar/home-active.png"
```

**方案2：使用字体图标**
```json
"iconPath": "static/images/tabbar/home.png"  # 暂时用PNG替代
```

**方案3：使用uview-plus的图标**
```vue
<template>
  <u-icon name="home"></u-icon>
</template>
```

### 3.2 up-popup组件未找到

**错误信息**：
```
[Vue warn]: Failed to resolve component: up-popup
```

**原因**：
- uview-plus组件库未正确引入

**解决方案**：

**修改main.ts，引入uview-plus**：
```typescript
// main.ts
import uviewPlus from 'uview-plus';

const app = createSSRApp(App);
app.use(uviewPlus);
```

**修改pages.json，添加easycom配置**：
```json
{
  "easycom": {
    "autoscan": true,
    "custom": {
      "^up-(.*)": "uview-plus/components/up-$1/up-$1.vue"
    }
  }
}
```

### 3.3 分类配置全是Mock数据

**问题位置**：`pages/profile/categories.vue`

**原因**：
- `loadCategories()` 函数全是Mock数据
- 没有调用真实API获取分类列表

**解决方案**：

需要后端提供分类列表接口：

```sql
-- 添加分类数据
INSERT INTO question_categories (name, parent_id, level, sort_order, status, created_at, updated_at)
VALUES
('网络协议', 0, 1, 1, 1, NOW(), NOW()),
('操作系统', 0, 1, 2, 1, NOW(), NOW()),
('数据结构', 0, 1, 3, 1, NOW(), NOW()),
('数据库', 0, 1, 4, 1, NOW(), NOW()),
('算法', 0, 1, 5, 1, NOW(), NOW());
```

**修改categories.vue，调用真实API**：
```typescript
async function loadCategories() {
  try {
    // 调用真实API（需要后端提供）
    const res = await request.get('/question-categories/all');
    const allCategories = res || [];

    // 过滤已绑定的分类
    const boundIds = new Set(bindings.value.map(b => b.categoryId));
    categories.value = allCategories.filter(cat => !boundIds.has(cat.id));
  } catch (error) {
    // 降级处理
    console.error('加载分类失败:', error);
  }
}
```

---

## 📋 需要添加的后端接口

### 1. 分类列表接口（移动端）

**接口**：`GET /api/v1/question-categories/all`

**Handler**：
```go
// question_category_handler.go
func (h *QuestionCategoryHandler) GetAll(c *gin.Context) {
    categories, err := h.service.GetAll()
    if err != nil {
        response.InternalError(c, "获取分类列表失败")
        return
    }
    response.Success(c, categories)
}
```

**Service**：
```go
func (s *QuestionCategoryService) GetAll() ([]model.QuestionCategory, error) {
    return s.repo.GetAll()
}
```

**Repository**：
```go
func (r *QuestionCategoryRepo) GetAll() ([]model.QuestionCategory, error) {
    var categories []model.QuestionCategory
    err := r.db.Where("status = ?", 1).Order("sort_order ASC").Find(&categories).Error
    return categories, err
}
```

**路由注册**：
```go
// router.go
authorized.GET("/question-categories/all", questionCategoryHandler.GetAll)
```

### 2. 首页数据接口

**接口**：`GET /api/v1/user/home`

**返回结构**：
```json
{
  "code": 0,
  "data": {
    "stats": {
      "totalQuestions": 15234,
      "todayPracticeCount": 45,
      "todayCorrectRate": 0.78,
      "continuousDays": 12
    },
    "recommended": [
      {
        "id": 1,
        "title": "TCP三次握手的目的是什么？",
        "questionType": "single_choice",
        "difficulty": 2,
        "categoryName": "网络协议"
      }
    ]
  }
}
```

---

## 🔧 立即可以做的修复

### 1. 添加uview-plus配置

**修改frontend-app/src/main.ts**：
```typescript
import uviewPlus from 'uview-plus';

const app = createSSRApp(App);
app.use(pinia);
app.use(uviewPlus);
```

**修改frontend-app/src/pages.json**：
```json
{
  "easycom": {
    "autoscan": true,
    "custom": {
      "^up-(.*)": "uview-plus/components/up-$1/up-$1.vue"
    }
  },
  "pages": [...]
}
```

### 2. 准备PNG图标

需要准备以下图标文件（48x48px）：
- `home.png` / `home-active.png`
- `question.png` / `question-active.png`
- `practice.png` / `practice-active.png`
- `my.png` / `my-active.png`

### 3. 添加测试数据

**SQL脚本**：
```sql
-- Banner数据
INSERT INTO banners ...;  -- 见上文

-- 分类数据
INSERT INTO question_categories ...;  -- 见上文

-- 题目数据
INSERT INTO questions (title, content, question_type, difficulty, answer, analysis, category_id)
VALUES ('TCP三次握手的目的是什么？', '...', 'single_choice', 2, '建立可靠连接', '...', 1);
```

---

## 📊 优先级排序

### 高优先级（立即执行）

1. ✅ **修复API路径重复** - 已完成
2. 🔧 **添加uview-plus配置** - 解决组件找不到问题
3. 🔧 **准备PNG图标** - 解决图标不显示问题
4. 🔧 **添加测试数据** - 让页面有真实数据

### 中等优先级（本周完成）

5. 🔧 **添加分类列表接口** - 后端开发
6. 🔧 **添加首页数据接口** - 后端开发
7. 🔧 **完善Banner管理** - 前端优化

### 低优先级（后续完善）

8. 图片上传功能优化
9. 拖拽排序功能
10. 批量操作功能

---

## 🚀 快速启动步骤

### 步骤1：刷新浏览器

移动端H5服务器已重启，刷新浏览器：
```
http://10.0.50.207:5173/
```

### 步骤2：添加uview-plus配置

修改以下文件：
- `frontend-app/src/main.ts`
- `frontend-app/src/pages.json`

### 步骤3：添加数据库数据

执行SQL脚本添加：
- Banner数据
- 分类数据
- 题目数据

### 步骤4：添加后台菜单

通过管理后台的菜单管理添加：
- 轮播图管理菜单

---

**下一步建议**：
1. 立即刷新浏览器测试API是否正常
2. 添加uview-plus配置解决组件问题
3. 添加数据库测试数据

需要我帮你执行哪个步骤？