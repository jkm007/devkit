import { requestClient } from '#/api/request';

/** 通知类型 */
export interface Notification {
  id: number;
  type: string;
  title: string;
  content: string;
  link: string;
  isRead: boolean;
  senderId: number;
  createdAt: string;
}

/** 获取通知列表 */
export function getNotifications(params?: { page?: number; pageSize?: number }) {
  return requestClient.get<{ items: Notification[]; total: number }>(
    '/notifications',
    { params },
  );
}

/** 获取未读通知数量 */
export function getUnreadCount() {
  return requestClient.get<{ count: number }>('/notifications/unread-count');
}

/** 标记单条已读 */
export function markRead(id: number) {
  return requestClient.put(`/notifications/${id}/read`);
}

/** 标记全部已读 */
export function markAllRead() {
  return requestClient.put('/notifications/read-all');
}

/** 删除通知 */
export function deleteNotification(id: number) {
  return requestClient.delete(`/notifications/${id}`);
}

/** 管理员：发布公告 */
export function publishAnnouncement(data: {
  title: string;
  content: string;
  link?: string;
}) {
  return requestClient.post('/system/notifications/announcement', data);
}

/** 管理员：获取通知列表 */
export function getAdminNotifications(params?: {
  page?: number;
  pageSize?: number;
  type?: string;
}) {
  return requestClient.get<{ items: Notification[]; total: number }>(
    '/system/notifications',
    { params },
  );
}
