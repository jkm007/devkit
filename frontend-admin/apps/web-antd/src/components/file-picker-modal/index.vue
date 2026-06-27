<script lang="ts" setup>
import type { FileApi } from '#/api/file';

import { computed, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Button, Empty, Input, Radio, Spin, Tabs } from 'ant-design-vue';

import { listFiles } from '#/api/file';
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
      loadFiles();
    }
  },
});

const activeTab = ref('image');
const keyword = ref('');
const loading = ref(false);
const files = ref<FileApi.FileEntry[]>([]);
const selectedFile = ref<FileApi.FileEntry | null>(null);

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

async function loadFiles() {
  loading.value = true;
  try {
    const params: FileApi.ListFilesParams = {
      page: 1,
      pageSize: 50,
      keyword: keyword.value,
      scope: 'own',
    };
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
    return appendToken(file.previewUrl);
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

      <Spin :spinning="loading">
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
