<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { KnowledgePointApi } from '#/api/question/knowledge';

import { ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal } from 'ant-design-vue';

import { useAccess } from '@vben/access';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  deleteKnowledgePoint,
  getKnowledgePointAll,
  getKnowledgePointList,
  updateKnowledgePoint,
} from '#/api/question/knowledge';
import {
  getExamAll,
  getQuestionCategoryAll,
  getSubjectAll,
} from '#/api/question/category';

import { useKnowledgePointColumns, useKnowledgePointFormSchema } from './data';
import KnowledgePointForm from './modules/form.vue';

const { hasAccessByCodes } = useAccess();

const examOptions = ref<any[]>([]);
const subjectOptions = ref<any[]>([]);
const categoryOptions = ref<any[]>([]);
const parentOptions = ref<any[]>([]);

const [FormDrawer, formDrawerApi] = useVbenDrawer({
  connectedComponent: KnowledgePointForm,
  destroyOnClose: true,
});

async function loadOptions() {
  try {
    const [exams, categories, knowledgePoints] = await Promise.all([
      getExamAll(),
      getQuestionCategoryAll(),
      getKnowledgePointAll(),
    ]);
    examOptions.value = (exams || []).map((item: any) => ({
      label: item.name,
      value: item.id,
    }));
    categoryOptions.value = (categories || []).map((item: any) => ({
      label: item.name,
      value: item.id,
    }));
    parentOptions.value = buildTree(knowledgePoints || []);
  } catch {
    message.error('加载选项数据失败');
  }
}

async function loadSubjectOptions(examId: number) {
  try {
    const res = await getSubjectAll(examId);
    subjectOptions.value = (res || []).map((item: any) => ({
      label: item.name,
      value: item.id,
    }));
  } catch {
    subjectOptions.value = [];
    message.error('加载科目选项失败');
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
    schema: useKnowledgePointFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useKnowledgePointColumns(onStatusChange),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getKnowledgePointList({
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
  } as VxeTableGridOptions<KnowledgePointApi.KnowledgePoint>,
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
  row: KnowledgePointApi.KnowledgePoint,
) {
  try {
    await confirm(
      `将【${row.name}】切换为 ${newStatus === 1 ? '启用' : '禁用'}？`,
      '切换状态',
    );
    await updateKnowledgePoint(row.id, { ...row, status: newStatus });
    return true;
  } catch {
    return false;
  }
}

function onEdit(row: KnowledgePointApi.KnowledgePoint) {
  formDrawerApi
    .setData({
      ...row,
      examOptions: examOptions.value,
      subjectOptions: subjectOptions.value,
      categoryOptions: categoryOptions.value,
      parentOptions: parentOptions.value,
    })
    .open();
}

async function onDelete(row: KnowledgePointApi.KnowledgePoint) {
  const hideLoading = message.loading({
    content: `正在删除 ${row.name}...`,
    duration: 0,
    key: 'action_process_msg',
  });
  try {
    await deleteKnowledgePoint(row.id);
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
  formDrawerApi
    .setData({
      examOptions: examOptions.value,
      subjectOptions: subjectOptions.value,
      categoryOptions: categoryOptions.value,
      parentOptions: parentOptions.value,
    })
    .open();
}
</script>
<template>
  <Page auto-content-height>
    <FormDrawer @success="onRefresh" />
    <Grid table-title="知识考点列表">
      <template #toolbar-tools>
        <Button
          v-if="hasAccessByCodes(['question:knowledge:add'])"
          type="primary"
          @click="onCreate"
        >
          <Plus class="size-5" />
          新增知识点
        </Button>
      </template>
      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: '编辑',
              icon: 'lucide:edit',
              onClick: () => onEdit(row),
              auth: ['question:knowledge:edit'],
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
              auth: ['question:knowledge:delete'],
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
  </Page>
</template>
