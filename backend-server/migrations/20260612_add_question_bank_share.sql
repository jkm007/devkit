-- 题库管理模块 - 题目分享菜单和权限

-- 题目分享菜单（题库管理子菜单，ID从290开始）
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(290, 220, 'QuestionShare', '/question/share', 'question/share/index', 'menu', 1, 'question:share:view', 'mdi:share-variant', 6, '{"title":"题目分享","order":6}', NOW(), NOW());

-- 题目分享按钮权限
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(291, 290, 'QuestionShareCreate', '', '', 'button', 1, 'question:share:create', '', 1, '{"title":"创建分享"}', NOW(), NOW()),
(292, 290, 'QuestionShareEdit', '', '', 'button', 1, 'question:share:edit', '', 2, '{"title":"编辑分享"}', NOW(), NOW()),
(293, 290, 'QuestionShareDisable', '', '', 'button', 1, 'question:share:disable', '', 3, '{"title":"禁用分享"}', NOW(), NOW()),
(294, 290, 'QuestionShareEnable', '', '', 'button', 1, 'question:share:enable', '', 4, '{"title":"启用分享"}', NOW(), NOW()),
(295, 290, 'QuestionShareDelete', '', '', 'button', 1, 'question:share:delete', '', 5, '{"title":"删除分享"}', NOW(), NOW());

-- 给超级管理员和管理员角色添加题目分享权限
-- 超级管理员(id=1)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:view') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:share:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:create') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:share:create"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:edit') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:share:edit"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:disable') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:share:disable"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:enable') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:share:enable"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:delete') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:share:delete"');

-- 管理员(id=2)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:view') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:share:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:create') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:share:create"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:edit') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:share:edit"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:disable') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:share:disable"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:enable') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:share:enable"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:share:delete') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:share:delete"');
