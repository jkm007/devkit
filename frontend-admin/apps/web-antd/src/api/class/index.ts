import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace ClassApi {
  export interface Class {
    id: number;
    name: string;
    code: string;
    description: string;
    status: number;
    createdBy: number;
    memberCount: number;
    creatorName: string;
    createdAt: string;
  }

  export interface ClassMember {
    id: number;
    userId: number;
    nickname: string;
    username: string;
    avatar: string;
    role: 'student' | 'monitor' | 'teacher';
    status: number;
    joinedAt: string;
  }

  export interface ClassInvitation {
    id: number;
    classId: number;
    code: string;
    expireAt: string;
    maxUses: number;
    usedCount: number;
    status: number;
    createdAt: string;
  }
}

export function getClassList(params: Recordable<any>) {
  return requestClient.get<{ items: ClassApi.Class[]; total: number }>(
    '/system/classes',
    { params },
  );
}

export function getClassDetail(id: number) {
  return requestClient.get<ClassApi.Class>(`/system/classes/${id}`);
}

export function createClass(data: { name: string; description?: string }) {
  return requestClient.post('/system/classes', data);
}

export function updateClass(
  id: number,
  data: { name?: string; description?: string; status?: number },
) {
  return requestClient.put(`/system/classes/${id}`, data);
}

export function deleteClass(id: number) {
  return requestClient.delete(`/system/classes/${id}`);
}

export function getClassMembers(id: number, params: Recordable<any>) {
  return requestClient.get<{ items: ClassApi.ClassMember[]; total: number }>(
    `/system/classes/${id}/members`,
    { params },
  );
}

export function addClassMember(
  id: number,
  data: { userId: number; role: string },
) {
  return requestClient.post(`/system/classes/${id}/members`, data);
}

export function updateClassMemberRole(
  classId: number,
  memberId: number,
  data: { role: string },
) {
  return requestClient.put(
    `/system/classes/${classId}/members/${memberId}/role`,
    data,
  );
}

export function removeClassMember(classId: number, memberId: number) {
  return requestClient.delete(
    `/system/classes/${classId}/members/${memberId}`,
  );
}

export function getClassInvitations(id: number) {
  return requestClient.get<ClassApi.ClassInvitation[]>(
    `/system/classes/${id}/invitations`,
  );
}

export function createClassInvitation(
  id: number,
  data: { expireAt?: string; maxUses?: number },
) {
  return requestClient.post(`/system/classes/${id}/invitations`, data);
}

export function disableClassInvitation(invitationId: number) {
  return requestClient.delete(`/system/classes/invitations/${invitationId}`);
}
