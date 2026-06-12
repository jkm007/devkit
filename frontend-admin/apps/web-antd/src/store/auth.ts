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
  // 权限版本检查定时器（纳入 store state，便于响应式管理和 $reset 时统一清理）
  const permissionCheckTimer = ref<ReturnType<typeof setInterval> | null>(null);
  let isLoggingOut = false;

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
    const [fetchUserInfoResult, accessCodes, permVersion] = await Promise.all([
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
    } catch (error: any) {
      console.error('Login failed:', error);
      // 403001 需要验证码处理，抛出让调用方处理
      if (error?.data?.code === 403001 || error?.code === 403001) {
        throw error;
      }
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
    } catch (error: any) {
      console.error('Email login failed:', error);
      // 403001 需要验证码处理，抛出让调用方处理
      if (error?.data?.code === 403001 || error?.code === 403001) {
        throw error;
      }
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

  async function logout(_redirect: boolean = true) {
    // 防止重复调用
    if (isLoggingOut) return;
    isLoggingOut = true;

    try {
      await logoutApi();
    } catch {
      // 不做任何处理
    }
    stopPermissionCheck();
    permissionVersion.value = '';

    // 构造目标路径
    const fullPath = router.currentRoute.value.fullPath;
    const targetPath =
      fullPath && fullPath !== LOGIN_PATH
        ? `${LOGIN_PATH}?redirect=${encodeURIComponent(fullPath)}`
        : LOGIN_PATH;

    // 先跳转，再重置所有 store，避免 resetAllStores 导致组件卸载取消导航
    try {
      await router.replace(targetPath);
    } catch {
      // router.replace 失败时用硬跳转兜底
    }

    // 跳转后再重置所有 store（不影响导航）
    resetAllStores();
    accessStore.setLoginExpired(false);

    // 如果路由跳转未生效（组件已卸载等），用硬跳转兜底
    if (router.currentRoute.value.fullPath !== LOGIN_PATH) {
      window.location.href = targetPath;
    }

    // 延迟重置标志，允许后续登录后再 logout
    setTimeout(() => {
      isLoggingOut = false;
    }, 1000);
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
    permissionCheckTimer.value = setInterval(checkPermissionVersion, 30_000);
  }

  /**
   * 停止权限版本检查
   */
  function stopPermissionCheck() {
    if (permissionCheckTimer.value) {
      clearInterval(permissionCheckTimer.value);
      permissionCheckTimer.value = null;
    }
  }

  function $reset() {
    loginLoading.value = false;
    stopPermissionCheck();
  }

  return {
    $reset,
    authLogin,
    handleLoginResult,
    authLoginByEmail,
    authLoginByPhone,
    checkPermissionVersion,
    fetchUserInfo,
    loginLoading,
    logout,
    refreshPermissions,
  };
});
