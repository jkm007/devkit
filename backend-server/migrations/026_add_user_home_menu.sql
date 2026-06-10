-- =======================================================
-- Migration 026: 添加用户首页菜单
-- 为普通用户添加独立的首页菜单项
-- auth_code 为空表示所有用户可见
-- pid=1 挂在 Dashboard 目录下
-- =======================================================

INSERT INTO sys_menus (pid, name, path, component, icon, auth_code, type, status, sort, meta, created_at, updated_at)
VALUES (1, 'UserHome', '/user-home', '/dashboard/user-home/index', 'lucide:home', '', 'menu', 1, 100, '{"title":"用户首页"}', NOW(), NOW());
