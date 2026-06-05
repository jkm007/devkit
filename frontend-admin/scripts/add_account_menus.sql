-- =============================================
-- 新增菜单：个人中心 + 系统管理（实名审核、安全日志）
-- 在 init_menu.sql 基础上追加
-- =============================================

USE backend_db;

-- -------------------------------------------
-- 一级目录: 个人中心
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(0, 'Account', '/account', '', 'catalog', 1, '', 'lucide:user-circle', '{"order":3,"title":"个人中心"}');

SET @account_id = LAST_INSERT_ID();

-- 个人中心子菜单
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@account_id, 'AccountIndex', '/account/index', '/account/index', 'menu', 1, '', 'lucide:user-circle', '{"order":1,"title":"个人中心"}');

-- -------------------------------------------
-- 系统管理下新增：实名审核
-- -------------------------------------------
-- 获取系统管理目录ID（假设已存在，ID=2）
SET @system_id = (SELECT id FROM sys_menus WHERE name = 'System' AND type = 'catalog' AND deleted_at IS NULL LIMIT 1);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@system_id, 'SystemRealName', '/system/real-name', '/system/real-name/list', 'menu', 1, 'system:realname:list', 'lucide:badge-check', '{"order":5,"title":"实名审核"}');
SET @realname_menu_id = LAST_INSERT_ID();

-- 实名审核按钮权限
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@realname_menu_id, 'SystemRealNameView',    '', '', 'button', 1, 'system:realname:view',    '', '{"title":"查看实名"}'),
(@realname_menu_id, 'SystemRealNameApprove', '', '', 'button', 1, 'system:realname:approve', '', '{"title":"审核通过"}'),
(@realname_menu_id, 'SystemRealNameReject',  '', '', 'button', 1, 'system:realname:reject',  '', '{"title":"审核拒绝"}');

-- -------------------------------------------
-- 系统管理下新增：安全日志
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@system_id, 'SystemSecurityLog', '/system/security-log', '/system/security-log/list', 'menu', 1, 'system:securitylog:list', 'lucide:shield-check', '{"order":6,"title":"安全日志"}');
SET @securitylog_menu_id = LAST_INSERT_ID();

-- 安全日志按钮权限
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@securitylog_menu_id, 'SystemSecurityLogView', '', '', 'button', 1, 'system:securitylog:view', '', '{"title":"查看安全日志"}');

-- -------------------------------------------
-- 给admin角色追加新权限（假设admin角色ID=2）
-- 注意：需要根据实际角色ID调整
-- -------------------------------------------
-- UPDATE sys_roles SET permissions = JSON_ARRAY_APPEND(
--   COALESCE permissions, '$',
--   'system:realname:view', 'system:realname:approve', 'system:realname:reject',
--   'system:securitylog:view'
-- ) WHERE id = 2;
