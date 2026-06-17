/**
 * 移动端配置 API
 */
import { request } from './request';

// 快捷菜单项
export interface QuickMenuItem {
  id: number;
  title: string;
  icon: string;
  link: string;
  linkType: string;
  sortOrder: number;
  status: string;
}

// 我的页面菜单项
export interface MyPageMenuItem {
  id: number;
  title: string;
  icon: string;
  link: string;
  showBadge: boolean;
  badgeText: string;
  sortOrder: number;
  status: string;
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

/**
 * 获取启用的快捷菜单列表（公开接口）
 */
export function getQuickMenus() {
  return request.get<QuickMenuItem[]>('/mobile/quick-menus');
}

/**
 * 获取启用的我的页面菜单列表（公开接口）
 */
export function getMyPageMenus() {
  return request.get<MyPageMenuItem[]>('/mobile/my-page-menus');
}

/**
 * 获取移动端设置（公开接口）
 */
export function getMobileSettings() {
  return request.get<MobileSettings>('/mobile/settings');
}
