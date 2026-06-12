import { requestClient } from '#/api/request';

/** 用户首页数据 */
export interface HomeData {
  storage: {
    used: number;
    quota: number; // 0 = 不限制
    usedPercent: number;
  };
  roles: Array<{
    id: number;
    name: string;
    code: string;
  }>;
  devices: Array<{
    id: number;
    deviceType: string;
    deviceName: string;
    browser: string;
    os: string;
    ip: string;
    platform: string;
    isCurrent: boolean;
    lastActiveAt: string;
  }>;
}

/** 获取用户首页数据 */
export function getHomeData() {
  return requestClient.get<HomeData>('/user/home');
}
