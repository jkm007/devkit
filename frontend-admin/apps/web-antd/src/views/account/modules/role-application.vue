<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Empty,
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
import { $t } from '#/locales';

const loading = ref(false);
const submitLoading = ref(false);
const modalVisible = ref(false);
const availableRoles = ref<RoleApplicationApi.AvailableRole[]>([]);
const applications = ref<RoleApplicationApi.RoleApplicationItem[]>([]);
const total = ref(0);
const pagination = reactive({ current: 1, pageSize: 10 });
const formState = reactive({
  roleId: undefined as number | undefined,
  reason: '',
});

const columns = [
  {
    title: $t('account.roleApplication.roleName'),
    dataIndex: 'roleName',
    key: 'roleName',
    width: 160,
  },
  {
    title: $t('account.roleApplication.reason'),
    dataIndex: 'reason',
    key: 'reason',
    ellipsis: true,
  },
  {
    title: $t('account.roleApplication.status'),
    dataIndex: 'status',
    key: 'status',
    width: 100,
  },
  {
    title: $t('account.roleApplication.reviewNote'),
    dataIndex: 'reviewNote',
    key: 'reviewNote',
    ellipsis: true,
  },
  {
    title: $t('account.roleApplication.createdAt'),
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 170,
  },
  {
    title: $t('account.roleApplication.reviewedAt'),
    dataIndex: 'reviewedAt',
    key: 'reviewedAt',
    width: 170,
  },
];

function getStatusText(status: number) {
  const statusMap: Record<number, string> = {
    0: $t('account.roleApplication.pending'),
    1: $t('account.roleApplication.approved'),
    2: $t('account.roleApplication.rejected'),
  };
  return statusMap[status] || $t('account.roleApplication.unknown');
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
    message.warning($t('account.roleApplication.selectRole'));
    return;
  }
  submitLoading.value = true;
  try {
    await createRoleApplication({
      roleId: formState.roleId,
      reason: formState.reason,
    });
    message.success($t('account.roleApplication.submitSuccess'));
    modalVisible.value = false;
    pagination.current = 1;
    await loadApplications();
  } finally {
    submitLoading.value = false;
  }
}

function handleTableChange(pag: any) {
  pagination.current = pag.current;
  pagination.pageSize = pag.pageSize;
  loadApplications();
}

onMounted(() => {
  loadApplications();
});
</script>

<template>
  <Page auto-content-height>
    <Card>
      <div class="mb-4 flex items-center justify-between">
        <div>
          <h3 class="text-base font-medium">
            {{ $t('account.roleApplication.title') }}
          </h3>
          <p class="text-foreground/50 mt-1 text-xs">
            {{ $t('account.roleApplication.description') }}
          </p>
        </div>
        <Space>
          <Button @click="loadApplications">
            <span class="lucide:refresh-cw mr-1 size-4" />
            刷新
          </Button>
          <Button type="primary" @click="openApplyModal">
            <span class="lucide:plus mr-1 size-4" />
            {{ $t('account.roleApplication.apply') }}
          </Button>
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
        size="middle"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'roleName'">
            <div class="font-medium">
              {{ record.roleName || `${$t('account.roleApplication.role')} ${record.roleId}` }}
            </div>
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
            <span class="text-foreground/70">{{ record.reason || '-' }}</span>
          </template>
          <template v-else-if="column.key === 'reviewNote'">
            <span class="text-foreground/70">{{ record.reviewNote || '-' }}</span>
          </template>
          <template v-else-if="column.key === 'createdAt'">
            <span class="text-foreground/50 text-xs">{{
              formatDate(record.createdAt)
            }}</span>
          </template>
          <template v-else-if="column.key === 'reviewedAt'">
            <span class="text-foreground/50 text-xs">{{
              formatDate(record.reviewedAt)
            }}</span>
          </template>
        </template>
        <template #emptyText>
          <Empty
            :description="$t('account.roleApplication.noApplications')"
            :image="Empty.PRESENTED_IMAGE_SIMPLE"
          />
        </template>
      </Table>

      <Modal
        v-model:open="modalVisible"
        :title="$t('account.roleApplication.applyTitle')"
        :confirm-loading="submitLoading"
        :ok-text="$t('account.roleApplication.submit')"
        @ok="submitApplication"
      >
        <Form layout="vertical" class="mt-4">
          <FormItem :label="$t('account.roleApplication.selectRoleLabel')" required>
            <Select
              v-model:value="formState.roleId"
              :placeholder="$t('account.roleApplication.selectRolePlaceholder')"
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
              {{ $t('account.roleApplication.noAvailableRoles') }}
            </div>
          </FormItem>
          <FormItem :label="$t('account.roleApplication.reasonLabel')">
            <Input.TextArea
              v-model:value="formState.reason"
              :maxlength="500"
              :rows="4"
              :placeholder="$t('account.roleApplication.reasonPlaceholder')"
              show-count
            />
          </FormItem>
        </Form>
      </Modal>
    </Card>
  </Page>
</template>
