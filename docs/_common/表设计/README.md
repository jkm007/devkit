# 数据库表设计

## 表清单

| 模块 | 表名 | 说明 | 状态 |
|------|------|------|------|
| 系统管理 | sys_users | 用户表（已升级，+17字段） | ✅ 已有 |
| 系统管理 | sys_roles | 角色表 | ✅ 已有 |
| 系统管理 | sys_user_roles | 用户角色关联表 | ✅ 已有 |
| 系统管理 | sys_menus | 菜单表 | ✅ 已有 |
| 系统管理 | sys_groups | 分组表 | ✅ 已有 |
| 系统管理 | sys_group_roles | 分组角色关联表 | ✅ 已有 |
| 用户认证 | sys_user_privacy | 用户隐私设置表 | ✅ 已建 |
| 用户认证 | sys_oauth_users | 第三方登录绑定表 | ✅ 已建 |
| 用户认证 | sys_login_devices | 登录设备表 | ✅ 已建 |
| 用户认证 | sys_security_logs | 安全日志表 | ✅ 已建 |
| 用户认证 | sys_password_history | 密码历史表 | ✅ 已建 |
| 用户认证 | sys_user_real_names | 实名认证表 | ✅ 已建 |
| 用户认证 | sys_role_applications | 角色申请表 | ✅ 已建 |
| 系统设置 | sys_system_settings | 系统配置表 | ❌ 待建 |

## 文档索引

- [当前表结构分析与升级方案](当前表结构分析.md)
- [用户认证模块表设计](用户认证模块.md)
- [系统设置模块表设计](系统设置模块.md)
