<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';
import { useAccessStore } from '@vben/stores';

import {
  Button,
  Card,
  Descriptions,
  DescriptionsItem,
  Empty,
  Form,
  FormItem,
  Image,
  Input,
  InputNumber,
  message,
  Modal,
  Progress,
  Space,
  Spin,
  Table,
  Tag,
  Tree,
  TreeSelect,
  Upload,
} from 'ant-design-vue';
import type { TreeProps } from 'ant-design-vue';

import {
  createFolder,
  createFileShare,
  createFolderShare,
  deleteFile,
  batchDeleteFiles,
  batchMoveFiles,
  deleteFolder,
  downloadFile,
  getFolderTree,
  listFiles,
  moveFile,
  renameFolder,
  simpleUpload,
} from '#/api/file';
import type { FileApi, UploadProgressCallback } from '#/api/file';
import { $t } from '#/locales';

defineOptions({ name: 'FileList' });

const accessStore = useAccessStore();

// ==================== 状态 ====================

const loading = ref(false);
const folderTree = ref<FileApi.Folder[]>([]);
const currentFolderId = ref<number | null>(null);
const fileList = ref<FileApi.FileEntry[]>([]);
const totalFiles = ref(0);
const pagination = ref({ current: 1, pageSize: 20 });
const selectedRowKeys = ref<number[]>([]);

// 新建文件夹
const newFolderModalVisible = ref(false);
const newFolderName = ref('');
const newFolderParentId = ref<number | null>(null);

// 重命名文件夹
const renameFolderModalVisible = ref(false);
const renameFolderId = ref<number | null>(null);
const renameFolderName = ref('');

// 删除文件夹
const deleteFolderModalVisible = ref(false);
const deleteFolderId = ref<number | null>(null);
const deleteFolderName = ref('');

// 移动文件
const moveFileModalVisible = ref(false);
const moveFileId = ref<number | null>(null);
const moveTargetFolderId = ref<number | null>(null);

// 分享
const shareModalVisible = ref(false);
const shareFileId = ref<number | null>(null);
const shareResult = ref<{ shareCode: string; shareUrl: string } | null>(null);
const shareExpireHours = ref(0);

// 预览
const previewVisible = ref(false);
const previewUrl = ref('');
const previewName = ref('');
const previewType = ref(''); // 'image' | 'video' | 'pdf'
const previewToken = ref('');

// 批量操作
const batchMoveModalVisible = ref(false);
const batchTargetFolderId = ref<number | null>(null);

// 上传进度
const uploading = ref(false);
const uploadProgress = ref(0);
const uploadFileName = ref('');

// ==================== 文件类型图标 ====================

function getFileIcon(type: string) {
  if (type?.startsWith('image/')) return 'i-ant-design:file-image-outlined';
  if (type?.startsWith('video/')) return 'i-ant-design:file-video-outlined';
  if (type?.startsWith('audio/')) return 'i-ant-design:sound-outlined';
  if (type?.includes('pdf')) return 'i-ant-design:file-pdf-outlined';
  if (type?.includes('word') || type?.includes('document')) return 'i-ant-design:file-word-outlined';
  if (type?.includes('excel') || type?.includes('spreadsheet')) return 'i-ant-design:file-excel-outlined';
  if (type?.includes('zip') || type?.includes('rar')) return 'i-ant-design:file-zip-outlined';
  return 'i-ant-design:file-outlined';
}

function formatFileSize(size: number | undefined | null) {
  if (size === undefined || size === null || isNaN(size)) return '-';
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`;
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`;
}

// ==================== 加载 ====================

async function loadFolderTree() {
  try {
    const result = await getFolderTree();
    folderTree.value = result || [];
  } catch {
    message.error('加载文件夹树失败');
  }
}

async function loadFileList() {
  loading.value = true;
  try {
    const result = await listFiles({
      folderId: currentFolderId.value ?? undefined,
      page: pagination.value.current,
      pageSize: pagination.value.pageSize,
    });
    const items = result?.items || [];
    items.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
    fileList.value = items;
    totalFiles.value = result?.total || 0;
    selectedRowKeys.value = [];
  } catch {
    message.error('加载文件列表失败');
  } finally {
    loading.value = false;
  }
}

function handleFolderSelect(keys: (string | number)[]) {
  currentFolderId.value = keys[0] as number | null;
  pagination.value.current = 1;
  loadFileList();
}

// ==================== 文件夹操作 ====================

// 文件夹操作菜单
const folderMenuVisible = ref(false);
const folderMenuId = ref<number | null>(null);
const folderMenuName = ref('');

function showFolderMenu(node: any) {
  folderMenuId.value = node.key as number;
  folderMenuName.value = node.title as string;
  folderMenuVisible.value = true;
}

function folderMenuAction(action: string) {
  folderMenuVisible.value = false;
  if (action === 'new') {
    openNewFolderModal(folderMenuId.value!);
  } else if (action === 'rename') {
    openRenameFolderModal(folderMenuId.value!, folderMenuName.value);
  } else if (action === 'delete') {
    openDeleteFolderModal(folderMenuId.value!, folderMenuName.value);
  } else if (action === 'share') {
    openFolderShareModal(folderMenuId.value!, folderMenuName.value);
  }
}

// 文件夹分享
const folderShareModalVisible = ref(false);
const folderShareId = ref<number | null>(null);
const folderShareName = ref('');
const folderShareResult = ref<{ shareCode: string; shareUrl: string } | null>(null);
const folderShareExpireHours = ref(0);

function openFolderShareModal(id: number, name: string) {
  folderShareId.value = id;
  folderShareName.value = name;
  folderShareExpireHours.value = 0;
  folderShareResult.value = null;
  folderShareModalVisible.value = true;
}

async function confirmFolderShare() {
  try {
    const result = await createFolderShare(folderShareId.value!, {
      expireHours: folderShareExpireHours.value || undefined,
    });
    folderShareResult.value = result;
    message.success('分享链接已生成');
  } catch {
    message.error('分享失败');
  }
}

function copyFolderShareUrl() {
  const url = `${window.location.origin}/share/${folderShareResult.value?.shareCode}`;
  navigator.clipboard.writeText(url).then(() => {
    message.success('链接已复制到剪贴板');
  }).catch(() => {
    const input = document.createElement('input');
    input.value = url;
    document.body.appendChild(input);
    input.select();
    document.execCommand('copy');
    document.body.removeChild(input);
    message.success('链接已复制');
  });
}

function openNewFolderModal(parentId?: number) {
  newFolderParentId.value = parentId ?? currentFolderId.value;
  newFolderName.value = '';
  newFolderModalVisible.value = true;
}

async function handleCreateFolder() {
  if (!newFolderName.value.trim()) {
    message.warning('请输入文件夹名称');
    return;
  }
  try {
    await createFolder({
      name: newFolderName.value.trim(),
      parentId: newFolderParentId.value || undefined,
    });
    message.success('创建成功');
    newFolderModalVisible.value = false;
    loadFolderTree();
  } catch {
    message.error('创建失败');
  }
}

function openRenameFolderModal(id: number, name: string) {
  renameFolderId.value = id;
  renameFolderName.value = name;
  renameFolderModalVisible.value = true;
}

async function handleRenameFolder() {
  if (!renameFolderName.value.trim()) return;
  try {
    await renameFolder(renameFolderId.value!, { name: renameFolderName.value.trim() });
    message.success('重命名成功');
    renameFolderModalVisible.value = false;
    loadFolderTree();
  } catch {
    message.error('重命名失败');
  }
}

function openDeleteFolderModal(id: number, name: string) {
  deleteFolderId.value = id;
  deleteFolderName.value = name;
  deleteFolderModalVisible.value = true;
}

async function handleDeleteFolder() {
  try {
    await deleteFolder(deleteFolderId.value!);
    message.success('删除成功');
    deleteFolderModalVisible.value = false;
    if (currentFolderId.value === deleteFolderId.value) {
      currentFolderId.value = null;
    }
    loadFolderTree();
    loadFileList();
  } catch {
    message.error('删除失败');
  }
}

// ==================== 文件操作 ====================

function openMoveFileModal(id: number) {
  moveFileId.value = id;
  moveTargetFolderId.value = currentFolderId.value;
  moveFileModalVisible.value = true;
}

async function handleMoveFile() {
  try {
    await moveFile({
      fileId: moveFileId.value!,
      targetFolderId: moveTargetFolderId.value || undefined,
    });
    message.success('移动成功');
    moveFileModalVisible.value = false;
    loadFileList();
  } catch {
    message.error('移动失败');
  }
}

async function handleDeleteFile(id: number, name: string) {
  try {
    await deleteFile(id);
    message.success(`已删除 ${name}`);
    loadFileList();
  } catch {
    message.error('删除失败');
  }
}

// 批量删除
async function handleBatchDelete() {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择文件');
    return;
  }

  try {
    const result = await batchDeleteFiles(selectedRowKeys.value);
    if (result.errors?.length > 0) {
      message.warning(`已删除 ${result.deleted} 个文件，${result.errors.length} 个失败`);
    } else {
      message.success(`已删除 ${result.deleted} 个文件`);
    }
    selectedRowKeys.value = [];
    loadFileList();
  } catch {
    message.error('批量删除失败');
  }
}

// 打开批量移动弹窗
function openBatchMoveModal() {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择文件');
    return;
  }
  batchTargetFolderId.value = currentFolderId.value;
  batchMoveModalVisible.value = true;
}

// 执行批量移动
async function handleBatchMove() {
  try {
    const result = await batchMoveFiles(selectedRowKeys.value, batchTargetFolderId.value || undefined);
    if (result.errors?.length > 0) {
      message.warning(`已移动 ${result.moved} 个文件，${result.errors.length} 个失败`);
    } else {
      message.success(`已移动 ${result.moved} 个文件`);
    }
    batchMoveModalVisible.value = false;
    selectedRowKeys.value = [];
    loadFileList();
  } catch {
    message.error('批量移动失败');
  }
}

async function handleDownload(file: FileApi.FileEntry) {
  try {
    // 直接使用 fetch 下载文件
    const response = await fetch(`/api/files/${file.id}/download`, {
      headers: {
        'Authorization': `Bearer ${useAccessStore().accessToken}`,
      },
    });

    if (!response.ok) {
      throw new Error('下载失败');
    }

    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = file.name;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  } catch {
    message.error('下载失败');
  }
}

async function handlePreview(file: FileApi.FileEntry) {
  previewName.value = file.name;
  previewUrl.value = '';
  previewType.value = '';

  // 支持预览的类型
  const isImage = file.contentType?.startsWith('image/');
  const isVideo = file.contentType?.startsWith('video/');
  const isPdf = file.contentType?.includes('pdf');

  if (!isImage && !isVideo && !isPdf) {
    // 其他类型 - 直接下载
    handleDownload(file);
    return;
  }

  const token = accessStore.accessToken;
  const viewUrl = `/api/files/${file.id}/view`;

  // 先显示 Modal 和加载状态
  previewVisible.value = true;

  // 视频使用流式播放（Range 请求）
  if (isVideo) {
    previewType.value = 'video';
    // 视频需要通过 fetch + blob URL（因为需要认证头）
    // 使用 ReadableStream 实现渐进式加载
    try {
      const response = await fetch(viewUrl, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      if (response.ok) {
        const blob = await response.blob();
        previewUrl.value = URL.createObjectURL(blob);
      } else {
        message.error('获取预览失败');
        previewVisible.value = false;
      }
    } catch {
      message.error('获取预览失败');
      previewVisible.value = false;
    }
    return;
  }

  // 图片和 PDF 需要先下载
  try {
    const response = await fetch(viewUrl, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (response.ok) {
      const blob = await response.blob();
      previewUrl.value = URL.createObjectURL(blob);
      previewType.value = isImage ? 'image' : 'pdf';
    } else {
      message.error('获取预览失败');
      previewVisible.value = false;
    }
  } catch {
    message.error('获取预览失败');
    previewVisible.value = false;
  }
}

async function handleShare(file: FileApi.FileEntry) {
  shareFileId.value = file.id;
  shareExpireHours.value = 0;
  shareResult.value = null;
  shareModalVisible.value = true;
}

async function confirmShare() {
  try {
    const result = await createFileShare(shareFileId.value!, {
      expireHours: shareExpireHours.value || undefined,
    });
    shareResult.value = result;
    message.success('分享链接已生成');
  } catch {
    message.error('分享失败');
  }
}

function copyShareUrl() {
  const url = `${window.location.origin}/share/${shareResult.value?.shareCode}`;
  navigator.clipboard.writeText(url).then(() => {
    message.success('链接已复制到剪贴板');
  }).catch(() => {
    // 备用方案
    const input = document.createElement('input');
    input.value = url;
    document.body.appendChild(input);
    input.select();
    document.execCommand('copy');
    document.body.removeChild(input);
    message.success('链接已复制');
  });
}

// ==================== 上传 ====================

const uploadSaving = ref(false);

async function handleUpload(file: File) {
  uploading.value = true;
  uploadProgress.value = 0;
  uploadFileName.value = file.name;
  uploadSaving.value = false;

  const onProgress: UploadProgressCallback = (event) => {
    uploadProgress.value = event.percent;
    // 当进度达到100%时，显示"正在保存到服务器..."
    if (event.percent >= 100) {
      uploadSaving.value = true;
    }
  };

  try {
    await simpleUpload(file, onProgress);
    message.success(`${file.name} 上传成功`);
    loadFileList();
  } catch (err) {
    message.error(`上传失败: ${err}`);
  } finally {
    uploading.value = false;
    uploadProgress.value = 0;
    uploadFileName.value = '';
    uploadSaving.value = false;
  }
  return false;
}

// ==================== 初始化 ====================

onMounted(() => {
  loadFolderTree();
  loadFileList();
});

// ==================== 表格列 ====================

// 存储类型标签映射
const storageTypeLabels: Record<string, { label: string; color: string }> = {
  local: { label: '本地', color: 'default' },
  minio: { label: 'MinIO', color: 'blue' },
  oss: { label: 'OSS', color: 'orange' },
  cos: { label: 'COS', color: 'green' },
};

const columns = [
  { title: '文件名', dataIndex: 'name', key: 'name', width: 200, ellipsis: true },
  { title: '大小', dataIndex: 'size', key: 'size', width: 80, customRender: ({ text }) => formatFileSize(text) },
  { title: '存储', key: 'storage', width: 80 },
  { title: '上传者', key: 'uploader', width: 80 },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 120 },
  { title: '操作', key: 'operation', width: 180, fixed: 'right' },
];

const treeData = computed<TreeProps['treeData']>(() => {
  const convert = (folders: FileApi.Folder[]): TreeProps['treeData'] =>
    folders.map((f) => ({
      key: f.id,
      title: f.name,
      type: f.type,
      children: f.children ? convert(f.children) : undefined,
    }));
  return convert(folderTree.value);
});

const folderSelectData = computed(() => {
  const convert = (folders: FileApi.Folder[]): any[] =>
    folders.map((f) => ({
      value: f.id,
      label: f.name,
      children: f.children ? convert(f.children) : undefined,
    }));
  return [{ value: null, label: '根目录' }, ...convert(folderTree.value)];
});
</script>

<template>
  <Page title="">
    <div class="flex gap-4">
      <!-- 左侧文件夹树 -->
      <div class="w-56 shrink-0 border rounded-lg p-3">
        <div class="flex items-center justify-between mb-3">
          <span class="font-medium">文件夹</span>
          <Button type="link" size="small" @click="openNewFolderModal()">+新建</Button>
        </div>

        <Tree
          :tree-data="treeData"
          :selected-keys="currentFolderId ? [currentFolderId] : []"
          default-expand-all
          @select="handleFolderSelect"
        >
          <template #title="node">
            <div class="flex items-center gap-1 py-1 group">
              <span :class="node.type === 'avatar' ? 'i-ant-design:user-outlined' : 'i-ant-design:folder-outlined'" />
              <span class="flex-1 truncate">{{ node.title as string }}</span>
              <button
                type="button"
                class="opacity-0 group-hover:opacity-100 ml-1 px-1 py-0.5 text-xs rounded hover:bg-gray-200"
                @click.stop
                @click="showFolderMenu(node)"
              >
                ⋯
              </button>
            </div>
          </template>
        </Tree>
      </div>

      <!-- 右侧文件列表 -->
      <div class="flex-1 border rounded-lg p-3">
        <!-- 工具栏 -->
        <div class="flex items-center justify-between mb-3">
          <Space>
            <Upload :show-upload-list="false" :before-upload="handleUpload" :multiple="true" :disabled="uploading">
              <Button type="primary" :loading="uploading">上传文件</Button>
            </Upload>
            <Button v-if="selectedRowKeys.length > 0" danger @click="handleBatchDelete">
              批量删除 ({{ selectedRowKeys.length }})
            </Button>
            <Button v-if="selectedRowKeys.length > 0" @click="openBatchMoveModal">
              批量移动 ({{ selectedRowKeys.length }})
            </Button>
          </Space>
          <span class="text-sm text-gray-500">共 {{ totalFiles }} 个文件</span>
        </div>

        <!-- 上传进度条 -->
        <div v-if="uploading" class="mb-4 p-3 bg-blue-50 rounded-lg border border-blue-200">
          <div class="flex items-center justify-between mb-2">
            <span class="text-sm text-blue-700">
              <span class="i-ant-design:loading-outlined animate-spin mr-1" />
              {{ uploadSaving ? '正在保存到服务器...' : `正在上传: ${uploadFileName}` }}
            </span>
            <span class="text-sm font-medium text-blue-700">{{ uploadSaving ? '处理中...' : `${uploadProgress}%` }}</span>
          </div>
          <Progress :percent="uploadSaving ? 100 : uploadProgress" :show-info="false" status="active" :stroke-color="{ from: '#108ee9', to: '#87d068' }" />
        </div>

        <!-- 文件表格 -->
        <Table
          :columns="columns"
          :data-source="fileList"
          :loading="loading"
          :pagination="{ current: pagination.current, pageSize: pagination.pageSize, total: totalFiles, showSizeChanger: true }"
          :scroll="{ x: 600 }"
          :row-selection="{ selectedRowKeys, onChange: (keys) => selectedRowKeys = keys }"
          row-key="id"
          @change="(pag) => { pagination = pag; loadFileList(); }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'name'">
              <span :class="getFileIcon(record.contentType)" class="mr-1" />
              {{ record.name }}
            </template>
            <template v-if="column.key === 'storage'">
              <Tag :color="storageTypeLabels[record.storageType]?.color || 'default'">
                {{ storageTypeLabels[record.storageType]?.label || record.storageType || '本地' }}
              </Tag>
            </template>
            <template v-if="column.key === 'uploader'">
              <span class="text-sm">{{ record.uploaderName || '-' }}</span>
            </template>
            <template v-if="column.key === 'operation'">
              <Space size="small">
                <Button type="link" size="small" @click="handlePreview(record)">预览</Button>
                <Button type="link" size="small" @click="handleDownload(record)">下载</Button>
                <Button type="link" size="small" @click="handleShare(record)">分享</Button>
                <Button type="link" size="small" danger @click="handleDeleteFile(record.id, record.name)">删除</Button>
              </Space>
            </template>
          </template>
        </Table>

        <Empty v-if="!loading && fileList.length === 0" description="暂无文件" />
      </div>
    </div>

    <!-- 新建文件夹 -->
    <Modal v-model:open="newFolderModalVisible" title="新建文件夹" @ok="handleCreateFolder">
      <Form layout="vertical">
        <FormItem label="名称">
          <Input v-model:value="newFolderName" placeholder="输入文件夹名称" />
        </FormItem>
      </Form>
    </Modal>

    <!-- 重命名文件夹 -->
    <Modal v-model:open="renameFolderModalVisible" title="重命名文件夹" @ok="handleRenameFolder">
      <Form layout="vertical">
        <FormItem label="新名称">
          <Input v-model:value="renameFolderName" placeholder="输入新名称" />
        </FormItem>
      </Form>
    </Modal>

    <!-- 删除文件夹 -->
    <Modal v-model:open="deleteFolderModalVisible" title="删除文件夹" @ok="handleDeleteFolder">
      <p>确定删除文件夹 "{{ deleteFolderName }}" 吗？</p>
      <p class="text-red-500">文件夹内的所有文件也将被删除！</p>
    </Modal>

    <!-- 移动文件 -->
    <Modal v-model:open="moveFileModalVisible" title="移动文件" @ok="handleMoveFile">
      <Form layout="vertical">
        <FormItem label="目标文件夹">
          <TreeSelect v-model:value="moveTargetFolderId" :tree-data="folderSelectData" placeholder="选择文件夹" allow-clear />
        </FormItem>
      </Form>
    </Modal>

    <!-- 批量移动文件 -->
    <Modal v-model:open="batchMoveModalVisible" title="批量移动" @ok="handleBatchMove">
      <p class="mb-2">将移动 {{ selectedRowKeys.length }} 个文件</p>
      <Form layout="vertical">
        <FormItem label="目标文件夹">
          <TreeSelect v-model:value="batchTargetFolderId" :tree-data="folderSelectData" placeholder="选择文件夹" allow-clear />
        </FormItem>
      </Form>
    </Modal>

    <!-- 分享 -->
    <Modal v-model:open="shareModalVisible" title="创建分享链接" @ok="confirmShare">
      <Form layout="vertical">
        <FormItem label="过期时间">
          <Space>
            <InputNumber v-model:value="shareExpireHours" :min="0" style="width: 100px" />
            <span>小时（0表示永久有效）</span>
          </Space>
        </FormItem>
      </Form>
      <div v-if="shareResult" class="mt-4 p-3 bg-gray-50 rounded">
        <p class="mb-2 font-medium">分享链接：</p>
        <Input.Group compact>
          <Input :value="shareResult ? `${window.location.origin}${shareResult.shareUrl}` : ''" style="width: 280px" readonly />
          <Button type="primary" @click="copyShareUrl">复制</Button>
        </Input.Group>
      </div>
    </Modal>

    <!-- 预览 -->
    <Modal v-model:open="previewVisible" :title="previewName" :footer="null" :width="previewType === 'video' ? 960 : 800" :bodyStyle="{ padding: previewType === 'video' ? '0' : '24px', textAlign: previewType === 'video' ? 'center' : 'left' }">
      <!-- 图片预览 -->
      <div v-if="previewType === 'image' && previewUrl" class="text-center">
        <Image :src="previewUrl" class="max-w-full" style="max-height: 600px" />
      </div>
      <!-- PDF 预览 -->
      <iframe v-else-if="previewType === 'pdf' && previewUrl" :src="previewUrl" sandbox="allow-scripts allow-same-origin" referrerpolicy="no-referrer" style="width: 100%; height: 600px; border: none;" frameborder="0" />
      <!-- 视频预览 -->
      <video v-else-if="previewType === 'video' && previewUrl" :src="previewUrl" controls autoplay style="width: 100%; max-height: 80vh; display: block;" />
      <!-- 加载中 -->
      <div v-else-if="previewVisible" class="py-12 text-center text-gray-500">
        <Spin size="large" />
        <p class="mt-4">加载中...</p>
      </div>
      <!-- 无预览 -->
      <div v-else class="py-12 text-center text-gray-500">该文件类型不支持预览</div>
    </Modal>

    <!-- 文件夹操作菜单 -->
    <Modal v-model:open="folderMenuVisible" title="文件夹操作" :footer="null">
      <div class="flex flex-col gap-2">
        <Button block @click="folderMenuAction('new')">新建子文件夹</Button>
        <Button block @click="folderMenuAction('rename')">重命名</Button>
        <Button block type="primary" @click="folderMenuAction('share')">分享文件夹</Button>
        <Button block danger @click="folderMenuAction('delete')">删除文件夹</Button>
      </div>
    </Modal>

    <!-- 文件夹分享 -->
    <Modal v-model:open="folderShareModalVisible" title="创建文件夹分享链接" @ok="confirmFolderShare">
      <Form layout="vertical">
        <FormItem label="文件夹">
          <Input :value="folderShareName" readonly />
        </FormItem>
        <FormItem label="过期时间">
          <Space>
            <InputNumber v-model:value="folderShareExpireHours" :min="0" style="width: 100px" />
            <span>小时（0表示永久有效）</span>
          </Space>
        </FormItem>
      </Form>
      <div v-if="folderShareResult" class="mt-4 p-3 bg-gray-50 rounded">
        <p class="mb-2 font-medium">分享链接：</p>
        <Input.Group compact>
          <Input :value="folderShareResult ? `${window.location.origin}${folderShareResult.shareUrl}` : ''" style="width: 280px" readonly />
          <Button type="primary" @click="copyFolderShareUrl">复制</Button>
        </Input.Group>
      </div>
    </Modal>
  </Page>
</template>

<style scoped>
.w-56 {
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}
</style>