<template>
  <view class="file-manage-page">
    <view class="header">
      <text class="title">我的文件</text>
      <view class="upload-btn" @click="showUpload = true">
        <text class="upload-icon">+</text>
      </view>
    </view>

    <!-- 文件类型筛选 -->
    <view class="filter-bar">
      <view v-for="tab in tabs" :key="tab.key" class="filter-tab" :class="{ active: activeTab === tab.key }" @click="activeTab = tab.key">
        <text class="tab-icon">{{ tab.icon }}</text>
        <text class="tab-label">{{ tab.label }}</text>
      </view>
    </view>

    <!-- 文件列表 -->
    <view class="file-list">
      <view v-if="loading" class="loading">加载中...</view>
      <view v-else-if="files.length === 0" class="empty">
        <text class="empty-icon">📁</text>
        <text class="empty-text">暂无文件</text>
      </view>
      <view v-else>
        <view v-for="file in files" :key="file.id" class="file-item" @click="previewFile(file)">
          <view class="file-icon">{{ getFileIcon(file.fileType) }}</view>
          <view class="file-info">
            <text class="file-name">{{ file.originalName }}</text>
            <view class="file-meta">
              <text class="file-size">{{ formatSize(file.fileSize) }}</text>
              <text class="file-date">{{ formatDate(file.uploadedAt) }}</text>
            </view>
          </view>
          <view class="file-actions" @click.stop>
            <text class="action-btn" @click="shareFile(file)">🔗</text>
            <text class="action-btn delete" @click="deleteFileItem(file)">🗑️</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 上传弹窗 -->
    <up-popup v-model:show="showUpload" mode="bottom" round="12">
      <view class="upload-panel">
        <text class="panel-title">上传文件</text>
        <FileUploader
          :max-count="9"
          :max-size="50"
          :tips="uploadTips"
          @success="onUploadSuccess"
          @error="onUploadError"
        />
        <button class="close-btn" @click="showUpload = false">关闭</button>
      </view>
    </up-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { getMyFiles, deleteFile } from '@/api/file';
import type { FileInfo } from '@/api/types';
import FileUploader from '@/components/FileUploader.vue';

const activeTab = ref('all');
const loading = ref(false);
const files = ref<FileInfo[]>([]);
const showUpload = ref(false);

const tabs = [
  { key: 'all', label: '全部', icon: '📁' },
  { key: 'image', label: '图片', icon: '🖼️' },
  { key: 'video', label: '视频', icon: '🎬' },
  { key: 'audio', label: '音频', icon: '🎵' },
  { key: 'document', label: '文档', icon: '📄' },
];

const uploadTips = '支持图片、视频、音频、文档，单文件最大 50MB';

onMounted(() => { loadFiles(); });
watch(activeTab, () => { loadFiles(); });

async function loadFiles() {
  loading.value = true;
  try {
    const params: any = { page: 1, pageSize: 50 };
    if (activeTab.value !== 'all') {
      params.fileType = activeTab.value;
    }
    const res = await getMyFiles(params);
    files.value = res.items || [];
  } catch {
    // Mock 数据
    files.value = [
      { id: 1, name: 'photo1.jpg', originalName: '风景照片.jpg', fileType: 'image/jpeg', fileSize: 2048000, url: '', uploadedAt: '2024-01-15T10:30:00Z', uploadedBy: 1 },
      { id: 2, name: 'video1.mp4', originalName: '教学视频.mp4', fileType: 'video/mp4', fileSize: 52428800, url: '', uploadedAt: '2024-01-14T14:20:00Z', uploadedBy: 1 },
      { id: 3, name: 'doc1.pdf', originalName: '考试大纲.pdf', fileType: 'application/pdf', fileSize: 1048576, url: '', uploadedAt: '2024-01-13T09:15:00Z', uploadedBy: 1 },
    ];
  } finally {
    loading.value = false;
  }
}

function getFileIcon(type: string): string {
  if (type?.startsWith('image')) return '🖼️';
  if (type?.startsWith('video')) return '🎬';
  if (type?.startsWith('audio')) return '🎵';
  if (type?.includes('pdf')) return '📕';
  if (type?.includes('word')) return '📘';
  if (type?.includes('excel')) return '📗';
  return '📄';
}

function formatSize(bytes: number): string {
  if (!bytes) return '0 B';
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  return `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
}

function previewFile(file: FileInfo) {
  uni.navigateTo({ url: `/pages/file/preview?id=${file.id}&name=${encodeURIComponent(file.originalName)}` });
}

function shareFile(file: FileInfo) {
  // TODO: 调用分享 API
  uni.showToast({ title: '分享功能开发中', icon: 'none' });
}

async function deleteFileItem(file: FileInfo) {
  uni.showModal({
    title: '确认删除',
    content: `确定要删除「${file.originalName}」吗？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await deleteFile(file.id);
          files.value = files.value.filter(f => f.id !== file.id);
          uni.showToast({ title: '已删除', icon: 'success' });
        } catch {
          uni.showToast({ title: '删除失败', icon: 'error' });
        }
      }
    },
  });
}

function onUploadSuccess(file: FileInfo) {
  files.value.unshift(file);
  uni.showToast({ title: '上传成功', icon: 'success' });
}

function onUploadError(error: Error) {
  uni.showToast({ title: error.message || '上传失败', icon: 'error' });
}
</script>

<style lang="scss" scoped>
.file-manage-page { min-height: 100vh; background: #f5f5f5; padding-bottom: 20px; }

.header { display: flex; justify-content: space-between; align-items: center; padding: 16px; background: #fff; }
.title { font-size: 18px; font-weight: bold; color: #333; }
.upload-btn { width: 32px; height: 32px; background: #1890ff; border-radius: 50%; display: flex; align-items: center; justify-content: center; }
.upload-icon { color: #fff; font-size: 20px; line-height: 1; }

.filter-bar { display: flex; background: #fff; margin-top: 12px; padding: 8px 0; overflow-x: auto; }
.filter-tab { flex: none; display: flex; flex-direction: column; align-items: center; padding: 8px 16px; margin: 0 4px; border-radius: 8px; min-width: 60px; }
.filter-tab.active { background: #e6f7ff; }
.tab-icon { font-size: 18px; margin-bottom: 2px; }
.tab-label { font-size: 11px; color: #666; }
.filter-tab.active .tab-label { color: #1890ff; }

.file-list { padding: 12px 16px; }
.loading, .empty { text-align: center; padding: 60px 0; color: #999; }
.empty-icon { font-size: 48px; display: block; margin-bottom: 12px; }
.empty-text { font-size: 14px; }

.file-item { display: flex; align-items: center; padding: 12px; background: #fff; border-radius: 8px; margin-bottom: 8px; }
.file-icon { font-size: 28px; margin-right: 12px; }
.file-info { flex: 1; overflow: hidden; }
.file-name { font-size: 14px; color: #333; display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.file-meta { display: flex; gap: 12px; margin-top: 4px; }
.file-size, .file-date { font-size: 12px; color: #999; }

.file-actions { display: flex; gap: 8px; }
.action-btn { font-size: 16px; padding: 4px; }
.action-btn.delete { color: #ff4d4f; }

.upload-panel { padding: 20px; }
.panel-title { font-size: 16px; font-weight: 500; margin-bottom: 16px; display: block; }
.close-btn { margin-top: 16px; height: 44px; line-height: 44px; border: none; border-radius: 22px; background: #f5f7fa; color: #666; font-size: 15px; }
</style>
