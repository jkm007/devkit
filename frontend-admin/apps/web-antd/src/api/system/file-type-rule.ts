import { requestClient } from '#/api/request';

// 文件类型规则类型定义
export interface FileTypeRule {
  id: number;
  extension: string;
  fileType: string;
  description: string;
  status: number;
  createdAt: string;
  updatedAt: string;
}

export interface FileTypeRuleGrouped {
  fileType: string;
  rules: FileTypeRule[];
}

// API 函数

/**
 * 获取所有文件类型规则
 */
export function getAllFileTypeRules() {
  return requestClient.get<FileTypeRule[]>('/system/file-type-rules');
}

/**
 * 获取按类型分组的文件类型规则
 */
export function getGroupedFileTypeRules() {
  return requestClient.get<FileTypeRuleGrouped[]>(
    '/system/file-type-rules/grouped',
  );
}

/**
 * 创建文件类型规则
 */
export function createFileTypeRule(data: Partial<FileTypeRule>) {
  return requestClient.post<FileTypeRule>('/system/file-type-rules', data);
}

/**
 * 更新文件类型规则
 */
export function updateFileTypeRule(id: number, data: Partial<FileTypeRule>) {
  return requestClient.put<FileTypeRule>(
    `/system/file-type-rules/${id}`,
    data,
  );
}

/**
 * 删除文件类型规则
 */
export function deleteFileTypeRule(id: number) {
  return requestClient.delete(`/system/file-type-rules/${id}`);
}

/**
 * 刷新 AutoTagger 的文件类型规则
 */
export function refreshAutoTagger() {
  return requestClient.post<{ count: number }>(
    '/system/file-type-rules/refresh',
  );
}
