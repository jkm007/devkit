import { requestClient } from '#/api/request';

export interface PageResponse<T> {
  items: T[];
  total: number;
}

export namespace FeedbackApi {
  export interface Feedback {
    id: number;
    questionId: number;
    questionTitle: string;
    userId: number;
    userNickname: string;
    feedbackType: string; // error / suggestion / other
    content: string;
    status: string; // pending / processing / resolved / closed
    reply: string;
    createdAt: string;
    updatedAt: string;
  }

  export interface UpdateFeedbackRequest {
    status?: string;
    reply?: string;
  }
}

/** 获取纠错列表 */
export function getFeedbackList(params: { page?: number; pageSize?: number; status?: string }) {
  return requestClient.get<PageResponse<FeedbackApi.Feedback>>('/system/feedbacks', { params });
}

/** 更新纠错状态 */
export function updateFeedback(id: number, data: FeedbackApi.UpdateFeedbackRequest) {
  return requestClient.put(`/system/feedbacks/${id}`, data);
}
