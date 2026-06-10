-- 迁移：更新所有 API 路径加上 /api/v1 前缀
-- 背景：后端路由统一添加了 /api/v1 前缀，数据库中存储的路径需要同步更新

-- 1. 更新限流规则路径
UPDATE sys_rate_limit_rules SET path_pattern = '/api/v1/auth/send-code' WHERE path_pattern = '/auth/send-code';
UPDATE sys_rate_limit_rules SET path_pattern = '/api/v1/auth/send-sms-code' WHERE path_pattern = '/auth/send-sms-code';
UPDATE sys_rate_limit_rules SET path_pattern = '/api/v1/auth/login-by-email' WHERE path_pattern = '/auth/login-by-email';
UPDATE sys_rate_limit_rules SET path_pattern = '/api/v1/auth/login-by-phone' WHERE path_pattern = '/auth/login-by-phone';
UPDATE sys_rate_limit_rules SET path_pattern = '/api/v1/auth/register' WHERE path_pattern = '/auth/register';
UPDATE sys_rate_limit_rules SET path_pattern = '/api/v1/auth/reset-password' WHERE path_pattern = '/auth/reset-password';
UPDATE sys_rate_limit_rules SET path_pattern = '/api/v1/auth/captcha' WHERE path_pattern = '/auth/captcha';
UPDATE sys_rate_limit_rules SET path_pattern = '/api/v1/auth/verify-code' WHERE path_pattern = '/auth/verify-code';
UPDATE sys_rate_limit_rules SET path_pattern = '/api/v1/auth/oauth/callback' WHERE path_pattern = '/auth/oauth/callback';
UPDATE sys_rate_limit_rules SET path_pattern = '/api/v1/share/*' WHERE path_pattern = '/share/*';

-- 2. 更新风险评分保护路径（逗号分隔格式，与 parseStringList 函数匹配）
UPDATE sys_system_settings
SET value = '"/api/v1/system/user,/api/v1/system/role,/api/v1/system/group,/api/v1/system/menu,/api/v1/system/settings,/api/v1/system/storage-buckets,/api/v1/system/storage-configs,/api/v1/system/scheduled-tasks,/api/v1/files,/api/v1/files/recycle,/api/v1/shares"'
WHERE `key` = 'risk_protected_paths' AND group_key = 'risk_score' AND deleted_at IS NULL;
