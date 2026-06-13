import type { RequestConfig } from './types';
import { enqueueOfflineRequest } from '@/utils/offline';

/**
 * Token 加密层（简单 XOR + Base64 混淆，防止明文存储）
 */
const TOKEN_SALT = 'devkit_token_salt_2026';

function xorWithSalt(text: string, salt: string): string {
  let result = '';
  for (let i = 0; i < text.length; i++) {
    result += String.fromCharCode(text.charCodeAt(i) ^ salt.charCodeAt(i % salt.length));
  }
  return result;
}

function encryptToken(text: string): string {
  try {
    const xored = xorWithSalt(text, TOKEN_SALT);
    return btoa(unescape(encodeURIComponent(xored)));
  } catch {
    return text; // 降级：如果加密失败直接存储原文
  }
}

function decryptToken(encrypted: string): string | null {
  try {
    const decoded = decodeURIComponent(escape(atob(encrypted)));
    return xorWithSalt(decoded, TOKEN_SALT);
  } catch {
    return null;
  }
}

/**
 * 统一响应结构
 */
interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

/**
 * 分页响应结构
 */
interface PageResponse<T = any> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

/**
 * 错误码常量
 */
export const ErrorCode = {
  VALIDATION_ERROR: 40001,
  NOT_FOUND: 40003,
  DELETED: 40004,
  NO_PERMISSION: 40005,
  STATE_ERROR: 40006,
  FILE_DELETED: 40007,
  SHARE_EXPIRED: 40012,
  CAPTCHA_REQUIRED: 403001,
} as const;

/**
 * Token 管理器
 */
class TokenManager {
  private static instance: TokenManager;
  private refreshing = false;
  private refreshPromise: Promise<string | null> | null = null;
  private subscribers: Array<(token: string | null) => void> = [];

  static getInstance(): TokenManager {
    if (!TokenManager.instance) {
      TokenManager.instance = new TokenManager();
    }
    return TokenManager.instance;
  }

  getAccessToken(): string | null {
    const raw = uni.getStorageSync('access_token');
    if (!raw) return null;
    return decryptToken(raw);
  }

  getRefreshToken(): string | null {
    const raw = uni.getStorageSync('refresh_token');
    if (!raw) return null;
    return decryptToken(raw);
  }

  setTokens(accessToken: string, refreshToken?: string) {
    uni.setStorageSync('access_token', encryptToken(accessToken));
    if (refreshToken) {
      uni.setStorageSync('refresh_token', encryptToken(refreshToken));
    }
    this.notifySubscribers(accessToken);
  }

  clearTokens() {
    uni.removeStorageSync('access_token');
    uni.removeStorageSync('refresh_token');
    this.notifySubscribers(null);
  }

  /**
   * 刷新 Token（带队列机制，确保并发请求只刷新一次）
   */
  async refreshToken(): Promise<string | null> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) {
      this.clearTokens();
      return null;
    }

    if (this.refreshing && this.refreshPromise) {
      return this.refreshPromise;
    }

    this.refreshing = true;
    this.refreshPromise = this.doRefresh(refreshToken).finally(() => {
      this.refreshing = false;
      this.refreshPromise = null;
    });

    return this.refreshPromise;
  }

  private async doRefresh(refreshToken: string): Promise<string | null> {
    try {
      const response = await new Promise<ApiResponse<{ accessToken: string; refreshToken?: string }>>(
        (resolve, reject) => {
          let settled = false;

          // 10秒超时
          const timeoutId = setTimeout(() => {
            if (!settled) {
              settled = true;
              reject(new Error('刷新 Token 超时'));
            }
          }, 10000);

          uni.request({
            url: `${getBaseURL()}/auth/refresh`,
            method: 'POST',
            header: {
              'X-Refresh-Token': refreshToken,
            },
            success: (res) => {
              if (!settled) {
                settled = true;
                clearTimeout(timeoutId);
                resolve(res.data as any);
              }
            },
            fail: (err) => {
              if (!settled) {
                settled = true;
                clearTimeout(timeoutId);
                reject(err);
              }
            },
          });
        }
      );

      if (response.code === 0 && response.data) {
        this.setTokens(response.data.accessToken, response.data.refreshToken);
        return response.data.accessToken;
      }

      this.clearTokens();
      return null;
    } catch {
      this.clearTokens();
      return null;
    }
  }

  onTokenChange(callback: (token: string | null) => void) {
    this.subscribers.push(callback);
  }

  private notifySubscribers(token: string | null) {
    this.subscribers.forEach((cb) => cb(token));
  }
}

/**
 * 获取 API 基础 URL
 */
function getBaseURL(): string {
  const url = import.meta.env.VITE_API_BASE_URL || '';
  // 生产环境强制要求 HTTPS
  if (!import.meta.env.DEV && url && url.startsWith('http://')) {
    throw new Error('生产环境必须使用 HTTPS 连接');
  }
  return url;
}

/**
 * 请求客户端
 */
class RequestClient {
  private tokenManager = TokenManager.getInstance();

  /**
   * 发起请求
   */
  async request<T = any>(config: RequestConfig): Promise<T> {
    const { url, method = 'GET', data, params, headers = {}, skipAuth = false } = config;

    // 构建完整 URL
    let fullUrl = url.startsWith('http') ? url : `${getBaseURL()}${url}`;

    // 添加 query 参数
    if (params) {
      const queryString = Object.entries(params)
        .filter(([, v]) => v !== undefined && v !== null)
        .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
        .join('&');
      if (queryString) {
        fullUrl += (fullUrl.includes('?') ? '&' : '?') + queryString;
      }
    }

    // 构建请求头
    const requestHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
      ...headers,
    };

    // 注入 Token
    if (!skipAuth) {
      const token = this.tokenManager.getAccessToken();
      if (token) {
        requestHeaders['Authorization'] = `Bearer ${token}`;
      }
    }

    // 注入设备信息
    requestHeaders['X-Client-Type'] = this.getClientType();
    requestHeaders['X-Device-ID'] = this.getDeviceId();

    // 发起请求
    const response = await this.doRequest<T>(fullUrl, method, data, requestHeaders);
    return response;
  }

  /**
   * GET 请求
   */
  async get<T = any>(url: string, config?: Omit<RequestConfig, 'url' | 'method'>): Promise<T> {
    return this.request<T>({ ...config, url, method: 'GET' });
  }

  /**
   * POST 请求
   */
  async post<T = any>(url: string, data?: any, config?: Omit<RequestConfig, 'url' | 'method'>): Promise<T> {
    return this.request<T>({ ...config, url, method: 'POST', data });
  }

  /**
   * PUT 请求
   */
  async put<T = any>(url: string, data?: any, config?: Omit<RequestConfig, 'url' | 'method'>): Promise<T> {
    return this.request<T>({ ...config, url, method: 'PUT', data });
  }

  /**
   * DELETE 请求
   */
  async delete<T = any>(url: string, config?: Omit<RequestConfig, 'url' | 'method'>): Promise<T> {
    return this.request<T>({ ...config, url, method: 'DELETE' });
  }

  /**
   * 执行请求（带 401 自动刷新）
   */
  private async doRequest<T>(
    url: string,
    method: string,
    data: any,
    headers: Record<string, string>,
    retry = false
  ): Promise<T> {
    return new Promise<T>((resolve, reject) => {
      uni.request({
        url,
        method: method as any,
        data,
        header: headers,
        success: async (res) => {
          const response = res.data as any;

          // 成功响应
          if (response?.code === 0) {
            resolve(response.data as T);
            return;
          }

          // 401 Token 过期，尝试刷新
          if (res.statusCode === 401 && !retry) {
            try {
              const newToken = await this.tokenManager.refreshToken();
              if (newToken) {
                headers['Authorization'] = `Bearer ${newToken}`;
                const result = await this.doRequest<T>(url, method, data, headers, true);
                resolve(result);
                return;
              }
            } catch {
              // 刷新失败（网络错误等），继续走登录逻辑
            }
            // 刷新失败或无新 Token，跳转登录页
            this.redirectToLogin();
            reject({ code: 401, message: 'Token 已过期，请重新登录' });
            return;
          }

          // 其他错误
          reject({
            code: response?.code || res.statusCode,
            message: response?.message || '请求失败',
            details: response?.details,
          });
        },
        fail: (err) => {
          // 网络错误：入队离线重试（仅非 GET 请求）
          if (method !== 'GET') {
            enqueueOfflineRequest(url, method, data);
          }
          reject({ code: -1, message: '网络错误，请检查网络连接' });
        },
      });
    });
  }

  /**
   * 获取客户端类型
   */
  private getClientType(): string {
    // #ifdef H5
    return 'h5';
    // #endif
    // #ifdef MP-WEIXIN
    return 'miniapp';
    // #endif
    // #ifdef APP-PLUS
    return 'app';
    // #endif
    return 'h5';
  }

  /**
   * 获取设备 ID
   */
  private getDeviceId(): string {
    let deviceId = uni.getStorageSync('device_id');
    if (!deviceId) {
      deviceId = `web_${Date.now()}_${Math.random().toString(36).slice(2, 11)}`;
      uni.setStorageSync('device_id', deviceId);
    }
    return deviceId;
  }

  /**
   * 跳转登录页
   */
  private redirectToLogin() {
    this.tokenManager.clearTokens();
    uni.reLaunch({ url: '/pages/login/index' });
  }
}

/**
 * 导出单例
 */
export const request = new RequestClient();
export const tokenManager = TokenManager.getInstance();
