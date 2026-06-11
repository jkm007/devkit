-- ============================================================
-- 迁移: 027
-- 描述: 修复菜单、角色、权限数据（全面对齐前后端代码）
-- 作者: Claude Code
-- 日期: 2026-06-11
-- ============================================================

USE backend_db;

-- ============================================================
-- 第一部分：清理混乱的菜单数据，重建正确结构
-- ============================================================

-- 1. 删除所有错误的/重复的菜单（保留基础结构）
-- 先标记要删除的菜单 ID
SET @delete_ids = (
    SELECT GROUP_CONCAT(id) FROM (
        SELECT id FROM sys_menus WHERE name IN (
            'Account', 'AccountIndex',
            'SystemSetting', 'SystemSettingView', 'SystemSettingEdit',
            'SystemTag', 'StorageManage', 'StorageManageConfigView',
            'StorageManageConfigEdit', 'StorageManageConfigDelete',
            'SecuritySettings', 'RateLimitRuleView', 'RateLimitRuleEdit',
            'RateLimitRuleDelete', 'CaptchaTest', 'SystemRisk',
            'SystemRiskView', 'SystemRiskEdit', 'SystemRealName',
            'SystemRealNameView', 'SystemRealNameApprove', 'SystemRealNameReject',
            'SystemSecurityLog', 'SystemSecurityLogView', 'TagView', 'TagEdit', 'TagDelete'
        ) AND deleted_at IS NULL
    ) AS t
);

-- 软删除这些菜单
UPDATE sys_menus SET deleted_at = NOW(3)
WHERE FIND_IN_SET(id, @delete_ids) > 0 AND deleted_at IS NULL;

-- 2. 修复 Analytics 和 Workspace 的 auth_code（错误地设为 system:user:list）
UPDATE sys_menus SET auth_code = '' WHERE name = 'Analytics' AND deleted_at IS NULL;
UPDATE sys_menus SET auth_code = '' WHERE name = 'Workspace' AND deleted_at IS NULL;

-- 3. 修复 FileList 的 auth_code（应为空，权限由按钮控制）
UPDATE sys_menus SET auth_code = '' WHERE name = 'FileList' AND deleted_at IS NULL;

-- 4. 修复 FileShare 的 auth_code
UPDATE sys_menus SET auth_code = '' WHERE name = 'FileShare' AND deleted_at IS NULL;

-- 5. 删除 FileList 下多余的按钮（这些权限由后端 Permission 中间件直接控制）
-- 保留 file:upload, file:manage, file:view:own, file:delete, file:share, file:download 对应的按钮
-- 但需要整理结构

-- 清理 FileList 下的重复/多余按钮
DELETE FROM sys_menus WHERE name IN ('FileListView', 'FileListViewAll', 'FileUpload', 'FileDownload', 'FileDelete', 'FileShareCreate', 'FileManage')
AND pid = 41 AND deleted_at IS NULL;

-- 清理 FileShare 下的多余按钮
DELETE FROM sys_menus WHERE name = 'FileShareViewOwn' AND pid = 42 AND deleted_at IS NULL;

-- ============================================================
-- 第二部分：重建正确的菜单结构
-- ============================================================

-- -------------------------------------------
-- 1. 系统管理目录下的菜单
-- -------------------------------------------
SET @system_id = (SELECT id FROM sys_menus WHERE name = 'System' AND type = 'catalog' AND deleted_at IS NULL LIMIT 1);

-- 系统设置菜单（缺失）
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @system_id, 'SystemSettings', '/system/settings', '/system/settings/index', 'menu', 1, 'system:setting:list', 'lucide:sliders-horizontal', 5, '{"title":"系统设置","icon":"lucide:sliders-horizontal","order":5}'
FROM DUAL WHERE @system_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'SystemSettings' AND deleted_at IS NULL);

SET @settings_menu_id = (SELECT id FROM sys_menus WHERE name = 'SystemSettings' AND deleted_at IS NULL LIMIT 1);

-- 系统设置按钮
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @settings_menu_id, 'SystemSettingEdit', '', '', 'button', 1, 'system:setting:edit', '', '{"title":"编辑设置"}', 0
FROM DUAL WHERE @settings_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'system:setting:edit' AND pid = @settings_menu_id);

-- 标签管理菜单（在系统管理下，用于标签CRUD）
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @system_id, 'SystemTag', '/system/tag', '/system/tag/index', 'menu', 1, 'storage:bucket:view', 'mdi:tag-multiple', 6, '{"title":"标签管理","icon":"mdi:tag-multiple","order":6}'
FROM DUAL WHERE @system_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'SystemTag' AND deleted_at IS NULL);

SET @tag_menu_id = (SELECT id FROM sys_menus WHERE name = 'SystemTag' AND deleted_at IS NULL LIMIT 1);

-- 标签管理按钮
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @tag_menu_id, 'TagEdit', '', '', 'button', 1, 'storage:bucket:edit', '', '{"title":"编辑标签"}', 0
FROM DUAL WHERE @tag_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:bucket:edit' AND pid = @tag_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @tag_menu_id, 'TagDelete', '', '', 'button', 1, 'storage:bucket:delete', '', '{"title":"删除标签"}', 0
FROM DUAL WHERE @tag_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:bucket:delete' AND pid = @tag_menu_id);

-- 定时任务菜单（如果不存在）
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @system_id, 'SystemScheduledTask', '/system/scheduled-task', '/system/scheduled-task/index', 'menu', 1, 'system:task:list', 'lucide:clock', 7, '{"title":"定时任务","icon":"lucide:clock","order":7}'
FROM DUAL WHERE @system_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'SystemScheduledTask' AND deleted_at IS NULL);

SET @task_menu_id = (SELECT id FROM sys_menus WHERE name = 'SystemScheduledTask' AND deleted_at IS NULL LIMIT 1);

-- 定时任务按钮补全
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @task_menu_id, 'SystemScheduledTaskView', '', '', 'button', 1, 'system:task:view', '', '{"title":"查看任务"}', 0
FROM DUAL WHERE @task_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'system:task:view' AND pid = @task_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @task_menu_id, 'SystemScheduledTaskDelete', '', '', 'button', 1, 'system:task:delete', '', '{"title":"删除任务"}', 0
FROM DUAL WHERE @task_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'system:task:delete' AND pid = @task_menu_id);

-- -------------------------------------------
-- 2. 存储管理目录下的菜单
-- -------------------------------------------
SET @storage_id = (SELECT id FROM sys_menus WHERE name = 'Storage' AND type = 'catalog' AND deleted_at IS NULL LIMIT 1);

-- 存储桶管理
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @storage_id, 'StorageBucket', '/storage/storage-bucket', '/storage/storage-bucket/index', 'menu', 1, 'storage:bucket:list', 'mdi:database', 1, '{"title":"存储桶管理","icon":"mdi:database","order":1}'
FROM DUAL WHERE @storage_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'StorageBucket' AND deleted_at IS NULL);

SET @bucket_menu_id = (SELECT id FROM sys_menus WHERE name = 'StorageBucket' AND deleted_at IS NULL LIMIT 1);

-- 存储桶按钮
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @bucket_menu_id, 'StorageBucketView', '', '', 'button', 1, 'storage:bucket:view', '', '{"title":"查看存储桶"}', 0
FROM DUAL WHERE @bucket_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:bucket:view' AND pid = @bucket_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @bucket_menu_id, 'StorageBucketEdit', '', '', 'button', 1, 'storage:bucket:edit', '', '{"title":"编辑存储桶"}', 0
FROM DUAL WHERE @bucket_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:bucket:edit' AND pid = @bucket_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @bucket_menu_id, 'StorageBucketDelete', '', '', 'button', 1, 'storage:bucket:delete', '', '{"title":"删除存储桶"}', 0
FROM DUAL WHERE @bucket_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:bucket:delete' AND pid = @bucket_menu_id);

-- 存储连接配置
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @storage_id, 'StorageConfig', '/storage/storage-config', '/storage/storage-config/index', 'menu', 1, 'storage:config:list', 'mdi:server-network', 2, '{"title":"存储连接配置","icon":"mdi:server-network","order":2}'
FROM DUAL WHERE @storage_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'StorageConfig' AND deleted_at IS NULL);

SET @config_menu_id = (SELECT id FROM sys_menus WHERE name = 'StorageConfig' AND deleted_at IS NULL LIMIT 1);

-- 存储配置按钮
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @config_menu_id, 'StorageConfigView', '', '', 'button', 1, 'storage:config:view', '', '{"title":"查看存储配置"}', 0
FROM DUAL WHERE @config_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:config:view' AND pid = @config_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @config_menu_id, 'StorageConfigEdit', '', '', 'button', 1, 'storage:config:edit', '', '{"title":"编辑存储配置"}', 0
FROM DUAL WHERE @config_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:config:edit' AND pid = @config_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @config_menu_id, 'StorageConfigDelete', '', '', 'button', 1, 'storage:config:delete', '', '{"title":"删除存储配置"}', 0
FROM DUAL WHERE @config_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:config:delete' AND pid = @config_menu_id);

-- 文件类型规则（如果不存在）
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @storage_id, 'StorageFileTypeRule', '/storage/file-type-rule', '/storage/file-type-rule/index', 'menu', 1, 'storage:file-type:list', 'mdi:file-check', 3, '{"title":"文件类型规则","icon":"mdi:file-check","order":3}'
FROM DUAL WHERE @storage_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'StorageFileTypeRule' AND deleted_at IS NULL);

SET @ftype_menu_id = (SELECT id FROM sys_menus WHERE name = 'StorageFileTypeRule' AND deleted_at IS NULL LIMIT 1);

-- 文件类型规则按钮补全
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @ftype_menu_id, 'StorageFileTypeRuleView', '', '', 'button', 1, 'storage:file-type:view', '', '{"title":"查看规则"}', 0
FROM DUAL WHERE @ftype_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:file-type:view' AND pid = @ftype_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @ftype_menu_id, 'StorageFileTypeRuleEdit', '', '', 'button', 1, 'storage:file-type:edit', '', '{"title":"编辑规则"}', 0
FROM DUAL WHERE @ftype_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:file-type:edit' AND pid = @ftype_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @ftype_menu_id, 'StorageFileTypeRuleDelete', '', '', 'button', 1, 'storage:file-type:delete', '', '{"title":"删除规则"}', 0
FROM DUAL WHERE @ftype_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'storage:file-type:delete' AND pid = @ftype_menu_id);

-- 标签路由管理
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @storage_id, 'StorageTagRouting', '/storage/tag-routing', '/storage/tag-routing/index', 'menu', 1, 'storage:bucket:view', 'lucide:git-branch', 4, '{"title":"标签路由","icon":"lucide:git-branch","order":4}'
FROM DUAL WHERE @storage_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'StorageTagRouting' AND deleted_at IS NULL);

-- -------------------------------------------
-- 3. 安全管理目录下的菜单
-- -------------------------------------------
SET @security_id = (SELECT id FROM sys_menus WHERE name = 'Security' AND type = 'catalog' AND deleted_at IS NULL LIMIT 1);

-- 实名审核
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @security_id, 'SecurityRealName', '/security/real-name', '/security/real-name/index', 'menu', 1, 'security:realname:list', 'lucide:badge-check', 1, '{"title":"实名审核","icon":"lucide:badge-check","order":1}'
FROM DUAL WHERE @security_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'SecurityRealName' AND deleted_at IS NULL);

SET @realname_menu_id = (SELECT id FROM sys_menus WHERE name = 'SecurityRealName' AND deleted_at IS NULL LIMIT 1);

-- 实名审核按钮
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @realname_menu_id, 'SecurityRealNameApprove', '', '', 'button', 1, 'security:realname:approve', '', '{"title":"审核通过"}', 0
FROM DUAL WHERE @realname_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'security:realname:approve' AND pid = @realname_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @realname_menu_id, 'SecurityRealNameReject', '', '', 'button', 1, 'security:realname:reject', '', '{"title":"审核拒绝"}', 0
FROM DUAL WHERE @realname_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'security:realname:reject' AND pid = @realname_menu_id);

-- 安全日志
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @security_id, 'SecurityLog', '/security/security-log', '/security/security-log/index', 'menu', 1, 'security:log:list', 'lucide:shield-check', 2, '{"title":"安全日志","icon":"lucide:shield-check","order":2}'
FROM DUAL WHERE @security_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'SecurityLog' AND deleted_at IS NULL);

SET @seclog_menu_id = (SELECT id FROM sys_menus WHERE name = 'SecurityLog' AND deleted_at IS NULL LIMIT 1);

-- 安全日志按钮
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @seclog_menu_id, 'SecurityLogView', '', '', 'button', 1, 'security:log:view', '', '{"title":"查看安全日志"}', 0
FROM DUAL WHERE @seclog_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'security:log:view' AND pid = @seclog_menu_id);

-- 风险评分监控
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @security_id, 'SecurityRisk', '/security/risk', '/security/risk/index', 'menu', 1, 'security:risk:list', 'lucide:shield-alert', 3, '{"title":"风险评分监控","icon":"lucide:shield-alert","order":3}'
FROM DUAL WHERE @security_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'SecurityRisk' AND deleted_at IS NULL);

SET @risk_menu_id = (SELECT id FROM sys_menus WHERE name = 'SecurityRisk' AND deleted_at IS NULL LIMIT 1);

-- 风险评分按钮
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @risk_menu_id, 'SecurityRiskView', '', '', 'button', 1, 'security:risk:view', '', '{"title":"查看风险评分"}', 0
FROM DUAL WHERE @risk_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'security:risk:view' AND pid = @risk_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @risk_menu_id, 'SecurityRiskEdit', '', '', 'button', 1, 'security:risk:edit', '', '{"title":"清除风险评分"}', 0
FROM DUAL WHERE @risk_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'security:risk:edit' AND pid = @risk_menu_id);

-- -------------------------------------------
-- 4. 文件管理目录下的菜单
-- -------------------------------------------
SET @file_id = (SELECT id FROM sys_menus WHERE name = 'File' AND type = 'catalog' AND deleted_at IS NULL LIMIT 1);

-- 文件列表
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @file_id, 'FileList', '/file/list', '/file/list/index', 'menu', 1, '', 'lucide:files', 1, '{"title":"文件列表","icon":"lucide:files","order":1}'
FROM DUAL WHERE @file_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'FileList' AND deleted_at IS NULL);

SET @filelist_menu_id = (SELECT id FROM sys_menus WHERE name = 'FileList' AND deleted_at IS NULL LIMIT 1);

-- 文件管理按钮（对应后端 Permission 中间件使用的权限码）
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @filelist_menu_id, 'FileUpload', '', '', 'button', 1, 'file:upload', '', '{"title":"上传文件"}', 0
FROM DUAL WHERE @filelist_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'file:upload' AND pid = @filelist_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @filelist_menu_id, 'FileManage', '', '', 'button', 1, 'file:manage', '', '{"title":"管理文件"}', 0
FROM DUAL WHERE @filelist_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'file:manage' AND pid = @filelist_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @filelist_menu_id, 'FileViewOwn', '', '', 'button', 1, 'file:view:own', '', '{"title":"查看自己的文件"}', 0
FROM DUAL WHERE @filelist_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'file:view:own' AND pid = @filelist_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @filelist_menu_id, 'FileDelete', '', '', 'button', 1, 'file:delete', '', '{"title":"删除文件"}', 0
FROM DUAL WHERE @filelist_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'file:delete' AND pid = @filelist_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @filelist_menu_id, 'FileShare', '', '', 'button', 1, 'file:share', '', '{"title":"分享文件"}', 0
FROM DUAL WHERE @filelist_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'file:share' AND pid = @filelist_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @filelist_menu_id, 'FileDownload', '', '', 'button', 1, 'file:download', '', '{"title":"下载文件"}', 0
FROM DUAL WHERE @filelist_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'file:download' AND pid = @filelist_menu_id);

-- 分享管理
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @file_id, 'FileShareManage', '/file/share', '/file/share/index', 'menu', 1, '', 'lucide:share-2', 2, '{"title":"分享管理","icon":"lucide:share-2","order":2}'
FROM DUAL WHERE @file_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'FileShareManage' AND deleted_at IS NULL);

-- 回收站
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, sort, meta)
SELECT @file_id, 'FileRecycle', '/file/recycle', '/file/recycle/index', 'menu', 1, '', 'lucide:trash-2', 3, '{"title":"回收站","icon":"lucide:trash-2","order":3}'
FROM DUAL WHERE @file_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = 'FileRecycle' AND deleted_at IS NULL);

SET @recycle_menu_id = (SELECT id FROM sys_menus WHERE name = 'FileRecycle' AND deleted_at IS NULL LIMIT 1);

-- 回收站按钮补全
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @recycle_menu_id, 'FileRecycleRestore', '', '', 'button', 1, 'file:recycle:restore', '', '{"title":"恢复文件"}', 0
FROM DUAL WHERE @recycle_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'file:recycle:restore' AND pid = @recycle_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @recycle_menu_id, 'FileRecycleDelete', '', '', 'button', 1, 'file:recycle:delete', '', '{"title":"永久删除"}', 0
FROM DUAL WHERE @recycle_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'file:recycle:delete' AND pid = @recycle_menu_id);

INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta, sort)
SELECT @recycle_menu_id, 'FileRecycleEmpty', '', '', 'button', 1, 'file:recycle:empty', '', '{"title":"清空回收站"}', 0
FROM DUAL WHERE @recycle_menu_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM sys_menus WHERE auth_code = 'file:recycle:empty' AND pid = @recycle_menu_id);

-- ============================================================
-- 第三部分：修复角色权限
-- ============================================================

-- super 角色：拥有所有权限（清空后重新设置）
UPDATE sys_roles SET permissions = JSON_ARRAY(
    'system:user:list', 'system:user:add', 'system:user:edit', 'system:user:delete', 'system:user:view',
    'system:role:list', 'system:role:add', 'system:role:edit', 'system:role:delete', 'system:role:view',
    'system:menu:list', 'system:menu:add', 'system:menu:edit', 'system:menu:delete', 'system:menu:view',
    'system:group:list', 'system:group:add', 'system:group:edit', 'system:group:delete', 'system:group:view',
    'system:setting:list', 'system:setting:edit',
    'system:task:list', 'system:task:view', 'system:task:edit', 'system:task:run', 'system:task:delete',
    'system:captcha:test',
    'system:roleapp:view', 'system:roleapp:review',
    'system:device:view', 'system:device:delete',
    'system:privacy:view',
    'system:oauth:view',
    'system:realname:view',
    'security:realname:list', 'security:realname:approve', 'security:realname:reject',
    'security:log:list', 'security:log:view',
    'security:risk:list', 'security:risk:view', 'security:risk:edit',
    'security:ratelimit:list', 'security:ratelimit:view', 'security:ratelimit:edit', 'security:ratelimit:delete',
    'storage:bucket:list', 'storage:bucket:view', 'storage:bucket:edit', 'storage:bucket:delete',
    'storage:config:list', 'storage:config:view', 'storage:config:edit', 'storage:config:delete',
    'storage:file-type:list', 'storage:file-type:view', 'storage:file-type:edit', 'storage:file-type:delete',
    'file:list', 'file:view:own', 'file:view:all', 'file:upload', 'file:download', 'file:delete', 'file:share', 'file:manage',
    'file:recycle:list', 'file:recycle:restore', 'file:recycle:delete', 'file:recycle:empty',
    'share:view:own', 'share:view:all', 'share:delete', 'share:manage'
) WHERE name = 'super' AND deleted_at IS NULL;

-- admin 角色：拥有大部分权限（与 super 基本相同）
UPDATE sys_roles SET permissions = JSON_ARRAY(
    'system:user:list', 'system:user:add', 'system:user:edit', 'system:user:delete', 'system:user:view',
    'system:role:list', 'system:role:add', 'system:role:edit', 'system:role:delete', 'system:role:view',
    'system:menu:list', 'system:menu:add', 'system:menu:edit', 'system:menu:delete', 'system:menu:view',
    'system:group:list', 'system:group:add', 'system:group:edit', 'system:group:delete', 'system:group:view',
    'system:setting:list', 'system:setting:edit',
    'system:task:list', 'system:task:view', 'system:task:edit', 'system:task:run', 'system:task:delete',
    'system:captcha:test',
    'system:roleapp:view', 'system:roleapp:review',
    'system:device:view', 'system:device:delete',
    'system:privacy:view',
    'system:oauth:view',
    'system:realname:view',
    'security:realname:list', 'security:realname:approve', 'security:realname:reject',
    'security:log:list', 'security:log:view',
    'security:risk:list', 'security:risk:view', 'security:risk:edit',
    'security:ratelimit:list', 'security:ratelimit:view', 'security:ratelimit:edit', 'security:ratelimit:delete',
    'storage:bucket:list', 'storage:bucket:view', 'storage:bucket:edit', 'storage:bucket:delete',
    'storage:config:list', 'storage:config:view', 'storage:config:edit', 'storage:config:delete',
    'storage:file-type:list', 'storage:file-type:view', 'storage:file-type:edit', 'storage:file-type:delete',
    'file:list', 'file:view:own', 'file:view:all', 'file:upload', 'file:download', 'file:delete', 'file:share', 'file:manage',
    'file:recycle:list', 'file:recycle:restore', 'file:recycle:delete', 'file:recycle:empty',
    'share:view:own', 'share:view:all', 'share:delete', 'share:manage'
) WHERE name = 'admin' AND deleted_at IS NULL;

-- user 角色：基础文件操作权限 + 个人功能
UPDATE sys_roles SET permissions = JSON_ARRAY(
    -- 文件管理
    'file:list', 'file:view:own', 'file:upload', 'file:download', 'file:delete', 'file:share', 'file:manage',
    'file:recycle:list', 'file:recycle:restore', 'file:recycle:delete', 'file:recycle:empty',
    -- 分享管理
    'share:view:own', 'share:delete', 'share:manage',
    -- 个人功能
    'system:security:view',
    'system:device:view', 'system:device:delete',
    'system:privacy:view',
    'system:oauth:view',
    'system:realname:view',
    'system:roleapp:view', 'system:roleapp:review'
) WHERE name = 'user' AND deleted_at IS NULL;

-- ============================================================
-- 完成
-- ============================================================
SELECT '菜单、角色、权限修复完成' AS status;
