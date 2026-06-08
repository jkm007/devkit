import type { Recordable, UserInfo } from '@vben/types';

import type { AuthApi } from '#/api/core/auth';
import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { LOGIN_PATH } from '@vben/constants';
import { preferences } from '@vben/preferences';
import { resetAllStores, useAccessStore, useUserStore } from '@vben/stores';

import { Modal, notification } from 'ant-design-vue';
import { defineStore } from 'pinia';

import {
  getAccessCodesApi,
  getPermissionVersionApi,
  getUserInfoApi,
  loginApi,
  loginByEmailApi,
  loginByPhoneApi,
  logoutApi,
} from '#/api';
import { $t } from '#/locales';

export const useAuthStore = defineStore('auth', () => {
  const accessStore = useAccessStore();
  const userStore = useUserStore();
  const router = useRouter();

  const loginLoading = ref(false);
  // 权限版本（内存中，不持久化，用于比对）
  const permissionVersion = ref('');
  let permissionCheckTimer: ReturnType<typeof setInterval> | null = null;

  /**
   * 登录后通用处理：存储 token、获取用户信息、跳转
   */
  async function handleLoginResult(
    result: AuthApi.LoginResult,
    onSuccess?: () => Promise<void> | void,
  ): Promise<{ userInfo: null | UserInfo }> {
    let userInfo: null | UserInfo = null;
    const { accessToken, refreshToken } = result;

    if (!accessToken) return { userInfo: null };

    accessStore.setAccessToken(accessToken);
    if (refreshToken) {
      accessStore.setRefreshToken(refreshToken);
    }

    // 获取用户信息、权限码、权限版本
    const [fetchUserInfoResult, accessCodes, permVersion] =
      await Promise.all([
        fetchUserInfo(),
        getAccessCodesApi(),
        getPermissionVersionApi(),
      ]);

    userInfo = fetchUserInfoResult;
    userStore.setUserInfo(userInfo);
    accessStore.setAccessCodes(accessCodes);
    permissionVersion.value = permVersion;

    // 启动权限版本检查
    startPermissionCheck();

    if (accessStore.loginExpired) {
      accessStore.setLoginExpired(false);
    } else {
      onSuccess
        ? await onSuccess?.()
        : await router.push(
            userInfo.homePath || preferences.app.defaultHomePath,
          );
    }

    if (userInfo?.realName) {
      notification.success({
        description: `${$t('authentication.loginSuccessDesc')}:${userInfo?.realName}`,
        duration: 3,
        message: $t('authentication.loginSuccess'),
      });
    }

    return { userInfo };
  }

  /**
   * 用户名密码登录
   */
  async function authLogin(
    params: Recordable<any>,
    onSuccess?: () => Promise<void> | void,
  ) {
    try {
      loginLoading.value = true;
      const result = await loginApi(params);
      return await handleLoginResult(result, onSuccess);
    } catch (error) {
      console.error('Login failed:', error);
      return { userInfo: null };
    } finally {
      loginLoading.value = false;
    }
  }

  /**
   * 邮箱验证码登录
   */
  async function authLoginByEmail(
    params: { code: string; email: string },
    onSuccess?: () => Promise<void> | void,
  ) {
    try {
      loginLoading.value = true;
      const result = await loginByEmailApi(params);
      return await handleLoginResult(result, onSuccess);
    } catch (error) {
      console.error('Email login failed:', error);
      return { userInfo: null };
    } finally {
      loginLoading.value = false;
    }
  }

  /**
   * 手机号验证码登录
   */
  async function authLoginByPhone(
    params: { code: string; phone: string },
    onSuccess?: () => Promise<void> | void,
  ) {
    try {
      loginLoading.value = true;
      const result = await loginByPhoneApi(params);
      return await handleLoginResult(result, onSuccess);
    } catch (error) {
      console.error('Phone login failed:', error);
      return { userInfo: null };
    } finally {
      loginLoading.value = false;
    }
  }

  async function logout(redirect: boolean = true) {
    try {
      await logoutApi();
    } catch {
      // 不做任何处理
    }
    stopPermissionCheck();
    resetAllStores();
    accessStore.setLoginExpired(false);
    permissionVersion.value = '';

    // 回登录页带上当前路由地址
    await router.replace({
      path: LOGIN_PATH,
      query: redirect
        ? {
            redirect: encodeURIComponent(router.currentRoute.value.fullPath),
          }
        : {},
    });
  }

  async function fetchUserInfo() {
    const userInfo = await getUserInfoApi();
    userStore.setUserInfo(userInfo);
    return userInfo;
  }

  /**
   * 检查权限版本是否变更
   * 如果变更，提示用户刷新页面
   */
  async function checkPermissionVersion() {
    if (!accessStore.accessToken) return;
    try {
      const latestVersion = await getPermissionVersionApi();
      if (!permissionVersion.value) {
        // 页面刷新后首次检测，初始化版本号（不弹窗）
        permissionVersion.value = latestVersion;
      } else if (latestVersion !== permissionVersion.value) {
        // 权限已变更，提示用户
        stopPermissionCheck();
        Modal.confirm({
          title: '权限变更提示',
          content: '您的权限已被管理员修改，需要刷新页面以获取最新权限。',
          okText: '刷新页面',
          cancelText: '稍后',
          onOk() {
            refreshPermissions();
          },
          onCancel() {
            startPermissionCheck();
          },
        });
      }
    } catch {
      // 检查失败不影响正常使用
    }
  }

  /**
   * 刷新权限码并更新版本
   */
  async function refreshPermissions() {
    try {
      const [accessCodes, permVersion] = await Promise.all([
        getAccessCodesApi(),
        getPermissionVersionApi(),
      ]);
      accessStore.setAccessCodes(accessCodes);
      permissionVersion.value = permVersion;
      // 刷新页面让路由和菜单重新加载
      window.location.reload();
    } catch (error) {
      console.error('Failed to refresh permissions:', error);
    }
  }

  /**
   * 启动权限版本定期检查（每 30 秒）
   */
  function startPermissionCheck() {
    stopPermissionCheck();
    permissionCheckTimer = setInterval(checkPermissionVersion, 30_000);
  }

  /**
   * 停止权限版本检查
   */
  function stopPermissionCheck() {
    if (permissionCheckTimer) {
      clearInterval(permissionCheckTimer);
      permissionCheckTimer = null;
    }
  }

  function $reset() {
    loginLoading.value = false;
    stopPermissionCheck();
  }

  return {
    $reset,
    authLogin,
    authLoginByEmail,
    authLoginByPhone,
    checkPermissionVersion,
    fetchUserInfo,
    loginLoading,
    logout,
    refreshPermissions,
  };
});
