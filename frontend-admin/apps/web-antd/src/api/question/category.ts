import type { Recordable } from '@vben/types';

import { requestClient } from '#/api/request';

// ==================== 考试大类 ====================

export namespace ExamCategoryApi {
  export interface ExamCategory {
    [key: string]: any;
    id: number;
    parentId: number;
    name: string;
    code: string;
    path: string;
    level: number;
    sortOrder: number;
    status: number;
    createdBy: number;
    createTime: string;
    children?: ExamCategory[];
  }
}

export async function getExamCategoryList(params: Recordable<any>) {
  return requestClient.get<Array<ExamCategoryApi.ExamCategory>>(
    '/system/exam-categories',
    { params },
  );
}

export async function getExamCategoryAll() {
  return requestClient.get<Array<ExamCategoryApi.ExamCategory>>(
    '/system/exam-categories/all',
  );
}

export async function getExamCategoryDetail(id: number) {
  return requestClient.get<ExamCategoryApi.ExamCategory>(
    `/system/exam-categories/${id}`,
  );
}

export async function createExamCategory(
  data: Omit<ExamCategoryApi.ExamCategory, 'id' | 'createTime'>,
) {
  return requestClient.post('/system/exam-categories', data);
}

export async function updateExamCategory(
  id: number,
  data: Omit<ExamCategoryApi.ExamCategory, 'id' | 'createTime'>,
) {
  return requestClient.put(`/system/exam-categories/${id}`, data);
}

export async function deleteExamCategory(id: number) {
  return requestClient.delete(`/system/exam-categories/${id}`);
}

// ==================== 具体考试 ====================

export namespace ExamApi {
  export interface Exam {
    [key: string]: any;
    id: number;
    examCategoryId: number;
    name: string;
    code: string;
    description: string;
    status: number;
    sortOrder: number;
    createdBy: number;
    createTime: string;
  }
}

export async function getExamList(params: Recordable<any>) {
  return requestClient.get<Array<ExamApi.Exam>>('/system/exams', { params });
}

export async function getExamAll(params?: Recordable<any>) {
  return requestClient.get<Array<ExamApi.Exam>>('/system/exams/all', {
    params,
  });
}

export async function getExamDetail(id: number) {
  return requestClient.get<ExamApi.Exam>(`/system/exams/${id}`);
}

export async function createExam(
  data: Omit<ExamApi.Exam, 'id' | 'createTime'>,
) {
  return requestClient.post('/system/exams', data);
}

export async function updateExam(
  id: number,
  data: Omit<ExamApi.Exam, 'id' | 'createTime'>,
) {
  return requestClient.put(`/system/exams/${id}`, data);
}

export async function deleteExam(id: number) {
  return requestClient.delete(`/system/exams/${id}`);
}

// ==================== 科目 ====================

export namespace SubjectApi {
  export interface Subject {
    [key: string]: any;
    id: number;
    examId: number;
    name: string;
    code: string;
    sortOrder: number;
    status: number;
    createdBy: number;
    createTime: string;
  }
}

export async function getSubjectList(params: Recordable<any>) {
  return requestClient.get<Array<SubjectApi.Subject>>('/system/subjects', {
    params,
  });
}

export async function getSubjectAll(examId: number) {
  return requestClient.get<Array<SubjectApi.Subject>>('/system/subjects/all', {
    params: { examId },
  });
}

export async function getSubjectDetail(id: number) {
  return requestClient.get<SubjectApi.Subject>(`/system/subjects/${id}`);
}

export async function createSubject(
  data: Omit<SubjectApi.Subject, 'id' | 'createTime'>,
) {
  return requestClient.post('/system/subjects', data);
}

export async function updateSubject(
  id: number,
  data: Omit<SubjectApi.Subject, 'id' | 'createTime'>,
) {
  return requestClient.put(`/system/subjects/${id}`, data);
}

export async function deleteSubject(id: number) {
  return requestClient.delete(`/system/subjects/${id}`);
}

// ==================== 章节分类 ====================

export namespace QuestionCategoryApi {
  export interface QuestionCategory {
    [key: string]: any;
    id: number;
    examId: number;
    subjectId: number;
    parentId: number;
    name: string;
    path: string;
    level: number;
    sortOrder: number;
    status: number;
    createdBy: number;
    createTime: string;
  }
}

export async function getQuestionCategoryList(params: Recordable<any>) {
  return requestClient.get<Array<QuestionCategoryApi.QuestionCategory>>(
    '/system/question-categories',
    { params },
  );
}

export async function getQuestionCategoryAll() {
  return requestClient.get<Array<QuestionCategoryApi.QuestionCategory>>(
    '/system/question-categories/all',
  );
}

export async function getQuestionCategoryDetail(id: number) {
  return requestClient.get<QuestionCategoryApi.QuestionCategory>(
    `/system/question-categories/${id}`,
  );
}

export async function createQuestionCategory(
  data: Omit<QuestionCategoryApi.QuestionCategory, 'id' | 'createTime'>,
) {
  return requestClient.post('/system/question-categories', data);
}

export async function updateQuestionCategory(
  id: number,
  data: Omit<QuestionCategoryApi.QuestionCategory, 'id' | 'createTime'>,
) {
  return requestClient.put(`/system/question-categories/${id}`, data);
}

export async function deleteQuestionCategory(id: number) {
  return requestClient.delete(`/system/question-categories/${id}`);
}
