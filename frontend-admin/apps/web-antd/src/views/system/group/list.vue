<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemGroupApi } from '#/api/system/group';

import { Page, useVbenModal } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { useAccess } from '@vben/access';

import { Button, message } from 'ant-design-vue';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import { deleteGroup, getGroupList } from '#/api/system/group';
import { $t } from '#/locales';

import { useColumns } from './data';
import Form from './modules/form.vue';

const { hasAccessByCodes } = useAccess();

const [FormModal, formModalApi] = useVbenModal({
  connectedComponent: Form,
  destroyOnClose: true,
});

/**
 * 编辑分组
 */
function onEdit(row: SystemGroupApi.SystemGroup) {
  formModalApi.setData(row).open();
}

/**
 * 添加下级分组
 */
function onAppend(row: SystemGroupApi.SystemGroup) {
  formModalApi.setData({ pid: row.id }).open();
}

/**
 * 创建新分组
 */
function onCreate() {
  formModalApi.setData(null).open();
}

/**
 * 删除分组
 */
function onDelete(row: SystemGroupApi.SystemGroup) {
  const hideLoading = message.loading({
    content: $t('ui.actionMessage.deleting', [row.name]),
    duration: 0,
    key: 'action_process_msg',
  });
  deleteGroup(row.id)
    .then(() => {
      message.success({
        content: $t('ui.actionMessage.deleteSuccess', [row.name]),
        key: 'action_process_msg',
      });
      refreshGrid();
    })
    .catch(() => {
      hideLoading();
    });
}

const [Grid, gridApi] = useVbenVxeGrid({
  gridEvents: {},
  gridOptions: {
    columns: useColumns(),
    height: 'auto',
    keepSource: true,
    pagerConfig: {
      enabled: false,
    },
    proxyConfig: {
      ajax: {
        query: async (_params) => {
          return await getGroupList();
        },
      },
    },
    toolbarConfig: {
      custom: true,
      export: false,
      refresh: true,
      zoom: true,
    },
    treeConfig: {
      parentField: 'pid',
      rowField: 'id',
      transform: false,
    },
  } as VxeTableGridOptions,
});

/**
 * 刷新表格
 */
function refreshGrid() {
  gridApi.query();
}
</script>
<template>
  <Page auto-content-height>
    <FormModal @success="refreshGrid" />
    <Grid table-title="分组列表">
      <template #toolbar-tools>
        <Button
          v-if="hasAccessByCodes(['system:group:add'])"
          type="primary"
          @click="onCreate"
        >
          <Plus class="size-5" />
          {{ $t('ui.actionTitle.create', [$t('system.group.name')]) }}
        </Button>
      </template>
      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: '添加下级',
              icon: 'lucide:plus',
              onClick: () => onAppend(row),
              auth: ['system:group:add'],
            },
            {
              text: $t('common.edit'),
              icon: 'lucide:edit',
              onClick: () => onEdit(row),
              auth: ['system:group:edit'],
            },
          ]"
          :dropdown-actions="[
            {
              text: $t('common.delete'),
              icon: 'lucide:trash-2',
              danger: true,
              popConfirm: {
                title: $t('ui.actionMessage.deleteConfirm', [row.name]),
                confirm: () => onDelete(row),
              },
              auth: ['system:group:delete'],
            },
          ]"
          align="center"
        />
      </template>
    </Grid>
  </Page>
</template>
