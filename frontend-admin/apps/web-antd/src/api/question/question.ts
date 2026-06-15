import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace QuestionApi {
  export interface Question {
    [key: string]: any;
    id: number;
    title: string;
    questionType: string;
    stem: string;
    content: string;
    answer: string;
    analysis: string;
    materials: string;
    scoreRule: string;
    examId: number;
    subjectId: number;
    categoryId: number;
    sourceId: number;
    difficulty: number;
    resourceType: string;
    status: string;
    currentVersionId: number;
    parentId: number;
    isGroup: number;
    subIndex: number;
    analysisVisiblePolicy: string;
    answerVisiblePolicy: string;
    createdBy: number;
    reviewedBy: number;
    reviewedAt: string;
    rejectReason: string;
    publishedAt: string;
    createTime: string;
  }
}

export async function getQuestionList(params: Recordable<any>) {
  return requestClient.get<Array<QuestionApi.Question>>('/system/questions', {
    params,
  });
}

export async function getQuestionDetail(id: number) {
  return requestClient.get<QuestionApi.Question>(`/system/questions/${id}`);
}

export async function createQuestion(
  data: Omit<QuestionApi.Question, 'id' | 'createTime'>,
) {
  return requestClient.post('/system/questions', data);
}

export async function updateQuestion(
  id: number,
  data: Omit<QuestionApi.Question, 'id' | 'createTime'>,
) {
  return requestClient.put(`/system/questions/${id}`, data);
}

export async function deleteQuestion(id: number) {
  return requestClient.delete(`/system/questions/${id}`);
}

export async function publishQuestion(id: number) {
  return requestClient.post(`/system/questions/${id}/publish`);
}

export async function archiveQuestion(id: number) {
  return requestClient.post(`/system/questions/${id}/archive`);
}

export async function submitAuditQuestion(id: number) {
  return requestClient.post(`/system/questions/${id}/submit-audit`);
}

export async function approveQuestion(id: number) {
  return requestClient.post(`/system/questions/${id}/audit/approve`);
}

export async function rejectQuestion(id: number, reason: string) {
  return requestClient.post(`/system/questions/${id}/audit/reject`, { reason });
}

export async function withdrawQuestion(id: number) {
  return requestClient.post(`/system/questions/${id}/withdraw`);
}

export async function reactivateQuestion(id: number) {
  return requestClient.post(`/system/questions/${id}/reactivate`);
}

export async function getQuestionTypes() {
  return requestClient.get<Array<any>>('/system/questions/types');
}

export async function getQuestionStats() {
  return requestClient.get<any>('/system/questions/stats');
}
