import type { Router } from 'vue-router';

import { LOGIN_PATH } from '@vben/constants';
import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { startProgress, stopProgress } from '@vben/utils';

import { accessRoutes, publicRouteNames } from '#/router/routes';
import { useAuthStore } from '#/store';

import { generateAccess } from './access';

/**
 * 验证重定向 URL 是否安全（防止开放重定向攻击）
 * 只允许同源路径，拒绝外部 URL 和协议级重定向
 */
function isValidRedirect(url: string): string {
  if (!url) return '';
  // 二次解码，防止双重编码绕过校验（如 https%253A%252F%252Fevil.com）
  let decoded = url;
  try {
    decoded = decodeURIComponent(decoded);
    decoded = decodeURIComponent(decoded);
  } catch {
    // 解码失败则使用已有结果，后续校验仍会拦截
  }
  // 拒绝外部协议（http://, https://, javascript:, data: 等）
  if (/^(https?:|javascript:|data:|vbscript:)/i.test(decoded)) {
    return '';
  }
  // 拒绝协议相对路径（//evil.com）
  if (decoded.startsWith('//')) {
    return '';
  }
  // 只允许 / 开头的同源路径
  if (decoded.startsWith('/')) {
    return decoded;
  }
  return '';
}

/**
 * 通用守卫配置
 * @param router
 */
function setupCommonGuard(router: Router) {
  // 记录已经加载的页面
  const loadedPaths = new Set<string>();

  router.beforeEach((to) => {
    to.meta.loaded = loadedPaths.has(to.path);

    // 页面加载进度条
    if (!to.meta.loaded && preferences.transition.progress) {
      startProgress();
    }
    return true;
  });

  router.afterEach((to) => {
    // 记录页面是否加载,如果已经加载，后续的页面切换动画等效果不在重复执行

    loadedPaths.add(to.path);

    // 关闭页面加载进度条
    if (preferences.transition.progress) {
      stopProgress();
    }
  });
}

/**
 * 权限访问守卫配置
 * @param router
 */
function setupAccessGuard(router: Router) {
  router.beforeEach(async (to, from) => {
    const accessStore = useAccessStore();
    const userStore = useUserStore();
    const authStore = useAuthStore();

    // 优先检查 ignoreAccess meta（公开页面如分享页面）
    if (to.meta.ignoreAccess) {
      return true;
    }

    // 公开路由（基本路由+外部路由），这些路由不需要进入权限拦截
    if (publicRouteNames.includes(to.name as string)) {
      if (to.path === LOGIN_PATH && accessStore.accessToken) {
        const redirectUrl = isValidRedirect(
          decodeURIComponent(
            (to.query?.redirect as string) ||
              userStore.userInfo?.homePath ||
              preferences.app.defaultHomePath,
          ),
        );
        return (
          redirectUrl ||
          userStore.userInfo?.homePath ||
          preferences.app.defaultHomePath
        );
      }
      return true;
    }

    // accessToken 检查
    if (!accessStore.accessToken) {
      // 没有访问权限，跳转登录页面
      if (to.fullPath !== LOGIN_PATH) {
        return {
          path: LOGIN_PATH,
          // 如不需要，直接删除 query
          query:
            to.fullPath === preferences.app.defaultHomePath
              ? {}
              : { redirect: encodeURIComponent(to.fullPath) },
          // 携带当前跳转的页面，登录后重新跳转该页面
          replace: true,
        };
      }
      return to;
    }

    // 是否已经生成过动态路由
    if (accessStore.isAccessChecked) {
      // 确保权限版本检查已启动（页面刷新后恢复）
      authStore.checkPermissionVersion();
      return true;
    }

    // 生成路由表
    // 当前登录用户拥有的角色标识列表
    const userInfo = userStore.userInfo || (await authStore.fetchUserInfo());
    const userRoles = userInfo.roles ?? [];

    // 生成菜单和路由
    const { accessibleMenus, accessibleRoutes } = await generateAccess({
      roles: userRoles,
      router,
      // 则会在菜单中显示，但是访问会被重定向到403
      routes: accessRoutes,
    });

    // 保存菜单信息和路由信息
    accessStore.setAccessMenus(accessibleMenus);
    accessStore.setAccessRoutes(accessibleRoutes);
    accessStore.setIsAccessChecked(true);
    const redirectPath = (from.query.redirect ??
      (to.path === preferences.app.defaultHomePath
        ? userInfo.homePath || preferences.app.defaultHomePath
        : to.fullPath)) as string;

    return {
      ...router.resolve(decodeURIComponent(redirectPath)),
      replace: true,
    };
  });
}

/**
 * 项目守卫配置
 * @param router
 */
function createRouterGuard(router: Router) {
  /** 通用 */
  setupCommonGuard(router);
  /** 权限访问 */
  setupAccessGuard(router);
}

export { createRouterGuard };
