-- 题库管理设计修正 - 版本表和字段补充

-- 1. 创建题目版本快照表
CREATE TABLE IF NOT EXISTS `qb_question_versions` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '版本ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `version` INT NOT NULL DEFAULT 1 COMMENT '版本号',
  `version_status` VARCHAR(20) DEFAULT 'active' COMMENT '版本状态 active/archived/deprecated',
  `change_log` VARCHAR(500) DEFAULT '' COMMENT '本次发布变更说明',

  -- 快照字段
  `title` VARCHAR(500) NOT NULL COMMENT '快照：题目标题/摘要',
  `question_type` VARCHAR(50) NOT NULL COMMENT '快照：题型',
  `stem` JSON NOT NULL COMMENT '快照：题干富媒体内容块',
  `content` JSON COMMENT '快照：题型结构内容',
  `answer` JSON COMMENT '快照：编辑态答案',
  `analysis` JSON COMMENT '快照：解析富媒体内容块',
  `materials` JSON COMMENT '快照：材料',
  `score_rule` JSON COMMENT '快照：判分规则',

  -- 分类和属性快照
  `exam_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '快照：具体考试',
  `subject_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '快照：科目/模块',
  `category_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '快照：章节分类',
  `source_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '快照：题目来源',
  `difficulty` TINYINT DEFAULT 1 COMMENT '快照：1简单 2中等 3困难',
  `resource_type` VARCHAR(20) DEFAULT 'private' COMMENT '快照：public/private/group/user',

  -- 可见策略快照
  `analysis_visible_policy` VARCHAR(30) DEFAULT 'after_answer' COMMENT '快照：解析可见策略',
  `answer_visible_policy` VARCHAR(30) DEFAULT 'after_answer' COMMENT '快照：答案可见策略',

  -- 组合题快照
  `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '快照：父题ID（组合题）',
  `parent_version_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '组合题父版本ID',
  `is_group` TINYINT DEFAULT 0 COMMENT '快照：是否组合题父题',
  `sub_index` INT DEFAULT 0 COMMENT '快照：子题排序',

  -- 指纹
  `stem_hash` VARCHAR(64) DEFAULT '' COMMENT '题干指纹',
  `content_hash` VARCHAR(64) DEFAULT '' COMMENT '内容指纹',
  `answer_hash` VARCHAR(64) DEFAULT '' COMMENT '答案指纹',

  -- 发布信息
  `published_by` BIGINT UNSIGNED DEFAULT 0 COMMENT '发布人ID',
  `attachment_snapshot` JSON COMMENT '附件快照',

  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) NULL COMMENT '软删除时间',

  INDEX `idx_qb_question_versions_question_id` (`question_id`),
  INDEX `idx_qb_question_versions_version` (`question_id`, `version`),
  INDEX `idx_qb_question_versions_status` (`version_status`),
  INDEX `idx_qb_question_versions_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目版本快照表';

-- 2. 添加 updated_by 字段到 qb_questions (如果不存在)
SET @column_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'qb_questions'
    AND COLUMN_NAME = 'updated_by'
);

SET @sql = IF(@column_exists = 0,
  'ALTER TABLE `qb_questions` ADD COLUMN `updated_by` BIGINT UNSIGNED DEFAULT 0 COMMENT \'最后修改人ID\' AFTER `created_by`, ADD INDEX `idx_qb_questions_updated_by` (`updated_by`)',
  'SELECT \'updated_by column already exists\''
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 3. 添加错误码字段到 qb_question_import_items (如果不存在)
SET @column_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'qb_question_import_items'
    AND COLUMN_NAME = 'error_code'
);

SET @sql = IF(@column_exists = 0,
  'ALTER TABLE `qb_question_import_items` ADD COLUMN `error_code` VARCHAR(50) DEFAULT \'\' COMMENT \'错误码\' AFTER `parse_status`, ADD COLUMN `duplicate_question_id` BIGINT UNSIGNED DEFAULT 0 COMMENT \'查重发现的疑似重复题ID\' AFTER `error_message`, ADD COLUMN `updated_at` DATETIME(3) NULL COMMENT \'更新时间\' AFTER `created_at`',
  'SELECT \'error_code column already exists\''
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 4. 创建关系表（如果不存在）
CREATE TABLE IF NOT EXISTS `qb_question_knowledge_points` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `knowledge_point_id` BIGINT UNSIGNED NOT NULL COMMENT '知识点ID',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL COMMENT '更新时间',

  UNIQUE KEY `uk_qb_question_knowledge` (`question_id`, `knowledge_point_id`),
  INDEX `idx_qb_question_knowledge_question_id` (`question_id`),
  INDEX `idx_qb_question_knowledge_point_id` (`knowledge_point_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目知识点关联表';

CREATE TABLE IF NOT EXISTS `qb_question_tags` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `tag_id` BIGINT UNSIGNED NOT NULL COMMENT '标签ID',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL COMMENT '更新时间',

  UNIQUE KEY `uk_qb_question_tag` (`question_id`, `tag_id`),
  INDEX `idx_qb_question_tag_question_id` (`question_id`),
  INDEX `idx_qb_question_tag_tag_id` (`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目标签关联表';

CREATE TABLE IF NOT EXISTS `qb_question_source_relations` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `source_id` BIGINT UNSIGNED NOT NULL COMMENT '来源ID',
  `question_no` VARCHAR(50) DEFAULT '' COMMENT '该题在特定来源中的题号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL COMMENT '更新时间',

  UNIQUE KEY `uk_qb_question_source` (`question_id`, `source_id`),
  INDEX `idx_qb_question_source_question_id` (`question_id`),
  INDEX `idx_qb_question_source_source_id` (`source_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题目来源关系表';

-- 5. 创建分享访问日志表
CREATE TABLE IF NOT EXISTS `qb_question_share_access_logs` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
  `share_id` BIGINT UNSIGNED NOT NULL COMMENT '分享ID',
  `question_id` BIGINT UNSIGNED NOT NULL COMMENT '题目ID',
  `share_code` VARCHAR(64) NOT NULL COMMENT '分享码',
  `ip` VARCHAR(50) DEFAULT '' COMMENT '访问IP',
  `user_agent` VARCHAR(500) DEFAULT '' COMMENT 'UserAgent',
  `user_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '访问者用户ID，未登录为0',
  `accessed_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '访问时间',
  `response_status` VARCHAR(20) DEFAULT 'success' COMMENT '响应状态 success/expired/disabled/not_found',

  INDEX `idx_qb_share_access_share_id` (`share_id`),
  INDEX `idx_qb_share_access_code` (`share_code`),
  INDEX `idx_qb_share_access_at` (`accessed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分享访问日志表';

-- 6. 创建题型配置表
CREATE TABLE IF NOT EXISTS `qb_question_type_configs` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '配置ID',
  `type` VARCHAR(50) NOT NULL COMMENT '题型标识如single_choice/programming',
  `label` VARCHAR(100) NOT NULL COMMENT '中文名称',
  `answer_format` JSON NULL COMMENT '答案结构模板',
  `scoring_mode` VARCHAR(30) NOT NULL DEFAULT 'auto' COMMENT '评分模式 auto/manual/auto_or_manual',
  `has_options` TINYINT DEFAULT 0 COMMENT '是否有选项',
  `has_analysis` TINYINT DEFAULT 1 COMMENT '是否有解析',
  `has_materials` TINYINT DEFAULT 0 COMMENT '是否有材料',
  `has_children` TINYINT DEFAULT 0 COMMENT '是否有子题',
  `enabled` TINYINT DEFAULT 1 COMMENT '是否启用',
  `sort_order` INT DEFAULT 0 COMMENT '排序',

  UNIQUE KEY `uk_qb_type_config_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='题型配置表';

-- 7. 插入默认题型配置
INSERT INTO `qb_question_type_configs` (`type`, `label`, `scoring_mode`, `has_options`, `has_analysis`, `sort_order`) VALUES
('single_choice', '单选题', 'auto', 1, 1, 1),
('multiple_choice', '多选题', 'auto', 1, 1, 2),
('true_false', '判断题', 'auto', 1, 1, 3),
('fill_blank', '填空题', 'auto', 0, 1, 4),
('short_answer', '简答题', 'manual', 0, 1, 5),
('essay_question', '论述题', 'manual', 0, 1, 6),
('material', '材料题', 'manual', 0, 1, 7),
('case_analysis', '案例分析题', 'manual', 0, 1, 8),
('programming', '编程题', 'auto_or_manual', 0, 1, 9),
('calculation', '计算题', 'manual', 0, 1, 10);
