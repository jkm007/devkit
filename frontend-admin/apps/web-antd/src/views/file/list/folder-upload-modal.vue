<script lang="ts" setup>
import { computed, defineComponent, h, ref, watch } from 'vue';
import type { DefineComponent, VNode } from 'vue';

import {
  Button,
  Modal,
  Progress,
  Tag,
  Tooltip,
} from 'ant-design-vue';

/** 文件夹节点 */
interface FolderNode {
  key: string;
  name: string;
  level: number;
  totalFiles: number;
  completedFiles: number;
  failedFiles: number;
  uploadingFiles: number;
  children: FolderNode[];
  isExpanded: boolean;
}

/** 文件上传状态 */
interface FileUploadStatus {
  id: string;
  name: string;
  folderPath: string;
  size: number;
  status: 'completed' | 'failed' | 'pending' | 'uploading';
  progress: number;
  error?: string;
}

const props = defineProps<{
  files: File[];
  folderName: string;
  visible: boolean;
}>();

const emit = defineEmits<{
  close: [];
  retryFailed: [files: File[]];
  startUpload: [];
}>();

// 递归文件夹树节点组件
const FolderTreeNode = defineComponent({
  name: 'FolderTreeNode',
  props: {
    node: {
      type: Object as () => FolderNode,
      required: true,
    },
  },
  emits: ['toggle'],
  setup(nodeProps, { emit: nodeEmit }) {
    const getFolderStatusIcon = (folderNode: FolderNode): string => {
      if (
        folderNode.failedFiles > 0 &&
        folderNode.completedFiles + folderNode.failedFiles === folderNode.totalFiles
      ) {
        return '⚠️';
      }
      if (folderNode.completedFiles === folderNode.totalFiles && folderNode.totalFiles > 0) {
        return '✅';
      }
      if (folderNode.uploadingFiles > 0) {
        return '⏳';
      }
      return '📁';
    };

    return (): VNode => {
      const node = nodeProps.node;
      const children: VNode[] =
        node.isExpanded && node.children.length > 0
          ? node.children.map((child) =>
              h(FolderTreeNode, {
                node: child,
                onToggle: (key: string) => nodeEmit('toggle', key),
              }),
            )
          : [];

      return h('div', { class: 'tree-node' }, [
        h(
          'div',
          {
            class: 'folder-row',
            style: { paddingLeft: `${node.level * 16 + 8}px` },
            onClick: () => nodeEmit('toggle', node.key),
          },
          [
            h(
              'span',
              { class: 'expand-icon' },
              node.children.length > 0
                ? (node.isExpanded ? '▼' : '▶')
                : '  ',
            ),
            h('span', { class: 'folder-icon' }, getFolderStatusIcon(node)),
            h('span', { class: 'folder-name' }, node.name),
            h(
              'span',
              { class: 'folder-count' },
              `(${node.completedFiles + node.failedFiles}/${node.totalFiles})`,
            ),
            node.uploadingFiles > 0
              ? h(
                  'span',
                  { class: 'uploading-badge' },
                  `⏳ ${node.uploadingFiles}`,
                )
              : null,
          ],
        ),
        ...children,
      ]);
    };
  },
}) as DefineComponent<{ node: FolderNode }>;

// 状态
const fileStatusMap = ref<Map<string, FileUploadStatus>>(new Map());
const folderTree = ref<FolderNode[]>([]);
const isUploading = ref(false);
const uploadProgress = ref(0);
const showFailedOnly = ref(false);
const isMinimized = ref(false);

// 统计
const stats = computed(() => {
  const statuses = [...fileStatusMap.value.values()];
  return {
    total: statuses.length,
    completed: statuses.filter((s) => s.status === 'completed').length,
    uploading: statuses.filter((s) => s.status === 'uploading').length,
    failed: statuses.filter((s) => s.status === 'failed').length,
    pending: statuses.filter((s) => s.status === 'pending').length,
  };
});

// 总进度百分比
const totalProgress = computed(() => {
  if (stats.value.total === 0) return 0;
  return Math.round(
    ((stats.value.completed + stats.value.failed) / stats.value.total) * 100,
  );
});

// 失败文件列表
const failedFiles = computed(() => {
  return [...fileStatusMap.value.values()].filter(
    (s) => s.status === 'failed',
  );
});

// 是否全部完成
const isAllDone = computed(() => {
  return (
    stats.value.total > 0 &&
    stats.value.completed + stats.value.failed === stats.value.total
  );
});

// 解析文件夹结构
function parseFolderStructure(files: File[]) {
  const folderMap = new Map<string, FolderNode>();
  const rootFolders = new Map<string, FolderNode>();

  // 初始化文件状态
  fileStatusMap.value.clear();

  for (const file of files) {
    const relativePath = file.webkitRelativePath || file.name;
    const parts = relativePath.split('/');

    // 创建文件状态
    const fileId = `${file.name}_${file.size}_${file.lastModified}`;
    fileStatusMap.value.set(fileId, {
      id: fileId,
      name: file.name,
      folderPath: parts.length > 1 ? parts.slice(0, -1).join('/') : '',
      size: file.size,
      status: 'pending',
      progress: 0,
    });

    // 解析文件夹路径
    if (parts.length > 1) {
      const folderParts = parts.slice(0, -1);

      // 创建每一级文件夹
      for (let i = 0; i < folderParts.length; i++) {
        const folderPath = folderParts.slice(0, i + 1).join('/');
        const folderName = folderParts[i] || '';
        const parentPath = i > 0 ? folderParts.slice(0, i).join('/') : '';

        if (!folderMap.has(folderPath)) {
          const node: FolderNode = {
            key: folderPath,
            name: folderName,
            level: i,
            totalFiles: 0,
            completedFiles: 0,
            failedFiles: 0,
            uploadingFiles: 0,
            children: [],
            isExpanded: i < 1, // 默认只展开第一级
          };
          folderMap.set(folderPath, node);

          // 添加到父文件夹
          if (parentPath && folderMap.has(parentPath)) {
            folderMap.get(parentPath)!.children.push(node);
          } else if (i === 0) {
            rootFolders.set(folderPath, node);
          }
        }

        // 更新文件计数
        folderMap.get(folderPath)!.totalFiles++;
      }
    }
  }

  folderTree.value = [...rootFolders.values()];
}

// 更新文件状态
function updateFileStatus(
  fileName: string,
  fileSize: number,
  status: FileUploadStatus['status'],
  progress?: number,
  error?: string,
) {
  // 查找匹配的文件状态
  for (const [, fileStatus] of fileStatusMap.value.entries()) {
    if (fileStatus.name === fileName && fileStatus.size === fileSize) {
      fileStatus.status = status;
      if (progress !== undefined) fileStatus.progress = progress;
      if (error) fileStatus.error = error;

      // 更新文件夹统计 - 递归更新所有父文件夹
      updateFolderStats(fileStatus.folderPath, status);
      break;
    }
  }
}

// 更新文件夹统计 - 递归更新所有父文件夹
function updateFolderStats(
  folderPath: string,
  status: FileUploadStatus['status'],
) {
  const updateNode = (nodes: FolderNode[]) => {
    for (const node of nodes) {
      // 匹配当前文件夹或其父文件夹
      if (folderPath === node.key || folderPath.startsWith(`${node.key}/`)) {
        switch (status) {
          case 'completed': {
            node.completedFiles++;
            break;
          }
          case 'failed': {
            node.failedFiles++;
            break;
          }
          case 'uploading': {
            node.uploadingFiles++;
            break;
          }
          default: {
            break;
          }
        }
        // 递归更新子文件夹
        updateNode(node.children);
      }
    }
  };
  updateNode(folderTree.value);
}

// 切换文件夹展开/折叠
function toggleFolder(key: string) {
  const toggleNode = (nodes: FolderNode[]) => {
    for (const node of nodes) {
      if (node.key === key) {
        node.isExpanded = !node.isExpanded;
        return true;
      }
      if (toggleNode(node.children)) return true;
    }
    return false;
  };
  toggleNode(folderTree.value);
}

// 展开所有文件夹
function expandAll() {
  const expandNode = (nodes: FolderNode[]) => {
    for (const node of nodes) {
      node.isExpanded = true;
      expandNode(node.children);
    }
  };
  expandNode(folderTree.value);
}

// 折叠所有文件夹
function collapseAll() {
  const collapseNode = (nodes: FolderNode[]) => {
    for (const node of nodes) {
      node.isExpanded = false;
      collapseNode(node.children);
    }
  };
  collapseNode(folderTree.value);
}

// 重试失败文件
function handleRetryFailed() {
  const failedFileNames = new Set(failedFiles.value.map((f) => f.name));
  emit(
    'retryFailed',
    props.files.filter((f) => failedFileNames.has(f.name)),
  );
}

// 最小化
function handleMinimize() {
  isMinimized.value = true;
  emit('close');
}

// 恢复弹窗
function handleRestore() {
  isMinimized.value = false;
}

// 关闭弹窗
function handleClose() {
  if (isUploading.value && !isAllDone.value) {
    // 上传中关闭，最小化到后台
    handleMinimize();
  } else {
    isMinimized.value = false;
    emit('close');
  }
}

// 监听文件变化
watch(
  () => props.files,
  (newFiles) => {
    if (newFiles.length > 0) {
      parseFolderStructure(newFiles);
    }
  },
  { immediate: true },
);

// 暴露方法给父组件
defineExpose({
  updateFileStatus,
  setUploading: (value: boolean) => {
    isUploading.value = value;
  },
  setProgress: (value: number) => {
    uploadProgress.value = value;
  },
  restore: handleRestore,
});
</script>

<template>
  <!-- 最小化浮动按钮 -->
  <div
    v-if="isMinimized && !visible"
    class="minimized-float"
    @click="handleRestore"
  >
    <div class="minimized-content">
      <div class="minimized-header">
        <span class="minimized-icon">📁</span>
        <span class="minimized-name">{{ folderName }}</span>
        <span class="minimized-progress">{{ totalProgress }}%</span>
      </div>
      <Progress
        :percent="totalProgress"
        :show-info="false"
        :stroke-width="4"
        :status="stats.failed > 0 ? 'exception' : isAllDone ? 'success' : 'active'"
      />
      <div class="minimized-stats">
        <span>✅ {{ stats.completed }}/{{ stats.total }}</span>
        <span v-if="stats.uploading > 0" class="uploading-count">⏳ {{ stats.uploading }}</span>
        <span v-if="stats.failed > 0" class="failed-count">❌ {{ stats.failed }}</span>
      </div>
    </div>
  </div>

  <!-- 主弹窗 -->
  <Modal
    :open="visible && !isMinimized"
    title="文件夹上传"
    :width="680"
    :footer="null"
    :mask-closable="false"
    @cancel="handleClose"
  >
    <div class="folder-upload-modal">
      <!-- 文件夹名称 -->
      <div class="folder-header">
        <span class="folder-icon">📁</span>
        <span class="folder-name">{{ folderName }}</span>
        <Tag v-if="isAllDone" color="success">已完成</Tag>
        <Tag v-else-if="isUploading" color="processing">上传中</Tag>
      </div>

      <!-- 统计卡片 -->
      <div class="stats-row">
        <div class="stat-item">
          <div class="stat-value">{{ stats.total }}</div>
          <div class="stat-label">总文件</div>
        </div>
        <div class="stat-item uploading">
          <div class="stat-value">{{ stats.uploading }}</div>
          <div class="stat-label">上传中</div>
        </div>
        <div class="stat-item completed">
          <div class="stat-value">{{ stats.completed }}</div>
          <div class="stat-label">已完成</div>
        </div>
        <div class="stat-item failed">
          <div class="stat-value">{{ stats.failed }}</div>
          <div class="stat-label">失败</div>
        </div>
      </div>

      <!-- 总进度条 -->
      <div class="progress-section">
        <div class="progress-header">
          <span>上传进度</span>
          <span>{{ totalProgress }}% ({{ stats.completed + stats.failed }}/{{ stats.total }})</span>
        </div>
        <Progress
          :percent="totalProgress"
          :status="stats.failed > 0 ? 'exception' : isAllDone ? 'success' : 'active'"
          :show-info="false"
        />
      </div>

      <!-- 文件夹树 -->
      <div class="folder-tree-section">
        <div class="section-header">
          <span>📂 文件夹结构</span>
          <div class="section-actions">
            <Button type="link" size="small" @click="expandAll">
              展开全部
            </Button>
            <Button type="link" size="small" @click="collapseAll">
              折叠全部
            </Button>
            <Button
              v-if="stats.failed > 0"
              type="link"
              size="small"
              @click="showFailedOnly = !showFailedOnly"
            >
              {{ showFailedOnly ? '显示全部' : '只看失败' }}
            </Button>
          </div>
        </div>
        <div class="folder-tree-scroll">
          <div v-if="showFailedOnly" class="failed-list">
            <div
              v-for="file in failedFiles"
              :key="file.id"
              class="failed-item"
            >
              <span class="file-icon">❌</span>
              <span class="file-name">{{ file.name }}</span>
              <Tooltip :title="file.error">
                <Tag color="error" class="error-tag">
                  {{ file.error?.substring(0, 20) || '失败' }}
                </Tag>
              </Tooltip>
            </div>
          </div>
          <div v-else class="folder-tree">
            <FolderTreeNode
              v-for="folderNode in folderTree"
              :key="folderNode.key"
              :node="folderNode"
              @toggle="toggleFolder"
            />
          </div>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="modal-footer">
        <Button
          v-if="isUploading && !isAllDone"
          @click="handleMinimize"
        >
          最小化
        </Button>
        <Button
          v-if="stats.failed > 0"
          type="primary"
          danger
          @click="handleRetryFailed"
        >
          重试失败 ({{ stats.failed }})
        </Button>
        <Button @click="handleClose">
          {{ isAllDone ? '关闭' : '取消' }}
        </Button>
      </div>
    </div>
  </Modal>
</template>

<style scoped>
.folder-upload-modal {
  padding: 0 8px;
}

.folder-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  font-size: 16px;
}

.folder-icon {
  font-size: 20px;
}

.folder-name {
  font-weight: 600;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.stats-row {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.stat-item {
  flex: 1;
  text-align: center;
  padding: 12px 8px;
  background: #f5f5f5;
  border-radius: 8px;
}

.stat-item.uploading {
  background: #e6f7ff;
}

.stat-item.completed {
  background: #f6ffed;
}

.stat-item.failed {
  background: #fff2f0;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #666;
}

.progress-section {
  margin-bottom: 16px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 13px;
  color: #666;
}

.folder-tree-section {
  margin-bottom: 16px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
}

.section-actions {
  display: flex;
  gap: 4px;
}

.folder-tree-scroll {
  max-height: 320px;
  overflow-y: auto;
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  padding: 8px;
}

.folder-tree,
.failed-list {
  font-size: 13px;
}

:deep(.tree-node) {
  /* 递归节点 */
}

:deep(.folder-row) {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  cursor: pointer;
  border-radius: 4px;
  transition: background 0.2s;
}

:deep(.folder-row:hover) {
  background: #f5f5f5;
}

:deep(.expand-icon) {
  width: 16px;
  text-align: center;
  font-size: 10px;
  color: #999;
  flex-shrink: 0;
}

:deep(.folder-icon) {
  font-size: 14px;
  flex-shrink: 0;
}

:deep(.folder-name) {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.folder-count) {
  color: #999;
  font-size: 12px;
  flex-shrink: 0;
}

:deep(.uploading-badge) {
  font-size: 11px;
  color: #1890ff;
  flex-shrink: 0;
}

.failed-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border-bottom: 1px solid #f0f0f0;
}

.failed-item:last-child {
  border-bottom: none;
}

.file-icon {
  font-size: 14px;
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.error-tag {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

/* 最小化浮动按钮 */
.minimized-float {
  position: fixed;
  bottom: 24px;
  right: 24px;
  z-index: 1000;
  width: 280px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  cursor: pointer;
  transition: all 0.3s;
  overflow: hidden;
}

.minimized-float:hover {
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.2);
  transform: translateY(-2px);
}

.minimized-content {
  padding: 12px 16px;
}

.minimized-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.minimized-icon {
  font-size: 18px;
}

.minimized-name {
  font-weight: 600;
  font-size: 13px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.minimized-progress {
  font-size: 14px;
  font-weight: 600;
  color: #1890ff;
}

.minimized-stats {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #999;
  margin-top: 6px;
}

.uploading-count {
  color: #1890ff;
}

.failed-count {
  color: #ff4d4f;
}
</style>
