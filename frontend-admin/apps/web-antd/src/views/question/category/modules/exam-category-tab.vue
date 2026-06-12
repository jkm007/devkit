<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { ExamCategoryApi } from '#/api/question/category';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal } from 'ant-design-vue';

import { useAccess } from '@vben/access';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  deleteExamCategory,
  getExamCategoryList,
  updateExamCategory,
} from '#/api/question/category';

import { useExamCategoryColumns, useExamCategoryFormSchema } from '../data';
import ExamCategoryForm from './exam-category-form.vue';

const { hasAccessByCodes } = useAccess();

const [FormDrawer, formDrawerApi] = useVbenDrawer({
  connectedComponent: ExamCategoryForm,
  destroyOnClose: true,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useExamCategoryFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useExamCategoryColumns(onStatusChange),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getExamCategoryList({
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
  } as VxeTableGridOptions<ExamCategoryApi.ExamCategory>,
});

function confirm(content: string, title: string) {
  return new Promise((resolve, reject) => {
    Modal.confirm({
      content,
      onCancel() {
        reject(new Error('已取消'));
      },
      onOk() {
        resolve(true);
      },
      title,
    });
  });
}

async function onStatusChange(
  newStatus: number,
  row: ExamCategoryApi.ExamCategory,
) {
  try {
    await confirm(
      `将【${row.name}】切换为 ${newStatus === 1 ? '启用' : '禁用'}？`,
      '切换状态',
    );
    await updateExamCategory(row.id, { ...row, status: newStatus });
    return true;
  } catch {
    return false;
  }
}

function onEdit(row: ExamCategoryApi.ExamCategory) {
  formDrawerApi.setData(row).open();
}

async function onDelete(row: ExamCategoryApi.ExamCategory) {
  const hideLoading = message.loading({
    content: `正在删除 ${row.name}...`,
    duration: 0,
    key: 'action_process_msg',
  });
  try {
    await deleteExamCategory(row.id);
    message.success({
      content: `${row.name} 删除成功`,
      key: 'action_process_msg',
    });
    onRefresh();
  } catch {
    hideLoading();
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
  <div>
    <FormDrawer @success="onRefresh" />
    <Grid table-title="考试大类列表">
      <template #toolbar-tools>
        <Button
          v-if="hasAccessByCodes(['question:category:add'])"
          type="primary"
          @click="onCreate"
        >
          <Plus class="size-5" />
          新增考试大类
        </Button>
      </template>
      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: '编辑',
              icon: 'lucide:edit',
              onClick: () => onEdit(row),
              auth: ['question:category:edit'],
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
              auth: ['question:category:delete'],
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
  </div>
</template>
