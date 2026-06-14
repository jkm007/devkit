/**
 * 用户 API
 */
import { request } from './request';

/**
 * 获取绑定的分类
 */
export function getCategoryBindings() {
  return request.get<any[]>('/user/category-bindings');
}

/**
 * 绑定分类
 */
export function bindCategory(data: { categoryId: number; isPrimary?: boolean }) {
  return request.post('/user/category-bindings', data);
}

/**
 * 解绑分类
 */
export function unbindCategory(id: number) {
  return request.delete(`/user/category-bindings/${id}`);
}

/**
 * 设为主分类
 */
export function setPrimaryCategory(id: number) {
  return request.put(`/user/category-bindings/${id}`);
}
