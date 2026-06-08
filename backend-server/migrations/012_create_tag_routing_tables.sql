-- ============================================================
-- 迁移: 012
-- 描述: 创建标签路由系统相关表
-- 作者: Claude Code
-- 日期: 2026-06-08
-- ============================================================

-- 1. 标签定义表
CREATE TABLE IF NOT EXISTS `sys_tag` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `tag_key`     VARCHAR(50) NOT NULL COMMENT '标签键(如: type, source, sensitivity)',
    `tag_value`   VARCHAR(50) NOT NULL COMMENT '标签值(如: image, video, user)',
    `tag_name`    VARCHAR(100) NOT NULL COMMENT '显示名称(如: 图片, 视频, 用户上传)',
    `icon`        VARCHAR(50) COMMENT '图标',
    `color`       VARCHAR(20) COMMENT '颜色',
    `description` VARCHAR(200) COMMENT '描述',
    `is_system`   TINYINT(1) DEFAULT 0 COMMENT '是否系统内置(不允许删除)',
    `sort_order`  INT DEFAULT 0 COMMENT '排序',
    `status`      TINYINT(1) DEFAULT 1 COMMENT '状态: 1=启用, 0=禁用',
    `created_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_key_value` (`tag_key`, `tag_value`),
    KEY `idx_key` (`tag_key`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签定义表';

-- 2. 标签路由规则表
CREATE TABLE IF NOT EXISTS `sys_tag_routing` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `rule_name`     VARCHAR(100) NOT NULL COMMENT '规则名称',
    `description`   VARCHAR(200) COMMENT '规则描述',
    `priority`      INT DEFAULT 0 COMMENT '优先级(越大越优先)',
    `match_type`    ENUM('all', 'any', 'exact') NOT NULL DEFAULT 'all' COMMENT '匹配类型',
    `conditions`    JSON NOT NULL COMMENT '匹配条件(标签键值对)',
    `driver`        VARCHAR(20) NOT NULL COMMENT '目标存储驱动(local/minio/oss/cos)',
    `bucket`        VARCHAR(100) COMMENT '目标桶名(云存储)',
    `path_prefix`   VARCHAR(200) COMMENT '路径前缀',
    `extra_config`  JSON COMMENT '额外配置(加密、压缩等)',
    `is_default`    TINYINT(1) DEFAULT 0 COMMENT '是否默认规则(兜底)',
    `status`        TINYINT(1) DEFAULT 1 COMMENT '状态: 1=启用, 0=禁用',
    `created_at`    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_priority` (`priority` DESC),
    KEY `idx_status` (`status`),
    KEY `idx_default` (`is_default`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签路由规则表';

-- 3. 文件标签关联表
CREATE TABLE IF NOT EXISTS `sys_file_tag` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `file_id`     BIGINT UNSIGNED NOT NULL COMMENT '文件ID(file_entry.id)',
    `tag_id`      BIGINT UNSIGNED NOT NULL COMMENT '标签ID(sys_tag.id)',
    `source`      ENUM('auto', 'manual') NOT NULL DEFAULT 'auto' COMMENT '来源: auto=自动, manual=手动',
    `created_at`  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_file_tag` (`file_id`, `tag_id`),
    KEY `idx_file_id` (`file_id`),
    KEY `idx_tag_id` (`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件标签关联表';

-- ============================================================
-- 插入默认标签数据
-- ============================================================

-- 文件类型标签
INSERT INTO `sys_tag` (`tag_key`, `tag_value`, `tag_name`, `icon`, `color`, `is_system`, `sort_order`) VALUES
('type', 'image', '图片', '🖼️', '#52c41a', 1, 1),
('type', 'video', '视频', '🎬', '#1890ff', 1, 2),
('type', 'audio', '音频', '🎵', '#722ed1', 1, 3),
('type', 'document', '文档', '📄', '#fa8c16', 1, 4),
('type', 'archive', '压缩包', '📦', '#13c2c2', 1, 5),
('type', 'other', '其他', '📎', '#8c8c8c', 1, 99);

-- 文件来源标签
INSERT INTO `sys_tag` (`tag_key`, `tag_value`, `tag_name`, `icon`, `color`, `is_system`, `sort_order`) VALUES
('source', 'user', '用户上传', '👤', '#1890ff', 1, 1),
('source', 'system', '系统生成', '⚙️', '#595959', 1, 2),
('source', 'import', '批量导入', '📥', '#faad14', 1, 3);

-- 敏感度标签
INSERT INTO `sys_tag` (`tag_key`, `tag_value`, `tag_name`, `icon`, `color`, `is_system`, `sort_order`) VALUES
('sensitivity', 'public', '公开', '🔓', '#52c41a', 1, 1),
('sensitivity', 'internal', '内部', '🔒', '#faad14', 1, 2),
('sensitivity', 'confidential', '机密', '🔐', '#f5222d', 1, 3);

-- 用途标签
INSERT INTO `sys_tag` (`tag_key`, `tag_value`, `tag_name`, `icon`, `color`, `is_system`, `sort_order`) VALUES
('purpose', 'avatar', '头像', '👤', '#1890ff', 1, 1),
('purpose', 'thumbnail', '缩略图', '🖼️', '#52c41a', 1, 2),
('purpose', 'backup', '备份', '💾', '#722ed1', 1, 3),
('purpose', 'temp', '临时文件', '⏱️', '#8c8c8c', 1, 4);

-- ============================================================
-- 插入默认路由规则
-- ============================================================

INSERT INTO `sys_tag_routing` (`rule_name`, `description`, `priority`, `match_type`, `conditions`, `driver`, `bucket`, `path_prefix`, `is_default`) VALUES
('图片存储', '所有图片文件存储到本地', 10, 'all',
  '{"tags": [{"key": "type", "value": "image"}]}',
  'local', NULL, 'images/', 0),

('视频存储', '所有视频文件存储到本地', 10, 'all',
  '{"tags": [{"key": "type", "value": "video"}]}',
  'local', NULL, 'videos/', 0),

('文档存储', '所有文档存储到本地', 10, 'all',
  '{"tags": [{"key": "type", "value": "document"}]}',
  'local', NULL, 'docs/', 0),

('音频存储', '所有音频文件存储到本地', 10, 'all',
  '{"tags": [{"key": "type", "value": "audio"}]}',
  'local', NULL, 'audios/', 0),

('压缩包存储', '所有压缩包存储到本地', 10, 'all',
  '{"tags": [{"key": "type", "value": "archive"}]}',
  'local', NULL, 'archives/', 0),

('备份存储', '系统备份存储到本地', 5, 'all',
  '{"tags": [{"key": "purpose", "value": "backup"}]}',
  'local', NULL, 'backup/', 0),

('临时文件', '临时文件存储到本地', 5, 'all',
  '{"tags": [{"key": "purpose", "value": "temp"}]}',
  'local', NULL, 'temp/', 0),

('默认规则', '其他文件存储到本地', 0, 'all',
  '{"tags": []}',
  'local', NULL, 'files/', 1);
