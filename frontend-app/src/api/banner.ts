/**
 * 轮播图 API
 */
import { request } from './request';

export interface BannerItem {
  id: number;
  title: string;
  image: string;
  link: string;
  linkType: 'external' | 'none' | 'page';
}

/**
 * 获取启用的轮播图列表（公开接口）
 */
export function getBanners() {
  return request.get<BannerItem[]>('/banners');
}
