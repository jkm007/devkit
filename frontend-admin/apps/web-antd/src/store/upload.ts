import { defineStore } from 'pinia';
import { ref } from 'vue';

import type { PartProgressCallback, UploadProgressCallback } from '#/api/file';
import { simpleUpload } from '#/api/file';

/** 分片详情 */
export interface UploadPartDetail {
  partNumber: number;
  status: 'pending' | 'uploading' | 'completed';
  startTime?: number;
  endTime?: number;
  duration?: number; // 毫秒
}

/** 路由信息 */
export interface RoutingInfo {
  driver: string;
  bucket?: string;
  pathPrefix?: string;
  ruleName?: string;
}

/** 上传任务 */
export interface UploadTaskItem {
  id: number;
  uploadId: string;
  fileName: string;
  fileSize: number;
  contentType: string;
  progress: number;
  status: 'uploading' | 'processing' | 'completed' | 'failed' | 'aborted';
  errorMessage?: string;
  totalParts: number;
  uploadedParts: number;
  partDetails: UploadPartDetail[];
  startTime: number;
  endTime?: number;
  totalDuration?: number; // 毫秒
  // 文件对象（用于继续上传）
  file?: File;
  // 路由信息（上传完成后显示）
  routing?: RoutingInfo;
}

export const useUploadStore = defineStore('upload', () => {
  const tasks = ref<UploadTaskItem[]>([]);
  const uploadingCount = ref(0);

  // 添加任务
  function addTask(task: UploadTaskItem) {
    tasks.value.unshift(task);
    // 限制显示数量
    if (tasks.value.length > 20) {
      tasks.value.pop();
    }
  }

  // 更新任务
  function updateTask(uploadId: string, updates: Partial<UploadTaskItem>) {
    const task = tasks.value.find(t => t.uploadId === uploadId);
    if (task) {
      Object.assign(task, updates);
    }
  }

  // 移除任务
  function removeTask(uploadId: string) {
    const index = tasks.value.findIndex(t => t.uploadId === uploadId);
    if (index !== -1) {
      tasks.value.splice(index, 1);
    }
  }

  // 获取任务
  function getTask(uploadId: string): UploadTaskItem | undefined {
    return tasks.value.find(t => t.uploadId === uploadId);
  }

  // 执行上传
  async function uploadFile(file: File): Promise<void> {
    const tempId = Date.now();
    const startTime = Date.now();
    const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB
    const totalParts = Math.ceil(file.size / CHUNK_SIZE);

    // 初始化分片详情
    const partDetails: UploadPartDetail[] = Array.from({ length: totalParts }, (_, i) => ({
      partNumber: i + 1,
      status: 'pending' as const,
    }));

    const onProgress: UploadProgressCallback = (event) => {
      updateTask(`temp-${tempId}`, {
        progress: event.percent,
        status: event.percent >= 100 ? 'processing' : 'uploading',
      });
    };

    const onPartProgress: PartProgressCallback = (event) => {
      const partIndex = event.partNumber - 1;
      if (partIndex >= 0 && partIndex < partDetails.length) {
        const part = partDetails[partIndex];
        if (part) {
          if (event.status === 'start') {
            part.status = 'uploading';
            part.startTime = event.startTime;
          } else if (event.status === 'completed') {
            part.status = 'completed';
            part.endTime = event.endTime;
            part.duration = event.duration!;
          }
        }
      }
      const uploadedCount = partDetails.filter(p => p.status === 'completed').length;
      updateTask(`temp-${tempId}`, {
        uploadedParts: uploadedCount,
        partDetails: [...partDetails],
      });
    };

    // 添加任务
    addTask({
      id: tempId,
      uploadId: `temp-${tempId}`,
      fileName: file.name,
      fileSize: file.size,
      contentType: file.type,
      progress: 0,
      status: 'uploading',
      totalParts,
      uploadedParts: 0,
      partDetails,
      startTime,
      file,
    });

    uploadingCount.value++;

    try {
      const result = await simpleUpload(file, onProgress, onPartProgress);
      console.log('上传成功:', result);

      const endTime = Date.now();
      updateTask(`temp-${tempId}`, {
        progress: 100,
        status: 'completed',
        endTime,
        totalDuration: endTime - startTime,
        uploadedParts: totalParts,
        file: undefined, // 清除文件引用
        routing: result.routing, // 保存路由信息
      });

      // 3秒后移除完成的任务
      setTimeout(() => {
        removeTask(`temp-${tempId}`);
      }, 3000);
    } catch (err) {
      console.error('上传失败:', err);

      const endTime = Date.now();
      updateTask(`temp-${tempId}`, {
        status: 'failed',
        errorMessage: String(err),
        endTime,
        totalDuration: endTime - startTime,
        file: undefined,
      });

      // 10秒后移除失败的任务
      setTimeout(() => {
        removeTask(`temp-${tempId}`);
      }, 10000);

      throw err;
    } finally {
      uploadingCount.value--;
    }
  }

  return {
    tasks,
    uploadingCount,
    addTask,
    updateTask,
    removeTask,
    getTask,
    uploadFile,
  };
});
