import { request } from './request';
import type { PageResponse } from '@/api/types';

export interface ClassInfo {
  id: number;
  name: string;
  code: string;
  description: string;
  status: number;
  createdBy: number;
  memberCount: number;
  creatorName: string;
  myRole: string;
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

export function getMyClasses() {
  return request.get<ClassInfo[]>('/my-classes');
}

export function getClassDetail(id: number) {
  return request.get<ClassInfo>(`/classes/${id}`);
}

export function joinClassByCode(code: string) {
  return request.post<ClassInfo>('/classes/join', { code });
}

export function getClassMembers(id: number, params: { page?: number; pageSize?: number }) {
  return request.get<PageResponse<ClassMember>>(`/classes/${id}/members`, { params });
}

export function getClassQuestions(id: number, params: { page?: number; pageSize?: number }) {
  return request.get<PageResponse<any>>(`/classes/${id}/questions`, { params });
}
