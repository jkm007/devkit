/**
 * 通知 API
 */
import { request } from './request';

export interface Notification {
  id: number;
  title: string;
  content: string;
  type: string;
  isRead: boolean;
  createdAt: string;
}

/**
 * 获取通知列表
 */
export function getNotifications(params: { page?: number; pageSize?: number }) {
  return request.get<{ items: Notification[]; total: number }>('/api/v1/notifications', { params });
}

/**
 * 获取未读数量
 */
export function getUnreadCount() {
  return request.get<{ count: number }>('/api/v1/notifications/unread-count');
}

/**
 * 标记已读
 */
export function markRead(id: number) {
  return request.put(`/api/v1/notifications/${id}/read`);
}

/**
 * 全部已读
 */
export function markAllRead() {
  return request.put('/api/v1/notifications/read-all');
}

/**
 * 删除通知
 */
export function deleteNotification(id: number) {
  return request.delete(`/api/v1/notifications/${id}`);
}
