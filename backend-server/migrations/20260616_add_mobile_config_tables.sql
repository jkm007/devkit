-- 移动端配置表
-- 创建时间: 2026-06-16

-- 1. 创建快捷菜单表
CREATE TABLE IF NOT EXISTS `mobile_quick_menus` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `title` VARCHAR(100) NOT NULL COMMENT '标题',
  `icon` VARCHAR(50) NOT NULL COMMENT '图标',
  `link` VARCHAR(512) DEFAULT '' COMMENT '链接地址',
  `link_type` VARCHAR(20) DEFAULT 'page' COMMENT '链接类型: page/url/function/none',
  `sort_order` INT DEFAULT 0 COMMENT '排序',
  `status` VARCHAR(20) DEFAULT 'enabled' COMMENT '状态: enabled/disabled',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_status_sort` (`status`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='移动端快捷菜单';

-- 2. 创建我的页面菜单表
CREATE TABLE IF NOT EXISTS `mobile_my_page_menus` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `title` VARCHAR(100) NOT NULL COMMENT '标题',
  `icon` VARCHAR(50) NOT NULL COMMENT '图标',
  `link` VARCHAR(512) NOT NULL COMMENT '链接地址',
  `show_badge` TINYINT(1) DEFAULT 0 COMMENT '是否显示角标',
  `badge_text` VARCHAR(50) DEFAULT '' COMMENT '角标文字',
  `sort_order` INT DEFAULT 0 COMMENT '排序',
  `status` VARCHAR(20) DEFAULT 'enabled' COMMENT '状态: enabled/disabled',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_status_sort` (`status`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='移动端我的页面菜单';

-- 3. 创建移动端设置表
CREATE TABLE IF NOT EXISTS `mobile_settings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `notice_enabled` TINYINT(1) DEFAULT 0 COMMENT '是否启用公告',
  `notice_content` TEXT COMMENT '公告内容',
  `app_download_url` VARCHAR(512) DEFAULT '' COMMENT 'APP下载地址',
  `customer_service_url` VARCHAR(512) DEFAULT '' COMMENT '客服链接',
  `about_us` TEXT COMMENT '关于我们',
  `agreement_url` VARCHAR(512) DEFAULT '' COMMENT '用户协议链接',
  `privacy_url` VARCHAR(512) DEFAULT '' COMMENT '隐私政策链接',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='移动端设置';

-- 4. 插入默认快捷菜单
INSERT INTO `mobile_quick_menus` (`title`, `icon`, `link`, `link_type`, `sort_order`, `status`) VALUES
('快速练习', '📝', '/pages/practice/quick', 'page', 1, 'enabled'),
('智能练习', '🎯', '/pages/practice/smart', 'page', 2, 'enabled'),
('错题本', '📖', '/pages/wrong-book/index', 'page', 3, 'enabled'),
('题库', '📚', '/pages/question-bank/index', 'page', 4, 'enabled'),
('收藏夹', '⭐', '/pages/favorites/index', 'page', 5, 'enabled'),
('练习历史', '📊', '/pages/history/index', 'page', 6, 'enabled'),
('公告', '📢', '/pages/announcement/index', 'page', 7, 'enabled'),
('意见反馈', '💬', '/pages/feedback/index', 'page', 8, 'enabled');

-- 5. 插入默认我的页面菜单
INSERT INTO `mobile_my_page_menus` (`title`, `icon`, `link`, `show_badge`, `badge_text`, `sort_order`, `status`) VALUES
('错题本', '📖', '/pages/wrong-book/index', 0, '', 1, 'enabled'),
('收藏夹', '⭐', '/pages/favorites/index', 0, '', 2, 'enabled'),
('练习历史', '📊', '/pages/history/index', 0, '', 3, 'enabled'),
('我的笔记', '📝', '/pages/notes/index', 0, '', 4, 'enabled'),
('题目纠错', '❓', '/pages/feedback/index', 1, 'NEW', 5, 'enabled'),
('通知消息', '🔔', '/pages/notification/index', 1, '', 6, 'enabled'),
('设置', '⚙️', '/pages/settings/index', 0, '', 7, 'enabled'),
('关于我们', 'ℹ️', '/pages/about/index', 0, '', 8, 'enabled');

-- 6. 插入默认移动端设置
INSERT INTO `mobile_settings` (`notice_enabled`, `notice_content`, `app_download_url`, `customer_service_url`, `about_us`, `agreement_url`, `privacy_url`) VALUES
(0, '', '', '', '题库助手 - 您的智能学习伙伴', '', '');

-- 7. 添加菜单项（移动端配置子菜单）
-- 获取移动端配置父菜单ID
SET @mobile_menu_id = (SELECT id FROM sys_menus WHERE name = '移动端配置' AND type = 'directory' LIMIT 1);

-- 如果父菜单不存在，创建它
INSERT INTO sys_menus (name, path, component, icon, pid, sort, type, auth_code, status, created_at, updated_at)
SELECT '移动端配置', '/mobile', NULL, 'cellphone', 0, 6, 'directory', NULL, 1, NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM sys_menus WHERE name = '移动端配置' AND type = 'directory');

SET @mobile_menu_id = (SELECT id FROM sys_menus WHERE name = '移动端配置' AND type = 'directory' LIMIT 1);

-- 快捷菜单管理
INSERT INTO sys_menus (name, path, component, icon, pid, sort, type, auth_code, status, created_at, updated_at)
VALUES ('快捷菜单', 'quick-menu', 'mobile/quick-menu/index', 'view-grid', @mobile_menu_id, 2, 'menu', 'system:banner:view', 1, NOW(), NOW());

-- 我的页面配置
INSERT INTO sys_menus (name, path, component, icon, pid, sort, type, auth_code, status, created_at, updated_at)
VALUES ('我的页面', 'my-page', 'mobile/my-page/index', 'account-cog', @mobile_menu_id, 3, 'menu', 'system:banner:view', 1, NOW(), NOW());

-- 移动端设置
INSERT INTO sys_menus (name, path, component, icon, pid, sort, type, auth_code, status, created_at, updated_at)
VALUES ('移动端设置', 'settings', 'mobile/settings/index', 'cog', @mobile_menu_id, 4, 'menu', 'system:banner:view', 1, NOW(), NOW());
