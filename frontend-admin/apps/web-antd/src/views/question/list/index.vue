<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { QuestionApi } from '#/api/question/question';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message } from 'ant-design-vue';

import { useAccess } from '@vben/access';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  archiveQuestion,
  deleteQuestion,
  getQuestionList,
  publishQuestion,
  submitAuditQuestion,
} from '#/api/question/question';

import { useQuestionColumns, useQuestionFormSchema } from './data';
import QuestionForm from './modules/form.vue';
import QuestionPreview from './modules/preview.vue';

const { hasAccessByCodes } = useAccess();

const [FormDrawer, formDrawerApi] = useVbenDrawer({
  connectedComponent: QuestionForm,
  destroyOnClose: true,
});

const [PreviewDrawer, previewDrawerApi] = useVbenDrawer({
  connectedComponent: QuestionPreview,
  destroyOnClose: true,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: useQuestionFormSchema(),
    submitOnChange: true,
  },
  gridOptions: {
    columns: useQuestionColumns(),
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getQuestionList({
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
  } as VxeTableGridOptions<QuestionApi.Question>,
});

function onEdit(row: QuestionApi.Question) {
  formDrawerApi.setData(row).open();
}

function onPreview(row: QuestionApi.Question) {
  previewDrawerApi.setData(row).open();
}

async function onDelete(row: QuestionApi.Question) {
  const hideLoading = message.loading({
    content: `正在删除 ${row.title}...`,
    duration: 0,
    key: 'action_process_msg',
  });
  try {
    await deleteQuestion(row.id);
    message.success({
      content: `${row.title} 删除成功`,
      key: 'action_process_msg',
    });
    onRefresh();
  } catch {
    hideLoading();
    message.error('操作失败');
  }
}

async function onPublish(row: QuestionApi.Question) {
  try {
    await publishQuestion(row.id);
    message.success('发布成功');
    onRefresh();
  } catch {
    message.error('操作失败');
  }
}

async function onSubmitAudit(row: QuestionApi.Question) {
  try {
    await submitAuditQuestion(row.id);
    message.success('已提交审核');
    onRefresh();
  } catch {
    message.error('操作失败');
  }
}

async function onArchive(row: QuestionApi.Question) {
  try {
    await archiveQuestion(row.id);
    message.success('已下架');
    onRefresh();
  } catch {
    message.error('操作失败');
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
    <PreviewDrawer />
    <Grid table-title="题目列表">
      <template #toolbar-tools>
        <Button
          v-if="hasAccessByCodes(['question:create'])"
          type="primary"
          @click="onCreate"
        >
          <Plus class="size-5" />
          新增题目
        </Button>
      </template>
      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: '预览',
              icon: 'lucide:eye',
              onClick: () => onPreview(row),
            },
            {
              text: '编辑',
              icon: 'lucide:edit',
              onClick: () => onEdit(row),
              auth: ['question:edit'],
            },
            {
              text: row.status === 'published' ? '下架' : '发布',
              icon:
                row.status === 'published'
                  ? 'lucide:archive'
                  : 'lucide:check-circle',
              onClick: () =>
                row.status === 'published'
                  ? onArchive(row)
                  : onPublish(row),
              auth: ['question:publish'],
              ifShow: row.status === 'draft' || row.status === 'published',
            },
            {
              text: '提交审核',
              icon: 'lucide:send',
              onClick: () => onSubmitAudit(row),
              auth: ['question:audit:submit'],
              ifShow: row.status === 'draft' || row.status === 'rejected',
            },
          ]"
          :dropdown-actions="[
            {
              text: '删除',
              icon: 'lucide:trash-2',
              danger: true,
              popConfirm: {
                title: `确定删除【${row.title}】？`,
                confirm: () => onDelete(row),
              },
              auth: ['question:delete'],
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
  </Page>
</template>
