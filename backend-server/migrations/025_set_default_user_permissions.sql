-- ============================================================
-- 迁移: 025
-- 描述: 添加文件管理菜单 + 为普通用户设置默认权限
-- 作者: Claude Code
-- 日期: 2026-06-10
-- ============================================================

-- 1. 添加文件管理目录（如果不存在）
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT 0, 'File', '/file', '', 'catalog', 1, '', 'lucide:folder', 9996, '{"title":"文件管理","order":9996}'
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'File' AND type = 'catalog' AND deleted_at IS NULL);

SET @file_catalog_id = (SELECT id FROM sys_menus WHERE name = 'File' AND type = 'catalog' AND deleted_at IS NULL LIMIT 1);

-- 2. 添加文件列表菜单
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @file_catalog_id, 'FileList', '/file/list', '/file/list/index', 'menu', 1, 'file:list', 'lucide:files', 1, '{"title":"文件列表","icon":"lucide:files","order":1}'
FROM DUAL WHERE @file_catalog_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'FileList' AND deleted_at IS NULL);

-- 3. 添加分享管理菜单
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @file_catalog_id, 'FileShare', '/file/share', '/file/share/index', 'menu', 1, 'share:view:own', 'lucide:share-2', 2, '{"title":"分享管理","icon":"lucide:share-2","order":2}'
FROM DUAL WHERE @file_catalog_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'FileShare' AND deleted_at IS NULL);

-- 4. 回收站菜单（migration 022 可能已添加，这里确保存在）
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @file_catalog_id, 'FileRecycle', '/file/recycle', '/file/recycle/index', 'menu', 1, 'file:recycle:list', 'lucide:trash-2', 3, '{"title":"回收站","icon":"lucide:trash-2","order":3}'
FROM DUAL WHERE @file_catalog_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'FileRecycle' AND deleted_at IS NULL);

-- 5. 回收站操作按钮（如果不存在）
SET @recycle_id = (SELECT id FROM sys_menus WHERE name = 'FileRecycle' AND deleted_at IS NULL LIMIT 1);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @recycle_id, 'FileRecycleRestore', '', '', 'button', 1, 'file:recycle:restore', '', 0, '{"title":"恢复文件"}'
FROM DUAL WHERE @recycle_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'FileRecycleRestore' AND deleted_at IS NULL);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @recycle_id, 'FileRecycleDelete', '', '', 'button', 1, 'file:recycle:delete', '', 0, '{"title":"永久删除"}'
FROM DUAL WHERE @recycle_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'FileRecycleDelete' AND deleted_at IS NULL);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @recycle_id, 'FileRecycleEmpty', '', '', 'button', 1, 'file:recycle:empty', '', 0, '{"title":"清空回收站"}'
FROM DUAL WHERE @recycle_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'FileRecycleEmpty' AND deleted_at IS NULL);

-- 6. 为普通用户(user)角色设置默认权限
UPDATE sys_roles SET permissions = '[
  "file:list",
  "file:recycle:list",
  "file:recycle:restore",
  "file:recycle:delete",
  "file:recycle:empty",
  "share:view:own",
  "system:security:view",
  "system:device:view",
  "system:privacy:view",
  "system:oauth:view"
]' WHERE name = 'user' AND deleted_at IS NULL;
