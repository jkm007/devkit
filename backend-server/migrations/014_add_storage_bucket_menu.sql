-- ============================================================
-- 迁移: 014
-- 描述: 添加存储桶管理菜单和权限
-- 作者: Claude Code
-- 日期: 2026-06-08
-- ============================================================

-- 获取系统管理目录 ID
SET @system_id = (SELECT id FROM sys_menus WHERE name = 'System' AND type = 'catalog' LIMIT 1);

-- -------------------------------------------
-- 存储桶管理菜单
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@system_id, 'SystemStorageBucket', '/system/storage-bucket', '/system/storage-bucket/index', 'menu', 1, 'system:setting:list', 'mdi:bucket', '{"order":11,"title":"存储桶管理"}');

SET @bucket_menu_id = LAST_INSERT_ID();

-- -------------------------------------------
-- 按钮权限: 存储桶管理（复用系统设置权限）
-- -------------------------------------------
-- 存储桶管理使用 system:setting:list 和 system:setting:edit 权限
-- 因为它属于系统设置的一部分
