-- 题库管理模块 - 题目管理菜单和权限

-- 题目管理菜单（题库管理子菜单，ID从250开始）
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(250, 220, 'QuestionList', '/question/list', 'question/list/index', 'menu', 1, 'question:view', 'mdi:file-question-outline', 0, '{"title":"题目管理","order":0}', NOW(), NOW());

-- 题目管理按钮权限
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(251, 250, 'QuestionCreate', '', '', 'button', 1, 'question:create', '', 1, '{"title":"创建题目"}', NOW(), NOW()),
(252, 250, 'QuestionEdit', '', '', 'button', 1, 'question:edit', '', 2, '{"title":"编辑题目"}', NOW(), NOW()),
(253, 250, 'QuestionDelete', '', '', 'button', 1, 'question:delete', '', 3, '{"title":"删除题目"}', NOW(), NOW()),
(254, 250, 'QuestionPublish', '', '', 'button', 1, 'question:publish', '', 4, '{"title":"发布题目"}', NOW(), NOW()),
(255, 250, 'QuestionArchive', '', '', 'button', 1, 'question:archive', '', 5, '{"title":"下架题目"}', NOW(), NOW()),
(256, 250, 'QuestionAuditSubmit', '', '', 'button', 1, 'question:audit:submit', '', 6, '{"title":"提交审核"}', NOW(), NOW()),
(257, 250, 'QuestionAuditApprove', '', '', 'button', 1, 'question:audit:approve', '', 7, '{"title":"审核通过"}', NOW(), NOW()),
(258, 250, 'QuestionAuditReject', '', '', 'button', 1, 'question:audit:reject', '', 8, '{"title":"审核驳回"}', NOW(), NOW());

-- 给超级管理员和管理员角色添加题目管理权限
-- 超级管理员(id=1)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:create') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:create"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:edit') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:edit"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:delete') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:delete"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:publish') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:publish"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:archive') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:archive"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:audit:submit') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:audit:submit"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:audit:approve') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:audit:approve"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:audit:reject') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:audit:reject"');

-- 管理员(id=2)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:create') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:create"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:edit') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:edit"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:delete') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:delete"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:publish') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:publish"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:archive') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:archive"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:audit:submit') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:audit:submit"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:audit:approve') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:audit:approve"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:audit:reject') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:audit:reject"');
