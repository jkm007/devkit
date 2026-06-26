-- 微信小程序配置
-- 执行此 SQL 后，后端才能启用微信小程序登录

INSERT INTO `sys_system_settings` (`group_key`, `key`, `value`, `label`, `type`, `description`, `sort_order`, `created_at`, `updated_at`) VALUES
('wechat', 'wechat_miniapp_enabled', 'true', '小程序登录开关', 'boolean', '是否启用微信小程序登录', 1, NOW(), NOW()),
('wechat', 'wechat_miniapp_appid', '"wxc7e0745dc22733bf"', '小程序 AppID', 'string', '微信小程序 AppID', 2, NOW(), NOW()),
('wechat', 'wechat_miniapp_secret', '"c932bdd8e816065140f5ca39ccc5ff5a"', '小程序 AppSecret', 'string', '微信小程序 AppSecret', 3, NOW(), NOW());
