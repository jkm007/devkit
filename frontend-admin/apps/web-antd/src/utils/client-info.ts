/**
 * 客户端信息工具 - 用于 H5/App/小程序 环境检测和设备信息采集
 */

export type ClientType = 'app' | 'h5' | 'miniapp' | 'web';

/**
 * 检测客户端类型
 * 优先级：URL参数 > UserAgent检测 > 默认web
 */
export function detectClientType(): ClientType {
  // 1. URL 参数指定（用于测试或 WebView 嵌入）
  const params = new URLSearchParams(window.location.search);
  const urlType = params.get('clientType') as ClientType;
  if (['web', 'h5', 'app', 'miniapp'].includes(urlType)) {
    return urlType;
  }

  // 2. 环境变量指定
  const envType = import.meta.env.VITE_CLIENT_TYPE as ClientType;
  if (envType && ['web', 'h5', 'app', 'miniapp'].includes(envType)) {
    return envType;
  }

  // 3. UserAgent 检测
  const ua = navigator.userAgent.toLowerCase();

  // 微信小程序 WebView
  if (ua.includes('miniprogram') || (window as any).__wxjs_environment === 'miniprogram') {
    return 'miniapp';
  }

  // App WebView（常见 App 标识）
  if (ua.includes('devkit-app') || ua.includes('hybrid') || (window as any).__APP_BRIDGE__) {
    return 'app';
  }

  // H5 移动端浏览器
  if (/android|iphone|ipad|ipod|mobile|webos|blackberry|iemobile|opera mini/i.test(ua)) {
    return 'h5';
  }

  return 'web';
}

/**
 * 获取设备元信息（用于移动端上报）
 */
export function getDeviceMeta() {
  const ua = navigator.userAgent;

  return {
    appVersion: (window as any).__APP_VERSION__ || '',
    systemVersion: getOSVersion(ua),
    deviceModel: getDeviceModel(ua),
    platform: getPlatform(ua),
    channel: (window as any).__APP_CHANNEL__ || '',
  };
}

/**
 * 获取操作系统版本
 */
function getOSVersion(ua: string): string {
  // iOS
  const iosMatch = ua.match(/OS (\d+_\d+(?:_\d+)?)/);
  if (iosMatch) {
    return `iOS ${(iosMatch[1] || '').replace(/_/g, '.')}`;
  }

  // Android
  const androidMatch = ua.match(/Android (\d+(?:\.\d+)*)/);
  if (androidMatch) {
    return `Android ${androidMatch[1]}`;
  }

  // Windows
  const windowsMatch = ua.match(/Windows NT (\d+\.\d+)/);
  if (windowsMatch) {
    const versions: Record<string, string> = {
      '10.0': '10',
      '6.3': '8.1',
      '6.2': '8',
      '6.1': '7',
    };
    return `Windows ${versions[windowsMatch[1] || ''] || windowsMatch[1] || ''}`;
  }

  // macOS
  const macMatch = ua.match(/Mac OS X (\d+_\d+(?:_\d+)?)/);
  if (macMatch) {
    return `macOS ${(macMatch[1] || '').replace(/_/g, '.')}`;
  }

  return '';
}

/**
 * 获取设备型号
 */
function getDeviceModel(ua: string): string {
  // iPhone
  if (ua.includes('iphone')) return 'iPhone';

  // iPad
  if (ua.includes('ipad')) return 'iPad';

  // Android 设备
  const androidMatch = ua.match(/;\s*(\w[\w\s-]*?)\s*(?:Build|[;)])/);
  if (androidMatch) {
    return (androidMatch[1] || '').trim();
  }

  return '';
}

/**
 * 获取平台标识
 */
function getPlatform(ua: string): string {
  if (/iphone|ipad|ipod/i.test(ua)) return 'ios';
  if (/android/i.test(ua)) return 'android';
  if (/windows/i.test(ua)) return 'windows';
  if (/macintosh|mac os x/i.test(ua)) return 'macos';
  if (/linux/i.test(ua)) return 'linux';
  return 'web';
}

/**
 * 获取设备信息摘要（用于显示）
 */
export function getDeviceSummary(): string {
  const ua = navigator.userAgent;
  const parts: string[] = [];

  const browser = getBrowserName(ua);
  if (browser) parts.push(browser);

  const os = getOSVersion(ua);
  if (os) parts.push(os);

  return parts.join(' · ') || '未知设备';
}

/**
 * 获取浏览器名称
 */
function getBrowserName(ua: string): string {
  if (ua.includes('Firefox/')) return 'Firefox';
  if (ua.includes('Edg/')) return 'Edge';
  if (ua.includes('Chrome/')) return 'Chrome';
  if (ua.includes('Safari/') && ua.includes('Version/')) return 'Safari';
  if (ua.includes('OPR/') || ua.includes('Opera/')) return 'Opera';
  return '';
}
