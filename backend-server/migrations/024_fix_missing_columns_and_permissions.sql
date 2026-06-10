-- ============================================================
-- 迁移: 024
-- 描述: 修复缺失的表列和菜单权限
-- 作者: Claude Code
-- 日期: 2026-06-10
-- ============================================================

-- 1. 补充 rate_limit_rules 缺失的 4 列
ALTER TABLE sys_rate_limit_rules
    ADD COLUMN IF NOT EXISTS `cooldown` INT NOT NULL DEFAULT 0 COMMENT '冷却时间（秒），触发限流后需等待多久恢复' AFTER `burst`,
    ADD COLUMN IF NOT EXISTS `block_duration` INT NOT NULL DEFAULT 0 COMMENT '封禁时长（秒），超过触发次数后封禁 IP' AFTER `cooldown`,
    ADD COLUMN IF NOT EXISTS `max_violations` INT NOT NULL DEFAULT 0 COMMENT '最大违规次数，超过后触发封禁（0=不封禁）' AFTER `block_duration`,
    ADD COLUMN IF NOT EXISTS `violation_score` INT NOT NULL DEFAULT 0 COMMENT '违规风险分，触发限流时累加到风险评分系统（0=不累加）' AFTER `max_violations`;

-- 2. 补充缺失的菜单按钮权限: system:task:view, system:task:delete
SET @task_menu_id = (SELECT id FROM sys_menus WHERE name = 'ScheduledTask' AND type = 'menu' LIMIT 1);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @task_menu_id, 'TaskView', '', '', 'button', 1, 'system:task:view', '', '{"title":"任务查看"}', 0
FROM DUAL WHERE @task_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'system:task:view');

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @task_menu_id, 'TaskDelete', '', '', 'button', 1, 'system:task:delete', '', '{"title":"任务删除"}', 0
FROM DUAL WHERE @task_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'system:task:delete');

-- 3. 补充缺失的菜单按钮权限: system:captcha:test
SET @captcha_menu_id = (SELECT id FROM sys_menus WHERE name = 'Captcha' AND type = 'menu' LIMIT 1);
SET @security_menu_id = (SELECT id FROM sys_menus WHERE name = 'Security' AND type = 'catalog' LIMIT 1);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT COALESCE(@captcha_menu_id, @security_menu_id), 'CaptchaTest', '', '', 'button', 1, 'system:captcha:test', '', '{"title":"验证码测试"}', 0
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'system:captcha:test');

-- 4. 给 admin 角色添加缺失的权限
UPDATE sys_roles SET permissions = JSON_ARRAY_APPEND(
    permissions, '$', 'system:task:view'
) WHERE id = 2 AND NOT JSON_CONTAINS(IFNULL(permissions, '[]'), '"system:task:view"');

UPDATE sys_roles SET permissions = JSON_ARRAY_APPEND(
    permissions, '$', 'system:task:delete'
) WHERE id = 2 AND NOT JSON_CONTAINS(IFNULL(permissions, '[]'), '"system:task:delete"');

UPDATE sys_roles SET permissions = JSON_ARRAY_APPEND(
    permissions, '$', 'system:captcha:test'
) WHERE id = 2 AND NOT JSON_CONTAINS(IFNULL(permissions, '[]'), '"system:captcha:test"');
