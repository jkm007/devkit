import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace SystemSecurityLogApi {
  export interface SecurityLog {
    id: number;
    userId: number;
    username: string;
    eventType: string;
    eventDetail: string;
    ip: string;
    userAgent: string;
    status: number;
    createdAt: string;
  }
}

/** 获取所有用户安全日志（管理员） */
export function getSystemSecurityLogs(params: Recordable<any>) {
  return requestClient.get<{
    items: SystemSecurityLogApi.SecurityLog[];
    total: number;
  }>('/system/security-logs', { params });
}
