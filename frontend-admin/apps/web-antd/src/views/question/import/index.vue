<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { QuestionImportApi } from '#/api/question/import';

import { useAccess } from '@vben/access';
import { Page, useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message } from 'ant-design-vue';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  deleteImportTask,
  getImportTaskList,
} from '#/api/question/import';

import { useImportColumns, useImportFormSchema } from './data';
import ImportForm from './modules/form.vue';

const { hasAccessByCodes } = useAccess();

const [FormDrawer, formDrawerApi] = useVbenDrawer({
  connectedComponent: ImportForm,
  destroyOnClose: true,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useImportFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useImportColumns(),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getImportTaskList({
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
  } as VxeTableGridOptions<QuestionImportApi.ImportTask>,
});

async function onDelete(row: QuestionImportApi.ImportTask) {
  const hideLoading = message.loading({
    content: `正在删除 ${row.fileName}...`,
    duration: 0,
    key: 'action_process_msg',
  });
  try {
    await deleteImportTask(row.id);
    message.success({
      content: `${row.fileName} 删除成功`,
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
    <Grid table-title="导入任务列表">
      <template #toolbar-tools>
        <Button
          v-if="hasAccessByCodes(['question:import'])"
          type="primary"
          @click="onCreate"
        >
          <Plus class="size-5" />
          新建导入任务
        </Button>
      </template>
      <template #action="{ row }">
        <VbenTableAction
          :dropdown-actions="[
            {
              text: '删除',
              icon: 'lucide:trash-2',
              danger: true,
              popConfirm: {
                title: `确定删除【${row.fileName}】？`,
                confirm: () => onDelete(row),
              },
              auth: ['question:import:delete'],
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
  </Page>
</template>
