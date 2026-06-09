-- 创建限流规则表
CREATE TABLE IF NOT EXISTS sys_rate_limit_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    path_pattern VARCHAR(255) NOT NULL COMMENT '路径模式，支持 * 通配符',
    method VARCHAR(10) DEFAULT '*' COMMENT 'HTTP 方法：GET/POST/PUT/DELETE/*',
    rate DOUBLE NOT NULL DEFAULT 10 COMMENT '每秒请求数',
    burst INT NOT NULL DEFAULT 20 COMMENT '突发容量',
    description VARCHAR(500) DEFAULT '' COMMENT '规则描述',
    enabled TINYINT(1) DEFAULT 1 COMMENT '是否启用',
    priority INT DEFAULT 0 COMMENT '优先级，数值越大越先匹配',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_path_pattern (path_pattern),
    INDEX idx_priority (priority),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='限流规则表';

-- 插入默认规则（从当前硬编码迁移）
INSERT INTO sys_rate_limit_rules (path_pattern, method, rate, burst, description, enabled, priority) VALUES
('/auth/send-code', 'POST', 1, 2, '邮箱验证码发送限流', 1, 100),
('/auth/send-sms-code', 'POST', 1, 2, '短信验证码发送限流', 1, 100),
('/auth/login-by-email', 'POST', 1, 3, '邮箱登录限流', 1, 90),
('/auth/login-by-phone', 'POST', 1, 3, '手机号登录限流', 1, 90),
('/auth/register', 'POST', 1, 3, '注册限流', 1, 90),
('/auth/reset-password', 'POST', 1, 3, '重置密码限流', 1, 90),
('/auth/captcha', 'GET', 5, 10, '验证码获取限流', 1, 80),
('/auth/verify-code', 'POST', 5, 10, '验证码验证限流', 1, 80),
('/auth/oauth/callback', 'GET', 2, 5, 'OAuth 回调限流', 1, 80),
('/share/*', 'GET', 10, 20, '分享链接访问限流', 1, 70);

-- 添加限流规则按钮权限（在安全管理目录下）
-- 查找安全管理目录的 ID
SET @security_id = (SELECT id FROM sys_menus WHERE name = 'Security' AND type = 'catalog' LIMIT 1);

-- 按钮权限：security:ratelimit:view, security:ratelimit:edit, security:ratelimit:delete
INSERT INTO sys_menus (pid, name, path, component, type, status, auth_code, icon, meta) VALUES
(@security_id, 'RateLimitRuleView', '', '', 'button', 1, 'security:ratelimit:view', '', '{"title":"限流规则查看"}'),
(@security_id, 'RateLimitRuleEdit', '', '', 'button', 1, 'security:ratelimit:edit', '', '{"title":"限流规则编辑"}'),
(@security_id, 'RateLimitRuleDelete', '', '', 'button', 1, 'security:ratelimit:delete', '', '{"title":"限流规则删除"}');

-- 给 admin 角色添加限流规则权限
UPDATE sys_roles SET permissions = JSON_ARRAY_APPEND(
    permissions, '$', 'security:ratelimit:view'
) WHERE id = 2 AND NOT JSON_CONTAINS(permissions, '"security:ratelimit:view"');

UPDATE sys_roles SET permissions = JSON_ARRAY_APPEND(
    permissions, '$', 'security:ratelimit:edit'
) WHERE id = 2 AND NOT JSON_CONTAINS(permissions, '"security:ratelimit:edit"');

UPDATE sys_roles SET permissions = JSON_ARRAY_APPEND(
    permissions, '$', 'security:ratelimit:delete'
) WHERE id = 2 AND NOT JSON_CONTAINS(permissions, '"security:ratelimit:delete"');
