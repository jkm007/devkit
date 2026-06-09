import { requestClient } from '#/api/request';
import { useAccessStore } from '@vben/stores';

export interface UploadProgressEvent {
  loaded: number;
  total: number;
  percent: number;
}

export type UploadProgressCallback = (event: UploadProgressEvent) => void;

/** 分片进度事件 */
export interface PartProgressEvent {
  partNumber: number;
  totalParts: number;
  status: 'start' | 'completed';
  startTime: number;
  endTime?: number;
  duration?: number; // 毫秒
}

export type PartProgressCallback = (event: PartProgressEvent) => void;

// ==================== 类型定义 ====================

export namespace FileApi {
  /** 文件夹 */
  export interface Folder {
    id: number;
    name: string;
    parentId: number | null;
    type?: string;
    createdAt: string;
    children?: Folder[];
  }

  /** 标签信息 */
  export interface TagInfo {
    id: number;
    key: string;
    value: string;
    name: string;
    icon: string;
    color: string;
  }

  /** 文件条目 */
  export interface FileEntry {
    id: number;
    name: string;
    folderId: number | null;
    size: number;
    contentType: string;
    storageType: string;
    createdAt: string;
    updatedAt: string;
    previewUrl?: string;
    uploaderName?: string;
    uploaderAvatar?: string;
    tags?: TagInfo[];
  }

  /** 文件资产 */
  export interface FileAsset {
    id: number;
    objectKey: string;
    contentType: string;
    fileSize: number;
    fileHash: string;
    storageType: string;
    createdAt: string;
  }

  /** 媒体信息 */
  export interface MediaInfo {
    id: number;
    fileId: number;
    duration: number;
    width: number;
    height: number;
    bitrate: number;
    codec: string;
    transcodeStatus: string;
    hlsPath: string;
    createdAt: string;
  }

  /** 秒传检查结果 */
  export interface CheckUploadResult {
    exists: boolean;
    fileId?: number;
    url?: string;
  }

  /** 初始化上传结果 */
  export interface InitUploadResult {
    uploadId: string;
    uploadedParts: number[];
    expiresAt: string;
  }

  /** 分片上传结果 */
  export interface UploadPartResult {
    partNumber: number;
    etag: string;
    uploaded: boolean;
  }

  /** 路由信息 */
  export interface RoutingInfo {
    driver: string;
    bucket?: string;
    pathPrefix?: string;
    ruleName?: string;
  }

  /** 完成上传结果 */
  export interface CompleteUploadResult {
    fileId: number;
    url: string;
    name: string;
    size: number;
    contentType: string;
    routing?: RoutingInfo;
  }

  /** 上传状态 */
  export interface UploadStatus {
    uploadId: string;
    status: string;
    uploadedParts: number[];
    totalParts: number;
    fileSize: number;
    uploadedSize: number;
    progress: number;
    expiresAt: string;
  }

  /** 文件列表请求参数 */
  export interface ListFilesParams {
    folderId?: number;
    page?: number;
    pageSize?: number;
    keyword?: string;
    contentType?: string;
    scope?: 'own' | 'all'; // own=自己的文件, all=所有文件（需要权限）
    tagKeys?: string; // 标签筛选，格式: "type:image,source:user"
  }

  /** 文件列表响应 */
  export interface ListFilesResponse {
    items: FileEntry[];
    total: number;
  }

  /** 流地址响应 */
  export interface StreamResponse {
    type: 'hls' | 'original';
    url: string;
  }

  /** 下载响应 */
  export interface DownloadResponse {
    url: string;
    fileName: string;
    contentType: string;
    size: number;
  }
}

// ==================== 分片上传 ====================

/** 秒传检查 */
export function checkUpload(data: {
  fileHash: string;
  fileSize: number;
}) {
  return requestClient.post<FileApi.CheckUploadResult>('/files/upload/check', data);
}

/** 初始化分片上传 */
export function initUpload(data: {
  fileName: string;
  fileSize: number;
  fileHash: string;
  contentType?: string;
  totalParts: number;
  folderId?: number;
}) {
  return requestClient.post<FileApi.InitUploadResult>('/files/upload/init', data);
}

/** 上传分片 - 使用原生 fetch 确保 FormData 正确发送 */
export async function uploadPart(data: {
  uploadId: string;
  partNumber: number;
  file: Blob | File;
  onProgress?: UploadProgressCallback;
}): Promise<FileApi.UploadPartResult> {
  const formData = new FormData();
  formData.append('uploadId', data.uploadId);
  formData.append('partNumber', String(data.partNumber));
  formData.append('file', data.file);

  const accessStore = useAccessStore();
  const token = accessStore.accessToken || '';

  // 如果有进度回调，使用 XMLHttpRequest
  if (data.onProgress) {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();

      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
          data.onProgress!({
            loaded: event.loaded,
            total: event.total,
            percent: Math.round((event.loaded / event.total) * 100),
          });
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const result = JSON.parse(xhr.responseText);
            if (result.code !== 0) {
              reject(new Error(result.message || result.error || '上传分片失败'));
            } else {
              resolve(result.data);
            }
          } catch {
            reject(new Error('解析响应失败'));
          }
        } else {
          reject(new Error(`上传失败: ${xhr.status}`));
        }
      });

      xhr.addEventListener('error', () => reject(new Error('上传失败')));
      xhr.addEventListener('abort', () => reject(new Error('上传已取消')));

      xhr.open('POST', '/api/files/upload/part');
      xhr.setRequestHeader('Authorization', `Bearer ${token}`);
      xhr.send(formData);
    });
  }

  const response = await fetch('/api/files/upload/part', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
    body: formData,
  });

  // 检查响应状态
  if (!response.ok) {
    const text = await response.text();
    // 如果返回 HTML 说明请求未正确到达后端
    if (text.startsWith('<') || text.startsWith('<!')) {
      throw new Error('请求未到达服务器，请检查网络或刷新页面重试');
    }
    try {
      const json = JSON.parse(text);
      throw new Error(json.message || json.error || `上传失败: ${response.status}`);
    } catch {
      throw new Error(`上传失败: ${response.status}`);
    }
  }

  const result = await response.json();
  if (result.code !== 0) {
    throw new Error(result.message || result.error || '上传分片失败');
  }
  return result.data;
}

/** 完成上传 - 使用原生 fetch 避免超时问题（大文件合并需要较长时间） */
export async function completeUpload(data: { uploadId: string }): Promise<FileApi.CompleteUploadResult> {
  const accessStore = useAccessStore();
  const token = accessStore.accessToken || '';

  const response = await fetch('/api/files/upload/complete', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const text = await response.text();
    if (text.startsWith('<') || text.startsWith('<!')) {
      throw new Error('请求未到达服务器，请检查网络或刷新页面重试');
    }
    try {
      const json = JSON.parse(text);
      throw new Error(json.message || json.error || `完成上传失败: ${response.status}`);
    } catch {
      throw new Error(`完成上传失败: ${response.status}`);
    }
  }

  const result = await response.json();
  if (result.code !== 0) {
    throw new Error(result.message || result.error || '完成上传失败');
  }
  return result.data;
}

/** 取消上传 */
export function abortUpload(data: { uploadId: string }) {
  return requestClient.post('/files/upload/abort', data);
}

/** 获取上传状态 */
export function getUploadStatus(params: { uploadId: string }) {
  return requestClient.get<FileApi.UploadStatus>('/files/upload/status', { params });
}

/** 上传任务状态 */
export interface UploadTaskStatus {
  id: number;
  uploadId: string;
  fileName: string;
  fileSize: number;
  contentType: string;
  totalParts: number;
  uploadedParts: number;
  progress: number;
  status: 'uploading' | 'processing' | 'completed' | 'failed' | 'aborted';
  errorMessage?: string;
  completedAt?: string;
  createdAt: string;
}

/** 获取用户的上传任务列表 */
export function getUploadTasks(limit?: number) {
  return requestClient.get<UploadTaskStatus[]>('/files/upload/tasks', {
    params: { limit: limit || 20 },
  });
}

/** 获取单个上传任务状态 */
export function getUploadTaskById(id: number) {
  return requestClient.get<UploadTaskStatus>(`/files/upload/tasks/${id}`);
}

// ==================== 文件夹管理 ====================

/** 创建文件夹 */
export function createFolder(data: { name: string; parentId?: number }) {
  return requestClient.post<FileApi.Folder>('/files/folder', data);
}

/** 获取文件夹树 */
export function getFolderTree() {
  return requestClient.get<FileApi.Folder[]>('/files/tree');
}

/** 重命名文件夹 */
export function renameFolder(id: number, data: { name: string }) {
  return requestClient.put(`/files/folder/${id}`, data);
}

/** 删除文件夹 */
export function deleteFolder(id: number) {
  return requestClient.delete(`/files/folder/${id}`);
}

// ==================== 文件管理 ====================

/** 获取文件列表 */
export function listFiles(params: FileApi.ListFilesParams) {
  return requestClient.get<FileApi.ListFilesResponse>('/files/list', { params });
}

/** 移动文件 */
export function moveFile(data: { fileId: number; targetFolderId?: number }) {
  return requestClient.post('/files/move', data);
}

/** 检查文件是否有活跃分享 */
export function checkFileShares(id: number) {
  return requestClient.get<{ shareCount: number }>(`/files/${id}/shares`);
}

/** 删除文件 */
export function deleteFile(id: number) {
  return requestClient.delete(`/files/${id}`);
}

/** 批量删除文件 */
export function batchDeleteFiles(fileIds: number[]) {
  return requestClient.post<{ deleted: number; errors: string[] }>('/files/batch-delete', { fileIds });
}

/** 批量移动文件 */
export function batchMoveFiles(fileIds: number[], targetFolderId?: number) {
  return requestClient.post<{ moved: number; errors: string[] }>('/files/batch-move', { fileIds, targetFolderId });
}

// ==================== 媒体文件 ====================

/** 获取媒体信息 */
export function getMediaInfo(id: number) {
  return requestClient.get<FileApi.MediaInfo>(`/files/${id}/metadata`);
}

/** 获取视频流地址 */
export function getStream(id: number) {
  return requestClient.get<FileApi.StreamResponse>(`/files/${id}/stream`);
}

/** 下载文件 */
export function downloadFile(id: number) {
  return requestClient.get<FileApi.DownloadResponse>(`/files/${id}/download`);
}

/** 查看文件（带认证） */
export function viewFile(id: number) {
  return requestClient.get(`/files/${id}/view`, { responseType: 'blob' });
}

/** 获取预签名 URL（用于视频流式播放） */
export function getPreviewURL(id: number, expires?: number) {
  return requestClient.get<{ url: string; contentType: string; name: string }>(`/files/${id}/preview-url`, {
    params: expires ? { expires } : undefined,
  });
}

/** 获取文件直链（presigned URL） */
export function getDirectUrl(id: number, expires?: number) {
  return requestClient.get<{ url: string; strategy: string; expiresIn?: number; contentType: string; name: string }>(`/files/${id}/direct-url`, {
    params: expires ? { expires } : undefined,
  });
}

// ==================== 分享 ====================

export interface ShareInfo {
  shareCode: string;
  type: 'file' | 'folder';
  fileName?: string;
  folderName?: string;
  fileSize?: number;
  contentType?: string;
  sharerName: string;
  sharerAvatar: string;
  createdAt: string;
  expireAt?: string;
}

/** 创建文件分享 */
export function createFileShare(id: number, data?: { expireHours?: number; maxAccess?: number }) {
  return requestClient.post<{ shareCode: string; shareUrl: string }>(`/files/${id}/share`, data || {});
}

/** 创建文件夹分享 */
export function createFolderShare(id: number, data?: { expireHours?: number; maxAccess?: number }) {
  return requestClient.post<{ shareCode: string; shareUrl: string }>(`/folders/${id}/share`, data || {});
}

/** 获取分享信息（公开） */
export function getShareInfo(code: string) {
  return requestClient.get<ShareInfo>(`/share/${code}`);
}

/** 获取分享文件夹内的文件列表（公开） */
export function getShareFolderFiles(code: string) {
  return requestClient.get<any[]>(`/share/${code}/files`);
}

/** 获取我的分享列表 */
export function getMyShares() {
  return requestClient.get(`/my-shares`);
}

/** 删除分享 */
export function deleteShare(id: number) {
  return requestClient.delete(`/shares/${id}`);
}

/** 分享列表项 */
export interface ShareListItem {
  id: number;
  shareCode: string;
  shareUrl: string;
  type: 'file' | 'folder';
  fileId?: number;
  folderId?: number;
  fileName?: string;
  folderName?: string;
  fileSize?: number;
  contentType?: string;
  status: number; // 1=有效, 2=已过期, 3=已禁用
  expireAt?: string;
  accessCount: number;
  maxAccess: number;
  accessedAt?: string;
  createdAt: string;
  // 分享人信息
  userId: number;
  userName?: string;
  userAvatar?: string;
}

/** 分享列表响应 */
export interface ShareListResponse {
  items: ShareListItem[];
  total: number;
}

/** 获取用户分享列表（带分页） */
export function getUserShares(params?: { page?: number; pageSize?: number; scope?: 'all' | 'own' }) {
  return requestClient.get<ShareListResponse>('/files/shares', { params });
}

/** 续签分享 */
export function renewShare(id: number, expireHours: number) {
  return requestClient.put(`/files/shares/${id}/renew`, { expireHours });
}

/** 立即过期分享 */
export function expireShare(id: number) {
  return requestClient.put(`/files/shares/${id}/expire`);
}

/** 修改分享到期时间 */
export function updateShareExpiry(id: number, expireAt?: string) {
  return requestClient.put(`/files/shares/${id}/expiry`, { expireAt });
}

/** 禁用分享 */
export function disableShare(id: number) {
  return requestClient.put(`/files/shares/${id}/disable`);
}

/** 启用分享 */
export function enableShare(id: number) {
  return requestClient.put(`/files/shares/${id}/enable`);
}

// ==================== 文件标签管理 ====================

/** 获取文件标签 */
export function getFileTags(fileId: number) {
  return requestClient.get<FileApi.TagInfo[]>(`/files/${fileId}/tags`);
}

/** 添加文件标签 */
export function addFileTag(fileId: number, tagId: number) {
  return requestClient.post(`/files/${fileId}/tags`, { tagId });
}

/** 移除文件标签 */
export function removeFileTag(fileId: number, tagId: number) {
  return requestClient.delete(`/files/${fileId}/tags/${tagId}`);
}

/** 批量更新文件标签 */
export function batchUpdateFileTags(fileId: number, tagIds: number[]) {
  return requestClient.put(`/files/${fileId}/tags`, { tagIds });
}

// ==================== 回收站 ====================

/** 回收站列表项 */
export interface RecycleBinItem {
  id: number;
  name: string;
  size: number;
  contentType: string;
  folderId: number;
  userId: number;
  userName?: string;
  deletedAt: string;
  recycleExpireAt?: string;
  daysRemaining: number;
}

/** 回收站列表响应 */
export interface RecycleBinListResponse {
  items: RecycleBinItem[];
  total: number;
}

/** 获取回收站列表 */
export function getRecycleBinList(params?: { page?: number; pageSize?: number; scope?: 'all' | 'own' }) {
  return requestClient.get<RecycleBinListResponse>('/files/recycle/list', { params });
}

/** 获取回收站文件数量 */
export function getRecycleBinCount() {
  return requestClient.get<{ count: number }>('/files/recycle/count');
}

/** 恢复文件 */
export function restoreFile(id: number) {
  return requestClient.post(`/files/recycle/restore/${id}`);
}

/** 批量恢复文件 */
export function batchRestoreFiles(fileIds: number[]) {
  return requestClient.post<{ restored: number; errors: string[] }>('/files/recycle/batch-restore', { fileIds });
}

/** 永久删除文件 */
export function permanentDeleteFile(id: number) {
  return requestClient.delete(`/files/recycle/${id}`);
}

/** 批量永久删除文件 */
export function batchPermanentDeleteFiles(fileIds: number[]) {
  return requestClient.post<{ deleted: number; errors: string[] }>('/files/recycle/batch-delete', { fileIds });
}

/** 清空回收站 */
export function emptyRecycleBin() {
  return requestClient.delete('/files/recycle/empty');
}

// ==================== 简单文件上传（用于头像等小文件） ====================

/** 计算文件 hash（兼容非 HTTPS 环境） */
async function calculateFileHash(file: File): Promise<string> {
  // crypto.subtle 只在安全上下文（HTTPS 或 localhost）中可用
  if (crypto?.subtle?.digest) {
    try {
      const arrayBuffer = await file.arrayBuffer();
      const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer);
      const hashArray = Array.from(new Uint8Array(hashBuffer));
      return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
    } catch {
      // fallback
    }
  }
  // 非 HTTPS 环境：使用文件名+大小+最后修改时间生成确定性标识
  return `${file.name}_${file.size}_${file.lastModified}`;
}

/** 分片大小：5MB */
const CHUNK_SIZE = 5 * 1024 * 1024;

/** 简单上传（小文件，不分片） */
export async function simpleUpload(
  file: File,
  onProgress?: UploadProgressCallback,
  onPartProgress?: PartProgressCallback,
  folderId?: number,
): Promise<FileApi.CompleteUploadResult> {
  // 计算文件 hash
  const fileHash = await calculateFileHash(file);

  // 先尝试秒传
  const checkResult = await checkUpload({ fileHash, fileSize: file.size });
  if (checkResult.exists && checkResult.fileId && checkResult.url) {
    onProgress?.({ loaded: file.size, total: file.size, percent: 100 });
    return {
      fileId: checkResult.fileId,
      url: checkResult.url,
      name: file.name,
      size: file.size,
      contentType: file.type,
    };
  }

  // 判断是否需要分片
  if (file.size <= CHUNK_SIZE) {
    // 小文件单分片上传
    const initResult = await initUpload({
      fileName: file.name,
      fileSize: file.size,
      fileHash,
      contentType: file.type,
      totalParts: 1,
      folderId,
    });

    const partStartTime = Date.now();
    onPartProgress?.({
      partNumber: 1,
      totalParts: 1,
      status: 'start',
      startTime: partStartTime,
    });

    await uploadPart({
      uploadId: initResult.uploadId,
      partNumber: 1,
      file,
      onProgress,
    });

    const partEndTime = Date.now();
    onPartProgress?.({
      partNumber: 1,
      totalParts: 1,
      status: 'completed',
      startTime: partStartTime,
      endTime: partEndTime,
      duration: partEndTime - partStartTime,
    });

    return completeUpload({ uploadId: initResult.uploadId });
  }

  // 大文件分片上传
  return chunkedUpload(file, fileHash, onProgress, onPartProgress, folderId);
}

/** 大文件分片上传 */
async function chunkedUpload(
  file: File,
  fileHash: string,
  onProgress?: UploadProgressCallback,
  onPartProgress?: PartProgressCallback,
  folderId?: number,
): Promise<FileApi.CompleteUploadResult> {
  // 计算分片数量
  const totalParts = Math.ceil(file.size / CHUNK_SIZE);

  // 初始化上传
  const initResult = await initUpload({
    fileName: file.name,
    fileSize: file.size,
    fileHash,
    contentType: file.type,
    totalParts,
    folderId,
  });

  // 已上传的分片
  const uploadedParts = new Set(initResult.uploadedParts || []);
  let uploadedSize = uploadedParts.size * CHUNK_SIZE;

  // 逐个上传分片
  for (let partNumber = 1; partNumber <= totalParts; partNumber++) {
    // 跳过已上传的分片
    if (uploadedParts.has(partNumber)) {
      continue;
    }

    // 计算分片范围
    const start = (partNumber - 1) * CHUNK_SIZE;
    const end = Math.min(start + CHUNK_SIZE, file.size);
    const chunk = file.slice(start, end);

    // 分片开始上传
    const partStartTime = Date.now();
    onPartProgress?.({
      partNumber,
      totalParts,
      status: 'start',
      startTime: partStartTime,
    });

    // 上传分片
    await uploadPart({
      uploadId: initResult.uploadId,
      partNumber,
      file: chunk,
      onProgress: (event) => {
        // 计算总体进度
        const chunkLoaded = event.loaded;
        const totalLoaded = uploadedSize + chunkLoaded;
        const percent = Math.round((totalLoaded / file.size) * 100);
        onProgress?.({
          loaded: totalLoaded,
          total: file.size,
          percent: Math.min(percent, 99), // 最多显示99%，完成后再显示100%
        });
      },
    });

    // 分片上传完成
    const partEndTime = Date.now();
    onPartProgress?.({
      partNumber,
      totalParts,
      status: 'completed',
      startTime: partStartTime,
      endTime: partEndTime,
      duration: partEndTime - partStartTime,
    });

    // 更新已上传大小
    uploadedSize += end - start;
  }

  // 完成上传
  onProgress?.({ loaded: file.size, total: file.size, percent: 100 });
  return completeUpload({ uploadId: initResult.uploadId });
}