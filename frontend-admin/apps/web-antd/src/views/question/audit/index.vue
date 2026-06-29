<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { QuestionApi } from '#/api/question/question';

import { computed, ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { useUserStore } from '@vben/stores';

import { Input, message, Modal } from 'ant-design-vue';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  approveQuestion,
  getQuestionList,
  rejectQuestion,
} from '#/api/question/question';

import QuestionPreview from '../list/modules/preview.vue';
import { useAuditColumns, useAuditFormSchema } from './data';

const userStore = useUserStore();
const currentUserId = computed(() => String(userStore.userInfo?.userId || '0'));

const [PreviewDrawer, previewDrawerApi] = useVbenDrawer({
  connectedComponent: QuestionPreview,
  destroyOnClose: true,
});

function onPreview(row: QuestionApi.Question) {
  previewDrawerApi.setData(row).open();
}

// 判断当前用户是否是题目创建者
function isCreator(row: QuestionApi.Question) {
  return currentUserId.value === String(row.createdBy);
}

const rejectVisible = ref(false);
const rejectRow = ref<QuestionApi.Question | null>(null);
const rejectReason = ref('');

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
            ...formValues,
            page: page.currentPage,
            pageSize: page.pageSize,
            status: formValues.status || 'pending',
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
    message.error('操作失败');
  }
}

function onReject(row: QuestionApi.Question) {
  rejectRow.value = row;
  rejectReason.value = '';
  rejectVisible.value = true;
}

function onRejectConfirm() {
  if (!rejectReason.value.trim()) {
    message.warning('请输入驳回原因');
    return;
  }
  Modal.confirm({
    title: '确认驳回',
    content: `确定要驳回该题目吗？原因：${rejectReason.value}`,
    async onOk() {
      try {
        await rejectQuestion(rejectRow.value!.id, rejectReason.value);
        message.success('已驳回');
        rejectVisible.value = false;
        onRefresh();
      } catch {
        message.error('操作失败');
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
    <PreviewDrawer />
    <Grid table-title="审核队列">
      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: '预览',
              icon: 'lucide:eye',
              onClick: () => onPreview(row),
            },
            {
              text: '通过',
              icon: 'lucide:check-circle',
              onClick: () => onApprove(row),
              auth: ['question:audit:approve'],
              ifShow: row.status === 'pending',
              disabled: isCreator(row),
              tooltip: isCreator(row) ? '不能审核自己创建的题目' : undefined,
            },
            {
              text: '驳回',
              icon: 'lucide:x-circle',
              onClick: () => onReject(row),
              auth: ['question:audit:reject'],
              ifShow: row.status === 'pending',
              danger: true,
              disabled: isCreator(row),
              tooltip: isCreator(row) ? '不能审核自己创建的题目' : undefined,
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
    <Modal
      v-model:open="rejectVisible"
      title="驳回题目"
      ok-text="确认驳回"
      @ok="onRejectConfirm"
    >
      <p>请输入驳回原因：</p>
      <Input.TextArea
        v-model:value="rejectReason"
        :rows="4"
        placeholder="请输入驳回原因"
      />
    </Modal>
  </Page>
</template>
