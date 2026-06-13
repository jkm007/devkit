-- 题库管理模块 - 分类科目菜单和权限

-- 题库管理一级菜单（与系统管理平级，pid=0）
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(220, 0, 'QuestionBank', '/question', '', 'catalog', 1, 'question:view', 'mdi:book-open-variant', 14, '{"title":"题库管理","order":14}', NOW(), NOW());

-- 分类科目菜单（题库管理子菜单）
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(221, 220, 'QuestionCategory', '/question/category', 'question/category/index', 'menu', 1, 'question:category', 'mdi:folder-tree', 1, '{"title":"分类科目","order":1}', NOW(), NOW());

-- 分类科目按钮权限
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(222, 221, 'QuestionCategoryAdd', '', '', 'button', 1, 'question:category:add', '', 1, '{"title":"新增分类"}', NOW(), NOW()),
(223, 221, 'QuestionCategoryEdit', '', '', 'button', 1, 'question:category:edit', '', 2, '{"title":"编辑分类"}', NOW(), NOW()),
(224, 221, 'QuestionCategoryDelete', '', '', 'button', 1, 'question:category:delete', '', 3, '{"title":"删除分类"}', NOW(), NOW()),
(225, 221, 'QuestionExamManage', '', '', 'button', 1, 'question:exam:manage', '', 4, '{"title":"管理考试"}', NOW(), NOW()),
(226, 221, 'QuestionSubjectManage', '', '', 'button', 1, 'question:subject:manage', '', 5, '{"title":"管理科目"}', NOW(), NOW());

-- 给超级管理员和管理员角色添加题库分类权限
-- 超级管理员(id=1)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:view') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:category') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:category"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:category:add') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:category:add"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:category:edit') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:category:edit"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:category:delete') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:category:delete"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:exam:manage') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:exam:manage"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:subject:manage') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:subject:manage"');

-- 管理员(id=2)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:view') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:category') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:category"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:category:add') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:category:add"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:category:edit') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:category:edit"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:category:delete') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:category:delete"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:exam:manage') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:exam:manage"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:subject:manage') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:subject:manage"');
