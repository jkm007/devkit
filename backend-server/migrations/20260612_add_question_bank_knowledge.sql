-- 题库管理模块 - 知识考点菜单和权限

-- 知识考点菜单（题库管理子菜单，ID从230开始）
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(230, 220, 'QuestionKnowledge', '/question/knowledge', 'question/knowledge/index', 'menu', 1, 'question:knowledge', 'mdi:lightbulb-outline', 2, '{"title":"知识考点","order":2}', NOW(), NOW());

-- 知识考点按钮权限
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(231, 230, 'QuestionKnowledgeAdd', '', '', 'button', 1, 'question:knowledge:add', '', 1, '{"title":"新增知识点"}', NOW(), NOW()),
(232, 230, 'QuestionKnowledgeEdit', '', '', 'button', 1, 'question:knowledge:edit', '', 2, '{"title":"编辑知识点"}', NOW(), NOW()),
(233, 230, 'QuestionKnowledgeDelete', '', '', 'button', 1, 'question:knowledge:delete', '', 3, '{"title":"删除知识点"}', NOW(), NOW());

-- 给超级管理员和管理员角色添加知识考点权限
-- 超级管理员(id=1)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:knowledge') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:knowledge"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:knowledge:add') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:knowledge:add"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:knowledge:edit') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:knowledge:edit"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:knowledge:delete') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"question:knowledge:delete"');

-- 管理员(id=2)
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:knowledge') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:knowledge"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:knowledge:add') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:knowledge:add"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:knowledge:edit') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:knowledge:edit"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'question:knowledge:delete') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"question:knowledge:delete"');
