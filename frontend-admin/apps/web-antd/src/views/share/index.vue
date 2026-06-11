<script lang="ts" setup>
import { onMounted, ref, computed } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Descriptions,
  DescriptionsItem,
  Image,
  Spin,
  Table,
} from 'ant-design-vue';

import { getShareInfo, getShareFolderFiles } from '#/api/file';

defineOptions({ name: 'SharePage' });

const route = useRoute();
const loading = ref(true);
const shareInfo = ref<any>(null);
const folderFiles = ref<any[]>([]);
const error = ref('');

const shareCode = route.params.code as string;

// 是否是文件夹分享
// @ts-expect-error - 暂时未使用，保留以备将来使用
const isFolderShare = computed(() => shareInfo.value?.type === 'folder');

async function loadShareInfo() {
  try {
    loading.value = true;
    const result = await getShareInfo(shareCode);
    shareInfo.value = result;

    // 如果是文件夹分享，加载文件列表
    if (result.type === 'folder') {
      const files = await getShareFolderFiles(shareCode);
      folderFiles.value = files || [];
    }
  } catch (err: any) {
    error.value = err.message || '分享不存在或已过期';
  } finally {
    loading.value = false;
  }
}

function viewFile(file: any) {
  // 文件夹分享的文件访问 URL
  const url = `/api/v1/share/${shareCode}/file/${file.fileId}`;
  window.open(url, '_blank', 'noopener,noreferrer');
}

function downloadFile(file: any) {
  const url = `/api/v1/share/${shareCode}/file/${file.fileId}`;
  const link = document.createElement('a');
  link.href = url;
  link.download = file.fileName;
  link.rel = 'noopener noreferrer';
  link.click();
}

// 文件分享的下载按钮
function downloadSharedFile() {
  const url = `/api/v1/share/${shareCode}/file`;
  const link = document.createElement('a');
  link.href = url;
  link.download = shareInfo.value?.fileName || 'download';
  link.rel = 'noopener noreferrer';
  link.click();
}

function formatDate(date: string) {
  if (!date) return '永久有效';
  return new Date(date).toLocaleString();
}

function formatFileSize(size: number) {
  if (!size) return '-';
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`;
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`;
}

// 文件列表表格列
const columns = [
  { title: '文件名', dataIndex: 'fileName', key: 'fileName', ellipsis: true },
  { title: '大小', dataIndex: 'fileSize', key: 'fileSize', width: 100 },
  { title: '类型', dataIndex: 'contentType', key: 'contentType', width: 120 },
  { title: '操作', key: 'action', width: 120 },
];

onMounted(() => {
  loadShareInfo();
});
</script>

<template>
  <Page title="文件分享">
    <div class="max-w-4xl mx-auto p-4">
      <Spin :spinning="loading">
        <!-- 文件分享 -->
        <Card
          v-if="shareInfo && shareInfo.type === 'file'"
          :title="shareInfo.fileName"
        >
          <!-- 图片预览 -->
          <div v-if="shareInfo.contentType?.startsWith('image/')">
            <Image
              :src="`/api/v1/share/${shareCode}/file`"
              class="max-w-full"
              style="max-height: 400px"
            />
          </div>

          <!-- PDF 预览 -->
          <div
            v-else-if="
              shareInfo.contentType?.includes('pdf') ||
              shareInfo.fileName?.toLowerCase().endsWith('.pdf')
            "
          >
            <iframe
              :src="`/api/v1/share/${shareCode}/file`"
              sandbox="allow-scripts allow-same-origin"
              referrerpolicy="no-referrer"
              style="width: 100%; height: 400px"
              frameborder="0"
            />
          </div>

          <!-- 视频预览 -->
          <div v-else-if="shareInfo.contentType?.startsWith('video/')">
            <video
              :src="`/api/v1/share/${shareCode}/file`"
              controls
              autoplay
              preload="auto"
              playsinline
              style="
                display: block;
                max-width: 100%;
                max-height: 400px;
                background: #000;
              "
            />
          </div>

          <!-- 音频预览 -->
          <div v-else-if="shareInfo.contentType?.startsWith('audio/')">
            <div class="py-6 text-center">
              <div class="mb-4 text-6xl text-blue-500">🎵</div>
              <p class="mb-4 text-lg text-foreground">
                {{ shareInfo.fileName }}
              </p>
              <audio
                :src="`/api/v1/share/${shareCode}/file`"
                controls
                autoplay
                preload="auto"
                style="width: 100%; max-width: 500px; margin: 0 auto"
              />
            </div>
          </div>

          <!-- 其他文件类型 -->
          <div v-else class="text-center py-4 text-gray-500">
            <span class="i-ant-design:file-outlined text-4xl mb-2" />
            <p>该文件类型不支持在线预览</p>
          </div>

          <!-- 文件信息 -->
          <Descriptions :column="2" class="mt-4">
            <DescriptionsItem label="文件名">{{
              shareInfo.fileName
            }}</DescriptionsItem>
            <DescriptionsItem label="文件大小">{{
              formatFileSize(shareInfo.fileSize)
            }}</DescriptionsItem>
            <DescriptionsItem label="分享者">
              <img
                v-if="shareInfo.sharerAvatar"
                :src="shareInfo.sharerAvatar"
                class="w-6 h-6 rounded-full inline-block mr-1"
              />
              {{ shareInfo.sharerName }}
            </DescriptionsItem>
            <DescriptionsItem label="过期时间">{{
              formatDate(shareInfo.expireAt)
            }}</DescriptionsItem>
          </Descriptions>

          <!-- 下载按钮 -->
          <div class="mt-4 text-center">
            <Button type="primary" size="large" @click="downloadSharedFile">
              下载文件
            </Button>
          </div>
        </Card>

        <!-- 文件夹分享 -->
        <Card
          v-if="shareInfo && shareInfo.type === 'folder'"
          :title="shareInfo.folderName"
        >
          <!-- 文件夹信息 -->
          <Descriptions :column="2" class="mb-4">
            <DescriptionsItem label="文件夹">{{
              shareInfo.folderName
            }}</DescriptionsItem>
            <DescriptionsItem label="文件数"
              >{{ folderFiles.length }} 个</DescriptionsItem
            >
            <DescriptionsItem label="分享者">
              <img
                v-if="shareInfo.sharerAvatar"
                :src="shareInfo.sharerAvatar"
                class="w-6 h-6 rounded-full inline-block mr-1"
              />
              {{ shareInfo.sharerName }}
            </DescriptionsItem>
            <DescriptionsItem label="过期时间">{{
              formatDate(shareInfo.expireAt)
            }}</DescriptionsItem>
          </Descriptions>

          <!-- 文件列表 -->
          <Table
            :columns="columns"
            :data-source="folderFiles"
            :pagination="false"
            row-key="fileId"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'fileSize'">
                {{ formatFileSize(record.fileSize) }}
              </template>
              <template v-if="column.key === 'contentType'">
                <span class="text-sm text-gray-500">{{
                  record.contentType || '未知'
                }}</span>
              </template>
              <template v-if="column.key === 'action'">
                <Button type="link" size="small" @click="viewFile(record)"
                  >预览</Button
                >
                <Button type="link" size="small" @click="downloadFile(record)"
                  >下载</Button
                >
              </template>
            </template>
          </Table>

          <div
            v-if="folderFiles.length === 0"
            class="text-center py-4 text-gray-500"
          >
            文件夹内暂无文件
          </div>
        </Card>

        <!-- 错误提示 -->
        <Card v-else-if="error">
          <div class="text-center py-8">
            <span
              class="i-ant-design:warning-outlined text-4xl text-red-500 mb-4"
            />
            <p class="text-lg">{{ error }}</p>
          </div>
        </Card>
      </Spin>
    </div>
  </Page>
</template>
