import { requestClient } from '#/api/request';
import { useAccessStore } from '@vben/stores';

export interface UploadProgressEvent {
  loaded: number;
  total: number;
  percent: number;
}

export type UploadProgressCallback = (event: UploadProgressEvent) => void;

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

  /** 文件条目 */
  export interface FileEntry {
    id: number;
    name: string;
    folderId: number | null;
    size: number;
    contentType: string;
    createdAt: string;
    updatedAt: string;
    previewUrl?: string;
    uploaderName?: string;
    uploaderAvatar?: string;
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

  /** 完成上传结果 */
  export interface CompleteUploadResult {
    fileId: number;
    url: string;
    name: string;
    size: number;
    contentType: string;
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

/** 完成上传 */
export function completeUpload(data: { uploadId: string }) {
  return requestClient.post<FileApi.CompleteUploadResult>('/files/upload/complete', data);
}

/** 取消上传 */
export function abortUpload(data: { uploadId: string }) {
  return requestClient.post('/files/upload/abort', data);
}

/** 获取上传状态 */
export function getUploadStatus(params: { uploadId: string }) {
  return requestClient.get<FileApi.UploadStatus>('/files/upload/status', { params });
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
  // 非 HTTPS 环境：使用文件名+大小+时间戳生成简单标识
  return `${file.name}_${file.size}_${Date.now()}`;
}

/** 简单上传（不分片） */
export async function simpleUpload(
  file: File,
  onProgress?: UploadProgressCallback,
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

  // 小文件单分片上传
  const initResult = await initUpload({
    fileName: file.name,
    fileSize: file.size,
    fileHash,
    contentType: file.type,
    totalParts: 1,
  });

  await uploadPart({
    uploadId: initResult.uploadId,
    partNumber: 1,
    file,
    onProgress,
  });

  return completeUpload({ uploadId: initResult.uploadId });
}