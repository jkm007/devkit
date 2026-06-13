/**
 * 平台适配工具
 *
 * uni-app 条件编译指南：
 * - #ifdef H5 / #endif：仅在 H5 生效
 * - #ifdef MP-WEIXIN / #endif：仅在微信小程序生效
 * - #ifdef APP-PLUS / #endif：仅在 App 生效
 *
 * 本文件提供跨平台统一接口
 */

/**
 * 获取平台信息
 */
export function getPlatform(): 'h5' | 'weapp' | 'app' {
  // #ifdef H5
  return 'h5';
  // #endif
  // #ifdef MP-WEIXIN
  return 'weapp';
  // #endif
  // #ifdef APP-PLUS
  return 'app';
  // #endif
  // #ifndef H5 || MP-WEIXIN || APP-PLUS
  return 'h5';
  // #endif
}

/**
 * 是否是 H5 环境
 */
export function isH5(): boolean {
  return getPlatform() === 'h5';
}

/**
 * 是否是微信小程序
 */
export function isWeapp(): boolean {
  return getPlatform() === 'weapp';
}

/**
 * 是否是 App
 */
export function isApp(): boolean {
  return getPlatform() === 'app';
}

/**
 * 获取状态栏高度
 */
export async function getStatusBarHeight(): Promise<number> {
  return new Promise((resolve) => {
    // #ifdef MP-WEIXIN
    const systemInfo = uni.getSystemInfoSync();
    resolve(systemInfo.statusBarHeight || 0);
    // #endif
    // #ifdef H5
    resolve(0); // H5 使用默认导航栏
    // #endif
    // #ifdef APP-PLUS
    const appSystemInfo = uni.getSystemInfoSync();
    resolve(appSystemInfo.statusBarHeight || 0);
    // #endif
  });
}

/**
 * 获取导航栏高度（状态栏 + 标题栏）
 */
export function getNavBarHeight(): number {
  const systemInfo = uni.getSystemInfoSync();
  const statusBarHeight = systemInfo.statusBarHeight || 0;
  // 微信小程序标题栏高度固定为 44px，H5 由框架处理
  const titleBarHeight = isWeapp() ? 44 : 0;
  return statusBarHeight + titleBarHeight;
}

/**
 * 复制到剪贴板（跨平台）
 */
export function copyToClipboard(text: string): Promise<void> {
  return new Promise((resolve, reject) => {
    // #ifdef H5
    if (navigator.clipboard) {
      navigator.clipboard.writeText(text).then(resolve, reject);
    } else {
      // 降级方案
      const textarea = document.createElement('textarea');
      textarea.value = text;
      document.body.appendChild(textarea);
      textarea.select();
      try {
        document.execCommand('copy');
        resolve();
      } catch {
        reject(new Error('复制失败'));
      }
      document.body.removeChild(textarea);
    }
    // #endif
    // #ifndef H5
    uni.setClipboardData({
      data: text,
      success: () => resolve(),
      fail: () => reject(new Error('复制失败')),
    });
    // #endif
  });
}

/**
 * 分享（跨平台）
 */
export function share(options: {
  title: string;
  content?: string;
  url?: string;
  imageUrl?: string;
}): Promise<void> {
  return new Promise((resolve, reject) => {
    // #ifdef MP-WEIXIN
    // 微信小程序需要用户主动触发分享，此处只提供提示
    uni.showToast({ title: '请使用右上角分享', icon: 'none' });
    resolve();
    // #endif
    // #ifdef H5
    if (navigator.share) {
      navigator.share({
        title: options.title,
        text: options.content,
        url: options.url,
      }).then(resolve, reject);
    } else {
      // 降级：复制链接
      const shareUrl = options.url || window.location.href;
      copyToClipboard(shareUrl).then(resolve, reject);
      uni.showToast({ title: '链接已复制', icon: 'success' });
    }
    // #endif
    // #ifdef APP-PLUS
    uni.share({
      provider: 'weixin',
      type: 0,
      title: options.title,
      summary: options.content,
      href: options.url,
      imageUrl: options.imageUrl,
      success: () => resolve(),
      fail: () => reject(new Error('分享失败')),
    });
    // #endif
  });
}

/**
 * 保存图片到相册
 */
export function saveImageToAlbum(url: string): Promise<void> {
  return new Promise((resolve, reject) => {
    // #ifdef MP-WEIXIN
    uni.downloadFile({
      url,
      success: (res) => {
        uni.saveImageToPhotosAlbum({
          filePath: res.tempFilePath,
          success: () => resolve(),
          fail: () => reject(new Error('保存失败')),
        });
      },
      fail: () => reject(new Error('下载失败')),
    });
    // #endif
    // #ifdef H5
    // H5 无法直接保存，提供长按提示
    uni.showToast({ title: '请长按图片保存', icon: 'none' });
    resolve();
    // #endif
    // #ifdef APP-PLUS
    uni.downloadFile({
      url,
      success: (res) => {
        uni.saveImageToPhotosAlbum({
          filePath: res.tempFilePath,
          success: () => resolve(),
          fail: () => reject(new Error('保存失败')),
        });
      },
      fail: () => reject(new Error('下载失败')),
    });
    // #endif
  });
}

/**
 * 获取安全区域高度（适配刘海屏/底部安全区）
 */
export function getSafeArea(): { top: number; bottom: number } {
  const systemInfo = uni.getSystemInfoSync();
  const safeArea = systemInfo.safeArea;
  return {
    top: safeArea?.top || 0,
    bottom: systemInfo.screenHeight - (safeArea?.bottom || systemInfo.screenHeight),
  };
}

/**
 * 平台差异化样式类名
 */
export function platformClass(base: string): string {
  const platform = getPlatform();
  return `${base} ${base}--${platform}`;
}
