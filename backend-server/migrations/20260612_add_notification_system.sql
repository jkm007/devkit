-- 通知消息表
CREATE TABLE IF NOT EXISTS `sys_notifications` (
  `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id`    bigint unsigned NOT NULL DEFAULT 0 COMMENT '目标用户ID 0=公告(全员)',
  `type`       varchar(50) NOT NULL COMMENT '通知类型',
  `title`      varchar(200) NOT NULL COMMENT '标题',
  `content`    text COMMENT '内容',
  `link`       varchar(500) DEFAULT '' COMMENT '跳转链接',
  `is_read`    tinyint(1) DEFAULT 0 COMMENT '是否已读',
  `sender_id`  bigint unsigned DEFAULT 0 COMMENT '发送者ID 0=系统',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_type` (`type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知消息';

-- 公告已读记录表
CREATE TABLE IF NOT EXISTS `sys_notification_reads` (
  `id`              bigint unsigned NOT NULL AUTO_INCREMENT,
  `notification_id` bigint unsigned NOT NULL COMMENT '通知ID',
  `user_id`         bigint unsigned NOT NULL COMMENT '用户ID',
  `read_at`         datetime(3) DEFAULT NULL COMMENT '阅读时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_notif_user` (`notification_id`, `user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知已读记录';

-- 添加公告管理菜单（系统管理子菜单）
-- 先查找系统管理菜单的ID（pid=4的catalog）
SET @system_menu_pid = 4;

-- 公告管理菜单
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(211, @system_menu_pid, 'SystemAnnouncement', '/system/announcement', '/system/announcement/index', 'menu', 1, 'system:notification:view', 'lucide:megaphone', 13, '{"title":"公告管理","order":13}', NOW(), NOW());

-- 公告管理按钮权限
INSERT INTO `sys_menus` (`id`, `pid`, `name`, `path`, `component`, `type`, `status`, `auth_code`, `icon`, `sort`, `meta`, `created_at`, `updated_at`) VALUES
(212, 211, 'SystemNotificationPublish', '', '', 'button', 1, 'system:notification:publish', '', 0, '{"title":"发布公告"}', NOW(), NOW());

-- 给超级管理员和管理员角色添加通知权限
-- 超级管理员(id=1)：添加所有通知权限
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'system:notification:view') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"system:notification:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'system:notification:publish') WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"system:notification:publish"');

-- 管理员(id=2)：添加所有通知权限
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'system:notification:view') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"system:notification:view"');
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(`permissions`, '$', 'system:notification:publish') WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"system:notification:publish"');
