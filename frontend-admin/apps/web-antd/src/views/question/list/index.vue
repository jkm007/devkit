<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { QuestionApi } from '#/api/question/question';

import { computed } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';
import { useUserStore } from '@vben/stores';

import { Button, message } from 'ant-design-vue';

import { useAccess } from '@vben/access';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  archiveQuestion,
  deleteQuestion,
  getQuestionList,
  publishQuestion,
  reactivateQuestion,
  submitAuditQuestion,
  withdrawQuestion,
} from '#/api/question/question';

import { useQuestionColumns, useQuestionFormSchema } from './data';
import QuestionForm from './modules/form.vue';
import QuestionPreview from './modules/preview.vue';

const { hasAccessByCodes } = useAccess();
const userStore = useUserStore();
const currentUserId = computed(() => Number(userStore.userInfo?.userId || 0));

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

async function onWithdraw(row: QuestionApi.Question) {
  try {
    await withdrawQuestion(row.id);
    message.success('已撤回到草稿');
    onRefresh();
  } catch {
    message.error('操作失败');
  }
}

async function onReactivate(row: QuestionApi.Question) {
  try {
    await reactivateQuestion(row.id);
    message.success('已重新上架');
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

// 判断当前用户是否是题目创建者
function isCreator(row: QuestionApi.Question) {
  return String(currentUserId.value) === String(row.createdBy);
}

// 根据 status + resourceType + 是否创建者 生成操作按钮
// 主按钮只显示 预览 + 1个主要操作，其余放更多下拉
function getActions(row: QuestionApi.Question) {
  const { status } = row;
  const creator = isCreator(row);
  const actions: any[] = [];

  // 预览 - 所有状态都有
  actions.push({
    text: '预览',
    icon: 'lucide:eye',
    onClick: () => onPreview(row),
  });

  switch (status) {
    case 'draft': {
      if (creator) {
        actions.push({
          text: '编辑',
          icon: 'lucide:edit',
          onClick: () => onEdit(row),
          auth: ['question:edit'],
        });
      }
      break;
    }

    case 'rejected': {
      if (creator) {
        actions.push({
          text: '编辑',
          icon: 'lucide:edit',
          onClick: () => onEdit(row),
          auth: ['question:edit'],
        });
      }
      break;
    }
  }

  return actions;
}

function getDropdownActions(row: QuestionApi.Question) {
  const { status, resourceType } = row;
  const isPrivate = resourceType === 'private';
  const creator = isCreator(row);
  const dropdown: any[] = [];

  switch (status) {
    case 'draft': {
      if (creator) {
        if (isPrivate) {
          dropdown.push({
            text: '发布',
            icon: 'lucide:check-circle',
            onClick: () => onPublish(row),
            auth: ['question:publish'],
          });
        } else {
          dropdown.push({
            text: '提交审核',
            icon: 'lucide:send',
            onClick: () => onSubmitAudit(row),
            auth: ['question:audit:submit'],
          });
        }
        dropdown.push({
          text: '删除',
          icon: 'lucide:trash-2',
          danger: true,
          popConfirm: {
            title: `确定删除【${row.title}】？`,
            confirm: () => onDelete(row),
          },
          auth: ['question:delete'],
        });
      }
      break;
    }

    case 'pending': {
      if (creator) {
        dropdown.push({
          text: '撤回到草稿',
          icon: 'lucide:undo',
          onClick: () => onWithdraw(row),
        });
      }
      break;
    }

    case 'approved': {
      if (creator) {
        dropdown.push({
          text: '发布',
          icon: 'lucide:check-circle',
          onClick: () => onPublish(row),
          auth: ['question:publish'],
        });
        dropdown.push({
          text: '撤回到草稿',
          icon: 'lucide:undo',
          onClick: () => onWithdraw(row),
        });
      }
      break;
    }

    case 'rejected': {
      if (creator) {
        dropdown.push({
          text: '重新提交审核',
          icon: 'lucide:send',
          onClick: () => onSubmitAudit(row),
          auth: ['question:audit:submit'],
        });
        dropdown.push({
          text: '删除',
          icon: 'lucide:trash-2',
          danger: true,
          popConfirm: {
            title: `确定删除【${row.title}】？`,
            confirm: () => onDelete(row),
          },
          auth: ['question:delete'],
        });
      }
      break;
    }

    case 'published': {
      if (creator) {
        dropdown.push({
          text: '下架',
          icon: 'lucide:archive',
          onClick: () => onArchive(row),
        });
        dropdown.push({
          text: '撤回到草稿',
          icon: 'lucide:undo',
          onClick: () => onWithdraw(row),
        });
      }
      break;
    }

    case 'archived': {
      if (creator) {
        dropdown.push({
          text: '重新上架',
          icon: 'lucide:refresh-cw',
          onClick: () => onReactivate(row),
        });
        dropdown.push({
          text: '删除',
          icon: 'lucide:trash-2',
          danger: true,
          popConfirm: {
            title: `确定删除【${row.title}】？`,
            confirm: () => onDelete(row),
          },
          auth: ['question:delete'],
        });
      }
      break;
    }
  }

  return dropdown;
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
          :actions="getActions(row)"
          :dropdown-actions="getDropdownActions(row)"
          align="center"
        />
      </template>
    </Grid>
  </Page>
</template>
