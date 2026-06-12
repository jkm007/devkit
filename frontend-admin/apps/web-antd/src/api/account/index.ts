import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export * from './role-application';

// ==================== 类型定义 ====================

export namespace AccountApi {
  /** 用户信息 */
  export interface UserInfo {
    id: number;
    username: string;
    nickname: string;
    realName: string;
    email: string;
    phone: string;
    avatar: string;
    gender: number;
    birthday: string;
    bio: string;
    roles: string[];
    homePath: string | null;
    isReal: number;
    registerSource: string;
    lastLoginAt: string;
    passwordChangedAt: string;
  }

  /** 更新用户资料请求 */
  export interface UpdateProfileRequest {
    nickname?: string;
    email?: string;
    phone?: string;
    gender?: number;
    birthday?: string;
    bio?: string;
  }

  /** 修改密码请求 */
  export interface ChangePasswordRequest {
    oldPassword: string;
    newPassword: string;
    confirmPassword: string;
    captchaId?: string;
    captchaCode?: string;
  }

  export type DeviceType = 'app' | 'h5' | 'miniapp' | 'web';

  /** 登录设备 */
  export interface LoginDevice {
    id: number;
    deviceId: string;
    deviceType: string;
    deviceName: string;
    browser: string;
    os: string;
    ip: string;
    location: string;
    appVersion: string;
    systemVersion: string;
    deviceModel: string;
    platform: string;
    channel: string;
    lastActiveAt: string;
    isCurrent: boolean;
    createdAt: string;
  }

  /** OAuth绑定 */
  export interface OAuthBinding {
    id: number;
    provider: string;
    providerUsername: string;
    providerAvatar: string;
    createdAt: string;
  }

  /** 实名认证状态 */
  export interface RealNameStatus {
    status: number;
    realName: string;
    idCard: string;
    rejectReason: string | null;
    submittedAt: string;
    reviewedAt: string;
  }

  /** 提交实名认证请求 */
  export interface SubmitRealNameRequest {
    realName: string;
    idCard: string;
  }

  /** 隐私设置 */
  export interface PrivacySettings {
    profileVisible: number;
    realnameVisible: number;
    emailVisible: number;
    statsVisible: number;
    classVisible: number;
  }

  /** 安全日志 */
  export interface SecurityLog {
    id: number;
    eventType: string;
    eventDetail: string;
    ip: string;
    userAgent: string;
    status: number;
    createdAt: string;
  }
}

// ==================== 个人信息 ====================

/** 获取当前用户详细信息 */
export function getUserInfo() {
  return requestClient.get<AccountApi.UserInfo>('/user/info');
}

/** 更新当前用户资料 */
export function updateProfile(data: AccountApi.UpdateProfileRequest) {
  return requestClient.put('/user/info', data);
}

/** 更新头像 */
export function updateAvatar(data: { avatar: string }) {
  return requestClient.put('/user/info', data);
}

// ==================== 修改密码 ====================

/** 修改密码 */
export function changePassword(data: AccountApi.ChangePasswordRequest) {
  return requestClient.put('/auth/change-password', data);
}

/** 检查密码是否与历史重复 */
export function checkPasswordHistory(data: { newPassword: string }) {
  return requestClient.post<{ isRepeated: boolean }>(
    '/auth/password-history/check',
    data,
  );
}

// ==================== 登录设备 ====================

/** 获取登录设备列表 */
export function getLoginDevices(params?: { deviceType?: AccountApi.DeviceType }) {
  return requestClient.get<AccountApi.LoginDevice[]>('/auth/devices', {
    params,
  });
}

/** 踢出指定设备 */
export function kickDevice(id: number) {
  return requestClient.delete(`/auth/devices/${id}`);
}

/** 踢出所有其他设备 */
export function kickAllOtherDevices() {
  return requestClient.delete<{ kickedCount: number }>(
    '/auth/devices/kick-all',
  );
}

// ==================== OAuth绑定 ====================

/** 获取第三方绑定列表 */
export function getOAuthBindings() {
  return requestClient.get<AccountApi.OAuthBinding[]>('/auth/oauth/bindings');
}

/** 获取第三方授权URL */
export function getOAuthBindUrl(params: {
  provider: string;
  redirectUri?: string;
}) {
  return requestClient.get<{ url: string }>('/auth/oauth/bind-url', {
    params,
  });
}

/** 解绑第三方账号 */
export function unbindOAuth(data: { provider: string }) {
  return requestClient.post('/auth/oauth/unbind', data);
}

// ==================== 实名认证 ====================

/** 获取当前用户实名认证状态 */
export function getRealNameStatus() {
  return requestClient.get<AccountApi.RealNameStatus>('/user/real-name');
}

/** 提交实名认证申请 */
export function submitRealName(data: AccountApi.SubmitRealNameRequest) {
  return requestClient.post('/user/real-name', data);
}

// ==================== 隐私设置 ====================

/** 获取隐私设置 */
export function getPrivacySettings() {
  return requestClient.get<AccountApi.PrivacySettings>('/user/privacy');
}

/** 更新隐私设置 */
export function updatePrivacySettings(
  data: Partial<AccountApi.PrivacySettings>,
) {
  return requestClient.put('/user/privacy', data);
}

// ==================== 安全日志（个人） ====================

/** 获取当前用户安全日志 */
export function getMySecurityLogs(params: Recordable<any>) {
  return requestClient.get<{ items: AccountApi.SecurityLog[]; total: number }>(
    '/auth/security-logs',
    { params },
  );
}
