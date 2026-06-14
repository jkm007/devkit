import { requestClient } from '#/api/request';

/** Banner类型 */
export interface Banner {
  id: number;
  title: string;
  image: string;
  link: string;
  linkType: string;
  sortOrder: number;
  status: string;
  createdAt: string;
  updatedAt: string;
}

/** Banner列表参数 */
export interface BannerListParams {
  page?: number;
  pageSize?: number;
}

/** Banner创建参数 */
export interface BannerCreateParams {
  title: string;
  image: string;
  link?: string;
  linkType?: string;
  sortOrder?: number;
}

/** Banner更新参数 */
export interface BannerUpdateParams {
  title?: string;
  image?: string;
  link?: string;
  linkType?: string;
  sortOrder?: number;
  status?: string;
}

/** 获取Banner列表 */
export function getBannerList(params?: BannerListParams) {
  return requestClient.get<Banner[]>('/system/banners', { params });
}

/** 创建Banner */
export function createBanner(data: BannerCreateParams) {
  return requestClient.post<Banner>('/system/banners', data);
}

/** 更新Banner */
export function updateBanner(id: number, data: BannerUpdateParams) {
  return requestClient.put<Banner>(`/system/banners/${id}`, data);
}

/** 删除Banner */
export function deleteBanner(id: number) {
  return requestClient.delete(`/system/banners/${id}`);
}

/** 更新Banner状态 */
export function updateBannerStatus(id: number, status: string) {
  return requestClient.put(`/system/banners/${id}/status`, { status });
}

/** 更新Banner排序 */
export function updateBannerSort(id: number, sortOrder: number) {
  return requestClient.put(`/system/banners/${id}/sort`, { sortOrder });
}