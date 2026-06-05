import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace SystemRealNameApi {
  export interface RealNameApplication {
    id: number;
    userId: number;
    username: string;
    realName: string;
    idCard: string;
    status: number;
    rejectReason?: string;
    submittedAt: string;
    reviewedAt?: string;
  }
}

/** 获取实名认证申请列表（管理员） */
export function getRealNameList(params: Recordable<any>) {
  return requestClient.get<{
    items: SystemRealNameApi.RealNameApplication[];
    total: number;
  }>('/system/real-name/list', { params });
}

/** 审核通过实名认证 */
export function approveRealName(id: number) {
  return requestClient.put(`/system/real-name/${id}/approve`);
}

/** 审核拒绝实名认证 */
export function rejectRealName(id: number, data: { reason: string }) {
  return requestClient.put(`/system/real-name/${id}/reject`, data);
}
