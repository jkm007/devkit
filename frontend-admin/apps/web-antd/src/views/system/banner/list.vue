<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { Banner } from '#/api/system/banner';

import { ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, message, Modal, Switch, Tag } from 'ant-design-vue';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import { getBannerList, deleteBanner, updateBannerStatus } from '#/api/system/banner';
import { $t } from '#/locales';

import BannerModal from './modules/banner-modal.vue';

const [BannerDrawer, bannerDrawerApi] = useVbenDrawer({
  connectedComponent: BannerModal,
  destroyOnClose: true,
});

const handleStatusChange = async (id: number, status: string) => {
  try {
    await updateBannerStatus(id, status);
    message.success('状态更新成功');
    gridApi.reload();
  } catch (error: any) {
    message.error(error.message || '状态更新失败');
  }
};

const columns: VxeTableGridOptions<Banner>['columns'] = [
  {
    title: 'ID',
    field: 'id',
    width: 80,
  },
  {
    title: '标题',
    field: 'title',
    width: 200,
  },
  {
    title: '图片',
    field: 'image',
    width: 150,
    slots: {
      default: 'image_slot',
    },
  },
  {
    title: '链接',
    field: 'link',
    width: 200,
    showOverflow: 'ellipsis',
  },
  {
    title: '链接类型',
    field: 'linkType',
    width: 100,
    slots: {
      default: 'linkType_slot',
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
      default: 'status_slot',
    },
  },
  {
    title: '创建时间',
    field: 'createdAt',
    width: 180,
    formatter: ({ cellValue }) => {
      return new Date(cellValue).toLocaleString();
    },
  },
  {
    title: '操作',
    field: 'action',
    width: 200,
    fixed: 'right',
    slots: {
      default: 'action_slot',
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
          const data = await getBannerList({
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
  } as VxeTableGridOptions<Banner>,
});

function handleCreate() {
  bannerDrawerApi.open({
    mode: 'create',
    data: null,
  });
}

function handleEdit(record: Banner) {
  bannerDrawerApi.open({
    mode: 'edit',
    data: record,
  });
}

async function handleDelete(id: number) {
  Modal.confirm({
    title: '确认删除',
    content: '确定删除此轮播图吗？',
    onOk: async () => {
      try {
        await deleteBanner(id);
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
    description="管理首页轮播图"
    title="轮播图管理"
  >
    <template #extra>
      <Button type="primary" @click="handleCreate">
        <Plus class="mr-2 size-4" />
        新增轮播图
      </Button>
    </template>

    <Grid>
      <template #image_slot="{ row }">
        <img
          :src="row.image"
          style="max-width: 100px; max-height: 50px; object-fit: cover;"
        />
      </template>
      <template #linkType_slot="{ row }">
        <Tag v-if="row.linkType === 'internal'" color="blue">内部链接</Tag>
        <Tag v-else-if="row.linkType === 'external'" color="orange">外部链接</Tag>
        <Tag v-else color="default">无链接</Tag>
      </template>
      <template #status_slot="{ row }">
        <Switch
          :checked="row.status === 'enabled'"
          checked-children="启用"
          un-checked-children="禁用"
          @change="(val: boolean) => handleStatusChange(row.id, val ? 'enabled' : 'disabled')"
        />
      </template>
      <template #action_slot="{ row }">
        <VbenTableAction
          :actions="[
            {
              label: '编辑',
              onClick: () => handleEdit(row),
            },
            {
              label: '删除',
              danger: true,
              onClick: () => handleDelete(row.id),
            },
          ]"
        />
      </template>
    </Grid>

    <BannerDrawer @success="handleModalSuccess" />
  </Page>
</template>