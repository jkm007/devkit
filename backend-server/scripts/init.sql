-- =============================================
-- 后端管理服务数据库初始化脚本
-- 说明: 按 GORM 模型定义生成，与 AutoMigrate 结构一致
-- =============================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS backend_db DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE backend_db;

-- -------------------------------------------
-- 用户表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_users` (
    `id`                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `name`                VARCHAR(50)  NOT NULL COMMENT '用户名',
    `nickname`            VARCHAR(50)  DEFAULT '' COMMENT '昵称',
    `email`               VARCHAR(100) DEFAULT '' COMMENT '邮箱',
    `phone`               VARCHAR(20)  DEFAULT '' COMMENT '手机号',
    `avatar`              VARCHAR(255) DEFAULT '' COMMENT '头像URL',
    `gender`              TINYINT      DEFAULT 0  COMMENT '性别 0未知 1男 2女',
    `birthday`            DATETIME     DEFAULT NULL COMMENT '生日',
    `bio`                 VARCHAR(500) DEFAULT '' COMMENT '个人简介',
    `password`            VARCHAR(255) NOT NULL COMMENT '密码(bcrypt)',
    `status`              TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 1:启用 0:禁用',
    `group_id`            BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '所属组ID',
    `real_name`           VARCHAR(50)  DEFAULT '' COMMENT '真实姓名',
    `id_card`             VARCHAR(255) DEFAULT '' COMMENT '身份证号(AES加密)',
    `is_real`             TINYINT      DEFAULT 0  COMMENT '是否已实名 0否 1是',
    `register_source`     VARCHAR(20)  DEFAULT 'web' COMMENT '注册来源',
    `last_login_at`       DATETIME     DEFAULT NULL COMMENT '最后登录时间',
    `last_login_ip`       VARCHAR(50)  DEFAULT '' COMMENT '最后登录IP',
    `last_login_device`   VARCHAR(200) DEFAULT '' COMMENT '最后登录设备',
    `login_fail_count`    INT          DEFAULT 0  COMMENT '连续登录失败次数',
    `lock_until`          DATETIME     DEFAULT NULL COMMENT '锁定截止时间',
    `password_changed_at` DATETIME     DEFAULT NULL COMMENT '密码修改时间',
    `remark`              VARCHAR(500) DEFAULT '' COMMENT '备注',
    `created_at`          DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`          DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    `deleted_at`          DATETIME(3)  DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`),
    KEY `idx_deleted_at` (`deleted_at`),
    KEY `idx_group_id` (`group_id`),
    KEY `idx_email` (`email`),
    KEY `idx_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- -------------------------------------------
-- 角色表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_roles` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
    `name`        VARCHAR(50)  NOT NULL COMMENT '角色名称',
    `status`      TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 1:启用 0:禁用',
    `permissions` TEXT         DEFAULT NULL COMMENT '权限码列表(JSON数组)',
    `remark`      VARCHAR(500) DEFAULT '' COMMENT '备注',
    `created_at`  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    `deleted_at`  DATETIME(3)  DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- -------------------------------------------
-- 用户-角色关联表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_user_roles` (
    `id`      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- -------------------------------------------
-- 菜单表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_menus` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '菜单ID',
    `pid`        BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父菜单ID',
    `name`       VARCHAR(100) NOT NULL COMMENT '菜单名称',
    `path`       VARCHAR(200) DEFAULT '' COMMENT '路由地址',
    `component`  VARCHAR(200) DEFAULT '' COMMENT '组件路径',
    `type`       VARCHAR(20)  NOT NULL COMMENT '类型 catalog/menu/embedded/link/button',
    `status`     TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 1:启用 0:禁用',
    `auth_code`  VARCHAR(100) DEFAULT '' COMMENT '权限码',
    `icon`       VARCHAR(100) DEFAULT '' COMMENT '图标',
    `meta`       TEXT         DEFAULT NULL COMMENT '元数据(JSON)',
    `sort`       INT          NOT NULL DEFAULT 0 COMMENT '排序',
    `created_at` DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    `deleted_at` DATETIME(3)  DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY `idx_pid` (`pid`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单表';

-- -------------------------------------------
-- 分组表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_groups` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分组ID',
    `pid`        BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父分组ID,顶级为0',
    `name`       VARCHAR(100) NOT NULL COMMENT '分组名称',
    `status`     TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 1:启用 0:禁用',
    `remark`     VARCHAR(500) DEFAULT '' COMMENT '备注',
    `created_at` DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    `deleted_at` DATETIME(3)  DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY `idx_pid` (`pid`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分组表';

-- -------------------------------------------
-- 分组-角色关联表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_group_roles` (
    `id`       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `group_id` BIGINT UNSIGNED NOT NULL COMMENT '分组ID',
    `role_id`  BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    PRIMARY KEY (`id`),
    KEY `idx_group_id` (`group_id`),
    KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分组角色关联表';

-- -------------------------------------------
-- 安全日志表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_security_logs` (
    `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id`      BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `event_type`   VARCHAR(30)  NOT NULL COMMENT '事件类型',
    `event_detail` VARCHAR(500) DEFAULT '' COMMENT '事件详情',
    `ip`           VARCHAR(50)  DEFAULT '' COMMENT 'IP地址',
    `user_agent`   VARCHAR(500) DEFAULT '' COMMENT 'User-Agent',
    `status`       TINYINT      DEFAULT 1  COMMENT '状态 0失败 1成功',
    `created_at`   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_event_type` (`event_type`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='安全日志表';

-- -------------------------------------------
-- 登录设备表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_login_devices` (
    `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id`        BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `device_id`      VARCHAR(100) NOT NULL COMMENT '设备唯一标识',
    `device_type`    VARCHAR(20)  NOT NULL COMMENT '设备类型 web/ios/android/miniapp',
    `device_name`    VARCHAR(100) DEFAULT '' COMMENT '设备名称',
    `browser`        VARCHAR(50)  DEFAULT '' COMMENT '浏览器',
    `os`             VARCHAR(50)  DEFAULT '' COMMENT '操作系统',
    `ip`             VARCHAR(50)  DEFAULT '' COMMENT 'IP地址',
    `location`       VARCHAR(100) DEFAULT '' COMMENT '登录地点',
    `token_jti`      VARCHAR(100) NOT NULL COMMENT 'JWT ID',
    `last_active_at` DATETIME     DEFAULT NULL COMMENT '最后活跃时间',
    `is_current`     TINYINT      DEFAULT 0  COMMENT '是否当前设备',
    `created_at`     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_device_id` (`device_id`),
    KEY `idx_token_jti` (`token_jti`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录设备表';

-- -------------------------------------------
-- 第三方登录绑定表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_oauth_users` (
    `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id`           BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `provider`          VARCHAR(20)  NOT NULL COMMENT '提供商 wechat/github/google',
    `provider_type`     VARCHAR(20)  DEFAULT '' COMMENT '提供商类型 oauth/oauth2/oidc',
    `provider_user_id`  VARCHAR(100) NOT NULL COMMENT '第三方用户ID',
    `provider_username` VARCHAR(100) DEFAULT '' COMMENT '第三方用户名',
    `provider_avatar`   VARCHAR(255) DEFAULT '' COMMENT '第三方头像',
    `access_token`      VARCHAR(500) DEFAULT '' COMMENT 'Access Token',
    `refresh_token`     VARCHAR(500) DEFAULT '' COMMENT 'Refresh Token',
    `expires_at`        DATETIME     DEFAULT NULL COMMENT 'Token过期时间',
    `created_at`        DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`        DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    UNIQUE KEY `uk_provider_user` (`provider`, `provider_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='第三方登录绑定表';

-- -------------------------------------------
-- 用户隐私设置表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_user_privacy` (
    `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id`          BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `profile_visible`  TINYINT DEFAULT 1 COMMENT '资料可见性 1全部 2仅班级 3仅自己',
    `realname_visible` TINYINT DEFAULT 1 COMMENT '真实姓名可见性',
    `email_visible`    TINYINT DEFAULT 1 COMMENT '邮箱可见性',
    `stats_visible`    TINYINT DEFAULT 1 COMMENT '统计可见性',
    `class_visible`    TINYINT DEFAULT 1 COMMENT '班级可见性',
    `created_at`       DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`       DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户隐私设置表';

-- -------------------------------------------
-- 密码历史表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_password_history` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id`    BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `password`   VARCHAR(255) NOT NULL COMMENT '历史密码(bcrypt)',
    `created_at` DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='密码历史表';

-- -------------------------------------------
-- 实名认证表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_user_real_names` (
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id`       BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `real_name`     VARCHAR(50)  NOT NULL COMMENT '真实姓名',
    `id_card`       VARCHAR(255) NOT NULL COMMENT '身份证号(AES加密)',
    `id_card_hash`  VARCHAR(64)  DEFAULT '' COMMENT '身份证号哈希',
    `status`        TINYINT      DEFAULT 0  COMMENT '状态 0待审 1已认证 2认证失败',
    `reject_reason` VARCHAR(200) DEFAULT '' COMMENT '拒绝原因',
    `reviewed_by`   BIGINT UNSIGNED DEFAULT NULL COMMENT '审核人ID',
    `reviewed_at`   DATETIME     DEFAULT NULL COMMENT '审核时间',
    `submitted_at`  DATETIME     DEFAULT NULL COMMENT '提交时间',
    `created_at`    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    `deleted_at`    DATETIME(3)  DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id` (`user_id`),
    KEY `idx_id_card_hash` (`id_card_hash`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实名认证表';

-- -------------------------------------------
-- 角色申请表
-- -------------------------------------------
CREATE TABLE IF NOT EXISTS `sys_role_applications` (
    `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id`      BIGINT UNSIGNED NOT NULL COMMENT '申请人ID',
    `role_id`      BIGINT UNSIGNED NOT NULL COMMENT '申请角色ID',
    `reason`       VARCHAR(500) DEFAULT '' COMMENT '申请理由',
    `status`       TINYINT      DEFAULT 0  COMMENT '状态 0待审 1通过 2驳回',
    `review_note`  VARCHAR(500) DEFAULT '' COMMENT '审核备注',
    `reviewed_by`  BIGINT UNSIGNED DEFAULT NULL COMMENT '审核人ID',
    `reviewed_at`  DATETIME     DEFAULT NULL COMMENT '审核时间',
    `created_at`   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    `deleted_at`   DATETIME(3)  DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_role_id` (`role_id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色申请表';
