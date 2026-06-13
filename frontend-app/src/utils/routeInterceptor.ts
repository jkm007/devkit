/**
 * 路由拦截器
 * 未登录时自动跳转登录页
 */
import { tokenManager } from '@/api/request';

// 不需要登录就能访问的页面白名单
const WHITE_LIST: string[] = [
  '/pages/login/index',
  '/pages/login/register',
  '/pages/login/forget',
  '/pages/share/view',
  '/pages/question/detail',
  '/pages/question/list',
  '/pages/question/search',
  '/pages/index/index',
];

/**
 * 检查是否需要登录
 */
function requiresLogin(pagePath: string): boolean {
  // 标准化路径
  const normalizedPath = pagePath.startsWith('/') ? pagePath : `/${pagePath}`;
  return !WHITE_LIST.includes(normalizedPath);
}

/**
 * 检查是否已登录（使用 TokenManager 统一来源）
 */
function isLoggedIn(): boolean {
  return !!tokenManager.getAccessToken();
}

/**
 * 注册路由拦截器
 */
export function setupRouteInterceptor() {
  uni.addInterceptor('navigateTo', {
    invoke(args) {
      const { url } = args;
      const pagePath = url.split('?')[0];

      if (requiresLogin(pagePath) && !isLoggedIn()) {
        uni.redirectTo({ url: '/pages/login/index' });
        return false;
      }
      return true;
    },
    fail(err) {
      console.error('navigateTo interceptor error:', err);
    },
  });

  uni.addInterceptor('redirectTo', {
    invoke(args) {
      const { url } = args;
      const pagePath = url.split('?')[0];

      if (requiresLogin(pagePath) && !isLoggedIn()) {
        uni.redirectTo({ url: '/pages/login/index' });
        return false;
      }
      return true;
    },
    fail(err) {
      console.error('redirectTo interceptor error:', err);
    },
  });

  uni.addInterceptor('reLaunch', {
    invoke(args) {
      const { url } = args;
      const pagePath = url.split('?')[0];

      if (requiresLogin(pagePath) && !isLoggedIn()) {
        // 直接修改 URL 到登录页，避免嵌套导航
        args.url = '/pages/login/index';
        return true;
      }
      return true;
    },
    fail(err) {
      console.error('reLaunch interceptor error:', err);
    },
  });

  uni.addInterceptor('switchTab', {
    invoke(args) {
      const { url } = args;
      const pagePath = url.split('?')[0];

      if (requiresLogin(pagePath) && !isLoggedIn()) {
        args.url = '/pages/login/index';
        return true;
      }
      return true;
    },
    fail(err) {
      console.error('switchTab interceptor error:', err);
    },
  });
}
