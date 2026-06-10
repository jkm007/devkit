<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import dayjs from 'dayjs';

import { Page } from '@vben/common-ui';
import { useAccessStore } from '@vben/stores';

import {
  Button,
  Card,
  DatePicker,
  Dropdown,
  InputNumber,
  Menu,
  MenuItem,
  message,
  Modal,
  Radio,
  Space,
  Table,
  Tag,
  Tooltip,
} from 'ant-design-vue';
import InputSearch from 'ant-design-vue/es/input/Search';

import {
  getUserShares,
  renewShare,
  expireShare,
  updateShareExpiry,
  disableShare,
  enableShare,
  deleteShare,
} from '#/api/file';
import type { ShareListItem } from '#/api/file';

defineOptions({ name: 'ShareList' });

const accessStore = useAccessStore();

// ==================== 权限检查 ====================

const permissions = computed(() => accessStore.accessCodes || []);
const hasViewAllPermission = computed(() => permissions.value.includes('file:view:all'));
// @ts-ignore - 暂时未使用，保留以备将来使用
const hasSharePermission = computed(() => permissions.value.includes('file:share'));
const hasDeletePermission = computed(() => permissions.value.includes('share:delete'));
const hasManagePermission = computed(() => permissions.value.includes('share:manage'));

// ==================== 状态 ====================

const loading = ref(false);
const shareList = ref<ShareListItem[]>([]);
const totalShares = ref(0);
const pagination = ref({ current: 1, pageSize: 20 });
const selectedRowKeys = ref<number[]>([]);

// 分享范围：own=自己的分享, all=所有分享
const shareScope = ref<'all' | 'own'>('own');

// 筛选
const statusFilter = ref<number | undefined>(undefined);
const searchText = ref('');

// 续签弹窗
const renewModalVisible = ref(false);
const renewShareId = ref<number | null>(null);
const renewExpireHours = ref(24);

// 修改到期时间弹窗
const expiryModalVisible = ref(false);
const expiryShareId = ref<number | null>(null);
const expiryDate = ref<any>(null);

// 批量修改状态弹窗
const batchStatusModalVisible = ref(false);
const batchStatusValue = ref<number>(3); // 3=禁用

// ==================== 计算属性 ====================

const hasBatchPermission = computed(() => hasDeletePermission.value || hasManagePermission.value);

const filteredShareList = computed(() => {
  let list = shareList.value;
  if (statusFilter.value !== undefined) {
    list = list.filter(item => item.status === statusFilter.value);
  }
  if (searchText.value) {
    const keyword = searchText.value.toLowerCase();
    list = list.filter(item =>
      (item.fileName || '').toLowerCase().includes(keyword) ||
      (item.folderName || '').toLowerCase().includes(keyword) ||
      (item.shareCode || '').toLowerCase().includes(keyword)
    );
  }
  return list;
});

const statusCounts = computed(() => {
  const counts = { total: shareList.value.length, active: 0, expired: 0, disabled: 0 };
  for (const item of shareList.value) {
    if (item.status === 1) counts.active++;
    else if (item.status === 2) counts.expired++;
    else if (item.status === 3) counts.disabled++;
  }
  return counts;
});

// ==================== 加载 ====================

async function loadShareList() {
  loading.value = true;
  try {
    const result = await getUserShares({
      page: pagination.value.current,
      pageSize: pagination.value.pageSize,
      scope: shareScope.value,
    });
    shareList.value = result?.items || [];
    totalShares.value = result?.total || 0;
  } catch {
    message.error('加载分享列表失败');
  } finally {
    loading.value = false;
  }
}

// ==================== 操作 ====================

// 复制分享链接
function copyShareUrl(share: any) {
  const url = getFullShareUrl(share);
  fallbackCopy(url);
}

// 获取完整的分享 URL（确保包含 origin）
function getFullShareUrl(share: any) {
  const url = share.shareUrl || `/share/${share.shareCode}`;
  if (url.startsWith('http://') || url.startsWith('https://')) {
    return url;
  }
  return `${window.location.origin}${url.startsWith('/') ? '' : '/'}${url}`;
}

function fallbackCopy(text: string) {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(text).then(
      () => message.success('链接已复制'),
      () => {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();
        try {
          document.execCommand('copy');
          message.success('链接已复制');
        } catch {
          message.error('复制失败，请手动复制');
        }
        document.body.removeChild(textarea);
      },
    );
  } else {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);
    textarea.select();
    try {
      document.execCommand('copy');
      message.success('链接已复制');
    } catch {
      message.error('复制失败，请手动复制');
    }
    document.body.removeChild(textarea);
  }
}

// 菜单操作处理
function handleMenuAction(key: string, record: any) {
  switch (key) {
    case 'renew':
      openRenewModal(record.id);
      break;
    case 'expiry':
      openExpiryModal(record);
      break;
    case 'expire':
      handleExpire(record.id);
      break;
    case 'disable':
      handleDisable(record.id);
      break;
    case 'enable':
      handleEnable(record.id);
      break;
    case 'delete':
      handleDelete(record.id);
      break;
  }
}

// 打开续签弹窗
function openRenewModal(id: number) {
  renewShareId.value = id;
  renewExpireHours.value = 24;
  renewModalVisible.value = true;
}

// 确认续签
async function confirmRenew() {
  if (!renewShareId.value) return;
  try {
    await renewShare(renewShareId.value, renewExpireHours.value);
    message.success('续签成功');
    renewModalVisible.value = false;
    loadShareList();
  } catch (err: any) {
    message.error(err.message || '续签失败');
  }
}

// 立即过期
async function handleExpire(id: number) {
  try {
    await expireShare(id);
    message.success('已设为过期');
    loadShareList();
  } catch (err: any) {
    message.error(err.message || '操作失败');
  }
}

// 打开修改到期时间弹窗
function openExpiryModal(share: ShareListItem) {
  expiryShareId.value = share.id;
  expiryDate.value = share.expireAt ? dayjs(share.expireAt) : null;
  expiryModalVisible.value = true;
}

// 确认修改到期时间
async function confirmExpiry() {
  if (!expiryShareId.value) return;
  try {
    const expireAtStr = expiryDate.value ? expiryDate.value.toISOString() : undefined;
    await updateShareExpiry(expiryShareId.value, expireAtStr);
    message.success('到期时间已更新');
    expiryModalVisible.value = false;
    loadShareList();
  } catch (err: any) {
    message.error(err.message || '操作失败');
  }
}

// 禁用分享
async function handleDisable(id: number) {
  try {
    await disableShare(id);
    message.success('已禁用');
    loadShareList();
  } catch (err: any) {
    message.error(err.message || '操作失败');
  }
}

// 启用分享
async function handleEnable(id: number) {
  try {
    await enableShare(id);
    message.success('已启用');
    loadShareList();
  } catch (err: any) {
    message.error(err.message || '操作失败');
  }
}

// 删除分享
async function handleDelete(id: number) {
  try {
    await deleteShare(id);
    message.success('已删除');
    loadShareList();
  } catch (err: any) {
    message.error(err.message || '删除失败');
  }
}

// 批量删除
async function handleBatchDelete() {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择要删除的分享');
    return;
  }

  try {
    let successCount = 0;
    let failCount = 0;

    for (const id of selectedRowKeys.value) {
      try {
        await deleteShare(id);
        successCount++;
      } catch {
        failCount++;
      }
    }

    if (failCount > 0) {
      message.warning(`已删除 ${successCount} 个，${failCount} 个失败`);
    } else {
      message.success(`已删除 ${successCount} 个分享`);
    }

    selectedRowKeys.value = [];
    loadShareList();
  } catch (err: any) {
    message.error(err.message || '批量删除失败');
  }
}

// 批量修改状态
async function handleBatchStatus() {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择要操作的分享');
    return;
  }

  try {
    let successCount = 0;
    let failCount = 0;

    for (const id of selectedRowKeys.value) {
      try {
        if (batchStatusValue.value === 2) {
          await expireShare(id);
        } else if (batchStatusValue.value === 3) {
          await disableShare(id);
        } else if (batchStatusValue.value === 1) {
          await enableShare(id);
        }
        successCount++;
      } catch {
        failCount++;
      }
    }

    if (failCount > 0) {
      message.warning(`已处理 ${successCount} 个，${failCount} 个失败`);
    } else {
      message.success(`已处理 ${successCount} 个分享`);
    }

    selectedRowKeys.value = [];
    batchStatusModalVisible.value = false;
    loadShareList();
  } catch (err: any) {
    message.error(err.message || '批量操作失败');
  }
}

// 打开批量状态弹窗
function openBatchStatusModal(status: number) {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择要操作的分享');
    return;
  }
  batchStatusValue.value = status;
  batchStatusModalVisible.value = true;
}

// ==================== 工具函数 ====================

function formatFileSize(size: number | undefined | null) {
  if (size === undefined || size === null || isNaN(size)) return '-';
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`;
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`;
}

function formatDate(date: string | undefined | null) {
  if (!date) return '永久';
  return new Date(date).toLocaleString();
}

function getStatusTag(status: number) {
  switch (status) {
    case 1:
      return { color: 'green', text: '有效' };
    case 2:
      return { color: 'orange', text: '已过期' };
    case 3:
      return { color: 'red', text: '已禁用' };
    default:
      return { color: 'default', text: '未知' };
  }
}

function getFileIcon(type: string | undefined) {
  if (!type) return 'i-ant-design:file-outlined';
  if (type.startsWith('image/')) return 'i-ant-design:file-image-outlined';
  if (type.startsWith('video/')) return 'i-ant-design:file-video-outlined';
  if (type.startsWith('audio/')) return 'i-ant-design:sound-outlined';
  if (type.includes('pdf')) return 'i-ant-design:file-pdf-outlined';
  if (type.includes('word') || type.includes('document')) return 'i-ant-design:file-word-outlined';
  if (type.includes('excel') || type.includes('spreadsheet')) return 'i-ant-design:file-excel-outlined';
  if (type.includes('zip') || type.includes('rar')) return 'i-ant-design:file-zip-outlined';
  return 'i-ant-design:file-outlined';
}

// @ts-ignore - 暂时未使用，保留以备将来使用
function getShareUrl(share: any) {
  return getFullShareUrl(share);
}

function isExpiringSoon(date: string | undefined) {
  if (!date) return false;
  const expireTime = new Date(date).getTime();
  const now = Date.now();
  const oneDay = 24 * 60 * 60 * 1000;
  return expireTime > now && expireTime - now < oneDay;
}

// ==================== 表格列 ====================

const columns = [
  {
    title: '分享内容',
    key: 'name',
    width: 220,
  },
  {
    title: '类型',
    key: 'type',
    width: 80,
    align: 'center' as const,
  },
  {
    title: '大小',
    key: 'size',
    width: 100,
    align: 'center' as const,
  },
  {
    title: '状态',
    key: 'status',
    width: 90,
    align: 'center' as const,
  },
  {
    title: '访问统计',
    key: 'access',
    width: 120,
    align: 'center' as const,
  },
  {
    title: '过期时间',
    key: 'expireAt',
    width: 160,
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: 160,
  },
  {
    title: '操作',
    key: 'operation',
    width: 150,
    fixed: 'right' as const,
    align: 'center' as const,
  },
];

// ==================== 初始化 ====================

onMounted(() => {
  loadShareList();
});
</script>

<template>
  <Page title="分享管理">
    <Card>
      <!-- 顶部统计卡片 -->
      <div class="mb-4 flex items-center gap-4 flex-wrap">
        <div
          class="flex-1 min-w-[120px] px-4 py-3 rounded-lg cursor-pointer transition-colors"
          :class="statusFilter === undefined ? 'bg-blue-50 border border-blue-200' : 'bg-gray-50 hover:bg-gray-100'"
          @click="statusFilter = undefined"
        >
          <div class="text-2xl font-bold text-gray-800">{{ statusCounts.total }}</div>
          <div class="text-sm text-gray-500">全部分享</div>
        </div>
        <div
          class="flex-1 min-w-[120px] px-4 py-3 rounded-lg cursor-pointer transition-colors"
          :class="statusFilter === 1 ? 'bg-green-50 border border-green-200' : 'bg-gray-50 hover:bg-gray-100'"
          @click="statusFilter = statusFilter === 1 ? undefined : 1"
        >
          <div class="text-2xl font-bold text-green-600">{{ statusCounts.active }}</div>
          <div class="text-sm text-gray-500">有效</div>
        </div>
        <div
          class="flex-1 min-w-[120px] px-4 py-3 rounded-lg cursor-pointer transition-colors"
          :class="statusFilter === 2 ? 'bg-orange-50 border border-orange-200' : 'bg-gray-50 hover:bg-gray-100'"
          @click="statusFilter = statusFilter === 2 ? undefined : 2"
        >
          <div class="text-2xl font-bold text-orange-500">{{ statusCounts.expired }}</div>
          <div class="text-sm text-gray-500">已过期</div>
        </div>
        <div
          class="flex-1 min-w-[120px] px-4 py-3 rounded-lg cursor-pointer transition-colors"
          :class="statusFilter === 3 ? 'bg-red-50 border border-red-200' : 'bg-gray-50 hover:bg-gray-100'"
          @click="statusFilter = statusFilter === 3 ? undefined : 3"
        >
          <div class="text-2xl font-bold text-red-500">{{ statusCounts.disabled }}</div>
          <div class="text-sm text-gray-500">已禁用</div>
        </div>
      </div>

      <!-- 工具栏 -->
      <div class="mb-4 flex items-center justify-between gap-4 flex-wrap">
        <div class="flex items-center gap-3">
          <!-- 分享范围切换 -->
          <Radio.Group v-if="hasViewAllPermission" v-model:value="shareScope" button-style="solid" @change="loadShareList">
            <Radio.Button value="own">我的分享</Radio.Button>
            <Radio.Button value="all">所有分享</Radio.Button>
          </Radio.Group>

          <!-- 搜索 -->
          <InputSearch
            v-model:value="searchText"
            placeholder="搜索文件名或分享码"
            allow-clear
            style="width: 240px"
          />
        </div>

        <!-- 批量操作 -->
        <Space v-if="selectedRowKeys.length > 0 && hasBatchPermission">
          <span class="text-blue-500">已选 {{ selectedRowKeys.length }} 项</span>
          <Button v-if="hasDeletePermission" size="small" danger @click="handleBatchDelete">
            批量删除
          </Button>
          <Button v-if="hasManagePermission" size="small" @click="openBatchStatusModal(3)">
            批量禁用
          </Button>
          <Button v-if="hasManagePermission" size="small" @click="openBatchStatusModal(1)">
            批量启用
          </Button>
        </Space>
      </div>

      <!-- 分享列表表格 -->
      <Table
        :columns="columns"
        :data-source="filteredShareList"
        :loading="loading"
        :pagination="{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: totalShares,
          showSizeChanger: true,
          showTotal: (total: number) => `共 ${total} 条`,
        }"
        :scroll="{ x: 1100 }"
        :row-selection="hasBatchPermission ? { selectedRowKeys, onChange: (keys: any) => selectedRowKeys = keys as number[] } : undefined"
        row-key="id"
        @change="(pag: any) => { pagination.current = pag.current; pagination.pageSize = pag.pageSize; loadShareList(); }"
      >
        <template #bodyCell="{ column, record }">
          <!-- 分享内容 -->
          <template v-if="column.key === 'name'">
            <div class="flex items-center gap-2">
              <span
                v-if="record.type === 'file'"
                :class="getFileIcon(record.contentType)"
                class="text-lg flex-shrink-0"
                :style="{ color: record.contentType?.startsWith('image/') ? '#8b5cf6' : record.contentType?.startsWith('video/') ? '#ef4444' : record.contentType?.startsWith('audio/') ? '#f59e0b' : record.contentType?.includes('pdf') ? '#ef4444' : '#6b7280' }"
              />
              <span v-else class="i-ant-design:folder-outlined text-lg text-yellow-500 flex-shrink-0" />
              <Tooltip :title="record.fileName || record.folderName">
                <span class="truncate max-w-[160px]">{{ record.fileName || record.folderName || '-' }}</span>
              </Tooltip>
            </div>
          </template>

          <!-- 类型 -->
          <template v-if="column.key === 'type'">
            <Tag :color="record.type === 'file' ? 'blue' : 'orange'" size="small">
              {{ record.type === 'file' ? '文件' : '文件夹' }}
            </Tag>
          </template>

          <!-- 大小 -->
          <template v-if="column.key === 'size'">
            <span class="text-gray-600">
              {{ record.type === 'file' ? formatFileSize(record.fileSize) : '-' }}
            </span>
          </template>

          <!-- 状态 -->
          <template v-if="column.key === 'status'">
            <Tag :color="getStatusTag(record.status).color" size="small">
              {{ getStatusTag(record.status).text }}
            </Tag>
          </template>

          <!-- 访问统计 -->
          <template v-if="column.key === 'access'">
            <Tooltip :title="record.accessedAt ? `最后访问: ${formatDate(record.accessedAt)}` : '暂无访问'">
              <div class="flex flex-col items-center">
                <span class="font-medium">{{ record.accessCount }}</span>
                <span v-if="record.maxAccess > 0" class="text-xs text-gray-400">
                  / {{ record.maxAccess }}
                </span>
              </div>
            </Tooltip>
          </template>

          <!-- 过期时间 -->
          <template v-if="column.key === 'expireAt'">
            <span
              :class="{
                'text-orange-500 font-medium': record.status === 2,
                'text-red-500': isExpiringSoon(record.expireAt) && record.status === 1,
              }"
            >
              <span v-if="isExpiringSoon(record.expireAt) && record.status === 1" class="i-ant-design:warning-outlined mr-1" />
              {{ formatDate(record.expireAt) }}
            </span>
          </template>

          <!-- 创建时间 -->
          <template v-if="column.key === 'createdAt'">
            <span class="text-gray-500">{{ formatDate(record.createdAt) }}</span>
          </template>

          <!-- 操作 -->
          <template v-if="column.key === 'operation'">
            <Space size="small">
              <Button type="link" size="small" @click="copyShareUrl(record)">
                复制链接
              </Button>

              <Dropdown :trigger="['click']">
                <Button type="link" size="small">
                  更多
                </Button>
                <template #overlay>
                  <Menu @click="({ key }: any) => handleMenuAction(key, record)">
                    <!-- 有效状态的操作 -->
                    <template v-if="record.status === 1 && hasManagePermission">
                      <MenuItem key="renew">续签</MenuItem>
                      <MenuItem key="expiry">修改到期时间</MenuItem>
                      <MenuItem key="expire">立即过期</MenuItem>
                      <MenuItem key="disable" class="text-red-500">禁用</MenuItem>
                    </template>

                    <!-- 过期/禁用状态的操作 -->
                    <template v-if="(record.status === 2 || record.status === 3) && hasManagePermission">
                      <MenuItem key="renew">续签</MenuItem>
                      <MenuItem key="expiry">修改到期时间</MenuItem>
                      <MenuItem v-if="record.status === 3" key="enable">启用</MenuItem>
                    </template>

                    <!-- 删除 -->
                    <MenuItem v-if="hasDeletePermission" key="delete" class="text-red-500">删除</MenuItem>
                  </Menu>
                </template>
              </Dropdown>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 续签弹窗 -->
    <Modal
      v-model:open="renewModalVisible"
      title="续签分享"
      @ok="confirmRenew"
    >
      <div class="py-4">
        <div class="mb-4">
          <span class="mr-2">续签时长：</span>
          <InputNumber
            v-model:value="renewExpireHours"
            :min="1"
            :max="8760"
            addon-after="小时"
          />
        </div>
        <div class="text-gray-500 text-sm">
          续签后，分享链接将在指定时长后过期。
        </div>
      </div>
    </Modal>

    <!-- 修改到期时间弹窗 -->
    <Modal
      v-model:open="expiryModalVisible"
      title="修改到期时间"
      @ok="confirmExpiry"
    >
      <div class="py-4">
        <div class="mb-4">
          <span class="mr-2">到期时间：</span>
          <DatePicker
            v-model:value="expiryDate"
            show-time
            placeholder="选择到期时间（留空为永久有效）"
            style="width: 300px"
          />
        </div>
        <div class="text-gray-500 text-sm">
          设置为永久有效请留空。
        </div>
      </div>
    </Modal>

    <!-- 批量修改状态弹窗 -->
    <Modal
      v-model:open="batchStatusModalVisible"
      title="批量修改状态"
      @ok="handleBatchStatus"
    >
      <div class="py-4">
        <p class="mb-4">确定要将选中的 {{ selectedRowKeys.length }} 个分享修改为以下状态吗？</p>
        <div class="flex items-center gap-4">
          <span>目标状态：</span>
          <Tag :color="batchStatusValue === 1 ? 'green' : batchStatusValue === 2 ? 'orange' : 'red'">
            {{ batchStatusValue === 1 ? '启用' : batchStatusValue === 2 ? '过期' : '禁用' }}
          </Tag>
        </div>
      </div>
    </Modal>
  </Page>
</template>
