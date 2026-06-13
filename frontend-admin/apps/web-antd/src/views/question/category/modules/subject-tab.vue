<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SubjectApi } from '#/api/question/category';

import { ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal } from 'ant-design-vue';

import { useAccess } from '@vben/access';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  deleteSubject,
  getExamAll,
  getSubjectList,
  updateSubject,
} from '#/api/question/category';

import { useSubjectColumns, useSubjectFormSchema } from '../data';
import SubjectForm from './subject-form.vue';

const { hasAccessByCodes } = useAccess();

const examOptions = ref<any[]>([]);

const [FormDrawer, formDrawerApi] = useVbenDrawer({
  connectedComponent: SubjectForm,
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
    message.error('加载考试列表失败');
  }
}

loadExamOptions();

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useSubjectFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useSubjectColumns(onStatusChange),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getSubjectList({
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
  } as VxeTableGridOptions<SubjectApi.Subject>,
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

async function onStatusChange(newStatus: number, row: SubjectApi.Subject) {
  try {
    await confirm(
      `将【${row.name}】切换为 ${newStatus === 1 ? '启用' : '禁用'}？`,
      '切换状态',
    );
    await updateSubject(row.id, { status: newStatus });
    return true;
  } catch {
    return false;
  }
}

function onEdit(row: SubjectApi.Subject) {
  formDrawerApi.setData({ ...row, examOptions: examOptions.value }).open();
}

async function onDelete(row: SubjectApi.Subject) {
  const hideLoading = message.loading({
    content: `正在删除 ${row.name}...`,
    duration: 0,
    key: 'action_process_msg',
  });
  try {
    await deleteSubject(row.id);
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
  <div>
    <FormDrawer @success="onRefresh" />
    <Grid table-title="科目列表">
      <template #toolbar-tools>
        <Button
          v-if="hasAccessByCodes(['question:subject:manage'])"
          type="primary"
          @click="onCreate"
        >
          <Plus class="size-5" />
          新增科目
        </Button>
      </template>
      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: '编辑',
              icon: 'lucide:edit',
              onClick: () => onEdit(row),
              auth: ['question:subject:manage'],
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
              auth: ['question:subject:manage'],
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
  </div>
</template>
