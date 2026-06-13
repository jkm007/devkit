/**
 * API 缓存工具
 * - 内存缓存 + 本地持久化
 * - 支持 TTL 过期
 * - 支持缓存 key 前缀
 */

interface CacheEntry<T = any> {
  data: T;
  expireAt: number; // 过期时间戳（毫秒）
  createdAt: number;
}

const MEMORY_CACHE = new Map<string, CacheEntry>();
const STORAGE_PREFIX = 'api_cache:';

/**
 * 获取缓存
 */
export function getCache<T = any>(key: string): T | null {
  // 先查内存
  const mem = MEMORY_CACHE.get(key);
  if (mem && mem.expireAt > Date.now()) {
    return mem.data as T;
  }
  if (mem) MEMORY_CACHE.delete(key);

  // 再查本地存储
  try {
    const raw = uni.getStorageSync(STORAGE_PREFIX + key);
    if (raw) {
      const entry = JSON.parse(raw) as CacheEntry;
      if (entry.expireAt > Date.now()) {
        // 回填内存
        MEMORY_CACHE.set(key, entry);
        return entry.data as T;
      }
      uni.removeStorageSync(STORAGE_PREFIX + key);
    }
  } catch {
    // ignore
  }

  return null;
}

/**
 * 设置缓存
 */
export function setCache<T = any>(key: string, data: T, ttlMs: number, persist = false): void {
  const entry: CacheEntry = {
    data,
    expireAt: Date.now() + ttlMs,
    createdAt: Date.now(),
  };

  // 写入内存
  MEMORY_CACHE.set(key, entry);

  // 可选持久化
  if (persist) {
    try {
      uni.setStorageSync(STORAGE_PREFIX + key, JSON.stringify(entry));
    } catch {
      // ignore: storage full
    }
  }
}

/**
 * 删除缓存
 */
export function deleteCache(key: string): void {
  MEMORY_CACHE.delete(key);
  try {
    uni.removeStorageSync(STORAGE_PREFIX + key);
  } catch {
    // ignore
  }
}

/**
 * 清除所有过期缓存
 */
export function clearExpiredCache(): void {
  const now = Date.now();

  // 清理内存
  for (const [key, entry] of MEMORY_CACHE.entries()) {
    if (entry.expireAt <= now) MEMORY_CACHE.delete(key);
  }

  // 清理本地存储（遍历 key 太贵，只在必要时调用）
}

/**
 * 清除所有 API 缓存
 */
export function clearAllCache(): void {
  MEMORY_CACHE.clear();
  try {
    const keys = uni.getStorageInfoSync().keys;
    for (const key of keys) {
      if (key.startsWith(STORAGE_PREFIX)) {
        uni.removeStorageSync(key);
      }
    }
  } catch {
    // ignore
  }
}

/**
 * 带缓存的请求包装器
 *
 * @example
 * const data = await cachedRequest('questions:list', () => fetchQuestions(), 60000);
 */
export async function cachedRequest<T>(
  cacheKey: string,
  fetcher: () => Promise<T>,
  ttlMs: number = 60000,
  persist: boolean = false,
): Promise<T> {
  const cached = getCache<T>(cacheKey);
  if (cached !== null) return cached;

  const data = await fetcher();
  setCache(cacheKey, data, ttlMs, persist);
  return data;
}
