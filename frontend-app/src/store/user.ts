import { defineStore } from 'pinia';
import { ref } from 'vue';
import { loginByUsername, loginByEmail, loginByPhone, loginByMiniProgram, logout as logoutApi, refreshTokenApi } from '@/api/auth';
import { tokenManager } from '@/api/request';

/**
 * 用户信息接口
 */
export interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  avatar: string;
  email: string;
  phone: string;
  roles: string[];
}

/**
 * 用户状态 Store
 */
export const useUserStore = defineStore('user', () => {
  // 状态
  const userInfo = ref<UserInfo | null>(null);
  const loginLoading = ref(false);

  /**
   * 登录（用户名密码）
   */
  async function loginWithPassword(username: string, password: string, captchaId?: string, captchaCode?: string) {
    loginLoading.value = true;
    try {
      const res = await loginByUsername(username, password, captchaId, captchaCode);
      tokenManager.setTokens(res.accessToken, res.refreshToken);
      return res;
    } finally {
      loginLoading.value = false;
    }
  }

  /**
   * 登录（邮箱验证码）
   */
  async function loginWithEmail(email: string, code: string) {
    loginLoading.value = true;
    try {
      const res = await loginByEmail(email, code);
      tokenManager.setTokens(res.accessToken, res.refreshToken);
      return res;
    } finally {
      loginLoading.value = false;
    }
  }

  /**
   * 登录（手机验证码）
   */
  async function loginWithPhone(phone: string, code: string) {
    loginLoading.value = true;
    try {
      const res = await loginByPhone(phone, code);
      tokenManager.setTokens(res.accessToken, res.refreshToken);
      return res;
    } finally {
      loginLoading.value = false;
    }
  }

  /**
   * 登录（微信小程序）
   */
  async function loginWithMiniProgram() {
    loginLoading.value = true;
    try {
      // 调用微信小程序登录获取 code
      const loginResult = await new Promise<UniApp.LoginRes>((resolve, reject) => {
        uni.login({
          provider: 'weixin',
          success: resolve,
          fail: reject,
        });
      });

      if (!loginResult.code) {
        throw new Error('获取登录凭证失败');
      }

      console.log('[微信登录] 获取到 code:', loginResult.code);

      // 用 code 换取后端 token
      const res = await loginByMiniProgram(loginResult.code);

      console.log('[微信登录] 后端响应:', JSON.stringify(res));
      console.log('[微信登录] accessToken:', res?.accessToken ? '存在' : '不存在');
      console.log('[微信登录] refreshToken:', res?.refreshToken ? '存在' : '不存在');

      if (!res?.accessToken) {
        throw new Error('登录响应中缺少 accessToken');
      }

      tokenManager.setTokens(res.accessToken, res.refreshToken);

      console.log('[微信登录] token 已保存，读取测试:', tokenManager.getAccessToken() ? '成功' : '失败');

      return res;
    } finally {
      loginLoading.value = false;
    }
  }

  /**
   * 登出
   */
  async function logout() {
    try {
      await logoutApi();
    } catch {
      // 忽略登出错误
    } finally {
      tokenManager.clearTokens();
      userInfo.value = null;
      uni.reLaunch({ url: '/pages/login/index' });
    }
  }

  /**
   * 检查登录状态
   */
  function isLoggedIn(): boolean {
    return !!tokenManager.getAccessToken();
  }

  return {
    userInfo,
    loginLoading,
    loginWithPassword,
    loginWithEmail,
    loginWithPhone,
    loginWithMiniProgram,
    logout,
    isLoggedIn,
  };
});
