<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import { useAccess } from '@vben/access';
import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Form,
  FormItem,
  Input,
  message,
  Modal,
  Select,
  SelectOption,
  Space,
  Table,
  Tag,
} from 'ant-design-vue';

import type { RoleApplicationApi } from '#/api';
import {
  approveRoleApplication,
  getRoleApplicationList,
  rejectRoleApplication,
} from '#/api';

const { hasAccessByCodes } = useAccess();
const loading = ref(false);
const reviewLoading = ref(false);
const reviewVisible = ref(false);
const reviewAction = ref<'approve' | 'reject'>('approve');
const currentRecord = ref<RoleApplicationApi.RoleApplicationItem>();
const applications = ref<RoleApplicationApi.RoleApplicationItem[]>([]);
const total = ref(0);
const pagination = reactive({ current: 1, pageSize: 20 });
const queryForm = reactive({ status: undefined as number | undefined, userId: '', roleId: '' });
const reviewForm = reactive({ note: '' });

const columns = [
  { title: '申请人', dataIndex: 'username', key: 'username', width: 160 },
  { title: '申请角色', dataIndex: 'roleName', key: 'roleName', width: 160 },
  { title: '申请理由', dataIndex: 'reason', key: 'reason', ellipsis: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 110 },
  { title: '审核人', dataIndex: 'reviewerName', key: 'reviewerName', width: 140 },
  { title: '审核备注', dataIndex: 'reviewNote', key: 'reviewNote', ellipsis: true },
  { title: '申请时间', dataIndex: 'createdAt', key: 'createdAt', width: 180 },
  { title: '审核时间', dataIndex: 'reviewedAt', key: 'reviewedAt', width: 180 },
  { title: '操作', key: 'action', fixed: 'right' as const, width: 150 },
];

function getStatusText(status: number) {
  const statusMap: Record<number, string> = {
    0: '待审核',
    1: '已通过',
    2: '已驳回',
  };
  return statusMap[status] || '未知';
}

function getStatusColor(status: number) {
  const colorMap: Record<number, string> = {
    0: 'processing',
    1: 'success',
    2: 'error',
  };
  return colorMap[status] || 'default';
}

function formatDate(date?: string) {
  return date ? new Date(date).toLocaleString() : '-';
}

async function loadApplications() {
  loading.value = true;
  try {
    const result = await getRoleApplicationList({
      page: pagination.current,
      pageSize: pagination.pageSize,
      status: queryForm.status,
      userId: queryForm.userId || undefined,
      roleId: queryForm.roleId || undefined,
    });
    applications.value = result?.items || [];
    total.value = result?.total || 0;
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  loadApplications();
}

function handleReset() {
  queryForm.status = undefined;
  queryForm.userId = '';
  queryForm.roleId = '';
  handleSearch();
}

function openReviewModal(
  record: RoleApplicationApi.RoleApplicationItem,
  action: 'approve' | 'reject',
) {
  currentRecord.value = record;
  reviewAction.value = action;
  reviewForm.note = '';
  reviewVisible.value = true;
}

async function submitReview() {
  if (!currentRecord.value) return;
  reviewLoading.value = true;
  try {
    if (reviewAction.value === 'approve') {
      await approveRoleApplication(currentRecord.value.id, { note: reviewForm.note });
      message.success('角色申请已通过');
    } else {
      await rejectRoleApplication(currentRecord.value.id, { note: reviewForm.note });
      message.success('角色申请已驳回');
    }
    reviewVisible.value = false;
    await loadApplications();
  } finally {
    reviewLoading.value = false;
  }
}

onMounted(() => {
  loadApplications();
});
</script>

<template>
  <Page auto-content-height title="角色申请审核">
    <Card class="mb-4">
      <Form layout="inline">
        <FormItem label="状态">
          <Select
            v-model:value="queryForm.status"
            allow-clear
            placeholder="全部状态"
            style="width: 140px"
          >
            <SelectOption :value="0">待审核</SelectOption>
            <SelectOption :value="1">已通过</SelectOption>
            <SelectOption :value="2">已驳回</SelectOption>
          </Select>
        </FormItem>
        <FormItem label="用户ID">
          <Input
            v-model:value="queryForm.userId"
            allow-clear
            placeholder="请输入用户ID"
            style="width: 160px"
          />
        </FormItem>
        <FormItem label="角色ID">
          <Input
            v-model:value="queryForm.roleId"
            allow-clear
            placeholder="请输入角色ID"
            style="width: 160px"
          />
        </FormItem>
        <FormItem>
          <Space>
            <Button type="primary" @click="handleSearch">查询</Button>
            <Button @click="handleReset">重置</Button>
          </Space>
        </FormItem>
      </Form>
    </Card>

    <Card>
      <Table
        :columns="columns"
        :data-source="applications"
        :loading="loading"
        :pagination="{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total,
          showSizeChanger: true,
          showTotal: (count: number) => `共 ${count} 条`,
        }"
        row-key="id"
        size="small"
        @change="
          (pag: any) => {
            pagination.current = pag.current;
            pagination.pageSize = pag.pageSize;
            loadApplications();
          }
        "
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'username'">
            <div>{{ record.nickname || record.username || `用户 ${record.userId}` }}</div>
            <div v-if="record.username" class="text-foreground/50 text-xs">
              {{ record.username }} · ID: {{ record.userId }}
            </div>
          </template>
          <template v-else-if="column.key === 'roleName'">
            <div>{{ record.roleName || `角色 ${record.roleId}` }}</div>
            <div v-if="record.roleRemark" class="text-foreground/50 text-xs">
              {{ record.roleRemark }}
            </div>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="getStatusColor(record.status)">
              {{ getStatusText(record.status) }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'reason'">
            {{ record.reason || '-' }}
          </template>
          <template v-else-if="column.key === 'reviewerName'">
            {{ record.reviewerName || '-' }}
          </template>
          <template v-else-if="column.key === 'reviewNote'">
            {{ record.reviewNote || '-' }}
          </template>
          <template v-else-if="column.key === 'createdAt'">
            {{ formatDate(record.createdAt) }}
          </template>
          <template v-else-if="column.key === 'reviewedAt'">
            {{ formatDate(record.reviewedAt) }}
          </template>
          <template v-else-if="column.key === 'action'">
            <Space v-if="record.status === 0 && hasAccessByCodes(['system:roleapp:review'])">
              <Button
                size="small"
                type="link"
                @click="openReviewModal(record as RoleApplicationApi.RoleApplicationItem, 'approve')"
              >
                通过
              </Button>
              <Button
                danger
                size="small"
                type="link"
                @click="openReviewModal(record as RoleApplicationApi.RoleApplicationItem, 'reject')"
              >
                驳回
              </Button>
            </Space>
            <span v-else class="text-foreground/40">-</span>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="reviewVisible"
      :confirm-loading="reviewLoading"
      :ok-text="reviewAction === 'approve' ? '通过' : '驳回'"
      :title="reviewAction === 'approve' ? '通过角色申请' : '驳回角色申请'"
      @ok="submitReview"
    >
      <div v-if="currentRecord" class="mb-4 rounded bg-gray-50 p-3 text-sm dark:bg-gray-800">
        <div>申请人：{{ currentRecord.nickname || currentRecord.username || currentRecord.userId }}</div>
        <div>申请角色：{{ currentRecord.roleName || currentRecord.roleId }}</div>
        <div>申请理由：{{ currentRecord.reason || '-' }}</div>
      </div>
      <Form layout="vertical">
        <FormItem label="审核备注">
          <Input.TextArea
            v-model:value="reviewForm.note"
            :maxlength="500"
            :rows="4"
            placeholder="请输入审核备注"
            show-count
          />
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>
