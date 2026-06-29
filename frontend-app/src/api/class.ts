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

export function leaveClass(id: number) {
  return request.post(`/classes/${id}/leave`);
}

// 以下接口仅 teacher/monitor 可调用

export function createClass(data: { name: string; description?: string }) {
  return request.post<ClassInfo>('/classes', data);
}

export function updateClass(id: number, data: { name?: string; description?: string; status?: number }) {
  return request.put(`/classes/${id}`, data);
}

export function addClassMember(id: number, data: { userId: number; role: string }) {
  return request.post(`/classes/${id}/members`, data);
}

export function removeClassMember(id: number, memberId: number) {
  return request.delete(`/classes/${id}/members/${memberId}`);
}

export function updateMemberRole(id: number, memberId: number, data: { role: string }) {
  return request.put(`/classes/${id}/members/${memberId}/role`, data);
}

export function getClassInvitations(id: number) {
  return request.get<ClassInvitation[]>(`/classes/${id}/invitations`);
}

export function createInvitation(id: number, data?: { maxUses?: number; expireAt?: string }) {
  return request.post<ClassInvitation>(`/classes/${id}/invitations`, data || {});
}

export function disableInvitation(classId: number, invitationId: number) {
  return request.delete(`/classes/${classId}/invitations/${invitationId}`);
}
