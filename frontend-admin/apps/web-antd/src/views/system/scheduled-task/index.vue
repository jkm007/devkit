<script lang="ts" setup>
import { onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Form,
  FormItem,
  Input,
  InputNumber,
  message,
  Modal,
  Space,
  Switch,
  Table,
  Tooltip,
} from 'ant-design-vue';

import {
  createScheduledTask,
  deleteScheduledTask,
  getScheduledTasks,
  runScheduledTask,
  updateScheduledTask,
  updateScheduledTaskEnabled,
} from '#/api/system/scheduled-task';
import type { ScheduledTask } from '#/api/system/scheduled-task';

defineOptions({ name: 'SystemScheduledTask' });

// ==================== 状态 ====================

const loading = ref(false);
const dataSource = ref<ScheduledTask[]>([]);
const editModalVisible = ref(false);
const editTask = ref<ScheduledTask | null>(null);
const editForm = ref({
  name: '',
  cronExpr: '',
  retentionDays: 7,
});

// 创建弹窗
const createModalVisible = ref(false);
const createForm = ref({
  name: '',
  taskType: 'recycle_cleanup',
  cronExpr: '0 3 * * *',
  retentionDays: 7,
});

// ==================== 表格列定义 ====================

const columns = [
  {
    title: '任务名称',
    dataIndex: 'name',
    width: 150,
  },
  {
    title: '任务类型',
    dataIndex: 'taskType',
    width: 120,
    customRender: ({ text }: { text: string }) => {
      const typeMap: Record<string, { label: string; color: string }> = {
        recycle_cleanup: { label: '回收站清理', color: 'blue' },
      };
      const type = typeMap[text] || { label: text, color: 'default' };
      return { children: type.label, props: { color: type.color } };
    },
  },
  {
    title: 'Cron 表达式',
    dataIndex: 'cronExpr',
    width: 120,
  },
  {
    title: '配置',
    dataIndex: 'config',
    width: 200,
    customRender: ({ text }: { text: Record<string, any> }) => {
      if (!text) return '-';
      const parts = [];
      if (text.retention_days !== undefined) parts.push(`保留 ${text.retention_days} 天`);
      return parts.join(', ') || '-';
    },
  },
  {
    title: '状态',
    dataIndex: 'enabled',
    width: 100,
    customRender: ({ text }: { text: boolean }) => ({
      children: text ? '启用' : '禁用',
      props: { color: text ? 'green' : 'red' },
    }),
  },
  {
    title: '运行状态',
    dataIndex: 'status',
    width: 100,
    customRender: ({ text }: { text: string }) => {
      const statusMap: Record<string, { label: string; color: string }> = {
        idle: { label: '空闲', color: 'default' },
        running: { label: '运行中', color: 'processing' },
        success: { label: '成功', color: 'green' },
        failed: { label: '失败', color: 'red' },
      };
      const status = statusMap[text] || { label: text, color: 'default' };
      return { children: status.label, props: { color: status.color } };
    },
  },
  {
    title: '执行次数',
    dataIndex: 'runCount',
    width: 100,
  },
  {
    title: '最后执行',
    dataIndex: 'lastRunAt',
    width: 180,
    customRender: ({ text }: { text: string }) => text ? new Date(text).toLocaleString('zh-CN') : '-',
  },
  {
    title: '最后结果',
    dataIndex: 'lastResult',
    width: 200,
    customRender: ({ text }: { text: string }) => {
      if (!text) return '-';
      return { children: text, props: { title: text } };
    },
  },
  {
    title: '下次执行',
    dataIndex: 'nextRunAt',
    width: 180,
    customRender: ({ text }: { text: string }) => text ? new Date(text).toLocaleString('zh-CN') : '-',
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
    const result = await getScheduledTasks();
    dataSource.value = result || [];
  } catch {
    message.error('加载定时任务失败');
  } finally {
    loading.value = false;
  }
}

// ==================== 操作 ====================

async function handleToggleEnabled(record: ScheduledTask) {
  try {
    await updateScheduledTaskEnabled(record.id, !record.enabled);
    message.success(`已${record.enabled ? '禁用' : '启用'}任务`);
    loadData();
  } catch {
    message.error('操作失败');
  }
}

function handleEdit(record: ScheduledTask) {
  editTask.value = record;
  editForm.value = {
    name: record.name,
    cronExpr: record.cronExpr,
    retentionDays: record.config?.retention_days || 7,
  };
  editModalVisible.value = true;
}

async function handleSaveEdit() {
  if (!editTask.value) return;

  try {
    await updateScheduledTask(editTask.value.id, {
      name: editForm.value.name,
      cronExpr: editForm.value.cronExpr,
      config: {
        retention_days: editForm.value.retentionDays,
      },
    });
    message.success('保存成功');
    editModalVisible.value = false;
    loadData();
  } catch (err: any) {
    message.error(err.message || '保存失败');
  }
}

async function handleRun(record: ScheduledTask) {
  Modal.confirm({
    title: '手动执行',
    content: `确定立即执行 "${record.name}" 任务吗？`,
    okText: '确定执行',
    cancelText: '取消',
    onOk: async () => {
      try {
        await runScheduledTask(record.id);
        message.success('任务已开始执行');
        // 延迟刷新，等待任务执行完成
        setTimeout(loadData, 2000);
      } catch {
        message.error('执行失败');
      }
    },
  });
}

function handleCreate() {
  createForm.value = {
    name: '',
    taskType: 'recycle_cleanup',
    cronExpr: '0 3 * * *',
    retentionDays: 7,
  };
  createModalVisible.value = true;
}

async function handleDelete(record: ScheduledTask) {
  Modal.confirm({
    title: '删除任务',
    content: `确定删除定时任务 "${record.name}" 吗？`,
    okText: '确定删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await deleteScheduledTask(record.id);
        message.success('删除成功');
        loadData();
      } catch {
        message.error('删除失败');
      }
    },
  });
}

async function handleSaveCreate() {
  if (!createForm.value.name) {
    message.warning('请输入任务名称');
    return;
  }

  try {
    await createScheduledTask({
      name: createForm.value.name,
      taskType: createForm.value.taskType,
      cronExpr: createForm.value.cronExpr,
      config: {
        retention_days: createForm.value.retentionDays,
      },
    });
    message.success('创建成功');
    createModalVisible.value = false;
    loadData();
  } catch (err: any) {
    message.error(err.message || '创建失败');
  }
}

// ==================== Cron 表达式帮助 ====================

const cronExamples = [
  { expr: '0 3 * * *', desc: '每天凌晨 3:00' },
  { expr: '0 */6 * * *', desc: '每 6 小时' },
  { expr: '0 0 * * 0', desc: '每周日 0:00' },
  { expr: '0 0 1 * *', desc: '每月 1 号 0:00' },
];

// ==================== 初始化 ====================

onMounted(() => {
  loadData();
});
</script>

<template>
  <Page auto-content-height>
    <div class="p-4">
      <!-- 说明卡片 -->
      <Card class="mb-4">
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-600">
            <p class="mb-2"><strong>定时任务管理</strong> - 配置系统自动执行的任务</p>
            <p>Cron 表达式格式：分 时 日 月 周（例：0 3 * * * = 每天凌晨 3:00）</p>
          </div>
          <Button type="primary" @click="handleCreate">
            新增任务
          </Button>
        </div>
      </Card>

      <!-- 表格 -->
      <Table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="false"
        row-key="id"
        :scroll="{ x: 1800 }"
      >
        <template #bodyCell="{ column, record }">
          <!-- 启用状态列 -->
          <template v-if="column.dataIndex === 'enabled'">
            <Switch
              :checked="(record as any).enabled"
              @change="() => handleToggleEnabled(record as any)"
            />
          </template>

          <!-- 最后结果列 -->
          <template v-if="column.dataIndex === 'lastResult'">
            <Tooltip v-if="(record as any).lastResult" :title="(record as any).lastResult">
              <span class="truncate max-w-[180px] block">{{ (record as any).lastResult }}</span>
            </Tooltip>
            <span v-else class="text-gray-400">-</span>
          </template>

          <!-- 操作列 -->
          <template v-if="column.dataIndex === 'action'">
            <Space>
              <Button type="link" size="small" @click="handleEdit(record as any)">
                编辑
              </Button>
              <Button type="link" size="small" @click="handleRun(record as any)">
                立即执行
              </Button>
              <Button type="link" size="small" danger @click="handleDelete(record as any)">
                删除
              </Button>
            </Space>
          </template>
        </template>
      </Table>
    </div>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="editModalVisible"
      title="编辑定时任务"
      @ok="handleSaveEdit"
      width="600px"
    >
      <Form v-if="editTask" :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <FormItem label="任务名称">
          <Input v-model:value="editForm.name" />
        </FormItem>

        <FormItem label="Cron 表达式">
          <Input v-model:value="editForm.cronExpr" placeholder="0 3 * * *" />
          <div class="mt-2 text-xs text-gray-500">
            <p class="mb-1">常用表达式：</p>
            <div v-for="example in cronExamples" :key="example.expr" class="flex items-center gap-2 mb-1">
              <code
                class="px-2 py-0.5 bg-gray-100 rounded cursor-pointer hover:bg-gray-200"
                @click="editForm.cronExpr = example.expr"
              >
                {{ example.expr }}
              </code>
              <span>= {{ example.desc }}</span>
            </div>
          </div>
        </FormItem>

        <FormItem v-if="editTask.taskType === 'recycle_cleanup'" label="保留天数">
          <InputNumber v-model:value="editForm.retentionDays" :min="1" :max="365" />
          <span class="ml-2 text-gray-500">天（超过此天数的文件将被自动清理）</span>
        </FormItem>
      </Form>
    </Modal>

    <!-- 创建弹窗 -->
    <Modal
      v-model:open="createModalVisible"
      title="新增定时任务"
      @ok="handleSaveCreate"
      width="600px"
    >
      <Form :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }">
        <FormItem label="任务名称">
          <Input v-model:value="createForm.name" placeholder="请输入任务名称" />
        </FormItem>

        <FormItem label="任务类型">
          <select v-model="createForm.taskType" class="w-full border rounded px-3 py-1.5">
            <option value="recycle_cleanup">回收站清理</option>
          </select>
        </FormItem>

        <FormItem label="Cron 表达式">
          <Input v-model:value="createForm.cronExpr" placeholder="0 3 * * *" />
          <div class="mt-2 text-xs text-gray-500">
            <p class="mb-1">常用表达式：</p>
            <div v-for="example in cronExamples" :key="example.expr" class="flex items-center gap-2 mb-1">
              <code
                class="px-2 py-0.5 bg-gray-100 rounded cursor-pointer hover:bg-gray-200"
                @click="createForm.cronExpr = example.expr"
              >
                {{ example.expr }}
              </code>
              <span>= {{ example.desc }}</span>
            </div>
          </div>
        </FormItem>

        <FormItem v-if="createForm.taskType === 'recycle_cleanup'" label="保留天数">
          <InputNumber v-model:value="createForm.retentionDays" :min="1" :max="365" />
          <span class="ml-2 text-gray-500">天（超过此天数的文件将被自动清理）</span>
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>
