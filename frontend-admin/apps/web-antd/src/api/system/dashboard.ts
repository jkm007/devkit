import { requestClient } from '#/api/request';

export interface DashboardStats {
  overview: {
    userCount: number;
    activeUsers: number;
    fileCount: number;
    totalStorage: number;
    todayLogins: number;
    todayEvents: number;
    onlineDevices: number;
  };
  eventsTrend: Array<{
    date: string;
    success: number;
    fail: number;
  }>;
  eventsByType: Array<{
    type: string;
    count: number;
  }>;
  deviceByType: Array<{
    type: string;
    count: number;
  }>;
  deviceByPlatform: Array<{
    type: string;
    count: number;
  }>;
  recentLogins: Array<{
    username: string;
    ip: string;
    device: string;
    location: string;
    status: number;
    createdAt: string;
  }>;
}

/**
 * 获取仪表盘统计数据
 */
async function getDashboardStats() {
  return requestClient.get<DashboardStats>('/system/dashboard/stats');
}

export { getDashboardStats };
