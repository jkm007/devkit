<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
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
  createRoleApplication,
  getAvailableRoles,
  getMyRoleApplications,
} from '#/api';

const loading = ref(false);
const submitLoading = ref(false);
const modalVisible = ref(false);
const availableRoles = ref<RoleApplicationApi.AvailableRole[]>([]);
const applications = ref<RoleApplicationApi.RoleApplicationItem[]>([]);
const total = ref(0);
const pagination = reactive({ current: 1, pageSize: 10 });
const formState = reactive({ roleId: undefined as number | undefined, reason: '' });

const columns = [
  { title: '申请角色', dataIndex: 'roleName', key: 'roleName' },
  { title: '申请理由', dataIndex: 'reason', key: 'reason', ellipsis: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 110 },
  { title: '审核备注', dataIndex: 'reviewNote', key: 'reviewNote', ellipsis: true },
  { title: '申请时间', dataIndex: 'createdAt', key: 'createdAt', width: 180 },
  { title: '审核时间', dataIndex: 'reviewedAt', key: 'reviewedAt', width: 180 },
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
    const result = await getMyRoleApplications({
      page: pagination.current,
      pageSize: pagination.pageSize,
    });
    applications.value = result?.items || [];
    total.value = result?.total || 0;
  } finally {
    loading.value = false;
  }
}

async function loadAvailableRoles() {
  availableRoles.value = await getAvailableRoles();
}

async function openApplyModal() {
  formState.roleId = undefined;
  formState.reason = '';
  await loadAvailableRoles();
  modalVisible.value = true;
}

async function submitApplication() {
  if (!formState.roleId) {
    message.warning('请选择要申请的角色');
    return;
  }
  submitLoading.value = true;
  try {
    await createRoleApplication({
      roleId: formState.roleId,
      reason: formState.reason,
    });
    message.success('角色申请已提交');
    modalVisible.value = false;
    pagination.current = 1;
    await loadApplications();
  } finally {
    submitLoading.value = false;
  }
}

onMounted(() => {
  loadApplications();
});
</script>

<template>
  <div>
    <div class="mb-4 flex items-center justify-between">
      <div>
        <h3 class="mb-1 text-base font-medium">角色申请</h3>
        <p class="text-foreground/50 text-sm">
          查看我的角色申请记录，或申请新的系统角色。
        </p>
      </div>
      <Space>
        <Button @click="loadApplications">刷新</Button>
        <Button type="primary" @click="openApplyModal">申请角色</Button>
      </Space>
    </div>

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
        <template v-if="column.key === 'roleName'">
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
        <template v-else-if="column.key === 'reviewNote'">
          {{ record.reviewNote || '-' }}
        </template>
        <template v-else-if="column.key === 'createdAt'">
          {{ formatDate(record.createdAt) }}
        </template>
        <template v-else-if="column.key === 'reviewedAt'">
          {{ formatDate(record.reviewedAt) }}
        </template>
      </template>
    </Table>

    <Modal
      v-model:open="modalVisible"
      title="申请角色"
      :confirm-loading="submitLoading"
      ok-text="提交申请"
      @ok="submitApplication"
    >
      <Form layout="vertical">
        <FormItem label="申请角色" required>
          <Select
            v-model:value="formState.roleId"
            placeholder="请选择角色"
            :options="undefined"
          >
            <SelectOption
              v-for="role in availableRoles"
              :key="role.id"
              :value="role.id"
            >
              {{ role.name }}{{ role.remark ? ` - ${role.remark}` : '' }}
            </SelectOption>
          </Select>
          <div
            v-if="availableRoles.length === 0"
            class="text-foreground/50 mt-2 text-xs"
          >
            当前没有可申请的角色，可能是您已拥有或已有待审核申请。
          </div>
        </FormItem>
        <FormItem label="申请理由">
          <Input.TextArea
            v-model:value="formState.reason"
            :maxlength="500"
            :rows="4"
            placeholder="请填写申请理由，便于管理员审核"
            show-count
          />
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>
