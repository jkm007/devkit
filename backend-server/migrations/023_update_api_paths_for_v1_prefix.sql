-- =======================================================
-- Migration 023: 更新风险防护路径为 /api/v1 前缀格式
-- =======================================================
-- 修复 risk_protected_paths 格式：从 JSON 数组改为逗号分隔
-- （匹配 parseStringList 函数的解析逻辑：按逗号分割、trim 引号和空格）
-- 注意：文件相关路径（/api/v1/files, /api/v1/shares）不纳入验证码防护，
-- 因为这些是普通用户的正常操作，不是管理后台操作。
UPDATE sys_system_settings
SET value = '/api/v1/system/user,/api/v1/system/role,/api/v1/system/group,/api/v1/system/menu,/api/v1/system/settings,/api/v1/system/storage-buckets,/api/v1/system/storage-configs,/api/v1/system/scheduled-tasks'
WHERE group_key = 'risk_score' AND `key` = 'risk_protected_paths';
