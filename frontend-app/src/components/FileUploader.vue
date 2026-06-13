<template>
  <view class="file-uploader">
    <view class="upload-area" @click="chooseFile">
      <text class="upload-icon">📎</text>
      <text class="upload-text">{{ placeholder || '点击选择文件' }}</text>
    </view>

    <!-- 文件列表 -->
    <view v-if="fileList.length" class="file-list">
      <view v-for="(file, index) in fileList" :key="index" class="file-item">
        <view class="file-info">
          <text class="file-icon">{{ getFileIcon(file.type) }}</text>
          <text class="file-name">{{ file.name }}</text>
          <text class="file-size">{{ formatSize(file.size) }}</text>
        </view>
        <view class="file-actions">
          <!-- 上传进度 -->
          <view v-if="file.status === 'uploading'" class="progress-bar">
            <view class="progress-fill" :style="{ width: file.progress + '%' }"></view>
            <text class="progress-text">{{ file.progress }}%</text>
          </view>
          <!-- 状态图标 -->
          <text v-else-if="file.status === 'success'" class="status-icon success">✅</text>
          <text v-else-if="file.status === 'error'" class="status-icon error" @click="retryUpload(index)">🔄</text>
          <!-- 删除按钮 -->
          <text class="delete-btn" @click="removeFile(index)">✕</text>
        </view>
      </view>
    </view>

    <!-- 提示 -->
    <view v-if="tips" class="upload-tips">
      <text class="tips-text">{{ tips }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { uploadFile, initChunkUpload, uploadChunk, completeChunkUpload } from '@/api/file';
import type { FileInfo } from '@/api/types';

interface UploadFile {
  name: string;
  size: number;
  type: string;
  path: string;
  status: 'pending' | 'uploading' | 'success' | 'error';
  progress: number;
  fileId?: number;
}

const props = defineProps<{
  accept?: string;          // 接受的文件类型，如 'image/*,video/*'
  maxSize?: number;         // 单文件最大大小（MB），默认 50MB
  maxCount?: number;        // 最大文件数，默认 9
  folder?: string;          // 上传到哪个文件夹
  chunkThreshold?: number;  // 分片上传阈值（MB），默认 10MB
  placeholder?: string;
  tips?: string;
}>();

const emit = defineEmits<{
  change: (files: FileInfo[]) => void;
  success: (file: FileInfo) => void;
  error: (error: Error) => void;
}>();

const fileList = ref<UploadFile[]>([]);
const MAX_CHUNK_SIZE = 5 * 1024 * 1024; // 5MB per chunk

const maxSizeBytes = computed(() => (props.maxSize || 50) * 1024 * 1024);
const maxCount = computed(() => props.maxCount || 9);
const chunkThreshold = computed(() => (props.chunkThreshold || 10) * 1024 * 1024);

function chooseFile() {
  if (fileList.value.length >= maxCount.value) {
    uni.showToast({ title: `最多上传 ${maxCount.value} 个文件`, icon: 'none' });
    return;
  }

  // #ifdef MP-WEIXIN
  uni.chooseMessageFile({
    count: maxCount.value - fileList.value.length,
    type: 'file',
    extension: getExtensions(),
    success: (res) => handleFiles(res.tempFiles),
  });
  // #endif

  // #ifdef H5
  // H5 使用 input 选择（通过 uni.chooseFile）
  uni.chooseFile({
    count: maxCount.value - fileList.value.length,
    extension: getExtensions(),
    success: (res) => handleFiles(res.tempFiles),
  });
  // #endif

  // #ifdef APP-PLUS
  uni.chooseFile({
    count: maxCount.value - fileList.value.length,
    extension: getExtensions(),
    success: (res) => handleFiles(res.tempFiles),
  });
  // #endif
}

function getExtensions(): string[] {
  if (!props.accept) return [];
  return props.accept.split(',').map(a => a.trim().replace('*.', ''));
}

async function handleFiles(files: any[]) {
  for (const f of files) {
    if (f.size > maxSizeBytes.value) {
      uni.showToast({ title: `文件 ${f.name} 超过大小限制`, icon: 'none' });
      continue;
    }

    const uploadFile: UploadFile = {
      name: f.name,
      size: f.size,
      type: getFileType(f.name),
      path: f.path || f.url,
      status: 'pending',
      progress: 0,
    };

    fileList.value.push(uploadFile);
    startUpload(fileList.value.length - 1);
  }
}

async function startUpload(index: number) {
  const file = fileList.value[index];
  if (!file) return;

  file.status = 'uploading';
  file.progress = 0;

  try {
    // 判断是否需要分片上传
    if (file.size > chunkThreshold.value) {
      await chunkUpload(index);
    } else {
      await simpleUpload(index);
    }
  } catch (error) {
    file.status = 'error';
    emit('error', error as Error);
  }
}

async function simpleUpload(index: number) {
  const file = fileList.value[index];
  const info = await uploadFile({ path: file.path, name: file.name } as any, props.folder);
  file.status = 'success';
  file.progress = 100;
  file.fileId = info.id;
  emit('success', info);
  emitNotifyChange();
}

async function chunkUpload(index: number) {
  const file = fileList.value[index];

  // 1. 初始化
  const initRes = await initChunkUpload({
    fileName: file.name,
    fileSize: file.size,
    chunkSize: MAX_CHUNK_SIZE,
    folder: props.folder,
  });

  // 2. 上传分片
  for (let i = 0; i < initRes.totalChunks; i++) {
    if (initRes.uploadedChunks.includes(i)) continue; // 跳过已上传的（断点续传）

    const start = i * initRes.chunkSize;
    const end = Math.min(start + initRes.chunkSize, file.size);
    const chunk = file.path.slice(start, end);

    await uploadChunk({ path: chunk } as any, {
      uploadId: initRes.uploadId,
      chunkIndex: i,
    });

    file.progress = Math.round(((i + 1) / initRes.totalChunks) * 100);
  }

  // 3. 完成
  const info = await completeChunkUpload(initRes.uploadId);
  file.status = 'success';
  file.progress = 100;
  file.fileId = info.id;
  emit('success', info);
  emitNotifyChange();
}

async function retryUpload(index: number) {
  fileList.value[index].status = 'pending';
  fileList.value[index].progress = 0;
  await startUpload(index);
}

function removeFile(index: number) {
  fileList.value.splice(index, 1);
  emitNotifyChange();
}

function getFileIcon(type: string): string {
  if (type.startsWith('image')) return '🖼️';
  if (type.startsWith('video')) return '🎬';
  if (type.startsWith('audio')) return '🎵';
  if (type.includes('pdf')) return '📕';
  if (type.includes('word') || type.includes('document')) return '📘';
  if (type.includes('excel') || type.includes('sheet')) return '📗';
  if (type.includes('zip') || type.includes('rar')) return '📦';
  return '📄';
}

function getFileType(filename: string): string {
  const ext = filename.split('.').pop()?.toLowerCase() || '';
  const map: Record<string, string> = {
    jpg: 'image/jpeg', jpeg: 'image/jpeg', png: 'image/png', gif: 'image/gif', webp: 'image/webp',
    mp4: 'video/mp4', webm: 'video/webm',
    mp3: 'audio/mpeg', wav: 'audio/wav',
    pdf: 'application/pdf',
    doc: 'application/msword', docx: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    xls: 'application/vnd.ms-excel', xlsx: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    zip: 'application/zip', rar: 'application/x-rar-compressed',
  };
  return map[ext] || 'application/octet-stream';
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

function emitNotifyChange() {
  const successFiles = fileList.value
    .filter(f => f.status === 'success' && f.fileId)
    .map(f => ({ id: f.fileId!, name: f.name, size: f.size, fileType: f.type, url: '', uploadedAt: '' } as FileInfo));
  emit('change', successFiles);
}
</script>

<style lang="scss" scoped>
.file-uploader { width: 100%; }

.upload-area { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 24px; border: 2px dashed #d9d9d9; border-radius: 8px; background: #fafafa; transition: border-color 0.3s; }
.upload-area:active { border-color: #1890ff; background: #f0f9ff; }
.upload-icon { font-size: 32px; margin-bottom: 8px; }
.upload-text { font-size: 14px; color: #666; }

.file-list { margin-top: 12px; }
.file-item { display: flex; flex-direction: column; padding: 12px; margin-bottom: 8px; background: #f9f9f9; border-radius: 8px; }
.file-info { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.file-icon { font-size: 20px; }
.file-name { flex: 1; font-size: 14px; color: #333; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.file-size { font-size: 12px; color: #999; }

.file-actions { display: flex; align-items: center; gap: 8px; }
.progress-bar { flex: 1; height: 4px; background: #e8e8e8; border-radius: 2px; position: relative; }
.progress-fill { height: 100%; background: #1890ff; border-radius: 2px; transition: width 0.3s; }
.progress-text { position: absolute; right: 0; top: -18px; font-size: 11px; color: #1890ff; }

.status-icon { font-size: 16px; }
.status-icon.success { color: #52c41a; }
.status-icon.error { color: #ff4d4f; }
.delete-btn { font-size: 16px; color: #999; padding: 4px; }

.upload-tips { margin-top: 8px; padding: 8px 12px; background: #fffbe6; border-radius: 4px; }
.tips-text { font-size: 12px; color: #d48806; }
</style>
