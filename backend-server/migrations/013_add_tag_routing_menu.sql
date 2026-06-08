-- ============================================================
-- 迁移: 013
-- 描述: 添加标签路由管理菜单和权限
-- 作者: Claude Code
-- 日期: 2026-06-08
-- ============================================================

-- 获取系统管理目录 ID
SET @system_id = (SELECT id FROM sys_menus WHERE name = 'System' AND type = 'catalog' LIMIT 1);

-- -------------------------------------------
-- 标签路由管理菜单
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@system_id, 'SystemTag', '/system/tag', '/system/tag/index', 'menu', 1, 'system:setting:list', 'mdi:tag-multiple', '{"order":10,"title":"标签路由"}');

SET @tag_menu_id = LAST_INSERT_ID();

-- -------------------------------------------
-- 按钮权限: 标签路由管理（复用系统设置权限）
-- -------------------------------------------
-- 标签路由管理使用 system:setting:list 和 system:setting:edit 权限
-- 因为它属于系统设置的一部分

-- 更新 admin 角色权限（假设 admin 角色 ID = 1）
-- 注意：如果权限已经存在，这步可以跳过

-- 插入权限说明（可选）
-- INSERT INTO sys_permissions (name, code, description) VALUES
-- ('查看标签路由', 'system:tag:view', '查看标签和路由规则'),
-- ('编辑标签路由', 'system:tag:edit', '创建、编辑、删除标签和路由规则');
