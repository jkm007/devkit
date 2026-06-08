<script lang="ts" setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue';

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
  Radio,
  Select,
  Space,
  Spin,
  Table,
  Tag,
  Tooltip,
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
} from '#/api/file';
import type { FileApi } from '#/api/file';
import { $t } from '#/locales';
import { useUploadStore } from '#/store/upload';
import type { UploadPartDetail, UploadTaskItem } from '#/store/upload';

defineOptions({ name: 'FileList' });

const accessStore = useAccessStore();
const uploadStore = useUploadStore();

// ==================== 权限检查 ====================

const permissions = computed(() => accessStore.accessCodes || []);
const hasViewAllPermission = computed(() => permissions.value.includes('file:view:all'));
const hasUploadPermission = computed(() => permissions.value.includes('file:upload'));
const hasDeletePermission = computed(() => permissions.value.includes('file:delete'));
const hasSharePermission = computed(() => permissions.value.includes('file:share'));
const hasManagePermission = computed(() => permissions.value.includes('file:manage'));

// ==================== 状态 ====================

const loading = ref(false);
const folderTree = ref<FileApi.Folder[]>([]);
const currentFolderId = ref<number | null>(null);
const fileList = ref<FileApi.FileEntry[]>([]);
const totalFiles = ref(0);
const pagination = ref({ current: 1, pageSize: 20 });
const selectedRowKeys = ref<number[]>([]);

// 文件范围：own=自己的文件, all=所有文件
const fileScope = ref<'own' | 'all'>('own');

// 标签筛选
const selectedTagKeys = ref<string[]>([]);
const availableTags = ref<{ id: number; key: string; value: string; name: string; icon: string; color: string }[]>([]);

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
const shareResult = ref<{ shareCode: string } | null>(null);
const shareExpireHours = ref(0);
const shareLoading = ref(false);

// 计算属性：是否有分享结果
const hasShareResult = computed(() => !!shareResult.value?.shareCode);

// 计算属性：分享链接完整URL
const shareFullUrl = computed(() => {
  if (!shareResult.value?.shareCode) return '';
  return `${window.location.origin}/share/${shareResult.value.shareCode}`;
});

// 计算属性：文件夹分享链接完整URL
const folderShareFullUrl = computed(() => {
  if (!folderShareResult.value?.shareCode) return '';
  return `${window.location.origin}/share/${folderShareResult.value.shareCode}`;
});

// 预览
const previewVisible = ref(false);
const previewUrl = ref('');
const previewName = ref('');
const previewType = ref(''); // 'image' | 'video' | 'pdf'
const previewToken = ref('');

// 监听预览 Modal 关闭，释放 Blob URL 防止内存泄漏
watch(previewVisible, (newVal) => {
  if (!newVal && previewUrl.value && previewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(previewUrl.value);
    previewUrl.value = '';
  }
});

// 批量操作
const batchMoveModalVisible = ref(false);
const batchTargetFolderId = ref<number | null>(null);

// 文件标签编辑
const tagEditModalVisible = ref(false);
const tagEditFileId = ref<number | null>(null);
const tagEditFileName = ref('');
const tagEditSelectedTags = ref<number[]>([]);
const fileTags = ref<FileApi.TagInfo[]>([]);

// 上传详情弹窗
const uploadDetailVisible = ref(false);
const uploadDetailTask = ref<UploadTaskItem | null>(null);

// 计算属性：上传任务列表
const uploadTasks = computed(() => uploadStore.tasks);
const uploading = computed(() => uploadStore.uploadingCount > 0);

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
      scope: fileScope.value,
      tagKeys: selectedTagKeys.value.length > 0 ? selectedTagKeys.value.join(',') : undefined,
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

// 加载可用标签
async function loadAvailableTags() {
  try {
    const { getAllTags } = await import('#/api/system/tag');
    const tags = await getAllTags();
    availableTags.value = tags.map(tag => ({
      id: tag.id,
      key: tag.tagKey,
      value: tag.tagValue,
      name: tag.tagName,
      icon: tag.icon,
      color: tag.color,
    }));
  } catch {
    // 静默失败
  }
}

function handleFolderSelect(keys: (string | number)[]) {
  const key = keys[0];
  // 点击"全部文件"或取消选择时，显示全部文件
  currentFolderId.value = key === '__all__' || key === undefined ? null : Number(key);
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
const folderShareResult = ref<{ shareCode: string } | null>(null);
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
  const url = folderShareFullUrl.value;
  if (!url) {
    message.error('分享链接不存在');
    return;
  }
  fallbackCopy(url);
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

// ==================== 文件标签编辑 ====================

async function openTagEditModal(file: FileApi.FileEntry) {
  tagEditFileId.value = file.id;
  tagEditFileName.value = file.name;
  tagEditSelectedTags.value = file.tags?.map(t => t.id) || [];
  tagEditModalVisible.value = true;

  // 加载文件当前标签
  try {
    const { getFileTags } = await import('#/api/file');
    fileTags.value = await getFileTags(file.id);
  } catch {
    fileTags.value = [];
  }
}

async function handleTagEditSubmit() {
  if (!tagEditFileId.value) return;

  try {
    const { batchUpdateFileTags } = await import('#/api/file');
    await batchUpdateFileTags(tagEditFileId.value, tagEditSelectedTags.value);
    message.success('标签更新成功');
    tagEditModalVisible.value = false;
    loadFileList();
  } catch (error: any) {
    message.error(error.message || '更新失败');
  }
}

// 获取有效的文件 ID（排除上传任务）
const validFileIds = computed(() => {
  return selectedRowKeys.value.filter((key) => typeof key === 'number') as number[];
});

// 批量删除
async function handleBatchDelete() {
  if (validFileIds.value.length === 0) {
    message.warning('请先选择文件');
    return;
  }

  try {
    const result = await batchDeleteFiles(validFileIds.value);
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
  if (validFileIds.value.length === 0) {
    message.warning('请先选择文件');
    return;
  }
  batchTargetFolderId.value = currentFolderId.value;
  batchMoveModalVisible.value = true;
}

// 执行批量移动
async function handleBatchMove() {
  try {
    const result = await batchMoveFiles(validFileIds.value, batchTargetFolderId.value || undefined);
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

  // 先显示 Modal 和加载状态
  previewVisible.value = true;

  // 视频和 PDF 使用预签名 URL（支持流式加载，不需要下载整个文件）
  if (isVideo || isPdf) {
    previewType.value = isVideo ? 'video' : 'pdf';
    try {
      const response = await fetch(`/api/files/${file.id}/preview-url`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });
      if (response.ok) {
        const result = await response.json();
        if (result.code === 0) {
          previewUrl.value = `/api${result.data.url}`;
          return;
        }
      }
      message.error('获取预览失败');
      previewVisible.value = false;
    } catch {
      message.error('获取预览失败');
      previewVisible.value = false;
    }
    return;
  }

  // 图片需要先下载（通常较小，blob URL 即可）
  const viewUrl = `/api/files/${file.id}/view`;
  try {
    const response = await fetch(viewUrl, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (response.ok) {
      const blob = await response.blob();
      previewUrl.value = URL.createObjectURL(blob);
      previewType.value = 'image';
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
  if (shareResult.value) {
    closeShareModal();
    return;
  }

  shareLoading.value = true;
  try {
    const result = await createFileShare(shareFileId.value!, {
      expireHours: shareExpireHours.value || undefined,
    });
    shareResult.value = { ...result };
    message.success('分享链接已生成');
  } catch (err) {
    console.error('Share error:', err);
    message.error('分享失败');
  } finally {
    shareLoading.value = false;
  }
}

function closeShareModal() {
  shareModalVisible.value = false;
  shareResult.value = null;
}

function copyShareUrl() {
  const url = shareFullUrl.value;
  fallbackCopy(url);
}

function fallbackCopy(text: string) {
  const textarea = document.createElement('textarea');
  textarea.value = text;
  textarea.style.cssText = 'position:fixed;left:0;top:0;opacity:0';
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  try {
    document.execCommand('copy');
    message.success('链接已复制到剪贴板');
  } catch {
    message.error('复制失败，请手动复制');
  }
  document.body.removeChild(textarea);
}

// ==================== 上传 ====================

async function handleUpload(file: File) {
  try {
    await uploadStore.uploadFile(file);
    message.success(`${file.name} 上传成功`);
    loadFileList();
  } catch (err) {
    message.error(`上传失败: ${err}`);
  }
  return false;
}

// 查看上传详情
function showUploadDetail(task: UploadTaskItem) {
  uploadDetailTask.value = task;
  uploadDetailVisible.value = true;
}

// 格式化耗时
function formatDuration(ms: number | undefined): string {
  if (!ms) return '-';
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

// ==================== 初始化 ====================

onMounted(() => {
  loadFolderTree();
  loadFileList();
  loadAvailableTags();
});

// ==================== 表格列 ====================

// 存储类型标签映射
const storageTypeLabels: Record<string, { label: string; color: string }> = {
  local: { label: '本地', color: 'default' },
  minio: { label: 'MinIO', color: 'blue' },
  oss: { label: 'OSS', color: 'orange' },
  cos: { label: 'COS', color: 'green' },
};

const columns = computed(() => {
  const cols = [
    { title: '文件名', dataIndex: 'name', key: 'name', width: 200, ellipsis: true },
    { title: '大小', dataIndex: 'size', key: 'size', width: 80, customRender: ({ text }: any) => formatFileSize(text) },
    { title: '状态', key: 'status', width: 150 },
    { title: '存储', key: 'storage', width: 80 },
    { title: '标签', key: 'tags', width: 200 },
  ];

  // 查看所有文件时显示上传者列
  if (fileScope.value === 'all') {
    cols.push({ title: '上传者', key: 'uploader', width: 100 });
  }

  cols.push(
    { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 120 },
    { title: '操作', key: 'operation', width: 180, fixed: 'right' as const },
  );

  return cols;
});

// 合并上传任务和文件列表为表格数据
const tableData = computed(() => {
  const tasks = uploadTasks.value.map(task => ({
    id: `task-${task.id}`,
    name: task.fileName,
    size: task.fileSize,
    contentType: task.contentType,
    storageType: 'uploading',
    isUploadTask: true,
    uploadTask: task,
    createdAt: new Date(task.startTime).toISOString(),
  }));

  const files = fileList.value.map(file => ({
    ...file,
    isUploadTask: false,
    uploadTask: null,
  }));

  // 上传中的任务放在最前面
  return [...tasks, ...files];
});

const treeData = computed<TreeProps['treeData']>(() => {
  const convert = (folders: FileApi.Folder[]): TreeProps['treeData'] =>
    folders.map((f) => ({
      key: f.id,
      title: f.name,
      type: f.type,
      children: f.children ? convert(f.children) : undefined,
    }));
  return [
    { key: '__all__', title: '全部文件', type: 'all' },
    ...convert(folderTree.value),
  ];
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
          :selected-keys="currentFolderId ? [currentFolderId] : ['__all__']"
          default-expand-all
          :style="{ '--ant-tree-node-selected-bg': '#e6f4ff' }"
          @select="handleFolderSelect"
        >
          <template #title="node">
            <div class="flex items-center gap-1 py-1 group">
              <span :class="node.type === 'all' ? 'i-ant-design:home-outlined' : node.type === 'avatar' ? 'i-ant-design:user-outlined' : 'i-ant-design:folder-outlined'" />
              <span class="flex-1 truncate">{{ node.title as string }}</span>
              <button
                v-if="node.type !== 'all'"
                type="button"
                class="opacity-0 group-hover:opacity-100 ml-1 px-1 py-0.5 text-xs rounded hover:bg-gray-200"
                @click.stop="showFolderMenu(node)"
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
            <Upload v-if="hasUploadPermission" :show-upload-list="false" :before-upload="handleUpload" :multiple="true" :disabled="uploading">
              <Button type="primary" :loading="uploading">上传文件</Button>
            </Upload>
            <Button v-if="validFileIds.length > 0 && hasDeletePermission" danger @click="handleBatchDelete">
              批量删除 ({{ validFileIds.length }})
            </Button>
            <Button v-if="validFileIds.length > 0 && hasManagePermission" @click="openBatchMoveModal">
              批量移动 ({{ validFileIds.length }})
            </Button>
            <!-- 文件范围切换 -->
            <div v-if="hasViewAllPermission" class="ml-4">
              <Radio.Group v-model:value="fileScope" button-style="solid" @change="loadFileList">
                <Radio.Button value="own">我的文件</Radio.Button>
                <Radio.Button value="all">所有文件</Radio.Button>
              </Radio.Group>
            </div>
          </Space>
          <Space>
            <!-- 标签筛选 -->
            <Select
              v-if="availableTags.length > 0"
              v-model:value="selectedTagKeys"
              mode="multiple"
              placeholder="按标签筛选"
              style="min-width: 200px"
              :options="availableTags.map(t => ({ label: `${t.icon} ${t.name}`, value: `${t.key}:${t.value}` }))"
              @change="() => { pagination.current = 1; loadFileList(); }"
              allow-clear
              max-tag-count="2"
            />
            <span class="text-sm text-gray-500">共 {{ totalFiles }} 个文件</span>
          </Space>
        </div>

        <!-- 文件表格 -->
        <Table
          :columns="columns"
          :data-source="tableData"
          :loading="loading"
          :pagination="{ current: pagination.current, pageSize: pagination.pageSize, total: totalFiles + uploadTasks.length, showSizeChanger: true }"
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
            <template v-if="column.key === 'status'">
              <!-- 上传任务状态 -->
              <template v-if="record.isUploadTask && record.uploadTask">
                <div class="cursor-pointer" @click="showUploadDetail(record.uploadTask)">
                  <div v-if="record.uploadTask.status === 'uploading'" class="flex items-center gap-2">
                    <Progress :percent="record.uploadTask.progress" :show-info="false" status="active" size="small" style="width: 80px; margin: 0;" />
                    <span class="text-xs text-blue-600">{{ record.uploadTask.progress }}%</span>
                  </div>
                  <div v-else-if="record.uploadTask.status === 'processing'" class="flex items-center gap-1">
                    <Spin size="small" />
                    <span class="text-xs text-yellow-600">处理中...</span>
                  </div>
                  <div v-else-if="record.uploadTask.status === 'completed'" class="text-xs text-green-600">
                    <span class="i-ant-design:check-circle-outlined mr-1" />上传成功
                    <div v-if="record.uploadTask.routing" class="text-xs text-gray-500 mt-1">
                      <Tag size="small" :color="storageTypeLabels[record.uploadTask.routing.driver]?.color || 'default'">
                        {{ storageTypeLabels[record.uploadTask.routing.driver]?.label || record.uploadTask.routing.driver }}
                      </Tag>
                      <span v-if="record.uploadTask.routing.ruleName">{{ record.uploadTask.routing.ruleName }}</span>
                    </div>
                  </div>
                  <div v-else-if="record.uploadTask.status === 'failed'" class="text-xs text-red-600">
                    <span class="i-ant-design:close-circle-outlined mr-1" />上传失败
                  </div>
                </div>
              </template>
              <!-- 正常文件状态 -->
              <template v-else>
                <Tag color="green">正常</Tag>
              </template>
            </template>
            <template v-if="column.key === 'storage'">
              <Tag :color="storageTypeLabels[record.storageType]?.color || 'default'">
                {{ storageTypeLabels[record.storageType]?.label || record.storageType || '本地' }}
              </Tag>
            </template>
            <template v-if="column.key === 'tags'">
              <template v-if="record.tags && record.tags.length > 0">
                <Tag
                  v-for="tag in record.tags.slice(0, 3)"
                  :key="tag.id"
                  :color="tag.color"
                  class="mr-1 mb-1"
                >
                  {{ tag.icon }} {{ tag.name }}
                </Tag>
                <Tooltip v-if="record.tags.length > 3" :title="record.tags.map(t => `${t.icon} ${t.name}`).join(', ')">
                  <Tag>+{{ record.tags.length - 3 }}</Tag>
                </Tooltip>
              </template>
              <span v-else class="text-gray-400">-</span>
            </template>
            <template v-if="column.key === 'uploader'">
              <span class="text-sm">{{ record.uploaderName || '-' }}</span>
            </template>
            <template v-if="column.key === 'operation'">
              <Space size="small">
                <template v-if="record.isUploadTask">
                  <Button type="link" size="small" @click="showUploadDetail(record.uploadTask)">详情</Button>
                </template>
                <template v-else>
                  <Button type="link" size="small" @click="handlePreview(record)">预览</Button>
                  <Button type="link" size="small" @click="handleDownload(record)">下载</Button>
                  <Button type="link" size="small" @click="openTagEditModal(record)">标签</Button>
                  <Button v-if="hasSharePermission" type="link" size="small" @click="handleShare(record)">分享</Button>
                  <Button v-if="hasDeletePermission" type="link" size="small" danger @click="handleDeleteFile(record.id, record.name)">删除</Button>
                </template>
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
      <p class="mb-2">将移动 {{ validFileIds.length }} 个文件</p>
      <Form layout="vertical">
        <FormItem label="目标文件夹">
          <TreeSelect v-model:value="batchTargetFolderId" :tree-data="folderSelectData" placeholder="选择文件夹" allow-clear />
        </FormItem>
      </Form>
    </Modal>

    <!-- 分享 -->
    <Modal v-model:open="shareModalVisible" title="创建分享链接" :closable="true" :maskClosable="false">
      <template #footer>
        <Button @click="closeShareModal">{{ hasShareResult ? '关闭' : '取消' }}</Button>
        <Button v-if="!hasShareResult" type="primary" :loading="shareLoading" @click="confirmShare">确定</Button>
      </template>
      <div v-if="!hasShareResult">
        <Form layout="vertical">
          <FormItem label="过期时间">
            <Space>
              <InputNumber v-model:value="shareExpireHours" :min="0" style="width: 100px" />
              <span>小时（0表示永久有效）</span>
            </Space>
          </FormItem>
        </Form>
      </div>
      <div v-else class="p-3 bg-gray-50 rounded">
        <p class="mb-2 font-medium">分享链接：</p>
        <div class="flex gap-2">
          <Input :value="shareFullUrl" readonly class="flex-1" />
          <Button type="primary" @click="copyShareUrl">复制</Button>
        </div>
      </div>
    </Modal>

    <!-- 预览 -->
    <Modal
      v-model:open="previewVisible"
      :title="previewName"
      :footer="null"
      :width="previewType === 'video' ? 960 : 800"
      :maskClosable="true"
      :keyboard="true"
      :destroyOnClose="true"
      class="preview-modal"
    >
      <!-- 图片预览 -->
      <div v-if="previewType === 'image' && previewUrl" class="text-center p-6">
        <Image :src="previewUrl" class="max-w-full" style="max-height: 600px" />
      </div>
      <!-- PDF 预览 -->
      <iframe v-else-if="previewType === 'pdf' && previewUrl" :src="previewUrl" style="width: 100%; height: 600px; border: none;" />
      <!-- 视频预览 -->
      <div v-else-if="previewType === 'video' && previewUrl" class="video-container">
        <video
          ref="videoPlayer"
          :src="previewUrl"
          controls
          autoplay
          preload="auto"
          playsinline
          webkit-playsinline
          style="width: 100%; max-height: 70vh; display: block; background: #000;"
        />
      </div>
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

    <!-- 文件标签编辑 -->
    <Modal
      v-model:open="tagEditModalVisible"
      :title="`编辑标签 - ${tagEditFileName}`"
      @ok="handleTagEditSubmit"
      width="500px"
    >
      <div class="mb-4">
        <p class="text-gray-500 mb-2">当前标签：</p>
        <div v-if="fileTags.length > 0" class="flex flex-wrap gap-2">
          <Tag
            v-for="tag in fileTags"
            :key="tag.id"
            :color="tag.color"
          >
            {{ tag.icon }} {{ tag.name }}
          </Tag>
        </div>
        <span v-else class="text-gray-400">暂无标签</span>
      </div>

      <div>
        <p class="text-gray-500 mb-2">选择标签：</p>
        <Select
          v-model:value="tagEditSelectedTags"
          mode="multiple"
          placeholder="选择标签"
          style="width: 100%"
          :options="availableTags.map(t => ({
            label: `${t.icon} ${t.name}`,
            value: t.id,
          }))"
        />
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
          <Input :value="folderShareFullUrl" style="width: 280px" readonly />
          <Button type="primary" @click="copyFolderShareUrl">复制</Button>
        </Input.Group>
      </div>
    </Modal>

    <!-- 上传详情弹窗 -->
    <Modal
      v-model:open="uploadDetailVisible"
      title="上传详情"
      :footer="null"
      width="700px"
    >
      <div v-if="uploadDetailTask" class="space-y-4">
        <!-- 基本信息 -->
        <Descriptions bordered size="small">
          <DescriptionsItem label="文件名">{{ uploadDetailTask.fileName }}</DescriptionsItem>
          <DescriptionsItem label="文件大小">{{ formatFileSize(uploadDetailTask.fileSize) }}</DescriptionsItem>
          <DescriptionsItem label="总分片数">{{ uploadDetailTask.totalParts }}</DescriptionsItem>
          <DescriptionsItem label="已上传">{{ uploadDetailTask.uploadedParts }} / {{ uploadDetailTask.totalParts }}</DescriptionsItem>
          <DescriptionsItem label="状态">
            <Tag :color="uploadDetailTask.status === 'completed' ? 'green' : uploadDetailTask.status === 'failed' ? 'red' : 'blue'">
              {{ uploadDetailTask.status === 'uploading' ? '上传中' : uploadDetailTask.status === 'processing' ? '处理中' : uploadDetailTask.status === 'completed' ? '已完成' : '失败' }}
            </Tag>
          </DescriptionsItem>
          <DescriptionsItem label="进度">{{ uploadDetailTask.progress }}%</DescriptionsItem>
          <DescriptionsItem label="开始时间">{{ new Date(uploadDetailTask.startTime).toLocaleString() }}</DescriptionsItem>
          <DescriptionsItem v-if="uploadDetailTask.endTime" label="结束时间">{{ new Date(uploadDetailTask.endTime).toLocaleString() }}</DescriptionsItem>
          <DescriptionsItem v-if="uploadDetailTask.totalDuration" label="总耗时">{{ formatDuration(uploadDetailTask.totalDuration) }}</DescriptionsItem>
          <DescriptionsItem v-if="uploadDetailTask.errorMessage" label="错误信息" :span="2">
            <span class="text-red-500">{{ uploadDetailTask.errorMessage }}</span>
          </DescriptionsItem>
        </Descriptions>

        <!-- 分片详情表格 -->
        <div v-if="uploadDetailTask.partDetails.length > 0">
          <h4 class="mb-2 font-medium">分片详情</h4>
          <Table
            :columns="[
              { title: '分片', dataIndex: 'partNumber', width: 80 },
              { title: '状态', key: 'partStatus', width: 100 },
              { title: '开始时间', key: 'partStart', width: 180 },
              { title: '耗时', key: 'partDuration', width: 100 },
            ]"
            :data-source="uploadDetailTask.partDetails"
            :pagination="false"
            size="small"
            :scroll="{ y: 300 }"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'partStatus'">
                <Tag :color="record.status === 'completed' ? 'green' : record.status === 'uploading' ? 'blue' : 'default'">
                  {{ record.status === 'completed' ? '已完成' : record.status === 'uploading' ? '上传中' : '待上传' }}
                </Tag>
              </template>
              <template v-if="column.key === 'partStart'">
                {{ record.startTime ? new Date(record.startTime).toLocaleTimeString() : '-' }}
              </template>
              <template v-if="column.key === 'partDuration'">
                {{ formatDuration(record.duration) }}
              </template>
            </template>
          </Table>
        </div>
      </div>
    </Modal>
  </Page>
</template>

<style scoped>
.w-56 {
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}

/* 目录树选中高亮 */
.w-56 :deep(.ant-tree .ant-tree-node-selected) {
  background-color: #e6f4ff !important;
  border-radius: 4px;
}

.w-56 :deep(.ant-tree .ant-tree-node-content-wrapper:hover) {
  background-color: #f0f5ff;
}

/* 视频预览容器 */
.video-container {
  position: relative;
  width: 100%;
  background: #000;
  overflow: hidden;
}

.video-container video {
  width: 100%;
  max-height: 70vh;
  display: block;
}
</style>

<style>
/* 全局样式 - 修复视频预览控件 */
.preview-modal .ant-modal-body {
  padding: 0 !important;
}

.preview-modal .ant-modal-close {
  z-index: 10;
  color: #fff;
  top: 8px;
  right: 8px;
}

.preview-modal .ant-modal-close:hover {
  color: rgba(255, 255, 255, 0.8);
}

/* 确保视频控件可点击 */
.preview-modal video::-webkit-media-controls {
  pointer-events: auto !important;
}

.preview-modal video::-webkit-media-controls-panel {
  pointer-events: auto !important;
}

.preview-modal video::-webkit-media-controls-play-button {
  pointer-events: auto !important;
}

.preview-modal video::-webkit-media-controls-timeline {
  pointer-events: auto !important;
}

.preview-modal video::-webkit-media-controls-volume-slider {
  pointer-events: auto !important;
}

.preview-modal video::-webkit-media-controls-fullscreen-button {
  pointer-events: auto !important;
}
</style>