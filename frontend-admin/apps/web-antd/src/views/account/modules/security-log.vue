<script lang="ts" setup>
import type { AccountApi } from '#/api';

import { onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Table,
  Tag,
  Tooltip,
  Empty,
} from 'ant-design-vue';

import { getMySecurityLogs } from '#/api';
import { $t } from '#/locales';

const loading = ref(false);
const securityLogs = ref<AccountApi.SecurityLog[]>([]);
const total = ref(0);
const pagination = reactive({ current: 1, pageSize: 15 });

const columns = [
  {
    title: $t('account.security.status'),
    dataIndex: 'status',
    key: 'status',
    width: 100,
  },
  {
    title: $t('account.security.eventType'),
    dataIndex: 'eventType',
    key: 'eventType',
    width: 150,
  },
  {
    title: $t('account.security.eventDetail'),
    dataIndex: 'eventDetail',
    key: 'eventDetail',
  },
  {
    title: 'IP',
    dataIndex: 'ip',
    key: 'ip',
    width: 140,
  },
  {
    title: $t('account.security.time'),
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 180,
  },
];

async function loadSecurityLogs() {
  loading.value = true;
  try {
    const res = await getMySecurityLogs({
      page: pagination.current,
      pageSize: pagination.pageSize,
    });
    securityLogs.value = res.items || [];
    total.value = res.total || 0;
  } catch {
    securityLogs.value = [];
  } finally {
    loading.value = false;
  }
}

function getEventTypeLabel(type: string) {
  return $t(`account.eventType.${type}`);
}

function getStatusTag(status: number) {
  return status === 1
    ? { color: 'success', text: $t('account.security.statusSuccess') }
    : { color: 'error', text: $t('account.security.statusFail') };
}

function handleTableChange(pag: any) {
  pagination.current = pag.current;
  pagination.pageSize = pag.pageSize;
  loadSecurityLogs();
}

onMounted(() => {
  loadSecurityLogs();
});
</script>

<template>
  <Page auto-content-height>
    <Card>
      <div class="mb-4 flex items-center justify-between">
        <div>
          <h3 class="text-base font-medium">
            {{ $t('account.security.securityLogs') }}
          </h3>
          <p class="text-foreground/50 mt-1 text-xs">
            {{ $t('account.security.logDescription') }}
          </p>
        </div>
        <Button @click="loadSecurityLogs">
          <span class="lucide:refresh-cw mr-1 size-4" />
          刷新
        </Button>
      </div>

      <Table
        :columns="columns"
        :data-source="securityLogs"
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
          <template v-if="column.key === 'status'">
            <Tag :color="getStatusTag(record.status).color">
              {{ getStatusTag(record.status).text }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'eventType'">
            <span class="font-medium">{{
              getEventTypeLabel(record.eventType)
            }}</span>
          </template>
          <template v-else-if="column.key === 'eventDetail'">
            <span class="text-foreground/70">{{ record.eventDetail }}</span>
          </template>
          <template v-else-if="column.key === 'ip'">
            <span class="font-mono text-foreground/60">{{ record.ip }}</span>
          </template>
          <template v-else-if="column.key === 'createdAt'">
            <Tooltip :title="record.createdAt">
              <span class="text-foreground/50 text-xs">{{
                record.createdAt
              }}</span>
            </Tooltip>
          </template>
        </template>
        <template #emptyText>
          <Empty
            :description="$t('account.security.noLogs')"
            :image="Empty.PRESENTED_IMAGE_SIMPLE"
          />
        </template>
      </Table>
    </Card>
  </Page>
</template>
