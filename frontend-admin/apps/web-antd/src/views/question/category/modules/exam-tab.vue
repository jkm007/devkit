<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { ExamApi } from '#/api/question/category';

import { ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal } from 'ant-design-vue';

import { useAccess } from '@vben/access';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  deleteExam,
  getExamCategoryAll,
  getExamList,
  updateExam,
} from '#/api/question/category';

import { useExamColumns, useExamFormSchema } from '../data';
import ExamForm from './exam-form.vue';

const { hasAccessByCodes } = useAccess();

const categoryOptions = ref<any[]>([]);

const [FormDrawer, formDrawerApi] = useVbenDrawer({
  connectedComponent: ExamForm,
  destroyOnClose: true,
});

async function loadCategoryOptions() {
  try {
    const res = await getExamCategoryAll();
    categoryOptions.value = (res || []).map((item: any) => ({
      label: item.name,
      value: item.id,
    }));
  } catch {
    message.error('加载考试大类失败');
  }
}

loadCategoryOptions();

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useExamFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useExamColumns(onStatusChange),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getExamList({
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
  } as VxeTableGridOptions<ExamApi.Exam>,
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

async function onStatusChange(newStatus: number, row: ExamApi.Exam) {
  try {
    await confirm(
      `将【${row.name}】切换为 ${newStatus === 1 ? '启用' : '禁用'}？`,
      '切换状态',
    );
    await updateExam(row.id, { status: newStatus });
    return true;
  } catch {
    return false;
  }
}

function onEdit(row: ExamApi.Exam) {
  formDrawerApi.setData({ ...row, categoryOptions: categoryOptions.value }).open();
}

async function onDelete(row: ExamApi.Exam) {
  const hideLoading = message.loading({
    content: `正在删除 ${row.name}...`,
    duration: 0,
    key: 'action_process_msg',
  });
  try {
    await deleteExam(row.id);
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
  formDrawerApi.setData({ categoryOptions: categoryOptions.value }).open();
}
</script>
<template>
  <div>
    <FormDrawer @success="onRefresh" />
    <Grid table-title="具体考试列表">
      <template #toolbar-tools>
        <Button
          v-if="hasAccessByCodes(['question:exam:manage'])"
          type="primary"
          @click="onCreate"
        >
          <Plus class="size-5" />
          新增考试
        </Button>
      </template>
      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: '编辑',
              icon: 'lucide:edit',
              onClick: () => onEdit(row),
              auth: ['question:exam:manage'],
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
              auth: ['question:exam:manage'],
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
  </div>
</template>
