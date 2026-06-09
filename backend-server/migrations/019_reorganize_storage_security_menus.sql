-- ============================================================
-- 迁移: 019
-- 描述: 拆分存储和安全为独立顶级目录，更新权限码
-- 作者: Claude Code
-- 日期: 2026-06-09
-- ============================================================

-- -------------------------------------------
-- 1. 创建「存储管理」顶级目录 (pid=0)
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(0, 'Storage', '/storage', '', 'catalog', 1, NULL, 'mdi:cloud-sync', '{"order":9995,"title":"存储管理"}');

SET @storage_id = LAST_INSERT_ID();

-- -------------------------------------------
-- 2. 创建「安全管理」顶级目录 (pid=0)
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(0, 'Security', '/security', '', 'catalog', 1, NULL, 'mdi:shield-lock', '{"order":9994,"title":"安全管理"}');

SET @security_id = LAST_INSERT_ID();

-- -------------------------------------------
-- 3. 移动存储相关菜单到「存储管理」目录
-- -------------------------------------------
-- 存储桶管理 (id=56)
UPDATE sys_menus SET pid = @storage_id, auth_code = 'storage:bucket:list' WHERE id = 56;

-- 存储配置 (id=69)
UPDATE sys_menus SET pid = @storage_id, auth_code = 'storage:config:list' WHERE id = 69;

-- 文件类型规则 (id=70)
UPDATE sys_menus SET pid = @storage_id, auth_code = 'storage:file-type:list' WHERE id = 70;

-- -------------------------------------------
-- 4. 移动安全相关菜单到「安全管理」目录
-- -------------------------------------------
-- 实名审核 (id=27)
UPDATE sys_menus SET pid = @security_id, auth_code = 'security:realname:list' WHERE id = 27;

-- 安全日志 (id=31)
UPDATE sys_menus SET pid = @security_id, auth_code = 'security:log:list' WHERE id = 31;

-- 风险评分监控 (id=39)
UPDATE sys_menus SET pid = @security_id, auth_code = 'security:risk:list' WHERE id = 39;

-- -------------------------------------------
-- 5. 更新已有的按钮权限 auth_code
-- -------------------------------------------
-- 实名审核按钮 (pid=27): 已有 system:realname:* -> 改为 security:realname:*
UPDATE sys_menus SET auth_code = 'security:realname:view' WHERE id = 28;
UPDATE sys_menus SET auth_code = 'security:realname:approve' WHERE id = 29;
UPDATE sys_menus SET auth_code = 'security:realname:reject' WHERE id = 30;

-- 安全日志按钮 (pid=31): 已有 system:security:view -> 改为 security:log:view
UPDATE sys_menus SET auth_code = 'security:log:view' WHERE id = 32;

-- -------------------------------------------
-- 6. 添加存储桶管理按钮权限
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(56, 'SystemStorageBucketView', '', '', 'button', 1, 'storage:bucket:view', '', '{}'),
(56, 'SystemStorageBucketEdit', '', '', 'button', 1, 'storage:bucket:edit', '', '{}'),
(56, 'SystemStorageBucketDelete', '', '', 'button', 1, 'storage:bucket:delete', '', '{}');

-- -------------------------------------------
-- 7. 添加存储配置按钮权限
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(69, 'SystemStorageConfigView', '', '', 'button', 1, 'storage:config:view', '', '{}'),
(69, 'SystemStorageConfigEdit', '', '', 'button', 1, 'storage:config:edit', '', '{}'),
(69, 'SystemStorageConfigDelete', '', '', 'button', 1, 'storage:config:delete', '', '{}');

-- -------------------------------------------
-- 8. 添加文件类型规则按钮权限
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(70, 'SystemFileTypeRuleView', '', '', 'button', 1, 'storage:file-type:view', '', '{}'),
(70, 'SystemFileTypeRuleEdit', '', '', 'button', 1, 'storage:file-type:edit', '', '{}'),
(70, 'SystemFileTypeRuleDelete', '', '', 'button', 1, 'storage:file-type:delete', '', '{}');

-- -------------------------------------------
-- 9. 添加风险评分按钮权限
-- -------------------------------------------
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(39, 'SystemRiskView', '', '', 'button', 1, 'security:risk:view', '', '{}'),
(39, 'SystemRiskEdit', '', '', 'button', 1, 'security:risk:edit', '', '{}');
