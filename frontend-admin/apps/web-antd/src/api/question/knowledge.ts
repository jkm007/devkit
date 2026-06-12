import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

export namespace KnowledgePointApi {
  export interface KnowledgePoint {
    [key: string]: any;
    id: number;
    examId: number;
    subjectId: number;
    categoryId: number;
    parentId: number;
    name: string;
    code: string;
    path: string;
    level: number;
    importance: number;
    description: string;
    sortOrder: number;
    status: number;
    createdBy: number;
    createTime: string;
  }
}

export async function getKnowledgePointList(params: Recordable<any>) {
  return requestClient.get<Array<KnowledgePointApi.KnowledgePoint>>(
    '/system/knowledge-points',
    { params },
  );
}

export async function getKnowledgePointAll() {
  return requestClient.get<Array<KnowledgePointApi.KnowledgePoint>>(
    '/system/knowledge-points/all',
  );
}

export async function getKnowledgePointDetail(id: number) {
  return requestClient.get<KnowledgePointApi.KnowledgePoint>(
    `/system/knowledge-points/${id}`,
  );
}

export async function createKnowledgePoint(
  data: Omit<KnowledgePointApi.KnowledgePoint, 'id' | 'createTime'>,
) {
  return requestClient.post('/system/knowledge-points', data);
}

export async function updateKnowledgePoint(
  id: number,
  data: Omit<KnowledgePointApi.KnowledgePoint, 'id' | 'createTime'>,
) {
  return requestClient.put(`/system/knowledge-points/${id}`, data);
}

export async function deleteKnowledgePoint(id: number) {
  return requestClient.delete(`/system/knowledge-points/${id}`);
}
