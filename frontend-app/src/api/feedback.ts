/**
 * 题目纠错 API
 */
import { request } from './request';

export interface FeedbackItem {
  id: number;
  questionId: number;
  feedbackType: string;
  description: string;
  suggestion: string;
  status: string;
  adminReply: string;
  createdAt: string;
}

/**
 * 提交纠错反馈
 */
export function submitFeedback(params: {
  questionId: number;
  feedbackType: string;
  description: string;
  suggestion?: string;
}) {
  return request.post('/study/feedback', params);
}

/**
 * 获取纠错列表
 */
export function getFeedbacks(params: { page?: number; pageSize?: number }) {
  return request.get<{ items: FeedbackItem[]; total: number }>('/study/feedback', { params });
}

/**
 * 获取纠错详情
 */
export function getFeedbackDetail(id: number) {
  return request.get<FeedbackItem>(`/study/feedback/${id}`);
}

/**
 * 删除纠错反馈
 */
export function deleteFeedback(id: number) {
  return request.delete(`/study/feedback/${id}`);
}
