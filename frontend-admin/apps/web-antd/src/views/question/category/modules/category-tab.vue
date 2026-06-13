<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { QuestionCategoryApi } from '#/api/question/category';

import { ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal } from 'ant-design-vue';

import { useAccess } from '@vben/access';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  deleteQuestionCategory,
  getExamAll,
  getQuestionCategoryAll,
  getQuestionCategoryList,
  getSubjectAll,
  updateQuestionCategory,
} from '#/api/question/category';

import { useCategoryColumns, useCategoryFormSchema } from '../data';
import CategoryForm from './category-form.vue';

const { hasAccessByCodes } = useAccess();

const examOptions = ref<any[]>([]);
const subjectOptions = ref<any[]>([]);
const parentOptions = ref<any[]>([]);

const [FormDrawer, formDrawerApi] = useVbenDrawer({
  connectedComponent: CategoryForm,
  destroyOnClose: true,
});

async function loadOptions() {
  try {
    const [exams, categories] = await Promise.all([
      getExamAll(),
      getQuestionCategoryAll(),
    ]);
    examOptions.value = (exams || []).map((item: any) => ({
      label: item.name,
      value: item.id,
    }));
    parentOptions.value = buildTree(categories || []);
  } catch {
    // ignore
  }
}

async function fetchSubjectOptions(examId: number): Promise<any[]> {
  try {
    const res = await getSubjectAll(examId);
    return (res || []).map((item: any) => ({
      label: item.name,
      value: item.id,
    }));
  } catch {
    return [];
  }
}

function buildTree(items: any[]): any[] {
  const map = new Map<number, any>();
  const roots: any[] = [];
  items.forEach((item) => {
    map.set(item.id, { ...item, children: [] });
  });
  items.forEach((item) => {
    const node = map.get(item.id)!;
    if (item.parentId && map.has(item.parentId)) {
      map.get(item.parentId)!.children.push(node);
    } else {
      roots.push(node);
    }
  });
  return roots;
}

loadOptions();

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useCategoryFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useCategoryColumns(onStatusChange),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getQuestionCategoryList({
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
  } as VxeTableGridOptions<QuestionCategoryApi.QuestionCategory>,
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
  row: QuestionCategoryApi.QuestionCategory,
) {
  try {
    await confirm(
      `将【${row.name}】切换为 ${newStatus === 1 ? '启用' : '禁用'}？`,
      '切换状态',
    );
    await updateQuestionCategory(row.id, { ...row, status: newStatus });
    return true;
  } catch {
    return false;
  }
}

function onEdit(row: QuestionCategoryApi.QuestionCategory) {
  formDrawerApi
    .setData({
      ...row,
      examOptions: examOptions.value,
      subjectOptions: subjectOptions.value,
      parentOptions: parentOptions.value,
      onExamChange: fetchSubjectOptions,
    })
    .open();
}

async function onDelete(row: QuestionCategoryApi.QuestionCategory) {
  const hideLoading = message.loading({
    content: `正在删除 ${row.name}...`,
    duration: 0,
    key: 'action_process_msg',
  });
  try {
    await deleteQuestionCategory(row.id);
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
  formDrawerApi
    .setData({
      examOptions: examOptions.value,
      subjectOptions: subjectOptions.value,
      parentOptions: parentOptions.value,
      onExamChange: fetchSubjectOptions,
    })
    .open();
}
</script>
<template>
  <div>
    <FormDrawer @success="onRefresh" />
    <Grid table-title="章节分类列表">
      <template #toolbar-tools>
        <Button
          v-if="hasAccessByCodes(['question:category:add'])"
          type="primary"
          @click="onCreate"
        >
          <Plus class="size-5" />
          新增分类
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
