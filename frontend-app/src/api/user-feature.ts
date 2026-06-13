/**
 * 用户功能 API（设备/隐私/OAuth/安全日志等）
 */
import { request } from './request';

/**
 * 获取登录设备列表
 */
export function getDevices() {
  return request.get<any[]>('/api/v1/auth/devices');
}

/**
 * 踢出其他设备
 */
export function kickAllOtherDevices() {
  return request.delete('/api/v1/auth/devices/kick-all');
}

/**
 * 踢出指定设备
 */
export function kickDevice(id: number) {
  return request.delete(`/api/v1/auth/devices/${id}`);
}

/**
 * 获取隐私设置
 */
export function getPrivacy() {
  return request.get<any>('/api/v1/user/privacy');
}

/**
 * 更新隐私设置
 */
export function updatePrivacy(params: {
  showProfile?: boolean;
  showStats?: boolean;
}) {
  return request.put('/api/v1/user/privacy', params);
}

/**
 * 获取 OAuth 绑定列表
 */
export function getOAuthBindings() {
  return request.get<any[]>('/api/v1/auth/oauth/bindings');
}

/**
 * 获取 OAuth 绑定 URL
 */
export function getOAuthBindUrl(provider: string) {
  return request.get<{ url: string }>('/api/v1/auth/oauth/bind-url', { params: { provider } });
}

/**
 * 解绑 OAuth
 */
export function unbindOAuth(provider: string) {
  return request.post('/api/v1/auth/oauth/unbind', { provider });
}

/**
 * 获取安全日志
 */
export function getSecurityLogs(params: { page?: number; pageSize?: number }) {
  return request.get<{ items: any[]; total: number }>('/api/v1/auth/security-logs', { params });
}

/**
 * 获取角色申请可用角色
 */
export function getAvailableRoles() {
  return request.get<any[]>('/api/v1/auth/role-applications/available-roles');
}

/**
 * 提交角色申请
 */
export function submitRoleApplication(params: {
  roleId: number;
  reason?: string;
}) {
  return request.post('/api/v1/auth/role-applications', params);
}

/**
 * 获取我的角色申请
 */
export function getMyRoleApplications(params: { page?: number; pageSize?: number }) {
  return request.get<{ items: any[]; total: number }>('/api/v1/auth/role-applications', { params });
}
