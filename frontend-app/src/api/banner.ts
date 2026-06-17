/**
 * 轮播图 API
 */
import { request } from './request';

export interface BannerItem {
  id: number;
  title: string;
  image: string;
  fileId?: number;
  link: string;
  linkType: 'external' | 'none' | 'page';
}

/**
 * 获取启用的轮播图列表（公开接口）
 */
export function getBanners() {
  return request.get<BannerItem[]>('/banners');
}

/**
 * 获取公开文件的预签名URL（无需认证）
 */
export function getPublicFileURL(fileId: number) {
  return request.get<{ url: string }>(`/files/${fileId}/public-url`);
}

/**
 * 批量获取公开文件的预签名URL
 */
export function batchGetPublicURLs(fileIds: number[]) {
  return request.post<{ urls: Record<number, string> }>('/files/batch-public-url', { fileIds });
}
