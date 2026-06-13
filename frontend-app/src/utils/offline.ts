/**
 * 离线缓存策略
 *
 * 1. 网络正常：请求 → 缓存 → 返回
 * 2. 网络断开：返回缓存 → 标记离线 → 请求入队
 * 3. 网络恢复：队列重放 → 更新缓存
 */

import { getCache, setCache, deleteCache } from './cache';

interface OfflineRequest {
  id: string;
  url: string;
  method: string;
  data?: any;
  timestamp: number;
  retryCount: number;
}

const OFFLINE_QUEUE_KEY = 'offline_queue';
const ONLINE_STATUS_KEY = 'is_online';
const MAX_RETRIES = 3;

let isOnline = true;
let queue: OfflineRequest[] = [];

/**
 * 初始化网络状态监听
 */
export function initNetworkListener() {
  // #ifdef H5
  window.addEventListener('online', () => { isOnline = true; flushQueue(); });
  window.addEventListener('offline', () => { isOnline = false; });
  // #endif

  // #ifdef MP-WEIXIN
  uni.onNetworkStatusChange((res) => {
    isOnline = res.isConnected;
    if (res.isConnected) flushQueue();
  });
  // #endif

  // #ifdef APP-PLUS
  uni.onNetworkStatusChange((res) => {
    isOnline = res.isConnected;
    if (res.isConnected) flushQueue();
  });
  // #endif

  // 初始状态
  // #ifdef H5
  isOnline = navigator.onLine;
  // #endif
}

/**
 * 是否在线
 */
export function checkOnline(): boolean {
  return isOnline;
}

/**
 * 缓存优先请求策略
 *
 * 1. 先返回缓存（如果有）
 * 2. 同时发起真实请求
 * 3. 请求成功后更新缓存
 * 4. 离线时直接返回缓存
 */
export async function cacheFirst<T>(
  cacheKey: string,
  fetcher: () => Promise<T>,
  ttlMs: number = 300000, // 默认 5 分钟
): Promise<{ data: T; fromCache: boolean }> {
  // 离线时：返回缓存
  if (!isOnline) {
    const cached = getCache<T>(cacheKey);
    if (cached) return { data: cached, fromCache: true };
    throw new Error('网络不可用，且无缓存数据');
  }

  // 在线时：先读缓存
  const cached = getCache<T>(cacheKey);

  // 发起真实请求
  try {
    const data = await fetcher();
    setCache(cacheKey, data, ttlMs, true); // 持久化
    return { data, fromCache: false };
  } catch (error) {
    // 请求失败但有缓存：返回缓存
    if (cached) return { data: cached, fromCache: true };
    throw error;
  }
}

/**
 * 网络优先请求策略
 *
 * 1. 先发起真实请求
 * 2. 失败时返回缓存
 */
export async function networkFirst<T>(
  cacheKey: string,
  fetcher: () => Promise<T>,
  ttlMs: number = 60000,
): Promise<{ data: T; fromCache: boolean }> {
  if (!isOnline) {
    const cached = getCache<T>(cacheKey);
    if (cached) return { data: cached, fromCache: true };
    throw new Error('网络不可用，且无缓存数据');
  }

  try {
    const data = await fetcher();
    setCache(cacheKey, data, ttlMs, true);
    return { data, fromCache: false };
  } catch {
    const cached = getCache<T>(cacheKey);
    if (cached) return { data: cached, fromCache: true };
    throw new Error('请求失败且无缓存');
  }
}

/**
 * 离线请求入队
 */
export function enqueueOfflineRequest(
  url: string,
  method: string,
  data?: any,
): void {
  const request: OfflineRequest = {
    id: `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
    url,
    method,
    data,
    timestamp: Date.now(),
    retryCount: 0,
  };

  queue.push(request);
  saveQueue();
}

/**
 * 刷新离线队列（网络恢复时调用）
 */
async function flushQueue(): Promise<void> {
  if (queue.length === 0) return;

  const toFlush = [...queue];
  queue = [];

  for (const req of toFlush) {
    if (req.retryCount >= MAX_RETRIES) continue; // 超过最大重试次数

    try {
      const token = uni.getStorageSync('access_token');
      const deviceId = uni.getStorageSync('device_id') || '';

      await new Promise<void>((resolve, reject) => {
        uni.request({
          url: req.url,
          method: req.method as any,
          data: req.data,
          header: {
            'Content-Type': 'application/json',
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
            'X-Device-ID': deviceId,
          },
          success: (res) => {
            const response = res.data as any;
            if (response?.code === 0) {
              resolve();
            } else {
              reject(new Error(response?.message || '请求失败'));
            }
          },
          fail: () => reject(new Error('网络错误')),
        });
      });
    } catch {
      // 失败重新入队
      req.retryCount++;
      queue.push(req);
    }
  }

  saveQueue();
}

/**
 * 保存队列到本地
 */
function saveQueue(): void {
  try {
    uni.setStorageSync(OFFLINE_QUEUE_KEY, JSON.stringify(queue));
  } catch {
    // ignore
  }
}

/**
 * 加载队列
 */
export function loadQueue(): void {
  try {
    const raw = uni.getStorageSync(OFFLINE_QUEUE_KEY);
    if (raw) queue = JSON.parse(raw);
  } catch {
    queue = [];
  }
}

/**
 * 获取队列状态
 */
export function getQueueStatus(): { pending: number; requests: OfflineRequest[] } {
  return { pending: queue.length, requests: queue };
}

/**
 * 清除队列
 */
export function clearQueue(): void {
  queue = [];
  uni.removeStorageSync(OFFLINE_QUEUE_KEY);
}
