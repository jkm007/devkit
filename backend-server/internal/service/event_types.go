package service

// 安全日志事件类型常量
const (
	// 认证相关
	EventLogin         = "login"          // 登录成功
	EventLoginFail     = "login_fail"     // 登录失败
	EventPasswordChange = "password_change" // 修改密码
	EventLogout        = "logout"         // 退出登录

	// 用户管理
	EventUserCreate = "user_create" // 创建用户
	EventUserUpdate = "user_update" // 更新用户
	EventUserDelete = "user_delete" // 删除用户

	// 角色管理
	EventRoleCreate = "role_create" // 创建角色
	EventRoleUpdate = "role_update" // 更新角色
	EventRoleDelete = "role_delete" // 删除角色

	// 菜单管理
	EventMenuCreate = "menu_create" // 创建菜单
	EventMenuUpdate = "menu_update" // 更新菜单
	EventMenuDelete = "menu_delete" // 删除菜单

	// 分组管理
	EventGroupCreate = "group_create" // 创建分组
	EventGroupUpdate = "group_update" // 更新分组
	EventGroupDelete = "group_delete" // 删除分组
)

// pathToEventType HTTP 方法+路径 映射到事件类型
func PathToEventType(method, path string) (string, string) {
	switch {
	// 用户管理
	case method == "POST" && path == "/system/user":
		return EventUserCreate, "创建用户"
	case method == "PUT" && len(path) > 13 && path[:13] == "/system/user/":
		return EventUserUpdate, "更新用户"
	case method == "DELETE" && len(path) > 13 && path[:13] == "/system/user/":
		return EventUserDelete, "删除用户"

	// 角色管理
	case method == "POST" && path == "/system/role":
		return EventRoleCreate, "创建角色"
	case method == "PUT" && len(path) > 13 && path[:13] == "/system/role/":
		return EventRoleUpdate, "更新角色"
	case method == "DELETE" && len(path) > 13 && path[:13] == "/system/role/":
		return EventRoleDelete, "删除角色"

	// 菜单管理
	case method == "POST" && path == "/system/menu":
		return EventMenuCreate, "创建菜单"
	case method == "PUT" && len(path) > 13 && path[:13] == "/system/menu/":
		return EventMenuUpdate, "更新菜单"
	case method == "DELETE" && len(path) > 13 && path[:13] == "/system/menu/":
		return EventMenuDelete, "删除菜单"

	// 分组管理
	case method == "POST" && path == "/system/group":
		return EventGroupCreate, "创建分组"
	case method == "PUT" && len(path) > 14 && path[:14] == "/system/group/":
		return EventGroupUpdate, "更新分组"
	case method == "DELETE" && len(path) > 14 && path[:14] == "/system/group/":
		return EventGroupDelete, "删除分组"

	default:
		return "", ""
	}
}
