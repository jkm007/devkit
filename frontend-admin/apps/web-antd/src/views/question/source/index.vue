<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { QuestionSourceApi } from '#/api/question/source';

import { ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal } from 'ant-design-vue';

import { useAccess } from '@vben/access';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import { getExamAll } from '#/api/question/category';
import {
  deleteQuestionSource,
  getQuestionSourceList,
} from '#/api/question/source';

import { useSourceColumns, useSourceFormSchema } from './data';
import SourceForm from './modules/form.vue';

const { hasAccessByCodes } = useAccess();

const examOptions = ref<any[]>([]);

const [FormDrawer, formDrawerApi] = useVbenDrawer({
  connectedComponent: SourceForm,
  destroyOnClose: true,
});

async function loadExamOptions() {
  try {
    const res = await getExamAll();
    examOptions.value = (res || []).map((item: any) => ({
      label: item.name,
      value: item.id,
    }));
  } catch {
    message.error('加载考试选项失败');
  }
}

loadExamOptions();

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useSourceFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useSourceColumns(),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getQuestionSourceList({
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
  } as VxeTableGridOptions<QuestionSourceApi.QuestionSource>,
});

function onEdit(row: QuestionSourceApi.QuestionSource) {
  formDrawerApi.setData({ ...row, examOptions: examOptions.value }).open();
}

async function onDelete(row: QuestionSourceApi.QuestionSource) {
  const hideLoading = message.loading({
    content: `正在删除 ${row.name}...`,
    duration: 0,
    key: 'action_process_msg',
  });
  try {
    await deleteQuestionSource(row.id);
    message.success({
      content: `${row.name} 删除成功`,
      key: 'action_process_msg',
    });
    onRefresh();
  } catch {
    hideLoading();
    message.error(`${row.name} 删除失败`);
  }
}

function onRefresh() {
  gridApi.query();
}

function onCreate() {
  formDrawerApi.setData({ examOptions: examOptions.value }).open();
}
</script>
<template>
  <Page auto-content-height>
    <FormDrawer @success="onRefresh" />
    <Grid table-title="来源列表">
      <template #toolbar-tools>
        <Button
          v-if="hasAccessByCodes(['question:source:manage'])"
          type="primary"
          @click="onCreate"
        >
          <Plus class="size-5" />
          新增来源
        </Button>
      </template>
      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: '编辑',
              icon: 'lucide:edit',
              onClick: () => onEdit(row),
              auth: ['question:source:manage'],
            },
          ]"
          :dropdown-actions="[
            {
              text: '删除',
              icon: 'lucide:trash-2',
              danger: true,
              popConfirm: {
                title: `确定删除【${row.name}】？`,
                confirm: () => onDelete(row),
              },
              auth: ['question:source:manage'],
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
  </Page>
</template>
