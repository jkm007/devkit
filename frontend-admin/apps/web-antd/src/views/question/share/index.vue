<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { QuestionShareApi } from '#/api/question/share';

import { useAccess } from '@vben/access';
import { Page, useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message } from 'ant-design-vue';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  deleteQuestionShare,
  disableQuestionShare,
  enableQuestionShare,
  getQuestionShareList,
} from '#/api/question/share';

import { useShareColumns, useShareFormSchema } from './data';
import ShareForm from './modules/form.vue';

const { hasAccessByCodes } = useAccess();

const [FormDrawer, formDrawerApi] = useVbenDrawer({
  connectedComponent: ShareForm,
  destroyOnClose: true,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useShareFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useShareColumns(),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getQuestionShareList({
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
  } as VxeTableGridOptions<QuestionShareApi.QuestionShare>,
});

async function onDisable(row: QuestionShareApi.QuestionShare) {
  try {
    await disableQuestionShare(row.id);
    message.success('已禁用');
    onRefresh();
  } catch {
    message.error('禁用失败');
  }
}

async function onEnable(row: QuestionShareApi.QuestionShare) {
  try {
    await enableQuestionShare(row.id);
    message.success('已启用');
    onRefresh();
  } catch {
    message.error('启用失败');
  }
}

async function onDelete(row: QuestionShareApi.QuestionShare) {
  const hideLoading = message.loading({
    content: `正在删除分享 ${row.shareCode}...`,
    duration: 0,
    key: 'action_process_msg',
  });
  try {
    await deleteQuestionShare(row.id);
    message.success({
      content: `分享 ${row.shareCode} 删除成功`,
      key: 'action_process_msg',
    });
    onRefresh();
  } catch {
    hideLoading();
    message.error('删除失败');
  }
}

function onRefresh() {
  gridApi.query();
}

function onCreate() {
  formDrawerApi.setData({}).open();
}
</script>
<template>
  <Page auto-content-height>
    <FormDrawer @success="onRefresh" />
    <Grid table-title="题目分享列表">
      <template #toolbar-tools>
        <Button
          v-if="hasAccessByCodes(['question:share:create'])"
          type="primary"
          @click="onCreate"
        >
          <Plus class="size-5" />
          创建分享
        </Button>
      </template>
      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: row.status === 1 ? '禁用' : '启用',
              icon: row.status === 1 ? 'lucide:ban' : 'lucide:check-circle',
              ifShow: () => row.status !== 2,
              onClick: () =>
                row.status === 1 ? onDisable(row) : onEnable(row),
              auth: [
                row.status === 1
                  ? 'question:share:disable'
                  : 'question:share:enable',
              ],
            },
          ]"
          :dropdown-actions="[
            {
              text: '删除',
              icon: 'lucide:trash-2',
              danger: true,
              popConfirm: {
                title: `确定删除此分享？`,
                confirm: () => onDelete(row),
              },
              auth: ['question:share:delete'],
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
  </Page>
</template>
