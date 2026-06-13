/**
 * 文件管理 API
 */
import { request } from './request';
import type { FileInfo, UploadInitResponse, UploadChunkResponse } from '@/api/types';

/**
 * 上传文件（单文件，小文件）
 */
export function uploadFile(file: File | UniApp.ChooseFileSuccessCallbackResultFile, folder?: string) {
  return new Promise<FileInfo>((resolve, reject) => {
    const token = uni.getStorageSync('access_token');
    uni.uploadFile({
      url: '/api/v1/files/upload',
      filePath: (file as any).path || (file as any).url,
      name: 'file',
      formData: { folder: folder || '' },
      header: { Authorization: `Bearer ${token}` },
      success: (res) => {
        try {
          const data = JSON.parse(res.data);
          if (data.code === 0) resolve(data.data);
          else reject(new Error(data.message));
        } catch { reject(new Error('解析响应失败')); }
      },
      fail: (err) => reject(err),
    });
  });
}

/**
 * 初始化分片上传（大文件）
 */
export function initChunkUpload(params: {
  fileName: string;
  fileSize: number;
  chunkSize: number;
  folder?: string;
}) {
  return request.post<UploadInitResponse>('/api/v1/files/upload/chunk/init', params);
}

/**
 * 上传分片
 */
export function uploadChunk(file: File, params: {
  uploadId: string;
  chunkIndex: number;
  chunkHash?: string;
}) {
  return new Promise<UploadChunkResponse>((resolve, reject) => {
    const token = uni.getStorageSync('access_token');
    uni.uploadFile({
      url: '/api/v1/files/upload/chunk',
      filePath: (file as any).path || (file as any).url,
      name: 'chunk',
      formData: {
        uploadId: params.uploadId,
        chunkIndex: String(params.chunkIndex),
        chunkHash: params.chunkHash || '',
      },
      header: { Authorization: `Bearer ${token}` },
      success: (res) => {
        try {
          const data = JSON.parse(res.data);
          if (data.code === 0) resolve(data.data);
          else reject(new Error(data.message));
        } catch { reject(new Error('解析响应失败')); }
      },
      fail: (err) => reject(err),
    });
  });
}

/**
 * 完成分片上传
 */
export function completeChunkUpload(uploadId: string) {
  return request.post<FileInfo>(`/api/v1/files/upload/chunk/complete/${uploadId}`);
}

/**
 * 获取文件信息
 */
export function getFileInfo(fileId: number) {
  return request.get<FileInfo>(`/api/v1/files/${fileId}`);
}

/**
 * 获取文件下载/预览链接
 */
export function getFileUrl(fileId: number, type: 'download' | 'preview' = 'preview') {
  return request.get<{ url: string; expiresAt: string }>(`/api/v1/files/${fileId}/${type}`);
}

/**
 * 删除文件（移入回收站）
 */
export function deleteFile(fileId: number) {
  return request.delete(`/api/v1/files/${fileId}`);
}

/**
 * 批量删除文件
 */
export function batchDeleteFiles(fileIds: number[]) {
  return request.post('/api/v1/files/batch/delete', { ids: fileIds });
}

/**
 * 获取我的文件列表
 */
export function getMyFiles(params: {
  page?: number;
  pageSize?: number;
  folder?: string;
  fileType?: string;
}) {
  return request.get<any>('/api/v1/files/my', { params });
}
