<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemRealNameApi } from '#/api';

import { h, ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Input, message, Modal } from 'ant-design-vue';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import { approveRealName, getRealNameList, rejectRealName } from '#/api';
import { $t } from '#/locales';

import { useColumns, useGridFormSchema } from './data';

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    fieldMappingTime: [['createTime', ['startTime', 'endTime']]],
    schema: useGridFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useColumns(),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getRealNameList({
            page: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      search: true,
      zoom: true,
    },
  } as VxeTableGridOptions<SystemRealNameApi.RealNameApplication>,
});

function onRefresh() {
  gridApi.query();
}

async function onApprove(row: SystemRealNameApi.RealNameApplication) {
  await approveRealName(row.id);
  message.success($t('system.realName.approveSuccess'));
  onRefresh();
}

function onReject(row: SystemRealNameApi.RealNameApplication) {
  const reasonRef = ref('');
  Modal.confirm({
    title: $t('system.realName.rejectTitle'),
    content: () =>
      h(Input.TextArea, {
        placeholder: $t('system.realName.rejectReasonPlaceholder'),
        rows: 3,
        onChange: (e: any) => {
          reasonRef.value = e.target?.value || '';
        },
      }),
    async onOk() {
      if (!reasonRef.value.trim()) {
        message.warning($t('system.realName.rejectReasonPlaceholder'));
        return Promise.reject();
      }
      await rejectRealName(row.id, { reason: reasonRef.value });
      message.success($t('system.realName.rejectSuccess'));
      onRefresh();
    },
  });
}
</script>

<template>
  <Page auto-content-height>
    <Grid :table-title="$t('system.realName.list')">
      <template #action="{ row }">
        <VbenTableAction
          v-if="row.status === 0"
          :actions="[
            {
              text: $t('system.realName.approve'),
              icon: 'lucide:check',
              onClick: () => onApprove(row),
            },
            {
              text: $t('system.realName.reject'),
              icon: 'lucide:x',
              danger: true,
              onClick: () => onReject(row),
            },
          ]"
          align="center"
        />
        <span v-else class="text-foreground/40">-</span>
      </template>
    </Grid>
  </Page>
</template>
