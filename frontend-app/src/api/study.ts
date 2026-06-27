/**
 * 移动端学习 API
 */
import { request } from './request';
import type { Question, PageResponse } from '@/api/types';

/**
 * 用户学习统计
 */
export interface UserStats {
  totalAnswered: number;
  totalCorrect: number;
  correctRate: number;
  continuousDays: number;
  favoritesCount: number;
  wrongCount: number;
  practiceDays: number;
}

/**
 * 获取题目列表
 */
export function getQuestions(params: {
  page?: number;
  pageSize?: number;
  questionType?: string;
  categoryId?: number;
  subjectId?: number;
  difficulty?: number;
  keyword?: string;
  knowledgePoint?: string;
}) {
  return request.get<PageResponse<Question>>('/study/questions', { params });
}

/**
 * 获取题目详情
 */
export function getQuestionDetail(id: number) {
  return request.get<Question>(`/study/questions/${id}`);
}

/**
 * 收藏题目
 */
export function addFavorite(id: number) {
  return request.post(`/study/questions/${id}/favorite`);
}

/**
 * 取消收藏
 */
export function removeFavorite(id: number) {
  return request.delete(`/study/questions/${id}/favorite`);
}

/**
 * 获取收藏列表
 */
export function getFavorites(params: { page?: number; pageSize?: number }) {
  return request.get<PageResponse<any>>('/user/favorites', { params });
}

/**
 * ==================== 分类收藏 API ====================
 */

/**
 * 获取分类收藏列表
 */
export function getCategoryFavorites(params: { page?: number; pageSize?: number }) {
  return request.get<PageResponse<any>>('/user/category-favorites', { params });
}

/**
 * 添加分类收藏
 */
export function addCategoryFavorite(data: { targetId: number; targetType: string }) {
  return request.post('/user/category-favorites', data);
}

/**
 * 取消分类收藏
 */
export function removeCategoryFavorite(id: number) {
  return request.delete(`/user/category-favorites/${id}`);
}

/**
 * 获取笔记列表
 */
export function getNotes(params: { page?: number; pageSize?: number }) {
  return request.get<PageResponse<any>>('/user/notes', { params });
}

/**
 * 创建/更新笔记
 */
export function saveNote(data: { questionId: number; content: string }) {
  return request.post('/user/notes', data);
}

/**
 * 更新笔记
 */
export function updateNote(id: number, data: { content: string }) {
  return request.put(`/user/notes/${id}`, data);
}

/**
 * 删除笔记
 */
export function deleteNote(id: number) {
  return request.delete(`/user/notes/${id}`);
}

/**
 * 获取练习题目（随机）
 */
export function getPracticeQuestions(data: {
  mode?: string;
  count?: number;
  types?: string[];
  categoryId?: number;
  subjectId?: number;
  difficulty?: number;
}) {
  return request.post<any>('/study/practice/questions', data);
}

/**
 * 提交练习结果
 */
export function submitPractice(data: {
  total: number;
  answered: number;
  correct: number;
  elapsed: number;
  answers: string[];
}) {
  return request.post('/study/practice/submit', data);
}

/**
 * 获取练习历史
 */
export function getPracticeHistory(params: { page?: number; pageSize?: number }) {
  return request.get<PageResponse<any>>('/study/practice/history', { params });
}

/**
 * 获取用户学习统计
 */
export function getUserStats() {
  return request.get<UserStats>('/user/stats');
}

/**
 * ==================== 错题本 API ====================
 */

/**
 * 获取错题列表
 */
export function getWrongBooks(params: {
  page?: number;
  pageSize?: number;
  categoryId?: number;
  isMastered?: boolean;
}) {
  return request.get<PageResponse<any>>('/study/wrong', { params });
}

/**
 * 标记已掌握
 */
export function markWrongMastered(questionId: number) {
  return request.put(`/study/wrong/${questionId}/mastered`);
}

/**
 * 批量标记已掌握
 */
export function batchMarkMastered(questionIds: number[]) {
  return request.post('/study/wrong/batch-mastered', { questionIds });
}

/**
 * 移除错题
 */
export function deleteWrongBook(questionId: number) {
  return request.delete(`/study/wrong/${questionId}`);
}

/**
 * 获取随机错题（重做）
 */
export function getWrongBookRandom(count = 20) {
  return request.get<any>('/study/wrong/random', { params: { count } });
}

/**
 * 获取错题统计
 */
export function getWrongBookStats() {
  return request.get<any>('/study/wrong/stats');
}

/**
 * ==================== 智能练习 API ====================
 */

/**
 * 智能练习
 */
export function getSmartPractice(data: {
  count?: number;
  categories?: number[];
  subjectIds?: number[];
  mode?: 'review' | 'weak' | 'mixed';
  difficulty?: number;
}) {
  return request.post<any>('/study/practice/smart', data);
}

/**
 * 练习分析
 */
export function getPracticeAnalysis() {
  return request.get<any>('/study/practice/analysis');
}
