-- 修复角色申请菜单 auth_code 重复问题
-- 问题：id=194 (菜单"角色申请") 和 id=195 (按钮"查看申请") 的 auth_code 都是 'system:roleapp:view'
-- 导致权限树中父子节点 key 重复，buildLeafKeysMap 无法解析，"系统管理"、"角色申请"、"查看申请" checkbox 无状态
-- 修复：菜单改为 'system:roleapp:list'（符合 menu 用 :list、button 用 :view 的命名规范）

-- 1. 修改菜单 auth_code
UPDATE `sys_menus` SET `auth_code` = 'system:roleapp:list' WHERE `id` = 194;

-- 2. 为已有角色补充新的菜单权限码
-- super 角色 (id=1)：已有 system:roleapp:view 和 system:roleapp:review，补 system:roleapp:list
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(
  `permissions`, '$', 'system:roleapp:list'
) WHERE `id` = 1 AND NOT JSON_CONTAINS(`permissions`, '"system:roleapp:list"');

-- admin 角色 (id=2)：补充 system:roleapp:list 和 system:roleapp:view、system:roleapp:review
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(
  JSON_ARRAY_APPEND(
    JSON_ARRAY_APPEND(`permissions`, '$', 'system:roleapp:list'),
    '$', 'system:roleapp:view'
  ),
  '$', 'system:roleapp:review'
) WHERE `id` = 2 AND NOT JSON_CONTAINS(`permissions`, '"system:roleapp:list"');

-- user 角色 (id=3)：已有 system:roleapp:view 和 system:roleapp:review，补 system:roleapp:list
UPDATE `sys_roles` SET `permissions` = JSON_ARRAY_APPEND(
  `permissions`, '$', 'system:roleapp:list'
) WHERE `id` = 3 AND NOT JSON_CONTAINS(`permissions`, '"system:roleapp:list"');
