INSERT INTO sys_system_settings (group_key, `key`, value, label, type, tip, sort, is_public, created_at, updated_at) VALUES
('auth', 'login_email_enabled', 'false', '邮箱验证码登录', 'boolean', '是否启用邮箱验证码登录', 1, 1, NOW(), NOW()),
('auth', 'login_phone_enabled', 'false', '手机验证码登录', 'boolean', '是否启用手机验证码登录', 2, 1, NOW(), NOW()),
('auth', 'login_oauth_enabled', 'false', '第三方 OAuth 登录', 'boolean', '是否启用第三方 OAuth 登录', 3, 1, NOW(), NOW()),
('auth', 'login_oauth_providers', '[]', 'OAuth 提供商', 'array', '启用的 OAuth 提供商列表，如 ["github","wechat","google"]', 4, 1, NOW(), NOW());