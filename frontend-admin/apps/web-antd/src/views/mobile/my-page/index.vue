<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { MyPageItem } from '#/api/system/mobile-config';

import { ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal, Switch, Tag } from 'ant-design-vue';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  getMyPageList,
  deleteMyPageMenu,
} from '#/api/system/mobile-config';

import MyPageModal from './modules/my-page-modal.vue';

const [MyPageDrawer, myPageDrawerApi] = useVbenDrawer({
  connectedComponent: MyPageModal,
  destroyOnClose: true,
});

const columns: VxeTableGridOptions<MyPageItem>['columns'] = [
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
    title: '角标',
    field: 'showBadge',
    width: 100,
    slots: {
      default: 'badge',
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
          const data = await getMyPageList({
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
  } as VxeTableGridOptions<MyPageItem>,
});

function handleCreate() {
  myPageDrawerApi.open({
    mode: 'create',
    data: null,
  });
}

function handleEdit(record: MyPageItem) {
  myPageDrawerApi.open({
    mode: 'edit',
    data: record,
  });
}

async function handleDelete(id: number) {
  Modal.confirm({
    title: '确认删除',
    content: '确定删除此菜单项吗？',
    onOk: async () => {
      try {
        await deleteMyPageMenu(id);
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
    description="管理移动端我的页面显示的功能菜单"
    title="我的页面配置"
  >
    <template #extra>
      <Button type="primary" @click="handleCreate">
        <Plus class="mr-2 size-4" />
        新增菜单项
      </Button>
    </template>

    <Grid>
      <template #icon="{ row }">
        <div class="flex items-center justify-center">
          <span class="text-2xl">{{ row.icon }}</span>
        </div>
      </template>
      <template #badge="{ row }">
        <Tag v-if="row.showBadge" color="red">{{ row.badgeText || 'NEW' }}</Tag>
        <span v-else class="text-gray-400">无</span>
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
                title: '确定删除此菜单项吗？',
                confirm: () => handleDelete(row.id),
              },
            },
          ]"
        />
      </template>
    </Grid>

    <MyPageDrawer @success="handleModalSuccess" />
  </Page>
</template>
