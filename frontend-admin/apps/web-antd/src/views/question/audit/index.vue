<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { QuestionApi } from '#/api/question/question';

import { Page } from '@vben/common-ui';

import { Button, Input, message, Modal } from 'ant-design-vue';

import { useAccess } from '@vben/access';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  approveQuestion,
  getQuestionList,
  rejectQuestion,
} from '#/api/question/question';

import { useAuditColumns, useAuditFormSchema } from './data';

const { hasAccessByCodes } = useAccess();

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useAuditFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useAuditColumns(),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getQuestionList({
            page: page.currentPage,
            pageSize: page.pageSize,
            status: formValues.status || 'pending',
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
  } as VxeTableGridOptions<QuestionApi.Question>,
});

async function onApprove(row: QuestionApi.Question) {
  try {
    await approveQuestion(row.id);
    message.success('审核通过');
    onRefresh();
  } catch {
    // ignore
  }
}

async function onReject(row: QuestionApi.Question) {
  Modal.confirm({
    title: '驳回题目',
    content: '请输入驳回原因：',
    input: true,
    async onOk() {
      try {
        await rejectQuestion(row.id, '不符合要求');
        message.success('已驳回');
        onRefresh();
      } catch {
        // ignore
      }
    },
  });
}

function onRefresh() {
  gridApi.query();
}
</script>
<template>
  <Page auto-content-height>
    <Grid table-title="审核队列">
      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: '通过',
              icon: 'lucide:check-circle',
              onClick: () => onApprove(row),
              auth: ['question:audit:approve'],
              ifShow: row.status === 'pending',
            },
            {
              text: '驳回',
              icon: 'lucide:x-circle',
              onClick: () => onReject(row),
              auth: ['question:audit:reject'],
              ifShow: row.status === 'pending',
              danger: true,
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
  </Page>
</template>
