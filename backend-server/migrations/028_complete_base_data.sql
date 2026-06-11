-- ============================================================
-- 迁移: 028
-- 描述: 补全完整的基础数据（菜单、按钮、权限）
-- 作者: Claude Code
-- 日期: 2026-06-11
-- ============================================================

USE backend_db;

-- ============================================================
-- 第一部分：更新 Analytics 和 Workspace 的权限
-- ============================================================

-- Analytics 和 Workspace 只对管理员可见（需要 system:user:list 权限）
UPDATE sys_menus SET auth_code = 'system:user:list' WHERE name IN ('Analytics', 'Workspace') AND type = 'menu' AND deleted_at IS NULL;

-- ============================================================
-- 第二部分：补全 FileShareManage 菜单和按钮
-- ============================================================

SET @file_catalog_id = (SELECT id FROM sys_menus WHERE name = 'File' AND type = 'catalog' AND deleted_at IS NULL LIMIT 1);

-- 分享管理菜单（如果不存在）
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @file_catalog_id, 'FileShareManage', '/file/share', '/file/share/index', 'menu', 1, '', 'lucide:share-2', 2, '{"title":"分享管理","icon":"lucide:share-2","order":2}'
FROM DUAL WHERE @file_catalog_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'FileShareManage' AND deleted_at IS NULL);

SET @share_menu_id = (SELECT id FROM sys_menus WHERE name = 'FileShareManage' AND deleted_at IS NULL LIMIT 1);

-- 如果 FileShareManage 不存在，就用 FileShare 的 ID
IF @share_menu_id IS NULL THEN
    SET @share_menu_id = (SELECT id FROM sys_menus WHERE name = 'FileShare' AND deleted_at IS NULL LIMIT 1);
END IF;

-- 补全分享管理按钮
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @share_menu_id, 'FileShareViewAll', '', '', 'button', 1, 'share:view:all', '', '{"title":"查看所有分享"}', 0
FROM DUAL WHERE @share_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'share:view:all' AND pid = @share_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @share_menu_id, 'FileShareDelete', '', '', 'button', 1, 'share:delete', '', '{"title":"删除分享"}', 0
FROM DUAL WHERE @share_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'share:delete' AND pid = @share_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @share_menu_id, 'FileShareManageBtn', '', '', 'button', 1, 'share:manage', '', '{"title":"管理分享"}', 0
FROM DUAL WHERE @share_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'share:manage' AND pid = @share_menu_id);

-- ============================================================
-- 第三部分：补全安全管理下的限速规则菜单
-- ============================================================

SET @security_id = (SELECT id FROM sys_menus WHERE name = 'Security' AND type = 'catalog' AND deleted_at IS NULL LIMIT 1);

-- 限速规则菜单（如果不存在）
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @security_id, 'SecurityRateLimit', '/security/rate-limit', '/security/rate-limit/index', 'menu', 1, 'security:ratelimit:list', 'lucide:sliders', 4, '{"title":"限速规则","icon":"lucide:sliders","order":4}'
FROM DUAL WHERE @security_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name IN ('SecurityRateLimit', 'SecuritySettings') AND auth_code = 'security:ratelimit:list' AND deleted_at IS NULL);

SET @ratelimit_menu_id = (SELECT id FROM sys_menus WHERE auth_code = 'security:ratelimit:list' AND type = 'menu' AND deleted_at IS NULL LIMIT 1);

-- 限速规则按钮
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @ratelimit_menu_id, 'RateLimitRuleView', '', '', 'button', 1, 'security:ratelimit:view', '', '{"title":"查看规则"}', 0
FROM DUAL WHERE @ratelimit_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'security:ratelimit:view' AND pid = @ratelimit_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @ratelimit_menu_id, 'RateLimitRuleEdit', '', '', 'button', 1, 'security:ratelimit:edit', '', '{"title":"编辑规则"}', 0
FROM DUAL WHERE @ratelimit_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'security:ratelimit:edit' AND pid = @ratelimit_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @ratelimit_menu_id, 'RateLimitRuleDelete', '', '', 'button', 1, 'security:ratelimit:delete', '', '{"title":"删除规则"}', 0
FROM DUAL WHERE @ratelimit_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'security:ratelimit:delete' AND pid = @ratelimit_menu_id);

-- ============================================================
-- 第四部分：补全存储管理菜单的正确结构
-- ============================================================

SET @storage_id = (SELECT id FROM sys_menus WHERE name = 'Storage' AND type = 'catalog' AND deleted_at IS NULL LIMIT 1);

-- 确保 StorageBucket 菜单正确
UPDATE sys_menus SET name = 'StorageBucket', path = '/storage/storage-bucket', component = '/storage/storage-bucket/index', auth_code = 'storage:bucket:list', icon = 'mdi:database', sort = 1, meta = '{"title":"存储桶管理","icon":"mdi:database","order":1}'
WHERE id = 93 AND deleted_at IS NULL;

-- 确保 StorageConfig 菜单正确
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @storage_id, 'StorageConfig', '/storage/storage-config', '/storage/storage-config/index', 'menu', 1, 'storage:config:list', 'mdi:server-network', 2, '{"title":"存储连接配置","icon":"mdi:server-network","order":2}'
FROM DUAL WHERE @storage_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'StorageConfig' AND auth_code = 'storage:config:list' AND deleted_at IS NULL);

SET @storage_config_id = (SELECT id FROM sys_menus WHERE name = 'StorageConfig' AND auth_code = 'storage:config:list' AND deleted_at IS NULL LIMIT 1);

-- 补全 StorageConfig 按钮（如果不存在）
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @storage_config_id, 'StorageConfigView', '', '', 'button', 1, 'storage:config:view', '', '{"title":"查看存储配置"}', 0
FROM DUAL WHERE @storage_config_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:config:view' AND pid = @storage_config_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @storage_config_id, 'StorageConfigEdit', '', '', 'button', 1, 'storage:config:edit', '', '{"title":"编辑存储配置"}', 0
FROM DUAL WHERE @storage_config_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:config:edit' AND pid = @storage_config_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @storage_config_id, 'StorageConfigDelete', '', '', 'button', 1, 'storage:config:delete', '', '{"title":"删除存储配置"}', 0
FROM DUAL WHERE @storage_config_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:config:delete' AND pid = @storage_config_id);

-- 删除重复的 StorageSettings 菜单（id=141）
UPDATE sys_menus SET deleted_at = NOW(3) WHERE id = 141 AND deleted_at IS NULL;

-- 清理 StorageBucket 下多余的按钮（已经有了，保持现状）

-- ============================================================
-- 第五部分：更新角色权限（确保所有权限都有）
-- ============================================================

-- super 角色：拥有所有完整权限
UPDATE sys_roles SET permissions = JSON_ARRAY(
    'system:user:list', 'system:user:add', 'system:user:edit', 'system:user:delete', 'system:user:view',
    'system:role:list', 'system:role:add', 'system:role:edit', 'system:role:delete', 'system:role:view',
    'system:menu:list', 'system:menu:add', 'system:menu:edit', 'system:menu:delete', 'system:menu:view',
    'system:group:list', 'system:group:add', 'system:group:edit', 'system:group:delete', 'system:group:view',
    'system:setting:list', 'system:setting:edit',
    'system:task:list', 'system:task:view', 'system:task:edit', 'system:task:run', 'system:task:delete',
    'system:captcha:test',
    'system:roleapp:view', 'system:roleapp:review',
    'system:device:view', 'system:device:delete',
    'system:privacy:view',
    'system:oauth:view',
    'system:realname:view',
    'security:realname:list', 'security:realname:approve', 'security:realname:reject',
    'security:log:list', 'security:log:view',
    'security:risk:list', 'security:risk:view', 'security:risk:edit',
    'security:ratelimit:list', 'security:ratelimit:view', 'security:ratelimit:edit', 'security:ratelimit:delete',
    'storage:bucket:list', 'storage:bucket:view', 'storage:bucket:edit', 'storage:bucket:delete',
    'storage:config:list', 'storage:config:view', 'storage:config:edit', 'storage:config:delete',
    'storage:file-type:list', 'storage:file-type:view', 'storage:file-type:edit', 'storage:file-type:delete',
    'file:list', 'file:view:own', 'file:view:all', 'file:upload', 'file:download', 'file:delete', 'file:share', 'file:manage',
    'file:recycle:list', 'file:recycle:restore', 'file:recycle:delete', 'file:recycle:empty',
    'share:view:own', 'share:view:all', 'share:delete', 'share:manage'
) WHERE name = 'super' AND deleted_at IS NULL;

-- admin 角色：拥有大部分完整权限
UPDATE sys_roles SET permissions = JSON_ARRAY(
    'system:user:list', 'system:user:add', 'system:user:edit', 'system:user:delete', 'system:user:view',
    'system:role:list', 'system:role:add', 'system:role:edit', 'system:role:delete', 'system:role:view',
    'system:menu:list', 'system:menu:add', 'system:menu:edit', 'system:menu:delete', 'system:menu:view',
    'system:group:list', 'system:group:add', 'system:group:edit', 'system:group:delete', 'system:group:view',
    'system:setting:list', 'system:setting:edit',
    'system:task:list', 'system:task:view', 'system:task:edit', 'system:task:run', 'system:task:delete',
    'system:captcha:test',
    'system:roleapp:view', 'system:roleapp:review',
    'system:device:view', 'system:device:delete',
    'system:privacy:view',
    'system:oauth:view',
    'system:realname:view',
    'security:realname:list', 'security:realname:approve', 'security:realname:reject',
    'security:log:list', 'security:log:view',
    'security:risk:list', 'security:risk:view', 'security:risk:edit',
    'security:ratelimit:list', 'security:ratelimit:view', 'security:ratelimit:edit', 'security:ratelimit:delete',
    'storage:bucket:list', 'storage:bucket:view', 'storage:bucket:edit', 'storage:bucket:delete',
    'storage:config:list', 'storage:config:view', 'storage:config:edit', 'storage:config:delete',
    'storage:file-type:list', 'storage:file-type:view', 'storage:file-type:edit', 'storage:file-type:delete',
    'file:list', 'file:view:own', 'file:view:all', 'file:upload', 'file:download', 'file:delete', 'file:share', 'file:manage',
    'file:recycle:list', 'file:recycle:restore', 'file:recycle:delete', 'file:recycle:empty',
    'share:view:own', 'share:view:all', 'share:delete', 'share:manage'
) WHERE name = 'admin' AND deleted_at IS NULL;

-- user 角色：基础文件操作权限 + 个人功能
UPDATE sys_roles SET permissions = JSON_ARRAY(
    -- 文件管理
    'file:list', 'file:view:own', 'file:upload', 'file:download', 'file:delete', 'file:share', 'file:manage',
    'file:recycle:list', 'file:recycle:restore', 'file:recycle:delete', 'file:recycle:empty',
    -- 分享管理
    'share:view:own', 'share:delete', 'share:manage',
    -- 个人功能
    'system:realname:view',
    'system:device:view', 'system:device:delete',
    'system:privacy:view',
    'system:oauth:view',
    'system:roleapp:view', 'system:roleapp:review'
) WHERE name = 'user' AND deleted_at IS NULL;

-- ============================================================
-- 第六部分：确保 FileList 和 FileShare 的权限正确
-- ============================================================

-- FileList 的 auth_code 应该为空（菜单本身无权限，权限由按钮控制）
UPDATE sys_menus SET auth_code = '' WHERE name = 'FileList' AND type = 'menu' AND deleted_at IS NULL;

-- FileShare 的 auth_code 应该为空（菜单本身无权限，权限由按钮控制）
UPDATE sys_menus SET auth_code = '' WHERE name IN ('FileShare', 'FileShareManage') AND type = 'menu' AND deleted_at IS NULL;

-- FileRecycle 的 auth_code 应该为空（菜单本身无权限，权限由按钮控制）
UPDATE sys_menus SET auth_code = '' WHERE name = 'FileRecycle' AND type = 'menu' AND deleted_at IS NULL;

-- ============================================================
-- 第七部分：确保 SystemTag 菜单正确（System 目录下）
-- ============================================================

SET @system_id = (SELECT id FROM sys_menus WHERE name = 'System' AND type = 'catalog' AND deleted_at IS NULL LIMIT 1);

-- 确保 SystemTag 在 System 目录下，且 auth_code 正确
UPDATE sys_menus SET pid = @system_id, name = 'SystemTag', path = '/system/tag', component = '/system/tag/index', auth_code = 'storage:bucket:view', icon = 'mdi:tag-multiple', sort = 6, meta = '{"title":"标签管理","icon":"mdi:tag-multiple","order":6}'
WHERE id = 104 AND deleted_at IS NULL;

-- 确保没有重复的 SystemTag
DELETE FROM sys_menus WHERE name = 'SystemTag' AND auth_code = 'system:setting:list' AND id != 104 AND deleted_at IS NULL;

-- ============================================================
-- 完成
-- ============================================================
SELECT '基础数据补全完成' AS status;
