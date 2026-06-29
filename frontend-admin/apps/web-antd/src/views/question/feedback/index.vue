<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { FeedbackApi } from '#/api/question/feedback';

import { ref } from 'vue';

import { Page } from '@vben/common-ui';

import { message, Tag } from 'ant-design-vue';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import { getFeedbackList, updateFeedback } from '#/api/question/feedback';

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: 'Select',
        fieldName: 'status',
        label: '状态',
        componentProps: {
          options: [
            { label: '待处理', value: 'pending' },
            { label: '处理中', value: 'processing' },
            { label: '已解决', value: 'resolved' },
            { label: '已关闭', value: 'closed' },
          ],
          placeholder: '全部',
          allowClear: true,
        },
      },
    ],
    submitOnChange: true,
  },
  gridOptions: {
    columns: [
      { field: 'id', title: 'ID', width: 60 },
      { field: 'questionId', title: '题目ID', width: 80 },
      { field: 'userNickname', title: '反馈用户', width: 120 },
      { field: 'feedbackType', title: '类型', width: 100, slots: { default: 'type' } },
      { field: 'content', title: '反馈内容', minWidth: 200 },
      { field: 'status', title: '状态', width: 100, slots: { default: 'status' } },
      { field: 'reply', title: '回复', minWidth: 150 },
      { field: 'createdAt', title: '创建时间', width: 180 },
      { field: 'action', title: '操作', width: 200, fixed: 'right', slots: { default: 'action' } },
    ],
    height: 'auto',
    keepSource: true,
    pagerConfig: { pageSize: 20, pageSizeOpts: [10, 20, 50] },
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getFeedbackList({
            page: page.currentPage,
            pageSize: page.pageSize,
            ...(formValues as any),
          });
        },
      },
    },
    rowConfig: { keyField: 'id' },
    toolbarConfig: {
      custom: true,
      refresh: true,
      search: true,
      zoom: true,
    },
  } as VxeTableGridOptions<FeedbackApi.Feedback>,
});

const typeLabel = (type: string) => {
  const map: Record<string, string> = { error: '纠错', suggestion: '建议', other: '其他' };
  return map[type] || type;
};

const statusTag = (status: string) => {
  const map: Record<string, { color: string; text: string }> = {
    pending: { color: 'orange', text: '待处理' },
    processing: { color: 'blue', text: '处理中' },
    resolved: { color: 'green', text: '已解决' },
    closed: { color: 'default', text: '已关闭' },
  };
  return map[status] || { color: 'default', text: status };
};

async function onStatusChange(row: FeedbackApi.Feedback, newStatus: string) {
  try {
    await updateFeedback(row.id, { status: newStatus });
    message.success('状态已更新');
    gridApi.reload();
  } catch {
    message.error('更新失败');
  }
}

const statusOptions = [
  { label: '待处理', value: 'pending' },
  { label: '处理中', value: 'processing' },
  { label: '已解决', value: 'resolved' },
  { label: '已关闭', value: 'closed' },
];
</script>

<template>
  <Page auto-content-height title="题目反馈">
    <Grid table-title="题目反馈/纠错列表" table-title-help="处理用户提交的题目纠错与建议">
      <template #type="{ row }">
        <Tag :color="row.feedbackType === 'error' ? 'red' : row.feedbackType === 'suggestion' ? 'blue' : 'default'">
          {{ typeLabel(row.feedbackType) }}
        </Tag>
      </template>

      <template #status="{ row }">
        <Tag :color="statusTag(row.status).color">
          {{ statusTag(row.status).text }}
        </Tag>
      </template>

      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: '标记处理中',
              icon: 'lucide:loader',
              ifShow: () => row.status === 'pending',
              onClick: () => onStatusChange(row, 'processing'),
            },
            {
              text: '标记已解决',
              icon: 'lucide:check',
              ifShow: () => row.status !== 'resolved' && row.status !== 'closed',
              onClick: () => onStatusChange(row, 'resolved'),
            },
            {
              text: '关闭',
              icon: 'lucide:x',
              ifShow: () => row.status !== 'closed',
              onClick: () => onStatusChange(row, 'closed'),
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
  </Page>
</template>
