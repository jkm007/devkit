-- =======================================================
-- Migration 026: 添加用户首页菜单 + 分离管理员/用户首页
-- =======================================================

-- 1. Analytics 和 Workspace 只对管理员可见（需要 system:user:list 权限）
UPDATE sys_menus SET auth_code = 'system:user:list' WHERE id IN (2, 3);

-- 2. 添加用户首页菜单（Dashboard 目录下，pid=1，所有用户可见）
INSERT INTO sys_menus (pid, name, path, component, icon, auth_code, type, status, sort, meta, created_at, updated_at)
VALUES (1, 'UserHome', '/user-home', '/dashboard/user-home/index', 'lucide:home', '', 'menu', 1, 100, '{"title":"用户首页"}', NOW(), NOW());
