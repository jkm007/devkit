import { requestClient } from '#/api/request';

// 标签类型定义
export interface Tag {
  id: number;
  tagKey: string;
  tagValue: string;
  tagName: string;
  icon: string;
  color: string;
  description: string;
  isSystem: boolean;
  sortOrder: number;
  status: number;
  createdAt: string;
  updatedAt: string;
}

export interface TagUsageStat {
  id: number;
  tagKey: string;
  tagValue: string;
  tagName: string;
  icon: string;
  color: string;
  fileCount: number;
}

// 路由规则类型定义
export interface TagCondition {
  key: string;
  value: string;
}

export interface TagRouting {
  id: number;
  ruleName: string;
  description: string;
  priority: number;
  matchType: 'all' | 'any' | 'exact';
  conditions: {
    tags: TagCondition[];
  };
  driver: string;
  bucket: string;
  pathPrefix: string;
  extraConfig: Record<string, any>;
  isDefault: boolean;
  status: number;
  createdAt: string;
  updatedAt: string;
}

// API 函数

// 标签管理（管理员接口，需要 storage:bucket:view 权限）
export function getAllTags() {
  return requestClient.get<Tag[]>('/system/tags');
}

// 标签查询（普通用户接口，用于文件管理的标签筛选）
export function getTagsForUser() {
  return requestClient.get<Tag[]>('/tags');
}

export function getGroupedTags() {
  return requestClient.get<Record<string, Tag[]>>('/system/tags/grouped');
}

export function getTagsByKey(key: string) {
  return requestClient.get<Tag[]>(`/system/tags/key/${key}`);
}

export function getTagById(id: number) {
  return requestClient.get<Tag>(`/system/tags/${id}`);
}

export function createTag(data: Partial<Tag>) {
  return requestClient.post<Tag>('/system/tags', data);
}

export function updateTag(id: number, data: Partial<Tag>) {
  return requestClient.put<Tag>(`/system/tags/${id}`, data);
}

export function deleteTag(id: number) {
  return requestClient.delete(`/system/tags/${id}`);
}

export function getTagUsageStats() {
  return requestClient.get<TagUsageStat[]>('/system/tags/stats');
}

// 文件标签管理
export function getFileTags(fileId: number) {
  return requestClient.get<Tag[]>(`/files/${fileId}/tags`);
}

export function addFileTag(fileId: number, tagId: number) {
  return requestClient.post(`/files/${fileId}/tags`, { tagId });
}

export function removeFileTag(fileId: number, tagId: number) {
  return requestClient.delete(`/files/${fileId}/tags/${tagId}`);
}

export function batchUpdateFileTags(fileId: number, tagIds: number[]) {
  return requestClient.put(`/files/${fileId}/tags`, { tagIds });
}

// 路由规则管理
export function getAllRoutingRules() {
  return requestClient.get<TagRouting[]>('/system/routing-rules');
}

export function getRoutingRuleById(id: number) {
  return requestClient.get<TagRouting>(`/system/routing-rules/${id}`);
}

export function createRoutingRule(data: Partial<TagRouting>) {
  return requestClient.post<TagRouting>('/system/routing-rules', data);
}

export function updateRoutingRule(id: number, data: Partial<TagRouting>) {
  return requestClient.put<TagRouting>(`/system/routing-rules/${id}`, data);
}

export function deleteRoutingRule(id: number) {
  return requestClient.delete(`/system/routing-rules/${id}`);
}

export function updateRoutingRuleStatus(id: number, status: number) {
  return requestClient.put(`/system/routing-rules/${id}/status`, { status });
}

export function updateRoutingRulePriority(id: number, priority: number) {
  return requestClient.put(`/system/routing-rules/${id}/priority`, {
    priority,
  });
}

export function batchUpdatePriority(priorities: Record<number, number>) {
  return requestClient.post('/system/routing-rules/batch-priority', {
    priorities,
  });
}

export function testRoutingRule(id: number, tags: TagCondition[]) {
  return requestClient.post<{ matched: boolean }>(
    `/system/routing-rules/${id}/test`,
    { tags },
  );
}

export function testRoute(
  fileName: string,
  contentType: string,
  source: string,
) {
  return requestClient.post('/system/routing-rules/test-route', {
    fileName,
    contentType,
    source,
  });
}
