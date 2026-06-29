/**
 * 媒体 URL 工具模块
 *
 * 核心思路：后端 JWTAuth 中间件支持 ?token= 查询参数认证，
 * 所以 <img> 和 <video> 标签可以直接加载带 token 的 URL，
 * 完全不需要 blob URL 转换。
 */

import { useAccessStore } from '@vben/stores';

const API_BASE = import.meta.env.VITE_GLOB_API_URL || '/api/v1';

/**
 * 获取当前用户的 access token
 */
function getToken(): string {
  const accessStore = useAccessStore();
  return accessStore.accessToken || '';
}

/**
 * 给 URL 附加 token 查询参数
 * - 跳过 blob: 和 data: URL
 * - 跳过已经是 http(s):// 但不是本站的 URL
 * - 已有 token 参数则替换
 */
export function appendToken(url: string): string {
  if (!url) return url;
  if (url.startsWith('blob:') || url.startsWith('data:')) return url;

  const token = getToken();
  if (!token) return url;

  try {
    // 处理相对路径和绝对路径
    const base = url.startsWith('http') ? url : `${window.location.origin}${url.startsWith('/') ? '' : '/'}${url}`;
    const u = new URL(base);
    u.searchParams.set('token', token);
    // 返回相对路径（如果是本站 URL）
    if (url.startsWith('/') || url.startsWith(API_BASE)) {
      return `${u.pathname}${u.search}`;
    }
    return u.toString();
  } catch {
    // URL 解析失败，手动拼接
    const separator = url.includes('?') ? '&' : '?';
    return `${url}${separator}token=${encodeURIComponent(token)}`;
  }
}

/**
 * 移除 URL 中的 token 查询参数（保存到数据库前清理）
 */
export function stripToken(url: string): string {
  if (!url || !url.includes('token=')) return url;
  try {
    const u = new URL(url, window.location.origin);
    u.searchParams.delete('token');
    const result = `${u.pathname}${u.search}`;
    return result || url;
  } catch {
    // 手动移除 token 参数
    return url
      .replace(/[?&]token=[^&]*/g, '')
      .replace(/\?$/, '')
      .replace(/&/, '?');
  }
}

/**
 * 规范化文件 URL：
 * - /files/{id}/direct-url → /files/{id}/view
 * - 补全 /api/v1 前缀
 */
export function normalizeFileUrl(url: string): string {
  if (!url) return url;
  let normalized = url.replace(/\/files\/(\d+)\/direct-url/g, '/files/$1/view');
  // 补全裸 /files/ 路径（不含 /api/v1 前缀的）
  if (/^\/files\/\d+\/view/.test(normalized) && !normalized.startsWith(API_BASE)) {
    normalized = `${API_BASE}${normalized}`;
  }
  return normalized;
}

/**
 * 处理 HTML 中所有媒体 URL：
 * 1. 规范化 URL (direct-url → view, 补全 /api/v1 前缀)
 * 2. 附加 token
 *
 * 用于预览和编辑模式下直接渲染 HTML
 */
export function processMediaHtml(html: any): string {
  if (!html) return '';
  // 如果是对象/数组，序列化为 JSON 字符串后再处理
  if (typeof html !== 'string') {
    html = JSON.stringify(html);
  }

  // 匹配 src/poster/href 属性中的文件 URL（支持单引号和双引号）
  return html.replace(
    /((?:src|poster|href)=["'])([^"']+)(["'])/g,
    (match: string, prefix: string, url: string, suffix: string) => {
      // 跳过 blob/data/http(s) 非本站 URL
      if (url.startsWith('blob:') || url.startsWith('data:')) return match;
      if (url.startsWith('http://') || url.startsWith('https://')) {
        // 只处理本站的 URL
        if (!url.includes('/files/')) return match;
      }

      // 规范化 URL
      let normalized = normalizeFileUrl(url);
      // 附加 token
      normalized = appendToken(normalized);
      return `${prefix}${normalized}${suffix}`;
    },
  );
}

/**
 * 保存前清理 HTML 中的 token 参数
 */
export function cleanMediaHtml(html: any): string {
  if (!html) return '';
  if (typeof html !== 'string') {
    html = JSON.stringify(html);
  }
  return html.replace(
    /((?:src|poster|href)=["'])([^"']+)(["'])/g,
    (match: string, prefix: string, url: string, suffix: string) => {
      if (!url.includes('token=')) return match;
      const cleaned = stripToken(url);
      return `${prefix}${cleaned}${suffix}`;
    },
  );
}

/**
 * 解析 JSON 字符串，处理双重编码
 * 返回解析后的值（对象、数组或字符串），失败返回原始字符串
 */
export function safeJsonParse(jsonStr: string): any {
  if (!jsonStr || jsonStr === 'null') return '';
  try {
    const parsed = JSON.parse(jsonStr);
    if (typeof parsed === 'string') {
      // 双重编码的字符串：再解一次
      try {
        const parsed2 = JSON.parse(parsed);
        // 无论 parsed2 是什么类型，都返回它（字符串、对象、数组）
        return parsed2;
      } catch {
        return parsed;
      }
    }
    // 对象/数组：直接返回
    return parsed;
  } catch {
    // 不是 JSON，返回原始字符串
    return jsonStr;
  }
}

/**
 * 检查 HTML 中是否有正在上传的媒体节点
 * 用于保存前的验证，阻止在上传过程中保存
 */
export function hasUploadingMedia(html: any): boolean {
  if (!html) return false;
  const str = typeof html === 'string' ? html : JSON.stringify(html);
  return str.includes('data-uploading="true"') || str.includes("data-uploading='true'");
}
