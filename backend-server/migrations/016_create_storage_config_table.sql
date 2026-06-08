-- 创建存储连接配置表
CREATE TABLE IF NOT EXISTS `sys_storage_config` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(100) NOT NULL,
  `driver` VARCHAR(20) NOT NULL,
  `endpoint` VARCHAR(500) DEFAULT '',
  `access_key` VARCHAR(500) DEFAULT '',
  `secret_key` VARCHAR(500) DEFAULT '',
  `bucket` VARCHAR(200) DEFAULT '',
  `region` VARCHAR(100) DEFAULT '',
  `use_ssl` TINYINT(1) DEFAULT 0,
  `cdn_domain` VARCHAR(500) DEFAULT '',
  `is_default` TINYINT(1) DEFAULT 0,
  `presigned_url_expiry` INT NOT NULL DEFAULT 3600 COMMENT '预签名URL默认过期时间(秒)',
  `status` TINYINT DEFAULT 1,
  `description` VARCHAR(500) DEFAULT '',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_driver` (`driver`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='存储连接配置表';

-- 插入默认本地存储配置
INSERT INTO `sys_storage_config` (`name`, `driver`, `status`, `is_default`, `presigned_url_expiry`, `description`)
SELECT '本地存储', 'local', 1, 1, 3600, '系统默认本地存储'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `sys_storage_config` WHERE `driver` = 'local');
