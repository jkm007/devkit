-- 题库管理模块 - 来源标签菜单和权限

-- 来源标签菜单（题库管理子菜单，ID从240开始）
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(240, 220, 'QuestionSource', '/question/source', 'question/source/index', 'menu', 1, 'question:source:view', 'mdi:source-branch', 7, '{"title":"来源标签","order":7}', NOW(), NOW());

-- 来源标签按钮权限
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(241, 240, 'QuestionSourceManage', '', '', 'button', 1, 'question:source:manage', '', 1, '{"title":"管理来源"}', NOW(), NOW()),
(242, 240, 'QuestionTagView', '', '', 'button', 1, 'question:tag:view', '', 2, '{"title":"查看标签"}', NOW(), NOW()),
(243, 240, 'QuestionTagManage', '', '', 'button', 1, 'question:tag:manage', '', 3, '{"title":"管理标签"}', NOW(), NOW());

-- 给超级管理员和管理员角色添加来源标签权限
-- 超级管理员(id=1)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:source:view') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:source:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:source:manage') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:source:manage"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:tag:view') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:tag:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:tag:manage') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:tag:manage"');

-- 管理员(id=2)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:source:view') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:source:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:source:manage') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:source:manage"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:tag:view') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:tag:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:tag:manage') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:tag:manage"');
