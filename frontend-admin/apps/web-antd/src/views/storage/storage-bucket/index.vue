<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import { useAccess } from '@vben/access';

import {
  Badge,
  Button,
  Form,
  FormItem,
  Input,
  message,
  Modal,
  Popconfirm,
  Select,
  SelectOption,
  Space,
  Switch,
  Table,
  Tag,
} from 'ant-design-vue';
import { IconifyIcon, Plus } from '@vben/icons';

import {
  createStorageBucketApi,
  deleteStorageBucketApi,
  getAllStorageBucketsApi,
  getEnabledDriversApi,
  setDefaultStorageBucketApi,
  testStorageBucketByDriverApi,
  testStorageBucketConnectionApi,
  updateStorageBucketApi,
  type StorageBucketApi,
} from '#/api/system/storage-bucket';

const { hasAccessByCodes } = useAccess();

// 数据列表
const bucketList = ref<StorageBucketApi.StorageBucket[]>([]);
const loading = ref(false);
const testLoading = ref(false);
const enabledDrivers = ref<
  Array<{ value: string; label: string; icon: string; enabled: boolean }>
>([]);

// 弹窗控制
const modalVisible = ref(false);
const modalTitle = ref('');
const isEditing = ref(false);
const currentBucket = ref<Partial<StorageBucketApi.StorageBucket> & StorageBucketApi.CreateStorageBucket>({
  name: '',
  driver: 'local',
  endpoint: '',
  bucket: '',
  accessKey: '',
  secretKey: '',
  region: '',
  useSsl: false,
  cdnDomain: '',
  pathPrefix: '',
  purpose: '',
  isDefault: false,
  status: 1,
  description: '',
});

// 驱动选项（含颜色）
const driverColors: Record<string, string> = {
  local: '#52c41a',
  minio: '#1677ff',
  oss: '#ff6a00',
  cos: '#006eff',
};

// 过滤后的驱动选项（只显示已启用的 + local）
const driverOptions = computed(() => {
  return enabledDrivers.value
    .filter((d) => d.enabled || d.value === 'local')
    .map((d) => ({
      value: d.value,
      label: `${d.icon} ${d.label}`,
      icon: d.icon,
      color: driverColors[d.value] || '#8c8c8c',
    }));
});

// 用途选项
const purposeOptions = [
  { value: 'file', label: '文件管理', color: '#1677ff' },
  { value: 'backup', label: '系统备份', color: '#faad14' },
  { value: 'avatar', label: '用户头像', color: '#722ed1' },
  { value: 'temp', label: '临时文件', color: '#8c8c8c' },
];

// 统计数据
const stats = computed(() => {
  const total = bucketList.value.length;
  const enabled = bucketList.value.filter((b) => b.status === 1).length;
  const disabled = total - enabled;
  const defaultBucket = bucketList.value.find((b) => b.isDefault);

  // 按驱动统计
  const byDriver: Record<string, number> = {};
  for (const b of bucketList.value) {
    byDriver[b.driver] = (byDriver[b.driver] || 0) + 1;
  }

  return { total, enabled, disabled, defaultBucket, byDriver };
});

// 表格列
const columns = [
  { title: '名称', dataIndex: 'name', width: 180 },
  { title: '存储驱动', key: 'driver', width: 140 },
  { title: '桶/路径', dataIndex: 'bucket', width: 180 },
  { title: '用途', key: 'purpose', width: 120 },
  { title: '状态', key: 'status', width: 100 },
  { title: '默认', key: 'isDefault', width: 100 },
  { title: '描述', dataIndex: 'description', ellipsis: true },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const },
];

// 加载数据
const fetchBuckets = async () => {
  loading.value = true;
  try {
    bucketList.value = await getAllStorageBucketsApi();
  } catch (error) {
    console.error('获取存储桶列表失败:', error);
    message.error('获取存储桶列表失败');
  } finally {
    loading.value = false;
  }
};

// 加载已启用的驱动
const loadEnabledDrivers = async () => {
  try {
    enabledDrivers.value = await getEnabledDriversApi();
  } catch (error) {
    console.error('获取启用驱动失败:', error);
  }
};

onMounted(() => {
  fetchBuckets();
  loadEnabledDrivers();
});

// 获取驱动信息
const getDriverInfo = (driver: string) => {
  const opt = enabledDrivers.value.find((d) => d.value === driver);
  if (opt) {
    return { label: opt.label, color: driverColors[opt.value] || '#8c8c8c', icon: opt.icon };
  }
  return { label: driver, color: '#8c8c8c', icon: '❓' };
};

// 测试连接（支持添加和编辑模式）
const handleTestConnection = async () => {
  if (currentBucket.value.driver === 'local') {
    message.info('本地存储无需测试连接');
    return;
  }
  if (!currentBucket.value.bucket) {
    message.warning('请先填写 Bucket 名称');
    return;
  }
  testLoading.value = true;
  try {
    if (isEditing.value && currentBucket.value.id) {
      // 编辑模式：用已保存的桶测试
      const res = await testStorageBucketConnectionApi(currentBucket.value.id);
      message.success(res.message || '连接成功');
    } else {
      // 添加模式：按驱动+桶名测试
      const res = await testStorageBucketByDriverApi({
        driver: currentBucket.value.driver,
        bucketName: currentBucket.value.bucket,
        region: currentBucket.value.region,
      });
      message.success(res.message || '连接成功');
    }
  } catch (error: any) {
    message.error(error?.message || '连接测试失败');
  } finally {
    testLoading.value = false;
  }
};

// 获取用途信息
const getPurposeInfo = (purpose: string) => {
  return purposeOptions.find((item) => item.value === purpose) || { label: purpose || '未指定', color: '#d9d9d9' };
};

// 重置表单
const resetForm = () => {
  currentBucket.value = {
    name: '',
    driver: 'local',
    endpoint: '',
    bucket: '',
    accessKey: '',
    secretKey: '',
    region: '',
    useSsl: false,
    cdnDomain: '',
    pathPrefix: '',
    purpose: '',
    isDefault: false,
    status: 1,
    description: '',
  };
};

// 添加
const handleAdd = () => {
  modalTitle.value = '添加存储桶';
  isEditing.value = false;
  resetForm();
  modalVisible.value = true;
};

// 编辑
const handleEdit = (record: any) => {
  modalTitle.value = '编辑存储桶';
  isEditing.value = true;
  currentBucket.value = { ...record };
  modalVisible.value = true;
};

// 保存
const handleSave = async () => {
  if (!currentBucket.value.name) {
    message.warning('请输入存储桶名称');
    return;
  }
  if (currentBucket.value.driver !== 'local' && !currentBucket.value.bucket) {
    message.warning('请输入 Bucket 名称');
    return;
  }

  try {
    if (isEditing.value && currentBucket.value.id) {
      await updateStorageBucketApi(currentBucket.value.id, currentBucket.value);
      message.success('更新成功');
    } else {
      await createStorageBucketApi(currentBucket.value);
      message.success('添加成功');
    }
    modalVisible.value = false;
    fetchBuckets();
  } catch (error: any) {
    message.error(error?.message || '操作失败');
  }
};

// 删除（由 Popconfirm 确认后直接执行）
const handleDelete = async (record: StorageBucketApi.StorageBucket) => {
  try {
    await deleteStorageBucketApi(record.id);
    message.success('删除成功');
    fetchBuckets();
  } catch (error: any) {
    message.error(error?.message || '删除失败');
  }
};

// 设置默认
const handleSetDefault = async (record: any) => {
  try {
    await setDefaultStorageBucketApi(record.id);
    message.success('设置成功');
    fetchBuckets();
  } catch (error: any) {
    message.error(error?.message || '设置失败');
  }
};

// 取消
const handleCancel = () => {
  modalVisible.value = false;
};
</script>

<template>
  <div class="p-4">
    <!-- 页面标题 -->
    <div class="mb-4">
      <h2 class="text-xl font-bold">存储桶管理</h2>
      <p class="text-gray-500 mt-1">
        配置可用的存储位置。存储桶会被「标签路由」引用，用于决定文件上传后存到哪里。
      </p>
    </div>

    <!-- 统计卡片 -->
    <div class="mb-4 grid grid-cols-2 gap-3 lg:grid-cols-4">
      <div class="rounded-lg border border-gray-200 bg-white p-3">
        <div class="text-xs text-gray-500">总存储桶</div>
        <div class="mt-1 text-2xl font-bold text-blue-600">{{ stats.total }}</div>
      </div>
      <div class="rounded-lg border border-gray-200 bg-white p-3">
        <div class="text-xs text-gray-500">已启用</div>
        <div class="mt-1 text-2xl font-bold text-green-600">{{ stats.enabled }}</div>
      </div>
      <div class="rounded-lg border border-gray-200 bg-white p-3">
        <div class="text-xs text-gray-500">已禁用</div>
        <div class="mt-1 text-2xl font-bold text-gray-400">{{ stats.disabled }}</div>
      </div>
      <div class="rounded-lg border border-gray-200 bg-white p-3">
        <div class="text-xs text-gray-500">默认桶</div>
        <div class="mt-1 text-sm font-bold text-amber-600 truncate">
          {{ stats.defaultBucket ? stats.defaultBucket.name : '未设置' }}
        </div>
      </div>
    </div>

    <!-- 使用说明 -->
    <div class="mb-4 rounded-lg border border-blue-200 bg-blue-50 p-3 text-sm text-blue-700">
      <div class="font-medium mb-1">💡 存储桶怎么用？</div>
      <ul class="list-disc pl-5 space-y-0.5">
        <li><b>存储桶</b> = 一个存储位置（本地、MinIO、阿里云 OSS、腾讯云 COS）</li>
        <li><b>标签路由</b> = 根据文件类型自动选择存到哪个桶（如：图片→OSS，文档→本地）</li>
        <li>添加存储桶后，去「标签路由」页面创建规则，即可实现自动分类存储</li>
      </ul>
    </div>

    <!-- 操作栏 -->
    <div class="mb-4 flex justify-between items-center">
      <Button v-if="hasAccessByCodes(['storage:bucket:edit'])" type="primary" @click="handleAdd">
        <Plus class="mr-1" />
        添加存储桶
      </Button>
      <Button @click="fetchBuckets">
        <IconifyIcon icon="mdi:reload" class="mr-1" />
        刷新
      </Button>
    </div>

    <!-- 数据表格 -->
    <Table
      :columns="columns"
      :data-source="bucketList"
      :loading="loading"
      row-key="id"
      :scroll="{ x: 1200 }"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'driver'">
          <Tag :color="getDriverInfo(record.driver).color">
            {{ getDriverInfo(record.driver).icon }} {{ getDriverInfo(record.driver).label }}
          </Tag>
        </template>
        <template v-else-if="column.key === 'purpose'">
          <Tag :color="getPurposeInfo(record.purpose).color">
            {{ getPurposeInfo(record.purpose).label }}
          </Tag>
        </template>
        <template v-else-if="column.key === 'status'">
          <Badge :status="record.status === 1 ? 'success' : 'default'" :text="record.status === 1 ? '启用' : '禁用'" />
        </template>
        <template v-else-if="column.key === 'isDefault'">
          <Tag v-if="record.isDefault" color="gold">⭐ 默认</Tag>
          <Button v-else-if="hasAccessByCodes(['storage:bucket:edit']) && record.status === 1" type="link" size="small" @click="handleSetDefault(record)">设为默认</Button>
        </template>
        <template v-else-if="column.key === 'action'">
          <Space>
            <Button v-if="hasAccessByCodes(['storage:bucket:edit'])" type="link" size="small" @click="handleEdit(record)">编辑</Button>
            <Popconfirm
              v-if="!record.isDefault && hasAccessByCodes(['storage:bucket:delete'])"
              title="确定删除?"
              @confirm="handleDelete(record)"
            >
              <Button type="link" size="small" danger>删除</Button>
            </Popconfirm>
          </Space>
        </template>
      </template>
    </Table>

    <!-- 添加/编辑弹窗 -->
    <Modal
      v-model:open="modalVisible"
      :title="modalTitle"
      :width="600"
      @cancel="handleCancel"
    >
      <Form layout="vertical" class="mt-4">
        <FormItem label="存储桶名称" required>
          <Input v-model:value="currentBucket.name" placeholder="如：生产环境图片桶" />
        </FormItem>
        <FormItem label="存储驱动" required>
          <Select v-model:value="currentBucket.driver" :options="driverOptions" placeholder="选择驱动" />
        </FormItem>

        <template v-if="currentBucket.driver !== 'local'">
          <FormItem label="Bucket 名称" required>
            <Input v-model:value="currentBucket.bucket" placeholder="如：my-bucket" />
          </FormItem>
          <div class="mb-4 flex items-center gap-2">
            <Button
              :loading="testLoading"
              size="small"
              @click="handleTestConnection"
            >
              测试连接
            </Button>
            <span class="text-xs text-gray-400">
              连接信息从「存储配置」自动获取
            </span>
          </div>
        </template>

        <FormItem label="路径前缀">
          <Input v-model:value="currentBucket.pathPrefix" placeholder="可选，如：files/" />
        </FormItem>

        <FormItem label="用途">
          <Select v-model:value="currentBucket.purpose" placeholder="选择用途" allow-clear>
            <SelectOption v-for="opt in purposeOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectOption>
          </Select>
        </FormItem>

        <FormItem label="设为默认">
          <Switch v-model:checked="currentBucket.isDefault" />
        </FormItem>

        <FormItem label="状态">
          <Switch v-model:checked="currentBucket.status" :checked-value="1" :un-checked-value="0" />
        </FormItem>

        <FormItem label="描述">
          <Input.TextArea v-model:value="currentBucket.description" placeholder="备注信息" :rows="2" />
        </FormItem>
      </Form>

      <template #footer>
        <Button @click="handleCancel">取消</Button>
        <Button type="primary" @click="handleSave">保存</Button>
      </template>
    </Modal>
  </div>
</template>
