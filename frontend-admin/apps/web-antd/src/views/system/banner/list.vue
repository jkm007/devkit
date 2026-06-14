<template>
  <div>
    <VbenTable
      :data="tableData"
      :columns="columns"
      :loading="loading"
      :pagination="pagination"
      @page-change="handlePageChange"
    >
      <template #toolbar>
        <VbenButton type="primary" @click="handleCreate">
          新增轮播图
        </VbenButton>
      </template>
    </VbenTable>

    <!-- 创建/编辑弹窗 -->
    <BannerModal
      v-model:visible="modalVisible"
      :banner="currentBanner"
      :mode="modalMode"
      @success="handleModalSuccess"
    />
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { VbenButton, VbenTable } from '#/components';
import type { VbenTableColumn } from '#/components/table/types';
import { getBannerList, deleteBanner, updateBannerStatus } from '#/api/system/banner';
import type { Banner } from '#/api/system/banner';
import BannerModal from './modules/banner-modal.vue';

const loading = ref(false);
const tableData = ref<Banner[]>([]);
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
});

const modalVisible = ref(false);
const modalMode = ref<'create' | 'edit'>('create');
const currentBanner = ref<Banner | null>(null);

const columns: VbenTableColumn[] = [
  {
    title: 'ID',
    dataIndex: 'id',
    width: 80,
  },
  {
    title: '标题',
    dataIndex: 'title',
    width: 200,
  },
  {
    title: '图片',
    dataIndex: 'image',
    width: 150,
    render: (text: string) => {
      return `<img src="${text}" style="max-width: 100px; max-height: 50px;" />`;
    },
  },
  {
    title: '链接',
    dataIndex: 'link',
    width: 200,
    ellipsis: true,
  },
  {
    title: '链接类型',
    dataIndex: 'linkType',
    width: 100,
    render: (text: string) => {
      const map = {
        internal: '内部链接',
        external: '外部链接',
        none: '无链接',
      };
      return map[text] || text;
    },
  },
  {
    title: '排序',
    dataIndex: 'sortOrder',
    width: 80,
  },
  {
    title: '状态',
    dataIndex: 'status',
    width: 100,
    render: (text: string) => {
      const map = {
        enabled: '启用',
        disabled: '禁用',
      };
      return map[text] || text;
    },
  },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    width: 180,
    render: (text: string) => {
      return new Date(text).toLocaleString();
    },
  },
  {
    title: '操作',
    dataIndex: 'action',
    width: 200,
    fixed: 'right',
    render: (_, record: Banner) => {
      return `
        <a onclick="handleEdit(${record.id})">编辑</a>
        <a-divider type="vertical" />
        <a onclick="handleToggleStatus(${record.id}, '${record.status}')">
          ${record.status === 'enabled' ? '禁用' : '启用'}
        </a>
        <a-divider type="vertical" />
        <a onclick="handleDelete(${record.id})" style="color: red;">删除</a>
      `;
    },
  },
];

onMounted(() => {
  loadTableData();
});

async function loadTableData() {
  loading.value = true;
  try {
    const data = await getBannerList({
      page: pagination.value.current,
      pageSize: pagination.value.pageSize,
    });
    tableData.value = data || [];
    pagination.value.total = data.length;
  } catch (error: any) {
    console.error('加载Banner列表失败:', error);
  } finally {
    loading.value = false;
  }
}

function handlePageChange(page: number, pageSize: number) {
  pagination.value.current = page;
  pagination.value.pageSize = pageSize;
  loadTableData();
}

function handleCreate() {
  modalMode.value = 'create';
  currentBanner.value = null;
  modalVisible.value = true;
}

function handleEdit(id: number) {
  const banner = tableData.value.find(item => item.id === id);
  if (banner) {
    modalMode.value = 'edit';
    currentBanner.value = banner;
    modalVisible.value = true;
  }
}

async function handleToggleStatus(id: number, status: string) {
  const newStatus = status === 'enabled' ? 'disabled' : 'enabled';
  try {
    await updateBannerStatus(id, newStatus);
    loadTableData();
  } catch (error: any) {
    console.error('更新状态失败:', error);
  }
}

async function handleDelete(id: number) {
  if (confirm('确定删除此轮播图吗？')) {
    try {
      await deleteBanner(id);
      loadTableData();
    } catch (error: any) {
      console.error('删除失败:', error);
    }
  }
}

function handleModalSuccess() {
  modalVisible.value = false;
  loadTableData();
}
</script>