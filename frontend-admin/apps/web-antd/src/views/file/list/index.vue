<script lang="ts" setup>
import type {
  VxeTableGridColumns,
  VxeTableGridOptions,
} from '#/adapter/vxe-table';
import type { FileApi } from '#/api/file';
import type { UploadTaskItem } from '#/store/upload';

import { computed, onMounted, ref, watch } from 'vue';

import { Page, Tree } from '@vben/common-ui';
import { Plus } from '@vben/icons';
import { useAccessStore } from '@vben/stores';

import {
  Button,
  Card,
  Descriptions,
  DescriptionsItem,
  Dropdown,
  Form,
  FormItem,
  Image,
  Input,
  InputNumber,
  InputPassword,
  Menu,
  MenuItem,
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
  TreeSelect,
  Upload,
} from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  batchDeleteFiles,
  batchMoveFiles,
  checkFileShares,
  createFileShare,
  createFolder,
  createFolderShare,
  deleteFile,
  deleteFolder,
  getFolderTree,
  listFiles,
  moveFile,
  renameFolder,
} from '#/api/file';
import { useUploadStore } from '#/store/upload';
import { fallbackCopy, formatFileSize, getFileIcon } from '#/utils/file-utils';

import FolderUploadModal from './folder-upload-modal.vue';

defineOptions({ name: 'FileList' });

/** 表格行数据（文件条目 + 上传任务） */
interface FileRowItem {
  id: number | string;
  name: string;
  size: number;
  contentType?: string;
  storageType: string;
  folderId?: null | number;
  createdAt: string;
  updatedAt?: string;
  uploaderName?: string;
  tags?: FileApi.TagInfo[];
  /** 是否为上传任务行 */
  isUploadTask: boolean;
  /** 关联的上传任务 */
  uploadTask: null | UploadTaskItem;
}

/** 文件夹树节点 */
interface FolderTreeNode {
  key: number | string;
  title: string;
  type?: string;
  children?: FolderTreeNode[];
}

/** 文件夹下拉选择节点 */
interface FolderSelectNode {
  value: number | string;
  label: string;
  children?: FolderSelectNode[];
}

const accessStore = useAccessStore();
const uploadStore = useUploadStore();

// ==================== 权限 ====================

const permissions = computed(() => accessStore.accessCodes || []);
const hasViewAllPermission = computed(() =>
  permissions.value.includes('file:view:all'),
);
const hasUploadPermission = computed(() =>
  permissions.value.includes('file:upload'),
);
const hasDeletePermission = computed(() =>
  permissions.value.includes('file:delete'),
);
const hasSharePermission = computed(() =>
  permissions.value.includes('file:share'),
);
const hasManagePermission = computed(() =>
  permissions.value.includes('file:manage'),
);

// ==================== 状态 ====================

const folderTree = ref<FileApi.Folder[]>([]);
const currentFolderId = ref<null | number>(null);
const selectedRowKeys = ref<number[]>([]);
const fileScope = ref<'all' | 'own'>('own');
const selectedTagKeys = ref<string[]>([]);
const availableTags = ref<
  {
    color: string;
    icon: string;
    id: number;
    key: string;
    name: string;
    value: string;
  }[]
>([]);

// 文件夹操作
const newFolderModalVisible = ref(false);
const newFolderName = ref('');
const newFolderParentId = ref<null | number>(null);
const renameFolderModalVisible = ref(false);
const renameFolderId = ref<null | number>(null);
const renameFolderName = ref('');
const deleteFolderModalVisible = ref(false);
const deleteFolderId = ref<null | number>(null);
const deleteFolderName = ref('');
const folderMenuVisible = ref(false);
const folderMenuId = ref<null | number>(null);
const folderMenuName = ref('');
const moveFileModalVisible = ref(false);
const moveFileId = ref<null | number>(null);
const moveTargetFolderId = ref<null | number>(null);
const batchMoveModalVisible = ref(false);
const batchTargetFolderId = ref<null | number>(null);

// 分享
const shareModalVisible = ref(false);
const shareFileId = ref<null | number>(null);
const shareResult = ref<null | { shareCode: string }>(null);
const shareExpireHours = ref(0);
const sharePassword = ref('');
const shareLoading = ref(false);
const folderShareModalVisible = ref(false);
const folderShareId = ref<null | number>(null);
const folderShareName = ref('');
const folderShareResult = ref<null | { shareCode: string }>(null);
const folderShareExpireHours = ref(0);
const folderSharePassword = ref('');

// 预览
const previewVisible = ref(false);
const previewUrl = ref('');
const previewName = ref('');
const previewType = ref('');

// 标签编辑
const tagEditModalVisible = ref(false);
const tagEditFileId = ref<null | number>(null);
const tagEditFileName = ref('');
const tagEditSelectedTags = ref<number[]>([]);
const fileTags = ref<FileApi.TagInfo[]>([]);

// 上传详情
const uploadDetailVisible = ref(false);
const uploadDetailTask = ref<null | UploadTaskItem>(null);

// 文件详情
const fileDetailVisible = ref(false);
const fileDetailData = ref<FileRowItem | null>(null);

// 文件夹上传
const folderInputRef = ref<HTMLInputElement | null>(null);
const folderUploadVisible = ref(false);
const folderUploadName = ref('');
const folderUploadFiles = ref<File[]>([]);
const folderUploadModalRef = ref<InstanceType<typeof FolderUploadModal> | null>(null);

// 下载进度
const downloadProgressVisible = ref(false);
const downloadProgress = ref(0);
const downloadFileName = ref('');
const downloadStatus = ref<'done' | 'downloading' | 'error' | 'saving'>(
  'downloading',
);

// ==================== 计算属性 ====================

const uploadTasks = computed(() => uploadStore.tasks);
const uploading = computed(() => uploadStore.uploadingCount > 0);
const hasShareResult = computed(() => !!shareResult.value?.shareCode);
const shareFullUrl = computed(() =>
  shareResult.value?.shareCode
    ? `${window.location.origin}/share/${shareResult.value.shareCode}`
    : '',
);
const folderShareFullUrl = computed(() =>
  folderShareResult.value?.shareCode
    ? `${window.location.origin}/share/${folderShareResult.value.shareCode}`
    : '',
);
const validFileIds = computed(
  () =>
    selectedRowKeys.value.filter((key) => typeof key === 'number') as number[],
);

watch(previewVisible, (newVal) => {
  if (!newVal && previewUrl.value && previewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(previewUrl.value);
    previewUrl.value = '';
  }
});

/** 标记：下次 query 仅使用缓存数据重新合并，不发起后端请求 */
const useCachedQuery = ref(false);

// 监听上传任务数量变化，仅更新本地合并结果，不触发后端请求
// 进度更新不需要刷新 — uploadTask 是 Pinia 响应式引用，模板自动更新
watch(
  () => uploadTasks.value.length,
  (newLen, oldLen) => {
    if (newLen !== (oldLen ?? 0)) {
      useCachedQuery.value = true;
      gridApi.query();
    }
  },
);

// ==================== 工具函数 ====================

// formatFileSize, getFileIcon, fallbackCopy are imported from '#/utils/file-utils'

function formatDuration(ms: number | undefined): string {
  if (!ms) return '-';
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`;
  const minutes = Math.floor(ms / 60_000);
  const seconds = Math.floor((ms % 60_000) / 1000);
  return `${minutes}m ${seconds}s`;
}

const storageTypeLabels: Record<string, { color: string; label: string; }> = {
  local: { label: '本地', color: 'default' },
  minio: { label: 'MinIO', color: 'blue' },
  oss: { label: 'OSS', color: 'orange' },
  cos: { label: 'COS', color: 'green' },
  uploading: { label: '上传中', color: 'blue' },
};

// ==================== VxeTable 配置 ====================

// 列定义
const showUploader = ref(false);

const tableColumns = computed(() => {
  const cols: VxeTableGridColumns<FileRowItem> = [
    { type: 'checkbox', width: 50, align: 'center' },
    {
      field: 'name',
      title: '文件名',
      minWidth: 200,
      slots: { default: 'name' },
    },
    {
      field: 'size',
      title: '大小',
      width: 100,
      align: 'center',
      formatter: ({ cellValue }: { cellValue: number }) =>
        formatFileSize(cellValue),
    },
    {
      field: 'status',
      title: '状态',
      width: 100,
      align: 'center',
      slots: { default: 'status' },
    },
    {
      field: 'storageType',
      title: '存储',
      width: 100,
      align: 'center',
      slots: { default: 'storage' },
    },
    { field: 'tags', title: '标签', minWidth: 150, slots: { default: 'tags' } },
  ];

  // 查看所有文件时显示上传者列
  if (showUploader.value) {
    cols.push({
      field: 'uploaderName',
      title: '上传者',
      width: 120,
      align: 'center',
    });
  }

  cols.push(
    { field: 'createdAt', title: '时间', width: 160, align: 'center' },
    {
      field: 'operation',
      title: '操作',
      width: 180,
      align: 'center',
      fixed: 'right',
      slots: { default: 'action' },
    },
  );

  return cols;
});

// 缓存最后一次 API 查询结果，用于上传进度更新时避免重复请求
const lastApiResult = ref<{ items: FileApi.FileEntry[]; total: number }>({
  items: [],
  total: 0,
});

// 合并上传任务与 API 数据
function mergeWithUploadTasks(apiResult: {
  items: FileApi.FileEntry[];
  total: number;
}) {
  const activeTasks = uploadTasks.value
    .filter(
      (task) => task.status === 'uploading' || task.status === 'processing',
    )
    .map((task) => ({
      id: `task-${task.id}`,
      name: task.fileName,
      size: task.fileSize,
      contentType: task.contentType,
      storageType: 'uploading',
      isUploadTask: true,
      uploadTask: task,
      createdAt: new Date(task.startTime).toISOString(),
    }));

  const files = apiResult.items.map((file) => ({
    ...file,
    isUploadTask: false,
    uploadTask: null,
  }));

  return {
    items: [...activeTasks, ...files],
    total: apiResult.total + activeTasks.length,
  };
}

const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions: {
    columns: tableColumns.value,
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }) => {
          // 上传任务变化时，仅使用缓存数据重新合并，不发起后端请求
          if (useCachedQuery.value) {
            useCachedQuery.value = false;
            return mergeWithUploadTasks(lastApiResult.value);
          }

          const result = await listFiles({
            folderId: currentFolderId.value ?? undefined,
            page: page.currentPage,
            pageSize: page.pageSize,
            scope: fileScope.value,
            tagKeys:
              selectedTagKeys.value.length > 0
                ? selectedTagKeys.value.join(',')
                : undefined,
          });

          // 缓存 API 结果
          lastApiResult.value = {
            items: result?.items || [],
            total: result?.total || 0,
          };

          return mergeWithUploadTasks(lastApiResult.value);
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    checkboxConfig: {
      highlight: true,
    },
    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      search: false,
      zoom: true,
    },
  } as VxeTableGridOptions<FileRowItem>,
  gridEvents: {
    checkboxChange({ records }: { records: FileRowItem[] }) {
      selectedRowKeys.value = records
        .map((r) => r.id)
        .filter((id): id is number => typeof id === 'number');
    },
    checkboxAll({ records }: { records: FileRowItem[] }) {
      selectedRowKeys.value = records
        .map((r) => r.id)
        .filter((id): id is number => typeof id === 'number');
    },
  },
});

// ==================== 数据加载 ====================

async function loadFolderTree() {
  try {
    const result = await getFolderTree(fileScope.value);
    folderTree.value = result || [];
  } catch {
    message.error('加载文件夹树失败');
  }
}

async function loadAvailableTags() {
  try {
    const { getTagsForUser } = await import('#/api/system/tag');
    const tags = await getTagsForUser();
    availableTags.value = tags.map((tag) => ({
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

function onRefresh() {
  gridApi.query();
}

function handleFolderClick(node: FolderTreeNode) {
  const key = node?.key;
  currentFolderId.value = !key || key === '__all__' ? null : Number(key);
  onRefresh();
}

function handleScopeChange() {
  // 更新列定义（显示/隐藏上传者列）
  showUploader.value = fileScope.value === 'all';
  gridApi.setGridOptions({ columns: tableColumns.value });
  // 重新加载文件夹树（不同范围可能有不同的文件夹）
  loadFolderTree();
  onRefresh();
}

function handleTagFilter() {
  onRefresh();
}

// 文件操作菜单处理
function handleFileAction(key: string, row: FileRowItem) {
  switch (key) {
    case 'delete': {
      handleDeleteWithShareCheck(row);
      break;
    }
    case 'detail': {
      openFileDetail(row);
      break;
    }
    case 'download': {
      handleDownload(row);
      break;
    }
    case 'move': {
      openMoveFileModal(row.id as number);
      break;
    }
    case 'preview': {
      handlePreview(row);
      break;
    }
    case 'share': {
      handleShare(row);
      break;
    }
    case 'tag': {
      openTagEditModal(row);
      break;
    }
  }
}

// ==================== 文件夹操作 ====================

function showFolderMenu(nodeData: FolderTreeNode) {
  folderMenuId.value = nodeData.key as number;
  folderMenuName.value = nodeData.title as string;
  folderMenuVisible.value = true;
}

function folderMenuAction(action: string) {
  folderMenuVisible.value = false;
  switch (action) {
  case 'delete': {
  openDeleteFolderModal(folderMenuId.value ?? 0, folderMenuName.value);
  break;
  }
  case 'new': {
  openNewFolderModal(folderMenuId.value ?? 0);
  break;
  }
  case 'rename': {
  openRenameFolderModal(folderMenuId.value ?? 0, folderMenuName.value);
  break;
  }
  case 'share': { {
  openFolderShareModal(folderMenuId.value ?? 0, folderMenuName.value);
  // No default
  }
  break;
  }
  }
}

function openFolderShareModal(id: number, name: string) {
  folderShareId.value = id;
  folderShareName.value = name;
  folderShareExpireHours.value = 0;
  folderSharePassword.value = '';
  folderShareResult.value = null;
  folderShareModalVisible.value = true;
}

async function confirmFolderShare() {
  try {
    const result = await createFolderShare(folderShareId.value ?? 0, {
      expireHours: folderShareExpireHours.value || undefined,
      password: folderSharePassword.value || undefined,
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
  if (!renameFolderName.value.trim()) {
    message.warning('请输入文件夹名称');
    return;
  }
  try {
    await renameFolder(renameFolderId.value ?? 0, {
      name: renameFolderName.value.trim(),
    });
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
    await deleteFolder(deleteFolderId.value ?? 0);
    message.success('删除成功');
    deleteFolderModalVisible.value = false;
    if (currentFolderId.value === deleteFolderId.value)
      currentFolderId.value = null;
    loadFolderTree();
    onRefresh();
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
      fileId: moveFileId.value ?? 0,
      targetFolderId: moveTargetFolderId.value || undefined,
    });
    message.success('移动成功');
    moveFileModalVisible.value = false;
    onRefresh();
  } catch {
    message.error('移动失败');
  }
}

async function handleDeleteWithShareCheck(row: FileRowItem) {
  try {
    const result = await checkFileShares(row.id as number);
    const shareCount = result?.shareCount || 0;

    if (shareCount > 0) {
      Modal.confirm({
        title: '文件有分享链接',
        content: `文件 "${row.name}" 当前有 ${shareCount} 个有效分享链接，移入回收站后分享链接将失效。确定继续吗？`,
        okText: '确定移入回收站',
        okType: 'danger',
        cancelText: '取消',
        onOk: async () => {
          await handleDeleteFile(row);
        },
      });
    } else {
      Modal.confirm({
        title: '移入回收站',
        content: `确定将文件 "${row.name}" 移入回收站吗？文件将在 7 天后自动清理。`,
        okText: '确定',
        okType: 'danger',
        cancelText: '取消',
        onOk: async () => {
          await handleDeleteFile(row);
        },
      });
    }
  } catch {
    // 检查失败时直接弹确认框
    Modal.confirm({
      title: '移入回收站',
      content: `确定将文件 "${row.name}" 移入回收站吗？文件将在 7 天后自动清理。`,
      okText: '确定',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        await handleDeleteFile(row);
      },
    });
  }
}

async function handleDeleteFile(row: FileRowItem) {
  try {
    await deleteFile(row.id as number);
    message.success(`已将 ${row.name} 移入回收站`);
    onRefresh();
  } catch {
    message.error('删除失败');
  }
}

async function openTagEditModal(row: FileRowItem) {
  tagEditFileId.value = row.id as number;
  tagEditFileName.value = row.name;
  tagEditSelectedTags.value = row.tags?.map((t) => t.id) || [];
  tagEditModalVisible.value = true;
  try {
    const { getFileTags } = await import('#/api/file');
    fileTags.value = await getFileTags(row.id as number);
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
    onRefresh();
  } catch (error: unknown) {
    message.error(error instanceof Error ? error.message : '更新失败');
  }
}

async function handleBatchDelete() {
  if (validFileIds.value.length === 0) {
    message.warning('请先选择文件');
    return;
  }

  // 检查选中文件是否有分享
  let totalShares = 0;
  try {
    const results = await Promise.all(
      validFileIds.value.map((id) =>
        checkFileShares(id).catch(() => ({ shareCount: 0 })),
      ),
    );
    totalShares = results.reduce((sum, r) => sum + (r?.shareCount || 0), 0);
  } catch {
    /* ignore */
  }

  const doDelete = async () => {
    try {
      const result = await batchDeleteFiles(validFileIds.value);
      if (result.errors?.length > 0)
        message.warning(
          `已移入回收站 ${result.deleted} 个文件，${result.errors.length} 个失败`,
        );
      else message.success(`已将 ${result.deleted} 个文件移入回收站`);
      selectedRowKeys.value = [];
      gridApi.grid?.clearCheckboxRow?.();
      onRefresh();
    } catch {
      message.error('批量删除失败');
    }
  };

  if (totalShares > 0) {
    Modal.confirm({
      title: '文件有分享链接',
      content: `选中的文件中有 ${totalShares} 个有效分享链接，移入回收站后这些分享链接将失效。确定继续吗？`,
      okText: '确定移入回收站',
      okType: 'danger',
      cancelText: '取消',
      onOk: doDelete,
    });
  } else {
    Modal.confirm({
      title: '批量移入回收站',
      content: `确定将 ${validFileIds.value.length} 个文件移入回收站吗？文件将在 7 天后自动清理。`,
      okText: '确定',
      okType: 'danger',
      cancelText: '取消',
      onOk: doDelete,
    });
  }
}

function openBatchMoveModal() {
  if (validFileIds.value.length === 0) {
    message.warning('请先选择文件');
    return;
  }
  batchTargetFolderId.value = currentFolderId.value;
  batchMoveModalVisible.value = true;
}

async function handleBatchMove() {
  try {
    const result = await batchMoveFiles(
      validFileIds.value,
      batchTargetFolderId.value || undefined,
    );
    if (result.errors?.length > 0)
      message.warning(
        `已移动 ${result.moved} 个文件，${result.errors.length} 个失败`,
      );
    else message.success(`已移动 ${result.moved} 个文件`);
    batchMoveModalVisible.value = false;
    selectedRowKeys.value = [];
    gridApi.grid?.clearCheckboxRow?.();
    onRefresh();
  } catch {
    message.error('批量移动失败');
  }
}

function handleDownload(row: FileRowItem) {
  const token = useAccessStore().accessToken;
  const url = `/api/v1/files/${row.id}/download`;

  // 显示下载进度弹窗
  downloadFileName.value = row.name;
  downloadProgress.value = 0;
  downloadStatus.value = 'downloading';
  downloadProgressVisible.value = true;

  const xhr = new XMLHttpRequest();
  xhr.open('GET', url, true);
  xhr.responseType = 'blob';
  xhr.setRequestHeader('Authorization', `Bearer ${token}`);

  // 进度回调
  xhr.addEventListener('progress', (event) => {
    if (event.lengthComputable) {
      downloadProgress.value = Math.round((event.loaded / event.total) * 100);
    }
  });

  xhr.addEventListener('load', () => {
    if (xhr.status >= 200 && xhr.status < 300) {
      downloadStatus.value = 'saving';
      downloadProgress.value = 100;
      const blob = xhr.response;
      const blobUrl = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = blobUrl;
      link.download = row.name;
      document.body.append(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(blobUrl);
      downloadStatus.value = 'done';
      message.success(`${row.name} 下载成功`);
      setTimeout(() => {
        downloadProgressVisible.value = false;
      }, 1500);
    } else {
      downloadStatus.value = 'error';
      message.error('下载失败');
    }
  });

  xhr.addEventListener('error', () => {
    downloadStatus.value = 'error';
    message.error('下载失败');
  });

  xhr.send();
}

async function handlePreview(row: FileRowItem) {
  previewName.value = row.name;
  previewUrl.value = '';
  previewType.value = '';

  const isImage = row.contentType?.startsWith('image/');
  const isVideo = row.contentType?.startsWith('video/');
  const isAudio = row.contentType?.startsWith('audio/');
  const isPdf = row.contentType?.includes('pdf');

  // 不支持预览的文件类型，弹出确认框询问是否下载
  if (!isImage && !isVideo && !isAudio && !isPdf) {
    Modal.confirm({
      title: '无法预览',
      content: `文件 "${row.name}" 不支持在线预览，是否下载查看？`,
      okText: '下载',
      cancelText: '取消',
      onOk: () => handleDownload(row),
    });
    return;
  }

  const token = accessStore.accessToken;
  previewVisible.value = true;

  // 视频、PDF、音频使用预签名 URL（预览场景使用较短过期时间 300 秒）
  if (isVideo || isPdf || isAudio) {
    previewType.value = isVideo ? 'video' : (isAudio ? 'audio' : 'pdf');
    try {
      const response = await fetch(
        `/api/v1/files/${row.id}/preview-url?expires=300`,
        {
          headers: { Authorization: `Bearer ${token}` },
        },
      );
      if (response.ok) {
        const result = await response.json();
        if (result.code === 0) {
          const url = result.data.url;
          // 云存储 presigned URL 是完整地址，本地存储是相对路径
          previewUrl.value = url.startsWith('http') ? url : `/api/v1${url}`;
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

  // 图片使用 blob URL
  try {
    const response = await fetch(`/api/v1/files/${row.id}/view`, {
      headers: { Authorization: `Bearer ${token}` },
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

async function handleShare(row: FileRowItem) {
  shareFileId.value = row.id as number;
  shareExpireHours.value = 0;
  sharePassword.value = '';
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
    const result = await createFileShare(shareFileId.value ?? 0, {
      expireHours: shareExpireHours.value || undefined,
      password: sharePassword.value || undefined,
    });
    shareResult.value = { ...result };
    message.success('分享链接已生成');
  } catch (error) {
    console.error('Share error:', error);
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
  fallbackCopy(shareFullUrl.value);
}

async function handleUpload(file: File) {
  try {
    await uploadStore.uploadFile(file, currentFolderId.value ?? undefined);
    message.success(`${file.name} 上传成功`);
    onRefresh();
  } catch (error: any) {
    const msg =
      error?.message || error?.response?.data?.message || error?.response?.data?.error || String(error);
    if (msg.includes('存储空间不足')) {
      message.warning(msg);
    } else {
      message.error(`上传失败: ${msg}`);
    }
  }
  return false;
}

function showUploadDetail(task: UploadTaskItem) {
  uploadDetailTask.value = task;
  uploadDetailVisible.value = true;
}

function openFileDetail(row: FileRowItem) {
  fileDetailData.value = row;
  fileDetailVisible.value = true;
}

function triggerFolderUpload() {
  folderInputRef.value?.click();
}

async function handleFolderSelect(event: Event) {
  const input = event.target as HTMLInputElement;
  const fileList = input.files;
  if (!fileList || fileList.length === 0) return;

  const files = [...fileList];

  // 获取根文件夹名称
  const firstFile = files[0]!;
  const relativePath = firstFile.webkitRelativePath;
  const rootFolderName = relativePath ? relativePath.split('/')[0] || '未知文件夹' : '未知文件夹';

  // 显示上传进度弹窗
  folderUploadName.value = rootFolderName;
  folderUploadFiles.value = files;
  folderUploadVisible.value = true;

  // 等待弹窗渲染
  await new Promise((resolve) => setTimeout(resolve, 100));

  // 开始上传
  await startFolderUpload(files);

  input.value = '';
}

async function startFolderUpload(files: File[]) {
  const modalRef = folderUploadModalRef.value;
  if (!modalRef) return;

  modalRef.setUploading(true);

  // 提取唯一的文件夹路径（相对于选中的根文件夹）
  const folderPathSet = new Set<string>();
  for (const file of files) {
    const relativePath = file.webkitRelativePath;
    if (!relativePath) continue;
    const parts = relativePath.split('/');
    parts.pop(); // 移除文件名
    parts.shift(); // 移除根文件夹名
    if (parts.length > 0) {
      for (let i = 1; i <= parts.length; i++) {
        folderPathSet.add(parts.slice(0, i).join('/'));
      }
    }
  }

  // 按深度排序（浅层优先）
  const sortedPaths = [...folderPathSet].toSorted(
    (a, b) => a.split('/').length - b.split('/').length,
  );

  // 创建文件夹并记录 ID
  const folderIdMap = new Map<string, number>();
  const rootId = currentFolderId.value ?? 0;
  folderIdMap.set('', rootId);

  // 辅助函数：在文件夹树中查找指定名称和父ID的文件夹
  const findFolderInTree = (
    folders: FileApi.Folder[],
    name: string,
    parentId: null | number,
  ): FileApi.Folder | undefined => {
    for (const f of folders) {
      if (f.name === name && (f.parentId ?? null) === parentId) {
        return f;
      }
      if (f.children) {
        const found = findFolderInTree(f.children, name, parentId);
        if (found) return found;
      }
    }
    return undefined;
  };

  // 创建文件夹
  for (const path of sortedPaths) {
    const parts = path.split('/');
    const folderName = parts[parts.length - 1] || '';
    const parentPath = parts.slice(0, -1).join('/');
    const parentId = folderIdMap.get(parentPath) ?? rootId;

    try {
      const result = await createFolder({
        name: folderName,
        parentId: parentId || undefined,
      });
      const folder = result as unknown as FileApi.Folder;
      if (folder?.id) {
        folderIdMap.set(path, folder.id);
      }
    } catch {
      const existingFolder = findFolderInTree(
        folderTree.value,
        folderName,
        parentId || null,
      );
      if (existingFolder?.id) {
        folderIdMap.set(path, existingFolder.id);
      }
    }
  }

  // 逐个上传文件
  for (const file of files) {
    const relativePath = file.webkitRelativePath;
    let folderId: number | undefined;

    if (relativePath) {
      const parts = relativePath.split('/');
      parts.pop();
      parts.shift();
      const folderPath = parts.join('/');
      folderId = folderIdMap.get(folderPath) ?? currentFolderId.value ?? undefined;
    } else {
      folderId = currentFolderId.value ?? undefined;
    }

    // 更新状态为上传中
    modalRef.updateFileStatus(file.name, file.size, 'uploading', 0);

    try {
      await uploadStore.uploadFile(file, folderId || undefined);

      // 更新状态为完成
      modalRef.updateFileStatus(file.name, file.size, 'completed', 100);
    } catch (error: any) {
      const errorMsg = error?.message || error?.response?.data?.message || '上传失败';
      modalRef.updateFileStatus(file.name, file.size, 'failed', 0, errorMsg);
    }
  }

  modalRef.setUploading(false);
  loadFolderTree();
  onRefresh();
}

// 重试失败文件
async function handleRetryFailed(retryFiles: File[]) {
  const modalRef = folderUploadModalRef.value;
  if (!modalRef) return;

  // 重置失败文件状态
  for (const file of retryFiles) {
    modalRef.updateFileStatus(file.name, file.size, 'pending', 0);
  }

  modalRef.setUploading(true);

  // 重新上传失败文件
  for (const file of retryFiles) {
    const relativePath = file.webkitRelativePath;
    let folderId: number | undefined;

    if (relativePath) {
      const parts = relativePath.split('/');
      parts.pop();
      parts.shift();
      const folderPath = parts.join('/');
      // 尝试从文件夹树中获取文件夹 ID
      const findFolderId = (folders: FileApi.Folder[]): number | undefined => {
        for (const f of folders) {
          if (f.name === folderPath || f.name === parts[parts.length - 1]) {
            return f.id;
          }
          if (f.children) {
            const found = findFolderId(f.children);
            if (found) return found;
          }
        }
        return undefined;
      };
      folderId = findFolderId(folderTree.value) ?? currentFolderId.value ?? undefined;
    } else {
      folderId = currentFolderId.value ?? undefined;
    }

    modalRef.updateFileStatus(file.name, file.size, 'uploading', 0);

    try {
      await uploadStore.uploadFile(file, folderId || undefined);
      modalRef.updateFileStatus(file.name, file.size, 'completed', 100);
    } catch (error: any) {
      const errorMsg = error?.message || error?.response?.data?.message || '上传失败';
      modalRef.updateFileStatus(file.name, file.size, 'failed', 0, errorMsg);
    }
  }

  modalRef.setUploading(false);
  loadFolderTree();
  onRefresh();
}

// ==================== 初始化 ====================

onMounted(() => {
  loadFolderTree();
  loadAvailableTags();
});

// ==================== Tree 数据 ====================

const treeData = computed((): FolderTreeNode[] => {
  const convert = (folders: FileApi.Folder[]): FolderTreeNode[] =>
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

const folderSelectData = computed((): FolderSelectNode[] => {
  const convert = (folders: FileApi.Folder[]): FolderSelectNode[] =>
    folders.map((f) => ({
      value: f.id,
      label: f.name,
      children: f.children ? convert(f.children) : undefined,
    }));
  return [{ value: 0, label: '根目录' }, ...convert(folderTree.value)];
});
</script>

<template>
  <Page auto-content-height>
    <div class="flex size-full">
      <!-- 左侧文件夹树 -->
      <Card class="w-1/6 min-w-[200px] flex flex-col">
        <div class="flex items-center justify-between mb-2 shrink-0">
          <span class="font-medium">文件夹</span>
          <Button type="link" size="small" @click="openNewFolderModal()">
            <Plus class="size-4" />
          </Button>
        </div>
        <div class="flex-1 overflow-y-auto">
          <Tree
            label-field="title"
            value-field="key"
            :tree-data="treeData"
            :default-expanded-level="1"
            :show-icon="false"
            :show-toggle-all="false"
            :model-value="currentFolderId ?? '__all__'"
          >
          <template #node="item">
            <Tooltip :title="item.value.title" :mouse-enter-delay="500">
              <div
                class="flex items-center gap-1 py-1 group cursor-pointer min-w-0"
                @click="handleFolderClick(item.value)"
              >
                <span
                  :class="
                    item.value.type === 'all'
                      ? 'i-ant-design:home-outlined'
                      : item.value.type === 'avatar'
                        ? 'i-ant-design:user-outlined'
                        : 'i-ant-design:folder-outlined'
                  "
                  class="shrink-0"
                ></span>
                <span class="flex-1 truncate min-w-0">{{ item.value.title }}</span>
                <button
                  v-if="item.value.type !== 'all'"
                  type="button"
                  class="opacity-0 group-hover:opacity-100 ml-1 px-1 py-0.5 text-xs rounded hover:bg-gray-200 shrink-0"
                  @click.stop="showFolderMenu(item.value)"
                >
                  ⋯
                </button>
              </div>
            </Tooltip>
          </template>
        </Tree>
        </div>
      </Card>

      <!-- 右侧文件列表 -->
      <div class="w-5/6 ml-4">
        <!-- 筛选工具栏 -->
        <div class="mb-3 flex items-center gap-3 flex-wrap">
          <Upload
            v-if="hasUploadPermission"
            :show-upload-list="false"
            :before-upload="handleUpload"
            :multiple="true"
            :disabled="uploading"
          >
            <Button type="primary" :loading="uploading">
              <Plus class="size-5" />
              上传文件
            </Button>
          </Upload>

          <Button
            v-if="hasUploadPermission"
            :disabled="uploading"
            @click="triggerFolderUpload"
          >
            <span class="i-ant-design:folder-add-outlined mr-1"></span>
            上传文件夹
          </Button>

          <input
            ref="folderInputRef"
            type="file"
            webkitdirectory
            style="display: none"
            @change="handleFolderSelect"
          />

          <Button
            v-if="validFileIds.length > 0 && hasDeletePermission"
            danger
            @click="handleBatchDelete"
          >
            批量删除 ({{ validFileIds.length }})
          </Button>
          <Button
            v-if="validFileIds.length > 0 && hasManagePermission"
            @click="openBatchMoveModal"
          >
            批量移动 ({{ validFileIds.length }})
          </Button>

          <div v-if="hasViewAllPermission" class="ml-auto">
            <Radio.Group
              v-model:value="fileScope"
              button-style="solid"
              @change="handleScopeChange"
            >
              <Radio.Button value="own">我的文件</Radio.Button>
              <Radio.Button value="all">所有文件</Radio.Button>
            </Radio.Group>
          </div>

          <Select
            v-if="availableTags.length > 0"
            v-model:value="selectedTagKeys"
            mode="multiple"
            placeholder="按标签筛选"
            style="min-width: 200px"
            :options="
              availableTags.map((t) => ({
                label: `${t.icon} ${t.name}`,
                value: `${t.key}:${t.value}`,
              }))
            "
            @change="handleTagFilter"
            allow-clear
            :max-tag-count="2"
          />
        </div>

        <!-- VxeTable -->
        <Grid>
          <!-- 文件名列 -->
          <template #name="{ row }">
            <div
              class="flex items-center gap-2 cursor-pointer"
              @click="handlePreview(row)"
            >
              <span
                :class="getFileIcon(row.contentType)"
                class="text-lg text-gray-500"
              ></span>
              <span class="truncate">{{ row.name }}</span>
            </div>
          </template>

          <!-- 状态列 -->
          <template #status="{ row }">
            <!-- 上传任务状态 -->
            <template v-if="row.isUploadTask && row.uploadTask">
              <div
                class="cursor-pointer"
                @click="showUploadDetail(row.uploadTask)"
              >
                <div
                  v-if="row.uploadTask.status === 'uploading'"
                  class="flex items-center gap-2"
                >
                  <Progress
                    :percent="row.uploadTask.progress"
                    :show-info="false"
                    status="active"
                    size="small"
                    style="width: 80px; margin: 0"
                  />
                  <span class="text-xs text-blue-600">{{ row.uploadTask.progress }}%</span>
                </div>
                <div
                  v-else-if="row.uploadTask.status === 'processing'"
                  class="flex items-center gap-1"
                >
                  <Spin size="small" />
                  <span class="text-xs text-yellow-600">处理中...</span>
                </div>
                <div
                  v-else-if="row.uploadTask.status === 'completed'"
                  class="text-xs text-green-600"
                >
                  <span
                    class="i-ant-design:check-circle-outlined mr-1"
                  ></span>上传成功
                </div>
                <div
                  v-else-if="row.uploadTask.status === 'failed'"
                  class="text-xs text-red-600"
                >
                  <span
                    class="i-ant-design:close-circle-outlined mr-1"
                  ></span>上传失败
                </div>
              </div>
            </template>
            <!-- 正常文件状态 -->
            <template v-else>
              <Tag color="green">正常</Tag>
            </template>
          </template>

          <!-- 存储列 -->
          <template #storage="{ row }">
            <Tag
              :color="storageTypeLabels[row.storageType]?.color || 'default'"
            >
              {{
                storageTypeLabels[row.storageType]?.label ||
                row.storageType ||
                '本地'
              }}
            </Tag>
          </template>

          <!-- 标签列 -->
          <template #tags="{ row }">
            <template v-if="row.tags && row.tags.length > 0">
              <Tag
                v-for="tag in row.tags.slice(0, 3)"
                :key="tag.id"
                :color="tag.color"
                class="mr-1 mb-1"
              >
                {{ tag.icon }} {{ tag.name }}
              </Tag>
              <Tooltip
                v-if="row.tags.length > 3"
                :title="row.tags.map((t) => `${t.icon} ${t.name}`).join(', ')"
              >
                <Tag>+{{ row.tags.length - 3 }}</Tag>
              </Tooltip>
            </template>
            <span v-else class="text-gray-400">-</span>
          </template>

          <!-- 操作列 -->
          <template #action="{ row }">
            <!-- 上传任务操作 -->
            <template v-if="row.isUploadTask">
              <Button
                type="link"
                size="small"
                @click="showUploadDetail(row.uploadTask!)"
                >
详情
</Button>
            </template>
            <!-- 正常文件操作 -->
            <template v-else>
              <div class="flex items-center gap-1">
                <Button type="link" size="small" @click="handlePreview(row)">
预览
</Button>
                <Button type="link" size="small" @click="handleDownload(row)">
下载
</Button>
                <Dropdown :trigger="['click']">
                  <Button type="link" size="small">
                    更多 <span class="i-ant-design:down-outlined ml-1"></span>
                  </Button>
                  <template #overlay>
                    <Menu
                      @click="
                        ({ key }: { key: string | number }) =>
                          handleFileAction(String(key), row)
                      "
                    >
                      <MenuItem key="detail">
                        <span
                          class="i-ant-design:info-circle-outlined mr-2"
                        ></span>详情
                      </MenuItem>
                      <MenuItem v-if="hasSharePermission" key="share">
                        <span
                          class="i-ant-design:share-alt-outlined mr-2"
                        ></span>分享
                      </MenuItem>
                      <MenuItem key="tag">
                        <span class="i-ant-design:tags-outlined mr-2"></span>标签
                      </MenuItem>
                      <MenuItem v-if="hasManagePermission" key="move">
                        <span
                          class="i-ant-design:folder-open-outlined mr-2"
                        ></span>移动
                      </MenuItem>
                      <MenuItem
                        v-if="hasDeletePermission"
                        key="delete"
                        class="text-red-500"
                      >
                        <span class="i-ant-design:delete-outlined mr-2"></span>删除
                      </MenuItem>
                    </Menu>
                  </template>
                </Dropdown>
              </div>
            </template>
          </template>
        </Grid>
      </div>
    </div>

    <!-- ==================== 弹窗 ==================== -->

    <!-- 新建文件夹 -->
    <Modal
      v-model:open="newFolderModalVisible"
      title="新建文件夹"
      @ok="handleCreateFolder"
    >
      <Form layout="vertical">
        <FormItem label="名称">
<Input v-model:value="newFolderName" placeholder="输入文件夹名称" />
</FormItem>
      </Form>
    </Modal>

    <!-- 重命名文件夹 -->
    <Modal
      v-model:open="renameFolderModalVisible"
      title="重命名文件夹"
      @ok="handleRenameFolder"
    >
      <Form layout="vertical">
        <FormItem label="新名称">
<Input v-model:value="renameFolderName" placeholder="输入新名称" />
</FormItem>
      </Form>
    </Modal>

    <!-- 删除文件夹 -->
    <Modal
      v-model:open="deleteFolderModalVisible"
      title="删除文件夹"
      @ok="handleDeleteFolder"
    >
      <p>确定删除文件夹 "{{ deleteFolderName }}" 吗？</p>
      <p class="text-red-500">文件夹内的所有文件也将被删除！</p>
    </Modal>

    <!-- 移动文件 -->
    <Modal
      v-model:open="moveFileModalVisible"
      title="移动文件"
      @ok="handleMoveFile"
    >
      <Form layout="vertical">
        <FormItem label="目标文件夹">
<TreeSelect
            v-model:value="moveTargetFolderId"
            :tree-data="folderSelectData"
            placeholder="选择文件夹"
            allow-clear
        />
</FormItem>
      </Form>
    </Modal>

    <!-- 批量移动 -->
    <Modal
      v-model:open="batchMoveModalVisible"
      title="批量移动"
      @ok="handleBatchMove"
    >
      <p class="mb-2">将移动 {{ validFileIds.length }} 个文件</p>
      <Form layout="vertical">
        <FormItem label="目标文件夹">
<TreeSelect
            v-model:value="batchTargetFolderId"
            :tree-data="folderSelectData"
            placeholder="选择文件夹"
            allow-clear
        />
</FormItem>
      </Form>
    </Modal>

    <!-- 分享 -->
    <Modal
      v-model:open="shareModalVisible"
      title="创建分享链接"
      :closable="true"
      :mask-closable="false"
    >
      <template #footer>
        <Button @click="closeShareModal">
{{
          hasShareResult ? '关闭' : '取消'
        }}
</Button>
        <Button
          v-if="!hasShareResult"
          type="primary"
          :loading="shareLoading"
          @click="confirmShare"
          >
确定
</Button>
      </template>
      <div v-if="!hasShareResult">
        <Form layout="vertical">
          <FormItem label="过期时间">
            <Space>
              <InputNumber
                v-model:value="shareExpireHours"
                :min="0"
                style="width: 100px"
              />
              <span>小时（0表示永久有效）</span>
            </Space>
          </FormItem>
          <FormItem label="访问密码（可选）">
            <InputPassword
              v-model:value="sharePassword"
              placeholder="留空表示无需密码"
              style="width: 100%"
            />
          </FormItem>
        </Form>
      </div>
      <div v-else class="p-3 bg-gray-50 rounded">
        <p class="mb-2 font-medium">分享链接：</p>
        <div class="flex gap-2">
          <Input :value="shareFullUrl" readonly class="flex-1" />
          <Button type="primary" @click="copyShareUrl">复制</Button>
        </div>
        <p v-if="sharePassword" class="mt-2 text-sm text-orange-500">密码：{{ sharePassword }}</p>
      </div>
    </Modal>

    <!-- 预览 -->
    <Modal
      v-model:open="previewVisible"
      :title="previewName"
      :footer="null"
      :width="
        previewType === 'video' ? 960 : previewType === 'audio' ? 500 : 800
      "
      :mask-closable="true"
      :keyboard="true"
      :destroy-on-close="true"
      class="preview-modal"
    >
      <div v-if="previewType === 'image' && previewUrl" class="text-center p-6">
        <Image :src="previewUrl" class="max-w-full" style="max-height: 600px" />
      </div>
      <iframe
        v-else-if="previewType === 'pdf' && previewUrl"
        :src="previewUrl"
        style="width: 100%; height: 600px; border: none"
      ></iframe>
      <div
        v-else-if="previewType === 'video' && previewUrl"
        class="video-container"
      >
        <video
          :src="previewUrl"
          controls
          autoplay
          preload="auto"
          playsinline
          style="
            display: block;
            width: 100%;
            max-height: 70vh;
            background: #000;
          "
        ></video>
      </div>
      <div v-else-if="previewType === 'audio' && previewUrl" class="p-6">
        <div class="text-center mb-4">
          <span class="i-ant-design:sound-outlined text-6xl text-blue-500"></span>
        </div>
        <audio
          :src="previewUrl"
          controls
          autoplay
          preload="auto"
          style="width: 100%"
        ></audio>
      </div>
      <div v-else-if="previewVisible" class="py-12 text-center text-gray-500">
        <Spin size="large" />
        <p class="mt-4">加载中...</p>
      </div>
      <div v-else class="py-12 text-center text-gray-500">
        该文件类型不支持预览
      </div>
    </Modal>

    <!-- 文件夹操作 -->
    <Modal v-model:open="folderMenuVisible" title="文件夹操作" :footer="null">
      <div class="flex flex-col gap-2">
        <Button block @click="folderMenuAction('new')">新建子文件夹</Button>
        <Button block @click="folderMenuAction('rename')">重命名</Button>
        <Button block type="primary" @click="folderMenuAction('share')">
分享文件夹
</Button>
        <Button block danger @click="folderMenuAction('delete')">
删除文件夹
</Button>
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
          <Tag v-for="tag in fileTags" :key="tag.id" :color="tag.color">
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
          :options="
            availableTags.map((t) => ({
              label: `${t.icon} ${t.name}`,
              value: t.id,
            }))
          "
        />
      </div>
    </Modal>

    <!-- 文件夹分享 -->
    <Modal
      v-model:open="folderShareModalVisible"
      title="创建文件夹分享链接"
      @ok="confirmFolderShare"
    >
      <Form layout="vertical">
        <FormItem label="文件夹">
<Input :value="folderShareName" readonly />
</FormItem>
        <FormItem label="过期时间">
          <Space>
            <InputNumber
              v-model:value="folderShareExpireHours"
              :min="0"
              style="width: 100px"
            />
            <span>小时（0表示永久有效）</span>
          </Space>
        </FormItem>
        <FormItem label="访问密码（可选）">
          <InputPassword
            v-model:value="folderSharePassword"
            placeholder="留空表示无需密码"
            style="width: 100%"
          />
        </FormItem>
      </Form>
      <div v-if="folderShareResult" class="mt-4 p-3 bg-gray-50 rounded">
        <p class="mb-2 font-medium">分享链接：</p>
        <Input.Group compact>
          <Input :value="folderShareFullUrl" style="width: 280px" readonly />
          <Button type="primary" @click="copyFolderShareUrl">复制</Button>
        </Input.Group>
        <p v-if="folderSharePassword" class="mt-2 text-sm text-orange-500">密码：{{ folderSharePassword }}</p>
      </div>
    </Modal>

    <!-- 上传详情 -->
    <Modal
      v-model:open="uploadDetailVisible"
      title="上传详情"
      :footer="null"
      width="700px"
    >
      <div v-if="uploadDetailTask" class="space-y-4">
        <Descriptions bordered size="small">
          <DescriptionsItem label="文件名">
{{
            uploadDetailTask.fileName
          }}
</DescriptionsItem>
          <DescriptionsItem label="文件大小">
{{
            formatFileSize(uploadDetailTask.fileSize)
          }}
</DescriptionsItem>
          <DescriptionsItem label="总分片数">
{{
            uploadDetailTask.totalParts
          }}
</DescriptionsItem>
          <DescriptionsItem label="已上传">
{{ uploadDetailTask.uploadedParts }} /
            {{ uploadDetailTask.totalParts }}
</DescriptionsItem>
          <DescriptionsItem label="状态">
            <Tag
              :color="
                uploadDetailTask.status === 'completed'
                  ? 'green'
                  : uploadDetailTask.status === 'failed'
                    ? 'red'
                    : 'blue'
              "
            >
              {{
                uploadDetailTask.status === 'uploading'
                  ? '上传中'
                  : uploadDetailTask.status === 'processing'
                    ? '处理中'
                    : uploadDetailTask.status === 'completed'
                      ? '已完成'
                      : '失败'
              }}
            </Tag>
          </DescriptionsItem>
          <DescriptionsItem label="进度">
{{ uploadDetailTask.progress }}%
</DescriptionsItem>
          <DescriptionsItem label="开始时间">
{{
            new Date(uploadDetailTask.startTime).toLocaleString()
          }}
</DescriptionsItem>
          <DescriptionsItem v-if="uploadDetailTask.endTime" label="结束时间">
{{
            new Date(uploadDetailTask.endTime).toLocaleString()
          }}
</DescriptionsItem>
          <DescriptionsItem
            v-if="uploadDetailTask.totalDuration"
            label="总耗时"
            >
{{
              formatDuration(uploadDetailTask.totalDuration)
            }}
</DescriptionsItem>
          <DescriptionsItem
            v-if="uploadDetailTask.errorMessage"
            label="错误信息"
            :span="2"
          >
            <span class="text-red-500">{{
              uploadDetailTask.errorMessage
            }}</span>
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
                <Tag
                  :color="
                    record.status === 'completed'
                      ? 'green'
                      : record.status === 'uploading'
                        ? 'blue'
                        : 'default'
                  "
                >
                  {{
                    record.status === 'completed'
                      ? '已完成'
                      : record.status === 'uploading'
                        ? '上传中'
                        : '待上传'
                  }}
                </Tag>
              </template>
              <template v-if="column.key === 'partStart'">
                {{
                  record.startTime
                    ? new Date(record.startTime).toLocaleTimeString()
                    : '-'
                }}
              </template>
              <template v-if="column.key === 'partDuration'">
                {{ formatDuration(record.duration) }}
              </template>
            </template>
          </Table>
        </div>
      </div>
    </Modal>

    <!-- 文件详情 -->
    <Modal
      v-model:open="fileDetailVisible"
      title="文件详情"
      :footer="null"
      width="600px"
    >
      <div v-if="fileDetailData">
        <Descriptions bordered size="small" :column="2">
          <DescriptionsItem label="文件名" :span="2">
            <div class="flex items-center gap-2">
              <span
                :class="getFileIcon(fileDetailData.contentType)"
                class="text-lg"
              ></span>
              <span class="break-all">{{ fileDetailData.name }}</span>
            </div>
          </DescriptionsItem>
          <DescriptionsItem label="文件大小">
            {{ formatFileSize(fileDetailData.size) }}
          </DescriptionsItem>
          <DescriptionsItem label="文件类型">
            {{ fileDetailData.contentType || '未知' }}
          </DescriptionsItem>
          <DescriptionsItem label="存储类型">
            <Tag
              :color="
                storageTypeLabels[fileDetailData.storageType]?.color ||
                'default'
              "
            >
              {{
                storageTypeLabels[fileDetailData.storageType]?.label ||
                fileDetailData.storageType ||
                '本地'
              }}
            </Tag>
          </DescriptionsItem>
          <DescriptionsItem label="文件ID">
            {{ fileDetailData.id }}
          </DescriptionsItem>
          <DescriptionsItem label="上传者">
            {{ fileDetailData.uploaderName || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="创建时间">
            {{
              fileDetailData.createdAt
                ? new Date(fileDetailData.createdAt).toLocaleString()
                : '-'
            }}
          </DescriptionsItem>
          <DescriptionsItem label="更新时间">
            {{
              fileDetailData.updatedAt
                ? new Date(fileDetailData.updatedAt).toLocaleString()
                : '-'
            }}
          </DescriptionsItem>
          <DescriptionsItem label="标签" :span="2">
            <template
              v-if="fileDetailData.tags && fileDetailData.tags.length > 0"
            >
              <Tag
                v-for="tag in fileDetailData.tags"
                :key="tag.id"
                :color="tag.color"
                class="mr-1 mb-1"
              >
                {{ tag.icon }} {{ tag.name }}
              </Tag>
            </template>
            <span v-else class="text-gray-400">暂无标签</span>
          </DescriptionsItem>
        </Descriptions>
      </div>
    </Modal>

    <!-- 下载进度 -->
    <Modal
      v-model:open="downloadProgressVisible"
      title="文件下载"
      :footer="null"
      :closable="downloadStatus !== 'downloading'"
      :mask-closable="false"
      width="400px"
    >
      <div class="py-4">
        <div class="mb-3 text-sm text-gray-600 truncate">
          {{ downloadFileName }}
        </div>
        <Progress
          :percent="downloadProgress"
          :status="
            downloadStatus === 'error'
              ? 'exception'
              : downloadStatus === 'done'
                ? 'success'
                : 'active'
          "
        />
        <div class="mt-2 text-center text-sm">
          <span v-if="downloadStatus === 'downloading'" class="text-blue-500">下载中... {{ downloadProgress }}%</span>
          <span v-else-if="downloadStatus === 'saving'" class="text-blue-500">正在保存文件...</span>
          <span v-else-if="downloadStatus === 'done'" class="text-green-500">下载完成</span>
          <span v-else-if="downloadStatus === 'error'" class="text-red-500">下载失败</span>
        </div>
      </div>
    </Modal>

    <!-- 文件夹上传进度弹窗 -->
    <FolderUploadModal
      ref="folderUploadModalRef"
      :visible="folderUploadVisible"
      :folder-name="folderUploadName"
      :files="folderUploadFiles"
      @close="folderUploadVisible = false"
      @retry-failed="handleRetryFailed"
    />
  </Page>
</template>

<style scoped>
.video-container {
  position: relative;
  width: 100%;
  overflow: hidden;
  background: #000;
}

.video-container video {
  display: block;
  width: 100%;
  max-height: 70vh;
}

:deep(.ant-card) {
  height: 100%;
}

:deep(.ant-card-body) {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 16px;
  overflow: hidden;
}
</style>

<style>
.preview-modal .ant-modal-body {
  padding: 0 !important;
}

.preview-modal .ant-modal-close {
  top: 8px;
  right: 8px;
  z-index: 10;
  color: #fff;
}

.preview-modal .ant-modal-close:hover {
  color: rgb(255 255 255 / 80%);
}

.preview-modal video::-webkit-media-controls {
  pointer-events: auto !important;
}
</style>
