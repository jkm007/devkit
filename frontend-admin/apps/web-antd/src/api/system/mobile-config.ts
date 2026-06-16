import { requestClient } from '#/api/request';

// 快捷菜单
export interface QuickMenuItem {
  id: number;
  title: string;
  icon: string;
  link: string;
  linkType: string;
  sortOrder: number;
  status: string;
  createdAt: string;
  updatedAt: string;
}

// 我的页面菜单
export interface MyPageItem {
  id: number;
  title: string;
  icon: string;
  link: string;
  showBadge: boolean;
  badgeText: string;
  sortOrder: number;
  status: string;
  createdAt: string;
  updatedAt: string;
}

// 移动端设置
export interface MobileSettings {
  noticeEnabled: boolean;
  noticeContent: string;
  appDownloadUrl: string;
  customerServiceUrl: string;
  aboutUs: string;
  agreementUrl: string;
  privacyUrl: string;
}

// ===== 快捷菜单 API =====
export function getQuickMenuList(params?: any) {
  return requestClient.get<QuickMenuItem[]>('/system/quick-menus', { params });
}

export function createQuickMenu(data: Partial<QuickMenuItem>) {
  return requestClient.post('/system/quick-menus', data);
}

export function updateQuickMenu(id: number, data: Partial<QuickMenuItem>) {
  return requestClient.put(`/system/quick-menus/${id}`, data);
}

export function deleteQuickMenu(id: number) {
  return requestClient.delete(`/system/quick-menus/${id}`);
}

// ===== 我的页面 API =====
export function getMyPageList(params?: any) {
  return requestClient.get<MyPageItem[]>('/system/my-page-menus', { params });
}

export function createMyPageMenu(data: Partial<MyPageItem>) {
  return requestClient.post('/system/my-page-menus', data);
}

export function updateMyPageMenu(id: number, data: Partial<MyPageItem>) {
  return requestClient.put(`/system/my-page-menus/${id}`, data);
}

export function deleteMyPageMenu(id: number) {
  return requestClient.delete(`/system/my-page-menus/${id}`);
}

// ===== 移动端设置 API =====
export function getMobileSettings() {
  return requestClient.get<MobileSettings>('/system/mobile-settings');
}

export function updateMobileSettings(data: Partial<MobileSettings>) {
  return requestClient.put('/system/mobile-settings', data);
}
