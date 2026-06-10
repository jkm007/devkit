<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';
import { useAccessStore } from '@vben/stores';

import {
  Button,
  Card,
  message,
  Modal,
  Progress,
  Radio,
  Space,
  Statistic,
  Table,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import {
  batchPermanentDeleteFiles,
  batchRestoreFiles,
  emptyRecycleBin,
  getRecycleBinCount,
  getRecycleBinList,
  permanentDeleteFile,
  restoreFile,
} from '#/api/file';
import type { RecycleBinItem } from '#/api/file';
import { getFileIcon, formatFileSize } from '#/utils/file-utils';

defineOptions({ name: 'FileRecycle' });

const accessStore = useAccessStore();

// ==================== 权限 ====================

const permissions = computed(() => accessStore.accessCodes || []);
const hasViewAllPermission = computed(() => permissions.value.includes('file:view:all'));
const hasRestorePermission = computed(() => permissions.value.includes('file:recycle:restore'));
const hasDeletePermission = computed(() => permissions.value.includes('file:recycle:delete'));
const hasEmptyPermission = computed(() => permissions.value.includes('file:recycle:empty'));

// ==================== 状态 ====================

const loading = ref(false);
const dataSource = ref<RecycleBinItem[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const selectedRowKeys = ref<number[]>([]);
const fileScope = ref<'own' | 'all'>('own');
const recycleCount = ref(0);

// ==================== 工具函数 ====================

// formatFileSize, getFileIcon are imported from '#/utils/file-utils'

function formatDate(date: string | undefined) {
  if (!date) return '-';
  return new Date(date).toLocaleString('zh-CN');
}

function getDaysColor(days: number) {
  if (days <= 1) return 'red';
  if (days <= 3) return 'orange';
  return 'green';
}

// ==================== 表格列定义 ====================

const columns = [
  {
    title: '',
    dataIndex: 'id',
    width: 50,
  },
  {
    title: '文件名',
    dataIndex: 'name',
    width: 250,
  },
  {
    title: '大小',
    dataIndex: 'size',
    width: 100,
    customRender: ({ text }: { text: number }) => formatFileSize(text),
  },
  {
    title: '删除时间',
    dataIndex: 'deletedAt',
    width: 180,
    customRender: ({ text }: { text: string }) => formatDate(text),
  },
  {
    title: '剩余天数',
    dataIndex: 'daysRemaining',
    width: 120,
    customRender: ({ text }: { text: number }) => ({
      children: `${text} 天`,
      props: { style: { color: getDaysColor(text === 0 ? 0 : text) } },
    }),
  },
  {
    title: '操作',
    dataIndex: 'action',
    width: 200,
    fixed: 'right' as const,
  },
];

// ==================== 数据加载 ====================

async function loadData() {
  loading.value = true;
  try {
    const result = await getRecycleBinList({
      page: page.value,
      pageSize: pageSize.value,
      scope: fileScope.value,
    });
    dataSource.value = result?.items || [];
    total.value = result?.total || 0;
  } catch {
    message.error('加载回收站数据失败');
  } finally {
    loading.value = false;
  }
}

async function loadCount() {
  try {
    const result = await getRecycleBinCount();
    recycleCount.value = result?.count || 0;
  } catch {
    // 静默失败
  }
}

function handlePageChange(newPage: number, newPageSize: number) {
  page.value = newPage;
  pageSize.value = newPageSize;
  loadData();
}

function handleScopeChange() {
  page.value = 1;
  loadData();
}

// ==================== 操作 ====================

async function handleRestore(record: RecycleBinItem) {
  try {
    await restoreFile(record.id);
    message.success(`已恢复 ${record.name}`);
    loadData();
    loadCount();
  } catch {
    message.error('恢复失败');
  }
}

async function handlePermanentDelete(record: RecycleBinItem) {
  Modal.confirm({
    title: '永久删除',
    content: `确定永久删除 "${record.name}" 吗？此操作不可恢复！`,
    okText: '确定删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await permanentDeleteFile(record.id);
        message.success(`已永久删除 ${record.name}`);
        loadData();
        loadCount();
      } catch {
        message.error('删除失败');
      }
    },
  });
}

async function handleBatchRestore() {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择文件');
    return;
  }

  try {
    const result = await batchRestoreFiles(selectedRowKeys.value);
    if (result.errors?.length > 0) {
      message.warning(`已恢复 ${result.restored} 个文件，${result.errors.length} 个失败`);
    } else {
      message.success(`已恢复 ${result.restored} 个文件`);
    }
    selectedRowKeys.value = [];
    loadData();
    loadCount();
  } catch {
    message.error('批量恢复失败');
  }
}

async function handleBatchPermanentDelete() {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择文件');
    return;
  }

  Modal.confirm({
    title: '批量永久删除',
    content: `确定永久删除选中的 ${selectedRowKeys.value.length} 个文件吗？此操作不可恢复！`,
    okText: '确定删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        const result = await batchPermanentDeleteFiles(selectedRowKeys.value);
        if (result.errors?.length > 0) {
          message.warning(`已删除 ${result.deleted} 个文件，${result.errors.length} 个失败`);
        } else {
          message.success(`已删除 ${result.deleted} 个文件`);
        }
        selectedRowKeys.value = [];
        loadData();
        loadCount();
      } catch {
        message.error('批量删除失败');
      }
    },
  });
}

async function handleEmptyRecycleBin() {
  Modal.confirm({
    title: '清空回收站',
    content: '确定清空回收站吗？所有文件将被永久删除，此操作不可恢复！',
    okText: '确定清空',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await emptyRecycleBin();
        message.success('回收站已清空');
        selectedRowKeys.value = [];
        loadData();
        loadCount();
      } catch {
        message.error('清空失败');
      }
    },
  });
}

// ==================== 初始化 ====================

onMounted(() => {
  loadData();
  loadCount();
});
</script>

<template>
  <Page auto-content-height>
    <div class="p-4">
      <!-- 统计卡片 -->
      <Card class="mb-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-8">
            <Statistic title="回收站文件数" :value="recycleCount" suffix="个" />
            <Statistic title="保留期限" :value="7" suffix="天" />
          </div>
          <Space>
            <Button v-if="hasEmptyPermission && recycleCount > 0" danger @click="handleEmptyRecycleBin">
              清空回收站
            </Button>
          </Space>
        </div>
      </Card>

      <!-- 工具栏 -->
      <div class="mb-4 flex items-center justify-between">
        <Space>
          <Button
            v-if="hasRestorePermission && selectedRowKeys.length > 0"
            type="primary"
            @click="handleBatchRestore"
          >
            批量恢复 ({{ selectedRowKeys.length }})
          </Button>
          <Button
            v-if="hasDeletePermission && selectedRowKeys.length > 0"
            danger
            @click="handleBatchPermanentDelete"
          >
            批量永久删除 ({{ selectedRowKeys.length }})
          </Button>
        </Space>

        <div v-if="hasViewAllPermission">
          <Radio.Group v-model:value="fileScope" button-style="solid" @change="handleScopeChange">
            <Radio.Button value="own">我的文件</Radio.Button>
            <Radio.Button value="all">所有文件</Radio.Button>
          </Radio.Group>
        </div>
      </div>

      <!-- 表格 -->
      <Table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="{
          current: page,
          pageSize: pageSize,
          total: total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (t: number) => `共 ${t} 条`,
          onChange: handlePageChange,
        }"
        :row-selection="{
          selectedRowKeys: selectedRowKeys,
          onChange: (keys: any[]) => selectedRowKeys = keys as number[],
        }"
        row-key="id"
        :scroll="{ x: 1000 }"
      >
        <template #bodyCell="{ column, record }">
          <!-- 文件名列 -->
          <template v-if="column.dataIndex === 'name'">
            <div class="flex items-center gap-2">
              <span :class="getFileIcon((record as any).contentType)" class="text-lg text-gray-500" />
              <Tooltip :title="(record as any).name">
                <span class="truncate max-w-[200px]">{{ (record as any).name }}</span>
              </Tooltip>
            </div>
          </template>

          <!-- 剩余天数列 -->
          <template v-if="column.dataIndex === 'daysRemaining'">
            <div class="flex items-center gap-2">
              <Progress
                :percent="((record as any).daysRemaining / 7) * 100"
                :show-info="false"
                :stroke-color="getDaysColor((record as any).daysRemaining)"
                size="small"
                style="width: 60px;"
              />
              <Tag :color="getDaysColor((record as any).daysRemaining)">
                {{ (record as any).daysRemaining }} 天
              </Tag>
            </div>
          </template>

          <!-- 操作列 -->
          <template v-if="column.dataIndex === 'action'">
            <Space>
              <Button
                v-if="hasRestorePermission"
                type="link"
                size="small"
                @click="handleRestore(record as any)"
              >
                恢复
              </Button>
              <Button
                v-if="hasDeletePermission"
                type="link"
                size="small"
                danger
                @click="handlePermanentDelete(record as any)"
              >
                永久删除
              </Button>
            </Space>
          </template>
        </template>
      </Table>
    </div>
  </Page>
</template>
