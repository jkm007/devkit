/**
 * 路由拦截器
 * 未登录时自动跳转登录页
 */

// 不需要登录就能访问的页面白名单
const WHITE_LIST: string[] = [
  '/pages/login/index',
  '/pages/login/register',
  '/pages/login/forget',
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
 * 检查是否已登录
 */
function isLoggedIn(): boolean {
  return !!uni.getStorageSync('access_token');
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
        uni.redirectTo({ url: '/pages/login/index' });
        return false;
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
        uni.redirectTo({ url: '/pages/login/index' });
        return false;
      }
      return true;
    },
    fail(err) {
      console.error('switchTab interceptor error:', err);
    },
  });
}
