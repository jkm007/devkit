-- ============================================================
-- 迁移: 015
-- 描述: 创建存储桶管理表
-- 作者: Claude Code
-- 日期: 2026-06-08
-- ============================================================

-- 存储桶配置表
CREATE TABLE `sys_storage_bucket` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL COMMENT '存储桶名称',
  `driver` VARCHAR(20) NOT NULL COMMENT '存储驱动: local, minio, oss, cos',
  `endpoint` VARCHAR(500) DEFAULT '' COMMENT '服务端点',
  `bucket` VARCHAR(200) DEFAULT '' COMMENT '桶名称',
  `access_key` VARCHAR(500) DEFAULT '' COMMENT '访问密钥ID',
  `secret_key` VARCHAR(500) DEFAULT '' COMMENT '访问密钥Secret',
  `region` VARCHAR(100) DEFAULT '' COMMENT '区域',
  `use_ssl` TINYINT(1) DEFAULT 0 COMMENT '是否使用SSL',
  `cdn_domain` VARCHAR(500) DEFAULT '' COMMENT 'CDN域名',
  `path_prefix` VARCHAR(500) DEFAULT '' COMMENT '路径前缀',
  `purpose` VARCHAR(100) DEFAULT '' COMMENT '用途: file, backup, avatar, temp, etc',
  `is_default` TINYINT(1) DEFAULT 0 COMMENT '是否默认存储桶',
  `status` TINYINT(1) DEFAULT 1 COMMENT '状态: 1=启用, 0=禁用',
  `description` VARCHAR(500) DEFAULT '' COMMENT '描述',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_driver` (`driver`),
  KEY `idx_purpose` (`purpose`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='存储桶配置表';

-- 插入默认存储桶配置（从 config.yaml 同步）
-- INSERT INTO sys_storage_bucket (name, driver, endpoint, bucket, access_key, secret_key, use_ssl, purpose, is_default, status, description) VALUES
-- ('默认本地存储', 'local', '', '', '', '', 0, 'file', 1, 1, '系统默认本地存储');
