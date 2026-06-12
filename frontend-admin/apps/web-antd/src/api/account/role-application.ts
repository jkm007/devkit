import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace RoleApplicationApi {
  export interface AvailableRole {
    id: number;
    name: string;
    remark?: string;
  }

  export interface RoleApplicationItem {
    id: number;
    userId: number;
    username?: string;
    nickname?: string;
    roleId: number;
    roleName?: string;
    roleRemark?: string;
    reason: string;
    status: 0 | 1 | 2;
    reviewNote?: string;
    reviewedBy?: number;
    reviewerName?: string;
    reviewedAt?: string;
    createdAt: string;
  }

  export interface CreateRoleApplicationRequest {
    roleId: number;
    reason?: string;
  }

  export interface ReviewRoleApplicationRequest {
    note?: string;
  }
}

export function getAvailableRoles() {
  return requestClient.get<RoleApplicationApi.AvailableRole[]>(
    '/auth/role-applications/available-roles',
  );
}

export function createRoleApplication(
  data: RoleApplicationApi.CreateRoleApplicationRequest,
) {
  return requestClient.post('/auth/role-applications', data);
}

export function getMyRoleApplications(params: Recordable<any>) {
  return requestClient.get<{
    items: RoleApplicationApi.RoleApplicationItem[];
    total: number;
  }>('/auth/role-applications', { params });
}

export function getRoleApplicationList(params: Recordable<any>) {
  return requestClient.get<{
    items: RoleApplicationApi.RoleApplicationItem[];
    total: number;
  }>('/system/role-applications/list', { params });
}

export function approveRoleApplication(
  id: number,
  data: RoleApplicationApi.ReviewRoleApplicationRequest,
) {
  return requestClient.put(`/system/role-applications/${id}/approve`, data);
}

export function rejectRoleApplication(
  id: number,
  data: RoleApplicationApi.ReviewRoleApplicationRequest,
) {
  return requestClient.put(`/system/role-applications/${id}/reject`, data);
}
