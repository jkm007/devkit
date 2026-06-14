/**
 * 轮播图 API
 */
import { request } from './request';

export interface BannerItem {
  id: number;
  title: string;
  image: string;
  link: string;
  linkType: 'internal' | 'external' | 'none';
}

/**
 * 获取启用的轮播图列表（公开接口）
 */
export function getBanners() {
  return request.get<BannerItem[]>('/banners');
}
