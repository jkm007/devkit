-- 题库管理模块 - 题目导入菜单和权限

-- 题目导入菜单（题库管理子菜单，ID从260开始）
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(260, 220, 'QuestionImport', '/question/import', 'question/import/index', 'menu', 1, 'question:import', 'mdi:import', 2, '{"title":"题目导入","order":2}', NOW(), NOW());

-- 题目导入按钮权限
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(261, 260, 'QuestionImportParse', '', '', 'button', 1, 'question:import:parse', '', 1, '{"title":"解析导入"}', NOW(), NOW()),
(262, 260, 'QuestionImportConfirm', '', '', 'button', 1, 'question:import:confirm', '', 2, '{"title":"确认入库"}', NOW(), NOW()),
(263, 260, 'QuestionImportPublish', '', '', 'button', 1, 'question:import:publish', '', 3, '{"title":"批量发布"}', NOW(), NOW()),
(264, 260, 'QuestionImportDelete', '', '', 'button', 1, 'question:import:delete', '', 4, '{"title":"删除导入任务"}', NOW(), NOW());

-- 给超级管理员和管理员角色添加题目导入权限
-- 超级管理员(id=1)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:import') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:import"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:import:parse') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:import:parse"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:import:confirm') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:import:confirm"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:import:publish') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:import:publish"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:import:delete') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:import:delete"');

-- 管理员(id=2)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:import') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:import"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:import:parse') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:import:parse"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:import:confirm') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:import:confirm"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:import:publish') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:import:publish"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:import:delete') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:import:delete"');
