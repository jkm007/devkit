-- ============================================================
-- 迁移: 018
-- 描述: 添加文件类型规则管理菜单
-- 作者: Claude Code
-- 日期: 2026-06-09
-- ============================================================

-- 获取系统管理目录 ID
SET @system_id = (SELECT id FROM sys_menus WHERE name = 'System' AND type = 'catalog' LIMIT 1);

-- -------------------------------------------
-- 文件类型规则管理菜单
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@system_id, 'SystemFileTypeRule', '/system/file-type-rule', '/system/file-type-rule/index', 'menu', 1, 'system:setting:list', 'mdi:file-check', '{"order":11,"title":"文件类型规则"}');
