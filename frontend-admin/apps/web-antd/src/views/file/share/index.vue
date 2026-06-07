<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';
import { useAccessStore } from '@vben/stores';

import {
  Button,
  Card,
  DatePicker,
  InputNumber,
  message,
  Modal,
  Popconfirm,
  Radio,
  Space,
  Table,
  Tag,
} from 'ant-design-vue';

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
const hasSharePermission = computed(() => permissions.value.includes('file:share'));
const hasDeletePermission = computed(() => permissions.value.includes('file:delete'));
const hasManagePermission = computed(() => permissions.value.includes('file:manage'));

// ==================== 状态 ====================

const loading = ref(false);
const shareList = ref<ShareListItem[]>([]);
const totalShares = ref(0);
const pagination = ref({ current: 1, pageSize: 20 });
const selectedRowKeys = ref<number[]>([]);

// 分享范围：own=自己的分享, all=所有分享
const shareScope = ref<'all' | 'own'>('own');

// 续签弹窗
const renewModalVisible = ref(false);
const renewShareId = ref<number | null>(null);
const renewExpireHours = ref(24);

// 修改到期时间弹窗
const expiryModalVisible = ref(false);
const expiryShareId = ref<number | null>(null);
const expiryDate = ref<string>('');

// 批量修改状态弹窗
const batchStatusModalVisible = ref(false);
const batchStatusValue = ref<number>(3); // 3=禁用

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
function copyShareUrl(share: ShareListItem) {
  const url = `${window.location.origin}/share/${share.shareCode}`;
  fallbackCopy(url);
}

function fallbackCopy(text: string) {
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
  expiryDate.value = share.expireAt || '';
  expiryModalVisible.value = true;
}

// 确认修改到期时间
async function confirmExpiry() {
  if (!expiryShareId.value) return;
  try {
    await updateShareExpiry(expiryShareId.value, expiryDate.value || undefined);
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
  if (!size) return '-';
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
  return 'i-ant-design:file-outlined';
}

// ==================== 表格列 ====================

const columns = [
  {
    title: '文件/文件夹',
    key: 'name',
    width: 200,
  },
  {
    title: '类型',
    key: 'type',
    width: 80,
  },
  {
    title: '大小',
    key: 'size',
    width: 100,
  },
  {
    title: '分享人',
    key: 'sharer',
    width: 120,
  },
  {
    title: '状态',
    key: 'status',
    width: 80,
  },
  {
    title: '访问次数',
    dataIndex: 'accessCount',
    key: 'accessCount',
    width: 100,
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
    width: 250,
    fixed: 'right' as const,
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
      <!-- 统计信息和批量操作 -->
      <div class="mb-4 flex items-center justify-between">
        <div class="flex items-center gap-4">
          <!-- 分享范围切换 -->
          <div v-if="hasViewAllPermission">
            <Radio.Group v-model:value="shareScope" button-style="solid" @change="loadShareList">
              <Radio.Button value="own">我的分享</Radio.Button>
              <Radio.Button value="all">所有分享</Radio.Button>
            </Radio.Group>
          </div>
          <span class="text-gray-500">共 {{ totalShares }} 个分享</span>
          <Space v-if="selectedRowKeys.length > 0 && (hasDeletePermission || hasManagePermission)">
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
            <Button v-if="hasManagePermission" size="small" @click="openBatchStatusModal(2)">
              批量过期
            </Button>
          </Space>
        </div>
      </div>

      <!-- 分享列表表格 -->
      <Table
        :columns="columns"
        :data-source="shareList"
        :loading="loading"
        :pagination="{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: totalShares,
          showSizeChanger: true,
        }"
        :scroll="{ x: 1200 }"
        :row-selection="(hasDeletePermission || hasManagePermission) ? { selectedRowKeys, onChange: (keys) => selectedRowKeys = keys } : undefined"
        row-key="id"
        @change="(pag) => { pagination = pag; loadShareList(); }"
      >
        <template #bodyCell="{ column, record }">
          <!-- 文件/文件夹名 -->
          <template v-if="column.key === 'name'">
            <div class="flex items-center">
              <span v-if="record.type === 'file'" :class="getFileIcon(record.contentType)" class="mr-2 text-lg" />
              <span v-else class="i-ant-design:folder-outlined mr-2 text-lg text-yellow-500" />
              <span class="truncate">{{ record.fileName || record.folderName || '-' }}</span>
            </div>
          </template>

          <!-- 类型 -->
          <template v-if="column.key === 'type'">
            <Tag :color="record.type === 'file' ? 'blue' : 'orange'">
              {{ record.type === 'file' ? '文件' : '文件夹' }}
            </Tag>
          </template>

          <!-- 大小 -->
          <template v-if="column.key === 'size'">
            {{ record.type === 'file' ? formatFileSize(record.fileSize) : '-' }}
          </template>

          <!-- 分享人 -->
          <template v-if="column.key === 'sharer'">
            <div class="flex items-center">
              <span v-if="record.userAvatar" class="mr-1 inline-block w-5 h-5 rounded-full bg-gray-200 overflow-hidden">
                <img :src="record.userAvatar" class="w-full h-full object-cover" />
              </span>
              <span v-else class="mr-1 inline-block w-5 h-5 rounded-full bg-blue-100 text-blue-600 text-xs flex items-center justify-center">
                {{ (record.userName || '?')[0] }}
              </span>
              <span class="text-sm">{{ record.userName || '-' }}</span>
            </div>
          </template>

          <!-- 状态 -->
          <template v-if="column.key === 'status'">
            <Tag :color="getStatusTag(record.status).color">
              {{ getStatusTag(record.status).text }}
            </Tag>
          </template>

          <!-- 过期时间 -->
          <template v-if="column.key === 'expireAt'">
            <span :class="{ 'text-orange-500': record.status === 2 }">
              {{ formatDate(record.expireAt) }}
            </span>
          </template>

          <!-- 创建时间 -->
          <template v-if="column.key === 'createdAt'">
            {{ formatDate(record.createdAt) }}
          </template>

          <!-- 操作 -->
          <template v-if="column.key === 'operation'">
            <Space size="small">
              <Button type="link" size="small" @click="copyShareUrl(record)">
                复制链接
              </Button>

              <!-- 有效状态的操作 -->
              <template v-if="record.status === 1 && hasManagePermission">
                <Button type="link" size="small" @click="openRenewModal(record.id)">
                  续签
                </Button>
                <Button type="link" size="small" @click="openExpiryModal(record)">
                  改期
                </Button>
                <Popconfirm
                  title="确定要立即过期此分享吗？"
                  @confirm="handleExpire(record.id)"
                >
                  <Button type="link" size="small" danger>
                    过期
                  </Button>
                </Popconfirm>
                <Popconfirm
                  title="确定要禁用此分享吗？"
                  @confirm="handleDisable(record.id)"
                >
                  <Button type="link" size="small" danger>
                    禁用
                  </Button>
                </Popconfirm>
              </template>

              <!-- 过期/禁用状态的操作 -->
              <template v-if="(record.status === 2 || record.status === 3) && hasManagePermission">
                <Button type="link" size="small" @click="openRenewModal(record.id)">
                  续签
                </Button>
                <Button type="link" size="small" @click="openExpiryModal(record)">
                  改期
                </Button>
                <Popconfirm
                  v-if="record.status === 3"
                  title="确定要启用此分享吗？"
                  @confirm="handleEnable(record.id)"
                >
                  <Button type="primary" size="small">
                    启用
                  </Button>
                </Popconfirm>
              </template>

              <!-- 删除 -->
              <Popconfirm
                v-if="hasDeletePermission"
                title="确定要删除此分享吗？删除后无法恢复。"
                @confirm="handleDelete(record.id)"
              >
                <Button type="link" size="small" danger>
                  删除
                </Button>
              </Popconfirm>
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
