-- 修复缺失的菜单和权限
-- 1. 给 admin 角色补全角色申请查看权限
-- 2. 修复移动端配置菜单结构和权限
-- 3. 修复班级管理菜单父级
-- 4. 补全存储管理子菜单

-- 给 admin 角色添加 system:roleapp:view（已存在则跳过）
UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(permissions, '$', 'system:roleapp:view'),
    updated_at = NOW()
WHERE name = 'admin'
  AND JSON_SEARCH(permissions, 'one', 'system:roleapp:view') IS NULL;

-- 修复移动端配置顶层菜单
UPDATE sys_menus
SET name = 'Mobile',
    path = '/mobile',
    component = NULL,
    meta = '{"title": "移动端配置"}',
    auth_code = NULL,
    sort = 6
WHERE id = 296;

-- 修复移动端子菜单
UPDATE sys_menus
SET pid = 296, name = 'MobileBanner', path = '/mobile/banner', component = 'system/banner/list',
    meta = '{"title": "轮播图管理"}', auth_code = 'system:banner:view', sort = 1
WHERE id = 297;

UPDATE sys_menus
SET pid = 296, name = 'MobileQuickMenu', path = '/mobile/quick-menu', component = 'mobile/quick-menu/index',
    meta = '{"title": "快捷菜单"}', auth_code = 'system:mobile:menu:view', sort = 2
WHERE id = 304;

UPDATE sys_menus
SET pid = 296, name = 'MobileMyPage', path = '/mobile/my-page', component = 'mobile/my-page/index',
    meta = '{"title": "我的页面"}', auth_code = 'system:mobile:menu:view', sort = 3
WHERE id = 305;

UPDATE sys_menus
SET pid = 296, name = 'MobileSettings', path = '/mobile/settings', component = 'mobile/settings/index',
    meta = '{"title": "移动端设置"}', auth_code = 'system:mobile:settings:view', sort = 4
WHERE id = 306;

UPDATE sys_menus
SET pid = 296, name = 'MobileFeedback', path = '/mobile/feedback', component = 'system/feedback/index',
    meta = '{"title": "用户反馈"}', auth_code = 'question:feedback:view', sort = 5
WHERE id = 303;

-- 修复班级管理菜单父级
UPDATE sys_menus SET pid = 0 WHERE name = 'Class';

-- 添加缺失的存储设置和存储管理菜单
INSERT INTO sys_menus (pid, name, path, component, type, status, icon, sort, meta, auth_code, created_at, updated_at)
SELECT 71, 'StorageSettings', '/storage/storage-settings', 'storage/storage-settings/index', 'menu', 1, 'lucide:settings', 5, '{"title": "存储设置"}', 'system:setting:list', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'StorageSettings');

INSERT INTO sys_menus (pid, name, path, component, type, status, icon, sort, meta, auth_code, created_at, updated_at)
SELECT 71, 'StorageManage', '/storage/storage-manage', 'storage/storage-manage/index', 'menu', 1, 'lucide:database', 6, '{"title": "存储管理"}', 'storage:config:list', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'StorageManage');

-- 给 admin 角色添加移动端相关权限（已存在则跳过）
UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(permissions, '$', 'system:mobile:menu:view'), updated_at = NOW()
WHERE name = 'admin' AND JSON_SEARCH(permissions, 'one', 'system:mobile:menu:view') IS NULL;

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(permissions, '$', 'system:mobile:menu:create'), updated_at = NOW()
WHERE name = 'admin' AND JSON_SEARCH(permissions, 'one', 'system:mobile:menu:create') IS NULL;

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(permissions, '$', 'system:mobile:menu:edit'), updated_at = NOW()
WHERE name = 'admin' AND JSON_SEARCH(permissions, 'one', 'system:mobile:menu:edit') IS NULL;

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(permissions, '$', 'system:mobile:menu:delete'), updated_at = NOW()
WHERE name = 'admin' AND JSON_SEARCH(permissions, 'one', 'system:mobile:menu:delete') IS NULL;

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(permissions, '$', 'system:mobile:settings:view'), updated_at = NOW()
WHERE name = 'admin' AND JSON_SEARCH(permissions, 'one', 'system:mobile:settings:view') IS NULL;

UPDATE sys_roles
SET permissions = JSON_ARRAY_APPEND(permissions, '$', 'system:mobile:settings:edit'), updated_at = NOW()
WHERE name = 'admin' AND JSON_SEARCH(permissions, 'one', 'system:mobile:settings:edit') IS NULL;
