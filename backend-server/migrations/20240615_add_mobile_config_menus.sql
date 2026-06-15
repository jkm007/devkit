-- 移动端配置菜单
-- 创建时间: 2024-06-15
-- sys_menus: pid(父ID), auth_code(权限码), type(directory/menu/button)

-- 1. 添加移动端配置父菜单
INSERT INTO sys_menus (name, path, component, icon, pid, sort, type, auth_code, status, created_at, updated_at)
VALUES ('移动端配置', '/system/mobile', NULL, 'mobile', 0, 90, 'directory', NULL, 1, NOW(), NOW());

SET @mobile_menu_id = LAST_INSERT_ID();

-- 2. 轮播图管理菜单
INSERT INTO sys_menus (name, path, component, icon, pid, sort, type, auth_code, status, created_at, updated_at)
VALUES ('轮播图管理', 'banner', '/system/banner/index', 'picture', @mobile_menu_id, 1, 'menu', 'system:banner:view', 1, NOW(), NOW());

SET @banner_menu_id = LAST_INSERT_ID();

-- 3. 轮播图管理按钮权限
INSERT INTO sys_menus (name, path, component, icon, pid, sort, type, auth_code, status, created_at, updated_at)
VALUES
('轮播图查看', NULL, NULL, NULL, @banner_menu_id, 1, 'button', 'system:banner:view', 1, NOW(), NOW()),
('轮播图新增', NULL, NULL, NULL, @banner_menu_id, 2, 'button', 'system:banner:add', 1, NOW(), NOW()),
('轮播图编辑', NULL, NULL, NULL, @banner_menu_id, 3, 'button', 'system:banner:edit', 1, NOW(), NOW()),
('轮播图删除', NULL, NULL, NULL, @banner_menu_id, 4, 'button', 'system:banner:delete', 1, NOW(), NOW());

-- 4. 考试分类管理菜单
INSERT INTO sys_menus (name, path, component, icon, pid, sort, type, auth_code, status, created_at, updated_at)
VALUES ('考试分类管理', 'exam-category', '/system/exam-category/index', 'folder', @mobile_menu_id, 2, 'menu', 'question:category', 1, NOW(), NOW());

-- 5. 题目管理菜单（如果不存在）
INSERT INTO sys_menus (name, path, component, icon, pid, sort, type, auth_code, status, created_at, updated_at)
SELECT '题目管理', 'question', '/system/question/index', 'file-text', @mobile_menu_id, 3, 'menu', 'question:view', 1, NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'question:view' AND type = 'menu');

-- 6. 通知管理菜单（如果不存在）
INSERT INTO sys_menus (name, path, component, icon, pid, sort, type, auth_code, status, created_at, updated_at)
SELECT '通知管理', 'notification', '/system/notification/index', 'bell', @mobile_menu_id, 4, 'menu', 'system:notification:view', 1, NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'system:notification:view' AND type = 'menu');

-- 7. 用户反馈菜单（如果不存在）
INSERT INTO sys_menus (name, path, component, icon, pid, sort, type, auth_code, status, created_at, updated_at)
SELECT '用户反馈', 'feedback', '/system/feedback/index', 'message', @mobile_menu_id, 5, 'menu', 'question:feedback:view', 1, NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'question:feedback:view' AND type = 'menu');

-- 8. 给超级管理员角色添加权限（更新 permissions JSON 字段）
UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(
    COALESCE(permissions, JSON_ARRAY()),
    '$',
    'system:banner:view'
)
WHERE name = 'super_admin'
AND NOT JSON_CONTAINS(COALESCE(permissions, JSON_ARRAY()), '"system:banner:view"');

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(
    COALESCE(permissions, JSON_ARRAY()),
    '$',
    'system:banner:add'
)
WHERE name = 'super_admin'
AND NOT JSON_CONTAINS(COALESCE(permissions, JSON_ARRAY()), '"system:banner:add"');

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(
    COALESCE(permissions, JSON_ARRAY()),
    '$',
    'system:banner:edit'
)
WHERE name = 'super_admin'
AND NOT JSON_CONTAINS(COALESCE(permissions, JSON_ARRAY()), '"system:banner:edit"');

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(
    COALESCE(permissions, JSON_ARRAY()),
    '$',
    'system:banner:delete'
)
WHERE name = 'super_admin'
AND NOT JSON_CONTAINS(COALESCE(permissions, JSON_ARRAY()), '"system:banner:delete"');

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(
    COALESCE(permissions, JSON_ARRAY()),
    '$',
    'question:category'
)
WHERE name = 'super_admin'
AND NOT JSON_CONTAINS(COALESCE(permissions, JSON_ARRAY()), '"question:category"');

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(
    COALESCE(permissions, JSON_ARRAY()),
    '$',
    'question:category:add'
)
WHERE name = 'super_admin'
AND NOT JSON_CONTAINS(COALESCE(permissions, JSON_ARRAY()), '"question:category:add"');

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(
    COALESCE(permissions, JSON_ARRAY()),
    '$',
    'question:category:edit'
)
WHERE name = 'super_admin'
AND NOT JSON_CONTAINS(COALESCE(permissions, JSON_ARRAY()), '"question:category:edit"');

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(
    COALESCE(permissions, JSON_ARRAY()),
    '$',
    'question:category:delete'
)
WHERE name = 'super_admin'
AND NOT JSON_CONTAINS(COALESCE(permissions, JSON_ARRAY()), '"question:category:delete"');
