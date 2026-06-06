import { requestClient } from '#/api/request';

// ==================== 类型定义 ====================

export namespace FileApi {
  /** 文件夹 */
  export interface Folder {
    id: number;
    name: string;
    parentId: number | null;
    createdAt: string;
    children?: Folder[];
  }

  /** 文件条目 */
  export interface FileEntry {
    id: number;
    name: string;
    folderId: number | null;
    fileType: string;
    fileSize: number;
    contentType: string;
    createdAt: string;
    updatedAt: string;
    thumbnailUrl?: string;
    previewUrl?: string;
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

/** 上传分片 */
export function uploadPart(data: {
  uploadId: string;
  partNumber: number;
  file: Blob | File;
}) {
  const formData = new FormData();
  formData.append('uploadId', data.uploadId);
  formData.append('partNumber', String(data.partNumber));
  formData.append('file', data.file);

  return requestClient.post<FileApi.UploadPartResult>('/files/upload/part', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
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

// ==================== 简单文件上传（用于头像等小文件） ====================

/** 简单上传（不分片） */
export async function simpleUpload(file: File): Promise<FileApi.CompleteUploadResult> {
  // 计算文件 hash
  const arrayBuffer = await file.arrayBuffer();
  const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  const fileHash = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');

  // 先尝试秒传
  const checkResult = await checkUpload({ fileHash, fileSize: file.size });
  if (checkResult.exists && checkResult.fileId && checkResult.url) {
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
  });

  return completeUpload({ uploadId: initResult.uploadId });
}