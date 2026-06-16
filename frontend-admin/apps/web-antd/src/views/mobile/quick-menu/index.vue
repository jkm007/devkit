<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { QuickMenuItem } from '#/api/system/mobile-config';

import { ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal, Switch, Tag } from 'ant-design-vue';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  getQuickMenuList,
  deleteQuickMenu,
} from '#/api/system/mobile-config';

import QuickMenuModal from './modules/quick-menu-modal.vue';

const [QuickMenuDrawer, quickMenuDrawerApi] = useVbenDrawer({
  connectedComponent: QuickMenuModal,
  destroyOnClose: true,
});

const columns: VxeTableGridOptions<QuickMenuItem>['columns'] = [
  {
    title: 'ID',
    field: 'id',
    width: 80,
  },
  {
    title: '图标',
    field: 'icon',
    width: 80,
    slots: {
      default: 'icon',
    },
  },
  {
    title: '标题',
    field: 'title',
    width: 150,
  },
  {
    title: '链接',
    field: 'link',
    minWidth: 200,
    showOverflow: 'ellipsis',
  },
  {
    title: '链接类型',
    field: 'linkType',
    width: 100,
    slots: {
      default: 'linkType',
    },
  },
  {
    title: '排序',
    field: 'sortOrder',
    width: 80,
  },
  {
    title: '状态',
    field: 'status',
    width: 100,
    slots: {
      default: 'status',
    },
  },
  {
    title: '操作',
    field: 'action',
    width: 200,
    fixed: 'right',
    slots: {
      default: 'action',
    },
  },
];

const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions: {
    columns,
    height: 'auto',
    keepSource: true,
    proxyConfig: {
      ajax: {
        query: async ({ page }) => {
          const data = await getQuickMenuList({
            page: page.currentPage,
            pageSize: page.pageSize,
          });
          return {
            items: data || [],
            total: (data || []).length,
          };
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
      zoom: true,
    },
  } as VxeTableGridOptions<QuickMenuItem>,
});

function handleCreate() {
  quickMenuDrawerApi.setData(null).open();
}

function handleEdit(record: QuickMenuItem) {
  quickMenuDrawerApi.setData(record).open();
}

async function handleDelete(id: number) {
  Modal.confirm({
    title: '确认删除',
    content: '确定删除此快捷菜单吗？',
    onOk: async () => {
      try {
        await deleteQuickMenu(id);
        message.success('删除成功');
        gridApi.reload();
      } catch (error: any) {
        message.error(error.message || '删除失败');
      }
    },
  });
}

function handleModalSuccess() {
  gridApi.reload();
}
</script>

<template>
  <Page
    auto-content-height
    description="管理移动端首页快捷入口菜单"
    title="快捷菜单管理"
  >
    <template #extra>
      <Button type="primary" @click="handleCreate">
        <Plus class="mr-2 size-4" />
        新增快捷菜单
      </Button>
    </template>

    <Grid>
      <template #icon="{ row }">
        <div class="flex items-center justify-center">
          <span class="text-2xl">{{ row.icon }}</span>
        </div>
      </template>
      <template #linkType="{ row }">
        <Tag v-if="row.linkType === 'page'" color="blue">页面跳转</Tag>
        <Tag v-else-if="row.linkType === 'url'" color="orange">外部链接</Tag>
        <Tag v-else-if="row.linkType === 'function'" color="green">功能</Tag>
        <Tag v-else color="default">无链接</Tag>
      </template>
      <template #status="{ row }">
        <Switch
          :checked="row.status === 'enabled'"
          checked-children="启用"
          un-checked-children="禁用"
          disabled
        />
      </template>
      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            {
              text: '编辑',
              icon: 'lucide:edit',
              onClick: () => handleEdit(row),
            },
          ]"
          :dropdown-actions="[
            {
              text: '删除',
              icon: 'lucide:trash-2',
              danger: true,
              popConfirm: {
                title: '确定删除此快捷菜单吗？',
                confirm: () => handleDelete(row.id),
              },
            },
          ]"
        />
      </template>
    </Grid>

    <QuickMenuDrawer @success="handleModalSuccess" />
  </Page>
</template>
