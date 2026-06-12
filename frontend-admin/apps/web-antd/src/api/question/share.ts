import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace QuestionShareApi {
  export interface QuestionShare {
    [key: string]: any;
    id: number;
    questionId: number;
    questionVersionId: number;
    shareCode: string;
    shareType: string;
    targetId: number;
    expireAt: string;
    maxAccess: number;
    accessCount: number;
    status: number;
    createdBy: number;
    createTime: string;
    accessedAt: string;
  }
}

export async function getQuestionShareList(params: Recordable<any>) {
  return requestClient.get<Array<QuestionShareApi.QuestionShare>>(
    '/system/question-shares',
    { params },
  );
}

export async function getQuestionShareDetail(id: number) {
  return requestClient.get<QuestionShareApi.QuestionShare>(
    `/system/question-shares/${id}`,
  );
}

export async function createQuestionShare(
  data: Omit<QuestionShareApi.QuestionShare, 'id' | 'createTime'>,
) {
  return requestClient.post('/system/question-shares', data);
}

export async function disableQuestionShare(id: number) {
  return requestClient.put(`/system/question-shares/${id}/disable`);
}

export async function enableQuestionShare(id: number) {
  return requestClient.put(`/system/question-shares/${id}/enable`);
}

export async function deleteQuestionShare(id: number) {
  return requestClient.delete(`/system/question-shares/${id}`);
}
