-- =============================================
-- 初始化菜单数据
-- 表名: sys_menus
-- 说明: 包含 Dashboard 和系统管理模块的基础菜单
-- 注意: 使用 LAST_INSERT_ID() 引用父级ID，避免硬编码
-- =============================================

USE backend_db;

-- -------------------------------------------
-- 一级目录: 概览
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(0, 'Dashboard', '/dashboard', '', 'catalog', 1, '', 'lucide:layout-dashboard', '{"order":1,"title":"概览"}');

SET @dashboard_id = LAST_INSERT_ID();

-- 概览子菜单
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@dashboard_id, 'Analytics', '/analytics', '/dashboard/analytics/index', 'menu', 1, '', 'lucide:area-chart', '{"order":1,"title":"分析页","affixTab":true}'),
(@dashboard_id, 'Workspace', '/workspace', '/dashboard/workspace/index', 'menu', 1, '', 'lucide:monitor', '{"order":2,"title":"工作台"}');

-- -------------------------------------------
-- 一级目录: 系统管理
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(0, 'System', '/system', '', 'catalog', 1, '', 'lucide:settings', '{"order":2,"title":"系统管理"}');

SET @system_id = LAST_INSERT_ID();

-- 系统管理子菜单
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@system_id, 'SystemUser', '/system/user', '/system/user/list', 'menu', 1, 'system:user:list', 'lucide:users', '{"order":1,"title":"用户管理"}');
SET @user_menu_id = LAST_INSERT_ID();

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@system_id, 'SystemRole', '/system/role', '/system/role/list', 'menu', 1, 'system:role:list', 'lucide:shield', '{"order":2,"title":"角色管理"}');
SET @role_menu_id = LAST_INSERT_ID();

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@system_id, 'SystemMenu', '/system/menu', '/system/menu/list', 'menu', 1, 'system:menu:list', 'lucide:list', '{"order":3,"title":"菜单管理"}');
SET @menu_menu_id = LAST_INSERT_ID();

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@system_id, 'SystemGroup', '/system/group', '/system/group/list', 'menu', 1, 'system:group:list', 'lucide:boxes', '{"order":4,"title":"分组管理"}');
SET @group_menu_id = LAST_INSERT_ID();

-- -------------------------------------------
-- 按钮权限: 用户管理
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@user_menu_id, 'SystemUserView',    '', '', 'button', 1, 'system:user:view',    '', '{"title":"查看用户"}'),
(@user_menu_id, 'SystemUserAdd',     '', '', 'button', 1, 'system:user:add',     '', '{"title":"添加用户"}'),
(@user_menu_id, 'SystemUserEdit',    '', '', 'button', 1, 'system:user:edit',    '', '{"title":"编辑用户"}'),
(@user_menu_id, 'SystemUserDelete',  '', '', 'button', 1, 'system:user:delete',  '', '{"title":"删除用户"}');

-- -------------------------------------------
-- 按钮权限: 角色管理
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@role_menu_id, 'SystemRoleView',    '', '', 'button', 1, 'system:role:view',    '', '{"title":"查看角色"}'),
(@role_menu_id, 'SystemRoleAdd',     '', '', 'button', 1, 'system:role:add',     '', '{"title":"添加角色"}'),
(@role_menu_id, 'SystemRoleEdit',    '', '', 'button', 1, 'system:role:edit',    '', '{"title":"编辑角色"}'),
(@role_menu_id, 'SystemRoleDelete',  '', '', 'button', 1, 'system:role:delete',  '', '{"title":"删除角色"}');

-- -------------------------------------------
-- 按钮权限: 菜单管理
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@menu_menu_id, 'SystemMenuView',    '', '', 'button', 1, 'system:menu:view',    '', '{"title":"查看菜单"}'),
(@menu_menu_id, 'SystemMenuAdd',     '', '', 'button', 1, 'system:menu:add',     '', '{"title":"添加菜单"}'),
(@menu_menu_id, 'SystemMenuEdit',    '', '', 'button', 1, 'system:menu:edit',    '', '{"title":"编辑菜单"}'),
(@menu_menu_id, 'SystemMenuDelete',  '', '', 'button', 1, 'system:menu:delete',  '', '{"title":"删除菜单"}');

-- -------------------------------------------
-- 按钮权限: 分组管理
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@group_menu_id, 'SystemGroupView',   '', '', 'button', 1, 'system:group:view',   '', '{"title":"查看分组"}'),
(@group_menu_id, 'SystemGroupAdd',    '', '', 'button', 1, 'system:group:add',    '', '{"title":"添加分组"}'),
(@group_menu_id, 'SystemGroupEdit',   '', '', 'button', 1, 'system:group:edit',   '', '{"title":"编辑分组"}'),
(@group_menu_id, 'SystemGroupDelete', '', '', 'button', 1, 'system:group:delete', '', '{"title":"删除分组"}');

-- -------------------------------------------
-- 一级目录: 用户认证
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(0, 'UserAuth', '/user-auth', '', 'catalog', 1, '', 'lucide:shield-check', '{"order":3,"title":"用户认证"}');

SET @userauth_id = LAST_INSERT_ID();

-- 用户认证子菜单
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@userauth_id, 'SecurityLog',     '/user-auth/security-log', '/user-auth/security-log/list',  'menu', 1, 'system:security:view',  'lucide:file-text',         '{"order":1,"title":"安全日志"}'),
(@userauth_id, 'LoginDevice',     '/user-auth/device',       '/user-auth/device/list',        'menu', 1, 'system:device:view',    'lucide:smartphone',        '{"order":2,"title":"登录设备"}'),
(@userauth_id, 'OAuthBinding',    '/user-auth/oauth',        '/user-auth/oauth/list',         'menu', 1, 'system:oauth:view',     'lucide:link',              '{"order":3,"title":"OAuth绑定"}'),
(@userauth_id, 'RealName',        '/user-auth/real-name',    '/user-auth/real-name/list',     'menu', 1, 'system:realname:view',  'lucide:user-check',        '{"order":4,"title":"实名认证"}'),
(@userauth_id, 'Privacy',         '/user-auth/privacy',      '/user-auth/privacy/index',      'menu', 1, 'system:privacy:view',   'lucide:eye-off',           '{"order":5,"title":"隐私设置"}'),
(@userauth_id, 'RoleApplication', '/user-auth/role-app',     '/user-auth/role-app/list',      'menu', 1, 'system:roleapp:view',   'lucide:clipboard-check',   '{"order":6,"title":"角色申请"}');

SET @security_menu_id = (SELECT id FROM sys_menus WHERE name = 'SecurityLog' AND deleted_at IS NULL LIMIT 1);
SET @device_menu_id = (SELECT id FROM sys_menus WHERE name = 'LoginDevice' AND deleted_at IS NULL LIMIT 1);
SET @realname_menu_id = (SELECT id FROM sys_menus WHERE name = 'RealName' AND deleted_at IS NULL LIMIT 1);
SET @roleapp_menu_id = (SELECT id FROM sys_menus WHERE name = 'RoleApplication' AND deleted_at IS NULL LIMIT 1);

-- 按钮权限: 安全日志
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@security_menu_id, 'SecurityLogView', '', '', 'button', 1, 'system:security:view', '', '{"title":"查看安全日志"}');

-- 按钮权限: 登录设备
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@device_menu_id, 'LoginDeviceView',   '', '', 'button', 1, 'system:device:view',   '', '{"title":"查看设备"}'),
(@device_menu_id, 'LoginDeviceDelete', '', '', 'button', 1, 'system:device:delete', '', '{"title":"踢出设备"}');

-- 按钮权限: 实名认证
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@realname_menu_id, 'RealNameReview', '', '', 'button', 1, 'system:realname:review', '', '{"title":"审核实名"}');

-- 按钮权限: 角色申请
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@roleapp_menu_id, 'RoleAppReview', '', '', 'button', 1, 'system:roleapp:review', '', '{"title":"审核角色申请"}');
