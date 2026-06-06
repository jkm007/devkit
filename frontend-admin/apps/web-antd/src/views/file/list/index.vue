<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Dropdown,
  Empty,
  Form,
  FormItem,
  Input,
  Menu,
  MenuItem,
  message,
  Modal,
  Popconfirm,
  Space,
  Table,
  Tree,
  TreeSelect,
  Upload,
} from 'ant-design-vue';
import type { TreeProps } from 'ant-design-vue';

import {
  createFolder,
  deleteFile,
  deleteFolder,
  getFolderTree,
  listFiles,
  moveFile,
  renameFolder,
  simpleUpload,
} from '#/api/file';
import type { FileApi } from '#/api/file';
import { $t } from '#/locales';

defineOptions({ name: 'FileList' });

// ==================== 状态 ====================

const loading = ref(false);
const folderTree = ref<FileApi.Folder[]>([]);
const currentFolderId = ref<number | null>(null);
const fileList = ref<FileApi.FileEntry[]>([]);
const totalFiles = ref(0);
const pagination = ref({ current: 1, pageSize: 20 });

// 新建文件夹
const newFolderModalVisible = ref(false);
const newFolderName = ref('');
const newFolderParentId = ref<number | null>(null);

// 重命名文件夹
const renameFolderModalVisible = ref(false);
const renameFolderId = ref<number | null>(null);
const renameFolderName = ref('');

// 移动文件
const moveFileModalVisible = ref(false);
const moveFileId = ref<number | null>(null);
const moveTargetFolderId = ref<number | null>(null);

// 上传状态
const uploadingFiles = ref<Map<string, { progress: number; status: string }>>(new Map());

// 预览
const previewVisible = ref(false);
const previewFile = ref<FileApi.FileEntry | null>(null);
const previewUrl = ref('');

// ==================== 文件类型图标 ====================

function getFileIcon(type: string) {
  if (type.startsWith('image/')) return 'i-ant-design:file-image-outlined';
  if (type.startsWith('video/')) return 'i-ant-design:file-video-outlined';
  if (type.startsWith('audio/')) return 'i-ant-design:sound-outlined';
  if (type.includes('pdf')) return 'i-ant-design:file-pdf-outlined';
  if (type.includes('word') || type.includes('document')) return 'i-ant-design:file-word-outlined';
  if (type.includes('excel') || type.includes('spreadsheet')) return 'i-ant-design:file-excel-outlined';
  if (type.includes('zip') || type.includes('rar') || type.includes('archive')) return 'i-ant-design:file-zip-outlined';
  return 'i-ant-design:file-outlined';
}

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`;
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`;
}

// ==================== 加载文件夹树 ====================

async function loadFolderTree() {
  try {
    const result = await getFolderTree();
    folderTree.value = result || [];
  } catch {
    message.error('加载文件夹树失败');
  }
}

// ==================== 加载文件列表 ====================

async function loadFileList() {
  loading.value = true;
  try {
    const result = await listFiles({
      folderId: currentFolderId.value ?? undefined,
      page: pagination.value.current,
      pageSize: pagination.value.pageSize,
    });
    fileList.value = result?.items || [];
    totalFiles.value = result?.total || 0;
  } catch {
    message.error('加载文件列表失败');
  } finally {
    loading.value = false;
  }
}

// ==================== 文件夹操作 ====================

function handleFolderSelect(selectedKeys: (string | number)[]) {
  currentFolderId.value = selectedKeys[0] as number | null;
  pagination.value.current = 1;
  loadFileList();
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
    newFolderName.value = '';
    loadFolderTree();
  } catch {
    message.error('创建失败');
  }
}

function openNewFolderModal(parentId?: number) {
  newFolderParentId.value = parentId ?? currentFolderId.value;
  newFolderName.value = '';
  newFolderModalVisible.value = true;
}

async function handleRenameFolder() {
  if (!renameFolderName.value.trim() || !renameFolderId.value) {
    message.warning('请输入文件夹名称');
    return;
  }
  try {
    await renameFolder(renameFolderId.value, { name: renameFolderName.value.trim() });
    message.success($t('file.list.renameSuccess'));
    renameFolderModalVisible.value = false;
    loadFolderTree();
  } catch {
    message.error('重命名失败');
  }
}

function openRenameFolderModal(folder: { id: number; name: string }) {
  renameFolderId.value = folder.id;
  renameFolderName.value = folder.name;
  renameFolderModalVisible.value = true;
}

async function handleDeleteFolder(folder: { id: number; name: string }) {
  try {
    await deleteFolder(folder.id);
    message.success($t('file.list.deleteSuccess'));
    loadFolderTree();
    if (currentFolderId.value === folder.id) {
      currentFolderId.value = null;
      loadFileList();
    }
  } catch {
    message.error('删除失败');
  }
}

// ==================== 文件操作 ====================

async function handleMoveFile() {
  if (!moveFileId.value) return;
  try {
    await moveFile({
      fileId: moveFileId.value,
      targetFolderId: moveTargetFolderId.value || undefined,
    });
    message.success($t('file.list.moveSuccess'));
    moveFileModalVisible.value = false;
    loadFileList();
  } catch {
    message.error('移动失败');
  }
}

function openMoveFileModal(file: any) {
  moveFileId.value = file.id;
  moveTargetFolderId.value = currentFolderId.value;
  moveFileModalVisible.value = true;
}

async function handleDeleteFile(file: any) {
  try {
    await deleteFile(file.id);
    message.success($t('file.list.deleteSuccess'));
    loadFileList();
  } catch {
    message.error('删除失败');
  }
}

// ==================== 上传 ====================

async function handleUpload(file: File) {
  const uploadId = `${file.name}-${Date.now()}`;
  uploadingFiles.value.set(uploadId, { progress: 0, status: 'uploading' });

  try {
    const result = await simpleUpload(file);
    uploadingFiles.value.set(uploadId, { progress: 100, status: 'done' });
    message.success($t('file.list.uploadSuccess'));
    loadFileList();
    return result;
  } catch (error) {
    uploadingFiles.value.set(uploadId, { progress: 0, status: 'error' });
    message.error($t('file.list.uploadError'));
    throw error;
  } finally {
    setTimeout(() => {
      uploadingFiles.value.delete(uploadId);
    }, 1000);
  }
}

// ==================== 预览 ====================

async function handlePreview(file: any) {
  previewFile.value = file;

  // 图片直接预览
  if (file.contentType.startsWith('image/')) {
    previewUrl.value = file.previewUrl || file.thumbnailUrl || '';
    previewVisible.value = true;
    return;
  }

  // 视频调用 stream API
  if (file.contentType.startsWith('video/')) {
    message.info('视频播放功能正在开发中');
    return;
  }

  message.info($t('file.list.noPreview'));
}

// ==================== 初始化 ====================

onMounted(() => {
  loadFolderTree();
  loadFileList();
});

// ==================== 表格列 ====================

const columns: any[] = [
  {
    title: $t('file.list.name'),
    dataIndex: 'name',
    key: 'name',
    ellipsis: true,
  },
  {
    title: $t('file.list.size'),
    dataIndex: 'fileSize',
    key: 'size',
    width: 120,
    customRender: ({ text }: { text: number }) => formatFileSize(text),
  },
  {
    title: $t('file.list.type'),
    dataIndex: 'contentType',
    key: 'type',
    width: 150,
  },
  {
    title: $t('file.list.createTime'),
    dataIndex: 'createdAt',
    key: 'createTime',
    width: 180,
  },
  {
    title: $t('file.list.operation'),
    key: 'operation',
    width: 200,
    fixed: 'right',
  },
];

// 文件夹树数据转换
const treeData = computed<TreeProps['treeData']>(() => {
  const convertTree = (folders: FileApi.Folder[]): TreeProps['treeData'] => {
    return folders.map((folder) => ({
      key: folder.id,
      title: folder.name,
      children: folder.children ? convertTree(folder.children) : undefined,
    }));
  };
  return convertTree(folderTree.value);
});

// 文件夹选择器数据（用于移动）
const folderSelectData = computed(() => {
  const convertTree = (folders: FileApi.Folder[]): any[] => {
    return folders.map((folder) => ({
      value: folder.id,
      label: folder.name,
      children: folder.children ? convertTree(folder.children) : undefined,
    }));
  };
  return [{ value: null, label: $t('file.list.rootFolder') }, ...convertTree(folderTree.value)];
});

// 分页改变
function handleTableChange(pag: any) {
  pagination.value = pag;
  loadFileList();
}
</script>

<template>
  <Page title="">
    <div class="file-manager flex gap-4">
      <!-- 左侧文件夹树 -->
      <div class="folder-tree w-64 shrink-0 border rounded-lg p-4 bg-background">
        <div class="flex items-center justify-between mb-4">
          <span class="font-medium">{{ $t('file.list.folder') }}</span>
          <Button type="link" size="small" @click="openNewFolderModal()">
            <span class="i-ant-design:folder-add-outlined" />
            {{ $t('file.list.newFolder') }}
          </Button>
        </div>

        <Tree
          :tree-data="treeData"
          :selected-keys="currentFolderId ? [currentFolderId] : []"
          :show-icon="true"
          default-expand-all
          @select="handleFolderSelect"
        >
          <template #title="{ title, key }">
            <div class="flex items-center gap-1 group">
              <span>{{ title }}</span>
              <Dropdown class="opacity-0 group-hover:opacity-100">
                <span class="i-ant-design:more-outlined cursor-pointer" />
                <template #overlay>
                  <Menu>
                    <MenuItem @click="openNewFolderModal(key as number)">
                      <span class="i-ant-design:folder-add-outlined mr-1" />
                      {{ $t('file.list.newFolder') }}
                    </MenuItem>
                    <MenuItem @click="openRenameFolderModal({ id: key as number, name: title as string })">
                      <span class="i-ant-design:edit-outlined mr-1" />
                      {{ $t('file.list.rename') }}
                    </MenuItem>
                    <MenuItem>
                      <Popconfirm
                        :title="$t('file.list.deleteConfirm', { name: title })"
                        @confirm="handleDeleteFolder({ id: key as number, name: title as string })"
                      >
                        <span class="i-ant-design:delete-outlined mr-1 text-red-500" />
                        {{ $t('file.list.delete') }}
                      </Popconfirm>
                    </MenuItem>
                  </Menu>
                </template>
              </Dropdown>
            </div>
          </template>
        </Tree>
      </div>

      <!-- 右侧文件列表 -->
      <div class="file-list flex-1 border rounded-lg p-4 bg-background">
        <!-- 工具栏 -->
        <div class="flex items-center justify-between mb-4">
          <Space>
            <Upload
              :show-upload-list="false"
              :before-upload="(file) => { handleUpload(file); return false; }"
              :multiple="true"
            >
              <Button type="primary">
                <span class="i-ant-design:upload-outlined mr-1" />
                {{ $t('file.list.upload') }}
              </Button>
            </Upload>
            <Button @click="openNewFolderModal()">
              <span class="i-ant-design:folder-add-outlined mr-1" />
              {{ $t('file.list.newFolder') }}
            </Button>
          </Space>

          <span class="text-muted-foreground text-sm">
            {{ $t('file.list.fileCount', { count: totalFiles }) }}
          </span>
        </div>

        <!-- 文件表格 -->
        <Table
          :columns="columns"
          :data-source="fileList"
          :loading="loading"
          :pagination="{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: totalFiles,
            showSizeChanger: true,
            showTotal: (total: number) => $t('file.list.fileCount', { count: total }),
          }"
          :scroll="{ x: 800 }"
          row-key="id"
          @change="handleTableChange"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'name'">
              <div class="flex items-center gap-2">
                <span :class="getFileIcon(record.contentType)" class="text-lg" />
                <span>{{ record.name }}</span>
              </div>
            </template>
            <template v-if="column.key === 'operation'">
              <Space>
                <Button type="link" size="small" @click="handlePreview(record)">
                  <span class="i-ant-design:eye-outlined" />
                </Button>
                <Button type="link" size="small" @click="openMoveFileModal(record)">
                  <span class="i-ant-design:folder-outlined" />
                </Button>
                <Popconfirm
                  :title="$t('file.list.deleteConfirm', { name: record.name })"
                  @confirm="handleDeleteFile(record)"
                >
                  <Button type="link" size="small" danger>
                    <span class="i-ant-design:delete-outlined" />
                  </Button>
                </Popconfirm>
              </Space>
            </template>
          </template>
        </Table>

        <!-- 空状态 -->
        <Empty v-if="!loading && fileList.length === 0" :description="$t('file.list.noFiles')" />
      </div>
    </div>

    <!-- 新建文件夹 Modal -->
    <Modal
      v-model:open="newFolderModalVisible"
      :title="$t('file.list.newFolder')"
      @ok="handleCreateFolder"
    >
      <Form layout="vertical">
        <FormItem :label="$t('file.list.folderName')">
          <Input
            v-model:value="newFolderName"
            :placeholder="$t('file.list.folderNamePlaceholder')"
          />
        </FormItem>
        <FormItem :label="$t('file.list.targetFolder')">
          <TreeSelect
            v-model:value="newFolderParentId"
            :tree-data="folderSelectData"
            :placeholder="`${$t('file.list.rootFolder')}`"
            allow-clear
          />
        </FormItem>
      </Form>
    </Modal>

    <!-- 重命名文件夹 Modal -->
    <Modal
      v-model:open="renameFolderModalVisible"
      :title="$t('file.list.rename')"
      @ok="handleRenameFolder"
    >
      <Form layout="vertical">
        <FormItem :label="$t('file.list.folderName')">
          <Input
            v-model:value="renameFolderName"
            :placeholder="$t('file.list.folderNamePlaceholder')"
          />
        </FormItem>
      </Form>
    </Modal>

    <!-- 移动文件 Modal -->
    <Modal
      v-model:open="moveFileModalVisible"
      :title="$t('file.list.move')"
      @ok="handleMoveFile"
    >
      <Form layout="vertical">
        <FormItem :label="$t('file.list.targetFolder')">
          <TreeSelect
            v-model:value="moveTargetFolderId"
            :tree-data="folderSelectData"
            :placeholder="`${$t('file.list.rootFolder')}`"
            allow-clear
          />
        </FormItem>
      </Form>
    </Modal>

    <!-- 图片预览 Modal -->
    <Modal
      v-model:open="previewVisible"
      :title="$t('file.list.previewImage')"
      :footer="null"
      :width="800"
    >
      <img v-if="previewUrl" :src="previewUrl" class="w-full" alt="preview" />
    </Modal>
  </Page>
</template>

<style scoped>
.file-manager {
  height: calc(100vh - 200px);
}

.folder-tree {
  max-height: calc(100vh - 240px);
  overflow-y: auto;
}

.file-list {
  max-height: calc(100vh - 240px);
  overflow-y: auto;
}
</style>