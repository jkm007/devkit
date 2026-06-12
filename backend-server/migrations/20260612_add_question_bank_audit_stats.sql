-- 题库管理模块 - 审核发布和题库统计菜单

-- 审核发布菜单（题库管理子菜单，ID从270开始）
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(270, 220, 'QuestionAudit', '/question/audit', 'question/audit/index', 'menu', 1, 'question:audit:view', 'mdi:check-decagram', 5, '{"title":"审核发布","order":5}', NOW(), NOW());

-- 题库统计菜单（题库管理子菜单，ID从280开始）
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(280, 220, 'QuestionStatistics', '/question/statistics', 'question/statistics/index', 'menu', 1, 'question:statistics:view', 'mdi:chart-bar', 8, '{"title":"题库统计","order":8}', NOW(), NOW());

-- 给超级管理员和管理员角色添加权限
-- 超级管理员(id=1)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:audit:view') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:audit:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:statistics:view') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:statistics:view"');

-- 管理员(id=2)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:audit:view') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:audit:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:statistics:view') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:statistics:view"');
