import { request } from './request';

/**
 * 登录响应
 */
export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

/**
 * 用户名密码登录
 */
export function loginByUsername(username: string, password: string, captchaId?: string, captchaCode?: string) {
  return request.post<LoginResponse>('/auth/login', {
    username,
    password,
    captchaId,
    captchaCode,
  });
}

/**
 * 邮箱验证码登录
 */
export function loginByEmail(email: string, code: string) {
  return request.post<LoginResponse>('/auth/login-by-email', { email, code });
}

/**
 * 手机验证码登录
 */
export function loginByPhone(phone: string, code: string) {
  return request.post<LoginResponse>('/auth/login-by-phone', { phone, code });
}

/**
 * 刷新 Token
 */
export function refreshTokenApi(token: string) {
  return request.post<LoginResponse>('/auth/refresh', { refreshToken: token });
}

/**
 * 用户注册
 */
export function register(params: {
  username: string;
  password: string;
  email?: string;
  phone?: string;
}) {
  return request.post('/auth/register', params);
}

/**
 * 发送验证码
 */
export function sendCode(email: string) {
  return request.post('/auth/send-code', { email });
}

/**
 * 获取验证码图片
 */
export function getCaptcha() {
  return request.get<{ id: string; image: string }>('/auth/captcha');
}

/**
 * 登出
 */
export function logout() {
  return request.post('/auth/logout');
}

/**
 * 发送邮箱验证码
 */
export function sendVerifyCode(email: string) {
  return request.post('/auth/send-code', { email });
}

/**
 * 验证邮箱验证码
 */
export function verifyCode(email: string, code: string) {
  return request.post('/auth/verify-code', { email, code });
}

/**
 * 重置密码（忘记密码）
 */
export function resetPassword(params: {
  email: string;
  code: string;
  newPassword: string;
}) {
  return request.post('/auth/reset-password', params);
}

/**
 * 修改密码（已登录）
 */
export function changePassword(params: {
  oldPassword: string;
  newPassword: string;
}) {
  return request.put('/auth/change-password', params);
}

/**
 * 获取用户信息
 */
export function getUserInfo() {
  return request.get<any>('/user/info');
}

/**
 * 更新用户信息
 */
export function updateUserInfo(params: { nickname?: string; avatar?: string }) {
  return request.put('/user/info', params);
}

