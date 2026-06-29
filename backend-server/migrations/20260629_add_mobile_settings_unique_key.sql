-- 20260629: 为 mobile_settings 表添加 setting_key 字段及唯一约束
ALTER TABLE mobile_settings ADD COLUMN setting_key VARCHAR(50) NOT NULL DEFAULT 'mobile';
ALTER TABLE mobile_settings ADD UNIQUE INDEX uk_mobile_settings_key (setting_key);
