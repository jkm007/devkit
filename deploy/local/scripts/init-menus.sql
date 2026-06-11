-- ============================================
-- 初始化菜单数据
-- 基于前端路由和后端权限码
-- ============================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- 一级目录 (pid=0)
-- ============================================
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(1, 0, 'Dashboard', '/dashboard', 'BasicLayout', 'catalog', 1, '', 'lucide:layout-dashboard', 1, '{"title":"概览","icon":"lucide:layout-dashboard","order":1}', NOW(), NOW()),
(4, 0, 'System', '/system', 'BasicLayout', 'catalog', 1, '', 'ion:settings-outline', 9997, '{"title":"系统管理","icon":"ion:settings-outline","order":9997}', NOW(), NOW()),
(25, 0, 'Account', '/account', 'BasicLayout', 'catalog', 1, '', 'lucide:user-circle', 3, '{"title":"个人中心","icon":"lucide:user-circle","hideInMenu":true,"order":3}', NOW(), NOW()),
(40, 0, 'File', '/file', 'BasicLayout', 'catalog', 1, '', 'lucide:folder', 9996, '{"title":"文件管理","icon":"lucide:folder","order":9996}', NOW(), NOW()),
(71, 0, 'Storage', '/storage', 'BasicLayout', 'catalog', 1, '', 'mdi:cloud-sync', 9995, '{"title":"存储管理","icon":"mdi:cloud-sync","order":9995}', NOW(), NOW()),
(72, 0, 'Security', '/security', 'BasicLayout', 'catalog', 1, '', 'mdi:shield-lock', 9994, '{"title":"安全管理","icon":"mdi:shield-lock","order":9994}', NOW(), NOW());

-- ============================================
-- Dashboard 子菜单
-- ============================================
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(2, 1, 'Analytics', '/analytics', 'dashboard/analytics/index', 'menu', 1, '', 'lucide:area-chart', 0, '{"title":"分析页","icon":"lucide:area-chart","affixTab":true}', NOW(), NOW()),
(3, 1, 'Workspace', '/workspace', 'dashboard/workspace/index', 'menu', 1, '', 'carbon:workspace', 0, '{"title":"工作台","icon":"carbon:workspace","order":2}', NOW(), NOW()),
(134, 1, 'UserHome', '/user-home', 'dashboard/user-home/index', 'menu', 1, '', 'lucide:home', 100, '{"title":"用户首页","icon":"lucide:home"}', NOW(), NOW());

-- ============================================
-- System 子菜单
-- ============================================
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(5, 4, 'SystemUser', '/system/user', 'system/user/list', 'menu', 1, 'system:user:list', 'mdi:user', 1, '{"title":"用户管理","icon":"mdi:user","order":1}', NOW(), NOW()),
(6, 4, 'SystemRole', '/system/role', 'system/role/list', 'menu', 1, 'system:role:list', 'mdi:account-group', 2, '{"title":"角色管理","icon":"mdi:account-group","order":2}', NOW(), NOW()),
(7, 4, 'SystemMenu', '/system/menu', 'system/menu/list', 'menu', 1, 'system:menu:list', 'mdi:menu', 3, '{"title":"菜单管理","icon":"mdi:menu","order":3}', NOW(), NOW()),
(8, 4, 'SystemGroup', '/system/group', 'system/group/list', 'menu', 1, 'system:group:list', 'charm:organisation', 4, '{"title":"分组管理","icon":"charm:organisation","order":4}', NOW(), NOW()),
(33, 4, 'SystemSettings', '/system/settings', 'system/settings/index', 'menu', 1, 'system:setting:list', 'lucide:sliders-horizontal', 5, '{"title":"系统设置","icon":"lucide:sliders-horizontal","order":5}', NOW(), NOW()),
(54, 4, 'SystemTag', '/system/tag', 'system/tag/list', 'menu', 1, 'storage:bucket:view', 'mdi:tag-multiple', 6, '{"title":"标签管理","icon":"mdi:tag-multiple","order":6}', NOW(), NOW()),
(131, 4, 'SystemScheduledTask', '/system/scheduled-task', 'system/scheduled-task/index', 'menu', 1, 'system:task:list', 'lucide:clock', 7, '{"title":"定时任务","icon":"lucide:clock","order":7}', NOW(), NOW());

-- ============================================
-- System 按钮权限
-- ============================================
-- 用户管理按钮
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(9, 5, 'SystemUserAdd', '', '', 'button', 1, 'system:user:add', '', 0, '{}', NOW(), NOW()),
(10, 5, 'SystemUserEdit', '', '', 'button', 1, 'system:user:edit', '', 0, '{}', NOW(), NOW()),
(11, 5, 'SystemUserDelete', '', '', 'button', 1, 'system:user:delete', '', 0, '{}', NOW(), NOW()),
(21, 5, 'SystemUserView', '', '', 'button', 1, 'system:user:view', '', 0, '{}', NOW(), NOW());

-- 角色管理按钮
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(12, 6, 'SystemRoleAdd', '', '', 'button', 1, 'system:role:add', '', 0, '{}', NOW(), NOW()),
(13, 6, 'SystemRoleEdit', '', '', 'button', 1, 'system:role:edit', '', 0, '{}', NOW(), NOW()),
(14, 6, 'SystemRoleDelete', '', '', 'button', 1, 'system:role:delete', '', 0, '{}', NOW(), NOW()),
(22, 6, 'SystemRoleView', '', '', 'button', 1, 'system:role:view', '', 0, '{}', NOW(), NOW());

-- 菜单管理按钮
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(15, 7, 'SystemMenuAdd', '', '', 'button', 1, 'system:menu:add', '', 0, '{}', NOW(), NOW()),
(16, 7, 'SystemMenuEdit', '', '', 'button', 1, 'system:menu:edit', '', 0, '{}', NOW(), NOW()),
(17, 7, 'SystemMenuDelete', '', '', 'button', 1, 'system:menu:delete', '', 0, '{}', NOW(), NOW()),
(23, 7, 'SystemMenuView', '', '', 'button', 1, 'system:menu:view', '', 0, '{}', NOW(), NOW());

-- 分组管理按钮
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(18, 8, 'SystemGroupAdd', '', '', 'button', 1, 'system:group:add', '', 0, '{}', NOW(), NOW()),
(19, 8, 'SystemGroupEdit', '', '', 'button', 1, 'system:group:edit', '', 0, '{}', NOW(), NOW()),
(20, 8, 'SystemGroupDelete', '', '', 'button', 1, 'system:group:delete', '', 0, '{}', NOW(), NOW()),
(24, 8, 'SystemGroupView', '', '', 'button', 1, 'system:group:view', '', 0, '{}', NOW(), NOW());

-- 系统设置按钮
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(34, 33, 'SystemSettingView', '', '', 'button', 1, 'system:setting:list', '', 0, '{}', NOW(), NOW()),
(35, 33, 'SystemSettingEdit', '', '', 'button', 1, 'system:setting:edit', '', 0, '{}', NOW(), NOW());

-- 定时任务按钮
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(132, 131, 'SystemScheduledTaskEdit', '', '', 'button', 1, 'system:task:edit', '', 0, '{}', NOW(), NOW()),
(133, 131, 'SystemScheduledTaskRun', '', '', 'button', 1, 'system:task:run', '', 0, '{}', NOW(), NOW()),
(135, 131, 'SystemScheduledTaskView', '', '', 'button', 1, 'system:task:view', '', 0, '{}', NOW(), NOW()),
(136, 131, 'SystemScheduledTaskDelete', '', '', 'button', 1, 'system:task:delete', '', 0, '{}', NOW(), NOW());

-- ============================================
-- Account 子菜单
-- ============================================
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(26, 25, 'AccountIndex', '/account/index', 'account/index', 'menu', 1, '', 'lucide:user-circle', 1, '{"title":"个人中心","icon":"lucide:user-circle","hideInMenu":true}', NOW(), NOW());

-- ============================================
-- File 子菜单
-- ============================================
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(41, 40, 'FileList', '/file/list', 'file/list/index', 'menu', 1, 'file:view:own', 'lucide:files', 1, '{"title":"文件列表","icon":"lucide:files","order":1}', NOW(), NOW()),
(42, 40, 'FileShare', '/file/share', 'file/share/index', 'menu', 1, 'share:view:own', 'lucide:share-2', 2, '{"title":"分享管理","icon":"lucide:share-2","order":2}', NOW(), NOW()),
(127, 40, 'FileRecycle', '/file/recycle', 'file/recycle/index', 'menu', 1, 'file:recycle:list', 'lucide:trash-2', 3, '{"title":"回收站","icon":"lucide:trash-2","order":3}', NOW(), NOW());

-- File 按钮权限
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(43, 41, 'FileListView', '', '', 'button', 1, 'file:view:own', '', 0, '{}', NOW(), NOW()),
(44, 41, 'FileListViewAll', '', '', 'button', 1, 'file:view:all', '', 0, '{}', NOW(), NOW()),
(45, 41, 'FileUpload', '', '', 'button', 1, 'file:upload', '', 0, '{}', NOW(), NOW()),
(46, 41, 'FileDownload', '', '', 'button', 1, 'file:download', '', 0, '{}', NOW(), NOW()),
(47, 41, 'FileDelete', '', '', 'button', 1, 'file:delete', '', 0, '{}', NOW(), NOW()),
(48, 41, 'FileShareCreate', '', '', 'button', 1, 'file:share', '', 0, '{}', NOW(), NOW()),
(49, 41, 'FileManage', '', '', 'button', 1, 'file:manage', '', 0, '{}', NOW(), NOW()),
(50, 42, 'FileShareViewOwn', '', '', 'button', 1, 'share:view:own', '', 0, '{}', NOW(), NOW()),
(128, 127, 'FileRecycleRestore', '', '', 'button', 1, 'file:recycle:restore', '', 0, '{}', NOW(), NOW()),
(129, 127, 'FileRecycleDelete', '', '', 'button', 1, 'file:recycle:delete', '', 0, '{}', NOW(), NOW()),
(130, 127, 'FileRecycleEmpty', '', '', 'button', 1, 'file:recycle:empty', '', 0, '{}', NOW(), NOW());

-- ============================================
-- Storage 子菜单
-- ============================================
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(93, 71, 'StorageBucket', '/storage/storage-bucket', 'storage/storage-bucket/index', 'menu', 1, 'storage:bucket:list', 'mdi:database', 1, '{"title":"存储桶管理","icon":"mdi:database","order":1}', NOW(), NOW()),
(140, 71, 'StorageConfig', '/storage/storage-config', 'storage/storage-config/index', 'menu', 1, 'storage:config:view', 'mdi:server-network', 2, '{"title":"存储配置","icon":"mdi:server-network","order":2}', NOW(), NOW()),
(55, 71, 'StorageTagRouting', '/storage/tag-routing', 'storage/tag-routing/index', 'menu', 1, 'storage:bucket:view', 'lucide:git-branch', 3, '{"title":"标签路由","icon":"lucide:git-branch","order":3}', NOW(), NOW()),
(141, 71, 'StorageSettings', '/storage/storage-settings', 'storage/storage-settings/index', 'menu', 1, 'storage:config:view', 'lucide:settings', 4, '{"title":"存储设置","icon":"lucide:settings","order":4}', NOW(), NOW()),
(70, 71, 'StorageFileTypeRule', '/storage/file-type-rule', 'storage/file-type-rule/index', 'menu', 1, 'storage:file-type:list', 'mdi:file-check', 5, '{"title":"文件类型规则","icon":"mdi:file-check","order":5}', NOW(), NOW());

-- Storage 按钮权限
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(97, 93, 'StorageManageBucketView', '', '', 'button', 1, 'storage:bucket:view', '', 0, '{}', NOW(), NOW()),
(98, 93, 'StorageManageBucketEdit', '', '', 'button', 1, 'storage:bucket:edit', '', 0, '{}', NOW(), NOW()),
(99, 93, 'StorageManageBucketDelete', '', '', 'button', 1, 'storage:bucket:delete', '', 0, '{}', NOW(), NOW()),
(84, 55, 'TagView', '', '', 'button', 1, 'storage:bucket:view', '', 0, '{}', NOW(), NOW()),
(85, 55, 'TagEdit', '', '', 'button', 1, 'storage:bucket:edit', '', 0, '{}', NOW(), NOW()),
(86, 55, 'TagDelete', '', '', 'button', 1, 'storage:bucket:delete', '', 0, '{}', NOW(), NOW()),
(79, 70, 'SystemFileTypeRuleView', '', '', 'button', 1, 'storage:file-type:view', '', 0, '{}', NOW(), NOW()),
(80, 70, 'SystemFileTypeRuleEdit', '', '', 'button', 1, 'storage:file-type:edit', '', 0, '{}', NOW(), NOW()),
(81, 70, 'SystemFileTypeRuleDelete', '', '', 'button', 1, 'storage:file-type:delete', '', 0, '{}', NOW(), NOW());

-- ============================================
-- Security 子菜单
-- ============================================
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(90, 72, 'SecuritySettings', '/security/security-settings', 'security/security-settings/index', 'menu', 1, 'security:risk:list', 'lucide:settings', 1, '{"title":"安全设置","icon":"lucide:settings","order":1}', NOW(), NOW()),
(27, 72, 'SecurityRealName', '/security/real-name', 'security/real-name/index', 'menu', 1, 'security:realname:list', 'lucide:badge-check', 2, '{"title":"实名审核","icon":"lucide:badge-check","order":2}', NOW(), NOW()),
(31, 72, 'SecurityLog', '/security/security-log', 'security/security-log/index', 'menu', 1, 'security:log:list', 'lucide:shield-check', 3, '{"title":"安全日志","icon":"lucide:shield-check","order":3}', NOW(), NOW()),
(39, 72, 'SecurityRisk', '/security/risk', 'security/risk/index', 'menu', 1, 'security:risk:list', 'lucide:shield-alert', 4, '{"title":"风控管理","icon":"lucide:shield-alert","order":4}', NOW(), NOW());

-- Security 按钮权限
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(101, 90, 'RateLimitRuleView', '', '', 'button', 1, 'security:ratelimit:view', '', 0, '{}', NOW(), NOW()),
(102, 90, 'RateLimitRuleEdit', '', '', 'button', 1, 'security:ratelimit:edit', '', 0, '{}', NOW(), NOW()),
(103, 90, 'RateLimitRuleDelete', '', '', 'button', 1, 'security:ratelimit:delete', '', 0, '{}', NOW(), NOW()),
(28, 27, 'SystemRealNameView', '', '', 'button', 1, 'security:realname:view', '', 0, '{}', NOW(), NOW()),
(29, 27, 'SystemRealNameApprove', '', '', 'button', 1, 'security:realname:approve', '', 0, '{}', NOW(), NOW()),
(30, 27, 'SystemRealNameReject', '', '', 'button', 1, 'security:realname:reject', '', 0, '{}', NOW(), NOW()),
(32, 31, 'SystemSecurityLogView', '', '', 'button', 1, 'security:log:view', '', 0, '{}', NOW(), NOW()),
(82, 39, 'SystemRiskView', '', '', 'button', 1, 'security:risk:view', '', 0, '{}', NOW(), NOW()),
(83, 39, 'SystemRiskEdit', '', '', 'button', 1, 'security:risk:edit', '', 0, '{}', NOW(), NOW());

-- ============================================
-- 其他按钮
-- ============================================
INSERT INTO sys_menus (id, pid, name, path, component, type, status, auth_code, icon, sort, meta, created_at, updated_at) VALUES
(143, 1, 'CaptchaTest', '', '', 'button', 1, 'system:captcha:test', '', 0, '{}', NOW(), NOW());

-- ============================================
-- 初始化角色
-- ============================================
-- super 角色（拥有所有权限）
INSERT INTO sys_roles (id, name, status, permissions, remark, created_at, updated_at) VALUES
(1, 'super', 1, '["system:user:list","system:user:view","system:user:add","system:user:edit","system:user:delete","system:role:list","system:role:view","system:role:add","system:role:edit","system:role:delete","system:menu:list","system:menu:view","system:menu:add","system:menu:edit","system:menu:delete","system:group:list","system:group:view","system:group:add","system:group:edit","system:group:delete","system:setting:list","system:setting:edit","system:task:list","system:task:view","system:task:edit","system:task:delete","system:task:run","system:captcha:test","storage:bucket:list","storage:bucket:view","storage:bucket:edit","storage:bucket:delete","storage:config:view","storage:config:edit","storage:config:delete","storage:file-type:list","storage:file-type:view","storage:file-type:edit","storage:file-type:delete","security:risk:list","security:risk:view","security:risk:edit","security:ratelimit:view","security:ratelimit:edit","security:ratelimit:delete","security:realname:list","security:realname:view","security:realname:approve","security:realname:reject","security:log:list","security:log:view","file:view:own","file:view:all","file:upload","file:download","file:delete","file:share","file:manage","share:view:own","share:view:all","share:delete","share:manage","file:recycle:list","file:recycle:restore","file:recycle:delete","file:recycle:empty"]', '超级管理员', NOW(), NOW());

-- admin 角色
INSERT INTO sys_roles (id, name, status, permissions, remark, created_at, updated_at) VALUES
(2, 'admin', 1, '["system:user:list","system:user:view","system:user:add","system:user:edit","system:user:delete","system:role:list","system:role:view","system:role:add","system:role:edit","system:role:delete","system:menu:list","system:menu:view","system:menu:add","system:menu:edit","system:menu:delete","system:group:list","system:group:view","system:group:add","system:group:edit","system:group:delete","system:setting:list","system:setting:edit","system:task:list","system:task:view","system:task:edit","system:task:delete","system:task:run","system:captcha:test","file:view:own","file:view:all","file:upload","file:download","file:delete","file:share","file:manage","share:view:own","share:view:all","share:delete","share:manage","file:recycle:list","file:recycle:restore","file:recycle:delete","file:recycle:empty"]', '管理员', NOW(), NOW());

-- user 角色（普通用户）
INSERT INTO sys_roles (id, name, status, permissions, remark, created_at, updated_at) VALUES
(3, 'user', 1, '["file:view:own","file:upload","file:download","file:delete","file:manage","file:share","share:view:own","file:recycle:list","file:recycle:restore","file:recycle:delete","file:recycle:empty"]', '普通用户', NOW(), NOW());

-- test-role 角色
INSERT INTO sys_roles (id, name, status, permissions, remark, created_at, updated_at) VALUES
(4, 'test-role', 1, '["system:user:list","system:user:view","system:user:add","system:user:edit","system:user:delete","system:group:list","system:group:view","file:view:own","file:upload","file:download","file:delete","file:share","file:manage","share:view:own","share:delete","share:manage"]', '测试角色', NOW(), NOW());

-- ============================================
-- 用户角色关联
-- ============================================
INSERT INTO sys_user_roles (user_id, role_id) VALUES
(1, 1),
(2, 2),
(3, 3),
(8, 3);

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- 验证数据
-- ============================================
SELECT '=== 菜单统计 ===' as info;
SELECT type, COUNT(*) as cnt FROM sys_menus WHERE deleted_at IS NULL GROUP BY type;

SELECT '=== 一级目录 ===' as info;
SELECT id, name, path FROM sys_menus WHERE pid = 0 AND deleted_at IS NULL ORDER BY sort, id;

SELECT '=== 角色统计 ===' as info;
SELECT id, name, JSON_LENGTH(permissions) as perm_count FROM sys_roles WHERE deleted_at IS NULL;
