-- ============================================
-- 权限菜单修复脚本
-- 优先级：P0 > P1 > P2 > P3
-- 执行顺序：先本地，再同步阿里云
-- ============================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- P0 修复：关键权限问题
-- ============================================

-- P0-1: Analytics/Workspace 不应使用 system:user:list
-- 这两个是仪表盘页面，所有登录用户都应该能看到
-- 清空 auth_code，让所有登录用户可见
UPDATE sys_menus SET auth_code = '' WHERE id IN (2, 3);

-- P0-2: FileList 应设置权限码
UPDATE sys_menus SET auth_code = 'file:view:own' WHERE id = 41;

-- P0-3: FileShare 应设置权限码
UPDATE sys_menus SET auth_code = 'share:view:own' WHERE id = 42;

-- ============================================
-- P1 修复：权限缺失和不匹配
-- ============================================

-- P1-1: 添加缺少的权限码到菜单按钮

-- 定时任务 - 添加缺少的 view 和 delete 按钮
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(135, 131, 'SystemScheduledTaskView', '', '', 'button', 1, 'system:task:view', '', 0, '{}', NOW(), NOW()),
(136, 131, 'SystemScheduledTaskDelete', '', '', 'button', 1, 'system:task:delete', '', 0, '{}', NOW(), NOW())
ON DUPLICATE KEY UPDATE auth_code = VALUES(auth_code);

-- 存储桶按钮权限（如果不存在）
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(137, 93, 'StorageManageBucketView', '', '', 'button', 1, 'storage:bucket:view', '', 0, '{}', NOW(), NOW()),
(138, 93, 'StorageManageBucketEdit', '', '', 'button', 1, 'storage:bucket:edit', '', 0, '{}', NOW(), NOW()),
(139, 93, 'StorageManageBucketDelete', '', '', 'button', 1, 'storage:bucket:delete', '', 0, '{}', NOW(), NOW())
ON DUPLICATE KEY UPDATE auth_code = VALUES(auth_code);

-- 存储配置按钮权限（已存在 97-99，确认权限码正确）
UPDATE sys_menus SET auth_code = 'storage:config:view' WHERE id = 97;
UPDATE sys_menus SET auth_code = 'storage:config:edit' WHERE id = 98;
UPDATE sys_menus SET auth_code = 'storage:config:delete' WHERE id = 99;

-- P1-2: SecuritySettings 应使用独立权限码
-- 添加新的权限码 storage:setting:list（如果需要）
-- 但当前后端没有这个权限码，先保持 security:risk:list
-- TODO: 后端需要添加 security:setting:list 权限码

-- P1-3: StorageTagRouting 使用 storage:bucket:view 语义不匹配
-- 标签路由管理需要独立权限码，但当前后端复用了 storage:bucket:*
-- TODO: 后端需要添加 storage:tag:view/edit 权限码

-- P1-4: super 角色添加缺少的 system:user:add
UPDATE sys_roles SET permissions = JSON_ARRAY_APPEND(permissions, '$', 'system:user:add')
WHERE id = 1 AND NOT JSON_CONTAINS(permissions, '"system:user:add"');

-- ============================================
-- P2 修复：前端路由缺失的菜单
-- ============================================

-- P2-1: 添加 StorageBucket 菜单（存储桶管理）
-- 前端路由: StorageBucket /storage/storage-bucket
-- 注意：当前 StorageManage (id=93) 可能就是这个功能，检查是否需要重命名
-- 先检查路径是否匹配
UPDATE sys_menus SET path = '/storage/storage-bucket', name = 'StorageBucket'
WHERE id = 93 AND path = '/storage/storage-manage';

-- P2-2: 添加 StorageConfig 菜单（存储配置）
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(140, 71, 'StorageConfig', '/storage/storage-config', 'storage/storage-config/index', 'menu', 1, 'storage:config:view', 'mdi:server-network', 2, '{"title":"存储配置","icon":"mdi:server-network","order":2}', NOW(), NOW())
ON DUPLICATE KEY UPDATE path = VALUES(path), component = VALUES(component);

-- P2-3: 添加 StorageSettings 菜单（存储设置）
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(141, 71, 'StorageSettings', '/storage/storage-settings', 'storage/storage-settings/index', 'menu', 1, 'storage:config:view', 'lucide:settings', 5, '{"title":"存储设置","icon":"lucide:settings","order":5}', NOW(), NOW())
ON DUPLICATE KEY UPDATE path = VALUES(path), component = VALUES(component);

-- P2-4: 添加 SystemTag 菜单（系统标签 - 前端 system.ts 中的路由）
-- 注意：这个与 StorageTagRouting 是不同的，SystemTag 在 /system/tag
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(142, 4, 'SystemTag', '/system/tag', 'system/tag/list', 'menu', 1, 'storage:bucket:view', 'mdi:tag-multiple', 6, '{"title":"标签管理","icon":"mdi:tag-multiple","order":6}', NOW(), NOW())
ON DUPLICATE KEY UPDATE path = VALUES(path), component = VALUES(component);

-- P2-5: user 角色添加基础导航权限
-- 让普通用户能看到仪表盘和文件菜单
UPDATE sys_roles SET permissions = '["file:view:own","file:upload","file:download","file:delete","file:manage","file:share","share:view:own","file:recycle:list","file:recycle:restore","file:recycle:delete","file:recycle:empty"]'
WHERE id = 3;

-- ============================================
-- P3 修复：清理和规范化
-- ============================================

-- P3-1: CaptchaTest 添加权限码
UPDATE sys_menus SET auth_code = 'system:captcha:test' WHERE id = 54;

-- P3-2: 确保所有目录级菜单 auth_code 为空（非 NULL）
UPDATE sys_menus SET auth_code = '' WHERE type = 'catalog' AND auth_code IS NULL;

-- ============================================
-- 验证修复结果
-- ============================================

SELECT '=== 修复后的菜单统计 ===' as info;
SELECT
  type,
  COUNT(*) as total,
  SUM(CASE WHEN auth_code IS NOT NULL AND auth_code != '' THEN 1 ELSE 0 END) as with_auth,
  SUM(CASE WHEN auth_code IS NULL OR auth_code = '' THEN 1 ELSE 0 END) as without_auth
FROM sys_menus
WHERE deleted_at IS NULL
GROUP BY type;

SELECT '=== 一级目录 ===' as info;
SELECT id, name, path, auth_code FROM sys_menus WHERE pid = 0 AND deleted_at IS NULL ORDER BY sort, id;

SELECT '=== super 角色权限数量 ===' as info;
SELECT id, name, JSON_LENGTH(permissions) as perm_count FROM sys_roles WHERE id = 1;

SET FOREIGN_KEY_CHECKS = 1;
