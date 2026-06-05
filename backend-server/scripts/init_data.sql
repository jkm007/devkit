-- =============================================
-- 初始化默认角色和用户
-- 说明: 创建默认角色(super/admin/user)和默认用户
-- 注意: 默认密码均为 123456，首次登录后应修改
-- =============================================

USE backend_db;

-- -------------------------------------------
-- 默认角色
-- -------------------------------------------
INSERT INTO sys_roles (name, status, permissions, remark) VALUES
('super', 1, '[]', '超级管理员'),
('admin', 1, '[]', '管理员'),
('user',  1, '[]', '普通用户');

-- -------------------------------------------
-- 默认用户 (密码: 123456, bcrypt 加密)
-- -------------------------------------------
INSERT INTO sys_users (name, password, status, remark) VALUES
('vben',  '$2a$10$aDnTQAcn2z/2f9otWlgcweoDJ9Y3i5qaWa895SEK3yP3UjAF2omqK', 1, '超级管理员'),
('admin', '$2a$10$aDnTQAcn2z/2f9otWlgcweoDJ9Y3i5qaWa895SEK3yP3UjAF2omqK', 1, '管理员'),
('jack',  '$2a$10$aDnTQAcn2z/2f9otWlgcweoDJ9Y3i5qaWa895SEK3yP3UjAF2omqK', 1, '普通用户');

-- -------------------------------------------
-- 用户-角色关联
-- vben  -> super
-- admin -> admin
-- jack  -> user
-- -------------------------------------------
INSERT INTO sys_user_roles (user_id, role_id) VALUES
((SELECT id FROM sys_users WHERE name = 'vben'),  (SELECT id FROM sys_roles WHERE name = 'super')),
((SELECT id FROM sys_users WHERE name = 'admin'), (SELECT id FROM sys_roles WHERE name = 'admin')),
((SELECT id FROM sys_users WHERE name = 'jack'),  (SELECT id FROM sys_roles WHERE name = 'user'));
