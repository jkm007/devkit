import { baseRequestClient, requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    password?: string;
    username?: string;
  }

  /** 登录接口返回值 */
  export interface LoginResult {
    accessToken: string;
    refreshToken?: string;
  }

  export interface RefreshTokenResult {
    data: string;
    status: number;
  }
}

/**
 * 登录
 */
export async function loginApi(data: AuthApi.LoginParams) {
  return requestClient.post<AuthApi.LoginResult>('/auth/login', data);
}

/**
 * 刷新accessToken
 */
export async function refreshTokenApi() {
  return baseRequestClient.post<AuthApi.RefreshTokenResult>('/auth/refresh', {
    withCredentials: true,
  });
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return baseRequestClient.post('/auth/logout', {
    withCredentials: true,
  });
}

/**
 * 获取用户权限码
 */
export async function getAccessCodesApi() {
  return requestClient.get<string[]>('/auth/codes');
}

/**
 * 获取权限版本（用于检测权限是否变更）
 */
export async function getPermissionVersionApi() {
  return requestClient.get<string>('/auth/permission-version');
}

/**
 * 发送邮箱验证码（需先完成图形验证码）
 */
export async function sendVerifyCodeApi(data: {
  email: string;
  purpose: 'login' | 'register' | 'reset_password';
  captchaId: string;
  captchaCode: string;
}) {
  return requestClient.post('/auth/send-code', data);
}

/**
 * 验证邮箱验证码
 */
export async function verifyEmailCodeApi(email: string, code: string, purpose: 'login' | 'register' | 'reset_password') {
  return requestClient.post('/auth/verify-code', { email, code, purpose });
}

/**
 * 邮箱验证码登录
 */
export async function loginByEmailApi(data: {
  email: string;
  code: string;
}) {
  return requestClient.post<AuthApi.LoginResult>('/auth/login-by-email', data);
}

/**
 * 用户注册
 */
export async function registerApi(data: {
  username: string;
  email: string;
  emailCode: string;
  password: string;
  confirmPassword: string;
}) {
  return requestClient.post('/auth/register', data);
}

/**
 * 重置密码
 */
export async function resetPasswordApi(data: {
  email: string;
  emailCode: string;
  newPassword: string;
  confirmPassword: string;
}) {
  return requestClient.post('/auth/reset-password', data);
}
