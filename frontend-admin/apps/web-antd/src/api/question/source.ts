import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace QuestionSourceApi {
  export interface QuestionSource {
    [key: string]: any;
    id: number;
    sourceType: string;
    name: string;
    examId: number;
    subjectId: number;
    year: number;
    region: string;
    paperName: string;
    questionNo: string;
    copyright: string;
    createdBy: number;
    createTime: string;
  }
}

export async function getQuestionSourceList(params: Recordable<any>) {
  return requestClient.get<Array<QuestionSourceApi.QuestionSource>>(
    '/system/question-sources',
    { params },
  );
}

export async function getQuestionSourceDetail(id: number) {
  return requestClient.get<QuestionSourceApi.QuestionSource>(
    `/system/question-sources/${id}`,
  );
}

export async function createQuestionSource(
  data: Omit<QuestionSourceApi.QuestionSource, 'id' | 'createTime'>,
) {
  return requestClient.post('/system/question-sources', data);
}

export async function updateQuestionSource(
  id: number,
  data: Omit<QuestionSourceApi.QuestionSource, 'id' | 'createTime'>,
) {
  return requestClient.put(`/system/question-sources/${id}`, data);
}

export async function deleteQuestionSource(id: number) {
  return requestClient.delete(`/system/question-sources/${id}`);
}
