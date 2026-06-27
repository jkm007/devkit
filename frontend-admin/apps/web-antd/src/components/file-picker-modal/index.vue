<script lang="ts" setup>
import type { FileApi } from '#/api/file';

import { computed, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Button, Empty, Input, Radio, Spin, Tabs } from 'ant-design-vue';

import { listFiles, getFolderTree, createFolder } from '#/api/file';
import { appendToken, normalizeFileUrl } from '#/utils/media-url';

interface FilePickerResult {
  fileId: number;
  url: string;
  name: string;
  contentType: string;
}

const emit = defineEmits<{
  select: [result: FilePickerResult];
}>();

const [Modal, modalApi] = useVbenModal({
  onOpenChange(isOpen) {
    if (isOpen) {
      activeTab.value = 'image';
      keyword.value = '';
      selectedFile.value = null;
      currentFolderId.value = null;
      loadFolders();
      loadFiles();
    }
  },
});

const activeTab = ref('image');
const keyword = ref('');
const loading = ref(false);
const files = ref<FileApi.FileEntry[]>([]);
const selectedFile = ref<FileApi.FileEntry | null>(null);
const folders = ref<FileApi.Folder[]>([]);
const currentFolderId = ref<number | null>(null);
const folderLoading = ref(false);

const tabs = [
  { key: 'image', label: '图片' },
  { key: 'video', label: '视频' },
  { key: 'all', label: '全部' },
];

const contentTypeFilter = computed(() => {
  if (activeTab.value === 'image') return 'image/';
  if (activeTab.value === 'video') return 'video/';
  return '';
});

// 面包屑路径
const breadcrumbPath = computed(() => {
  const path: { id: number | null; name: string }[] = [{ id: null, name: '根目录' }];
  function findInTree(list: FileApi.Folder[], targetId: number): FileApi.Folder[] {
    for (const f of list) {
      if (f.id === targetId) return [f];
      if (f.children) {
        const found = findInTree(f.children, targetId);
        if (found.length > 0) return [f, ...found];
      }
    }
    return [];
  }
  if (currentFolderId.value) {
    const chain = findInTree(folders.value, currentFolderId.value);
    path.push(...chain.map(f => ({ id: f.id, name: f.name })));
  }
  return path;
});

async function loadFolders() {
  try {
    const res = await getFolderTree('own');
    folders.value = Array.isArray(res) ? res : [];
  } catch {
    folders.value = [];
  }
}

async function loadFiles() {
  loading.value = true;
  try {
    const params: FileApi.ListFilesParams = {
      page: 1,
      pageSize: 50,
      keyword: keyword.value,
      scope: 'own',
    };
    if (currentFolderId.value) {
      params.folderId = currentFolderId.value;
    }
    if (contentTypeFilter.value) {
      params.contentType = contentTypeFilter.value;
    }
    const res = await listFiles(params);
    files.value = res?.items || [];
  } catch (e: any) {
    files.value = [];
  } finally {
    loading.value = false;
  }
}

function enterFolder(folder: FileApi.Folder) {
  currentFolderId.value = folder.id;
  selectedFile.value = null;
  loadFiles();
}

function goBackToRoot() {
  currentFolderId.value = null;
  selectedFile.value = null;
  loadFiles();
}

watch(activeTab, () => {
  loadFiles();
});

function onSearch() {
  loadFiles();
}

function isImage(contentType: string) {
  return contentType?.startsWith('image/');
}

function isVideo(contentType: string) {
  return contentType?.startsWith('video/');
}

function getThumbnail(file: FileApi.FileEntry) {
  if (file.previewUrl) {
    return appendToken(normalizeFileUrl(file.previewUrl));
  }
  if (isImage(file.contentType)) {
    return appendToken(normalizeFileUrl(`/files/${file.id}/view`));
  }
  return '';
}

function getInsertUrl(file: FileApi.FileEntry) {
  return appendToken(normalizeFileUrl(`/files/${file.id}/view`));
}

function handleSelect(file: FileApi.FileEntry) {
  selectedFile.value = file;
}

function handleConfirm() {
  if (!selectedFile.value) return;
  emit('select', {
    fileId: selectedFile.value.id,
    url: getInsertUrl(selectedFile.value),
    name: selectedFile.value.name,
    contentType: selectedFile.value.contentType,
  });
  modalApi.close();
}

function handleCancel() {
  modalApi.close();
}

function handleImageError(e: Event) {
  const img = e.target as HTMLImageElement;
  if (img && img.parentElement) {
    const placeholder = document.createElement('div');
    placeholder.className = 'placeholder';
    placeholder.textContent = getFileIcon(img.alt || 'image');
    img.style.display = 'none';
    img.parentElement.appendChild(placeholder);
  }
}

function getFileIcon(contentType: string) {
  if (isImage(contentType)) return '🖼️';
  if (isVideo(contentType)) return '🎬';
  return '📄';
}

function formatSize(size: number) {
  if (size < 1024) return `${size}B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)}KB`;
  return `${(size / (1024 * 1024)).toFixed(1)}MB`;
}

function open() {
  modalApi.open();
}

defineExpose({ open });
</script>

<template>
  <Modal title="从素材库选择" class="w-[800px]" :footer="false">
    <div class="file-picker">
      <!-- 文件夹面包屑 + 操作 -->
      <div class="mb-3 flex items-center gap-2">
        <Button v-if="currentFolderId" size="small" @click="goBackToRoot">📁 根目录</Button>
        <span v-if="breadcrumbPath.length > 1" class="text-xs text-gray-400">
          / {{ breadcrumbPath.map(p => p.name).slice(1).join(' / ') }}
        </span>
        <span v-if="files.length > 0" class="text-xs text-gray-400 ml-auto">
          {{ files.length }} 个文件
        </span>
      </div>

      <div class="mb-4 flex gap-3">
        <Tabs v-model:activeKey="activeTab" class="flex-1">
          <Tabs.TabPane
            v-for="tab in tabs"
            :key="tab.key"
            :tab="tab.label"
          />
        </Tabs>
        <Input
          v-model:value="keyword"
          placeholder="搜索文件名"
          style="width: 200px"
          @pressEnter="onSearch"
        />
        <Button type="primary" @click="onSearch">搜索</Button>
      </div>

      <Spin :spinning="loading || folderLoading">
        <div v-if="files.length === 0" class="py-10">
          <Empty description="暂无文件" />
        </div>
        <div v-else class="file-grid">
          <div
            v-for="file in files"
            :key="file.id"
            class="file-card"
            :class="{ selected: selectedFile?.id === file.id }"
            @click="handleSelect(file)"
          >
            <div class="thumb">
              <img
                v-if="getThumbnail(file)"
                :src="getThumbnail(file)"
                :alt="file.name"
                @error="handleImageError"
              />
              <div v-else class="placeholder">{{ getFileIcon(file.contentType) }}</div>
              <div v-if="isVideo(file.contentType)" class="video-badge">▶</div>
            </div>
            <div class="info">
              <div class="name" :title="file.name">{{ file.name }}</div>
              <div class="size">{{ formatSize(file.size) }}</div>
            </div>
            <Radio
              class="absolute right-2 top-2"
              :checked="selectedFile?.id === file.id"
            />
          </div>
        </div>
      </Spin>

      <div class="mt-4 flex justify-end gap-3">
        <Button @click="handleCancel">取消</Button>
        <Button type="primary" :disabled="!selectedFile" @click="handleConfirm">
          插入
        </Button>
      </div>
    </div>
  </Modal>
</template>

<style lang="scss" scoped>
.file-picker {
  min-height: 300px;
}

.file-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
  padding: 4px;
}

.file-card {
  position: relative;
  border: 2px solid #f0f0f0;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s;

  &:hover,
  &.selected {
    border-color: #1890ff;
  }

  .thumb {
    aspect-ratio: 1;
    background: #f5f5f5;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    .placeholder {
      font-size: 32px;
    }

    .video-badge {
      position: absolute;
      width: 32px;
      height: 32px;
      background: rgba(0, 0, 0, 0.5);
      border-radius: 50%;
      color: #fff;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 14px;
    }
  }

  .info {
    padding: 8px;

    .name {
      font-size: 12px;
      color: #333;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .size {
      font-size: 11px;
      color: #999;
      margin-top: 2px;
    }
  }
}
</style>
