import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace QuestionImportApi {
  export interface ImportTask {
    [key: string]: any;
    id: number;
    fileId: number;
    fileName: string;
    fileType: string;
    status: string;
    totalCount: number;
    successCount: number;
    failedCount: number;
    errorReport: string;
    targetCategoryId: number;
    targetResourceType: string;
    targetScopeType: string;
    targetScopeId: number;
    createdBy: number;
    confirmedAt: string;
    createTime: string;
  }

  export interface ImportItem {
    [key: string]: any;
    id: number;
    taskId: number;
    rowNo: number;
    questionNo: string;
    parseStatus: string;
    questionId: number;
    errorMessage: string;
    rawContent: string;
    createTime: string;
  }
}

export async function getImportTaskList(params: Recordable<any>) {
  return requestClient.get<Array<QuestionImportApi.ImportTask>>(
    '/system/question-imports',
    { params },
  );
}

export async function getImportTaskDetail(id: number) {
  return requestClient.get<QuestionImportApi.ImportTask>(
    `/system/question-imports/${id}`,
  );
}

export async function getImportTaskItems(id: number) {
  return requestClient.get<Array<QuestionImportApi.ImportItem>>(
    `/system/question-imports/${id}/items`,
  );
}

export async function createImportTask(
  data: Omit<QuestionImportApi.ImportTask, 'id' | 'createTime'>,
) {
  return requestClient.post('/system/question-imports', data);
}

export async function deleteImportTask(id: number) {
  return requestClient.delete(`/system/question-imports/${id}`);
}
