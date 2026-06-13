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
export function refreshToken(refreshToken: string) {
  return request.post<LoginResponse>('/auth/refresh', { refreshToken });
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
