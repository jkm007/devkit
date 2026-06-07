-- 分享管理权限按钮菜单
-- 添加到 sys_menus 表，pid=42 是 FileShare 菜单的 ID

INSERT INTO sys_menus (name, path, pid, type, icon, auth_code, status, created_at, updated_at) VALUES
('ShareViewOwn', 'share:view:own', 42, 'button', '', 'share:view:own', 1, NOW(), NOW()),
('ShareViewAll', 'share:view:all', 42, 'button', '', 'share:view:all', 1, NOW(), NOW()),
('ShareDelete', 'share:delete', 42, 'button', '', 'share:delete', 1, NOW(), NOW()),
('ShareManage', 'share:manage', 42, 'button', '', 'share:manage', 1, NOW(), NOW());

-- 给角色添加权限（使用 JSON_ARRAY_APPEND）
-- super 角色 (id=1)
UPDATE sys_roles SET permissions = JSON_ARRAY_APPEND(
  permissions,
  '$', 'share:view:own',
  '$', 'share:view:all',
  '$', 'share:delete',
  '$', 'share:manage'
) WHERE id = 1;

-- admin 角色 (id=2)
UPDATE sys_roles SET permissions = JSON_ARRAY_APPEND(
  permissions,
  '$', 'share:view:own',
  '$', 'share:view:all',
  '$', 'share:delete',
  '$', 'share:manage'
) WHERE id = 2;

-- user 角色 (id=3) - 只能查看自己的分享
UPDATE sys_roles SET permissions = JSON_ARRAY_APPEND(
  permissions,
  '$', 'share:view:own'
) WHERE id = 3;
