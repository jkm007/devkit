<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import { useAccess } from '@vben/access';
import { Page } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import {
  Badge,
  Button,
  Card,
  Col,
  Input,
  InputNumber,
  message,
  Modal,
  Popconfirm,
  Row,
  Select,
  SelectOption,
  Space,
  Switch,
  Table,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import {
  createStorageConfigApi,
  deleteStorageConfigApi,
  getAllStorageConfigsApi,
  setDefaultStorageConfigApi,
  testStorageConfigByDataApi,
  testStorageConfigConnectionApi,
  updateStorageConfigApi,
  type StorageConfigApi,
} from '#/api/system/storage-config';
import {
  createStorageBucketApi,
  deleteStorageBucketApi,
  getAllStorageBucketsApi,
  setDefaultStorageBucketApi,
  testStorageBucketByDriverApi,
  testStorageBucketConnectionApi,
  updateStorageBucketApi,
  type StorageBucketApi,
} from '#/api/system/storage-bucket';

const { hasAccessByCodes } = useAccess();
const canEdit = computed(() => hasAccessByCodes(['storage:bucket:edit', 'storage:config:edit']));

// ==================== State ====================
const activeTab = ref<'bucket' | 'config'>('config');
const loading = ref(false);
const saveLoading = ref(false);
const testLoading = ref(false);

const configList = ref<StorageConfigApi.StorageConfig[]>([]);
const bucketList = ref<StorageBucketApi.StorageBucket[]>([]);
const modalVisible = ref(false);
const editingId = ref<number | null>(null);

// ==================== Forms ====================
const configForm = reactive({
  name: '',
  driver: 'minio',
  endpoint: '',
  accessKey: '',
  secretKey: '',
  bucket: '',
  region: '',
  useSsl: false,
  cdnDomain: '',
  isDefault: false,
  presignedUrlExpiry: 3600,
  status: 1,
  description: '',
});

const bucketForm = reactive({
  name: '',
  driver: 'minio',
  bucket: '',
  pathPrefix: '',
  purpose: 'file',
  isDefault: false,
  status: 1,
  description: '',
});

// ==================== Driver Options ====================
const driverOptions = [
  { value: 'local', label: '本地存储', icon: '💾', color: 'green' },
  { value: 'minio', label: 'MinIO', icon: '🪣', color: 'blue' },
  { value: 'oss', label: '阿里云 OSS', icon: '☁️', color: 'orange' },
  { value: 'cos', label: '腾讯云 COS', icon: '🌐', color: 'cyan' },
];

const purposeOptions = [
  { value: 'file', label: '📁 文件管理' },
  { value: 'backup', label: '💾 系统备份' },
  { value: 'avatar', label: '👤 用户头像' },
  { value: 'temp', label: '📂 临时文件' },
];

// ==================== Config Table Columns ====================
const configColumns = [
  { title: '名称', dataIndex: 'name', width: 160 },
  { title: '驱动', dataIndex: 'driver', width: 100, key: 'driver' },
  { title: '连接信息', dataIndex: 'endpoint', width: 200, key: 'connection' },
  { title: '桶名', dataIndex: 'bucket', width: 140 },
  { title: '状态', dataIndex: 'status', width: 80, key: 'status' },
  { title: '默认', dataIndex: 'isDefault', width: 70, key: 'isDefault' },
  { title: '说明', dataIndex: 'description', ellipsis: true },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const },
];

// ==================== Bucket Table Columns ====================
const bucketColumns = [
  { title: '名称', dataIndex: 'name', width: 160 },
  { title: '驱动', dataIndex: 'driver', width: 100, key: 'driver' },
  { title: '桶 / 路径', dataIndex: 'bucket', width: 180, key: 'bucketPath' },
  { title: '用途', dataIndex: 'purpose', width: 120, key: 'purpose' },
  { title: '状态', dataIndex: 'status', width: 80, key: 'status' },
  { title: '默认', dataIndex: 'isDefault', width: 70, key: 'isDefault' },
  { title: '说明', dataIndex: 'description', ellipsis: true },
  { title: '操作', key: 'action', width: 200, fixed: 'right' as const },
];

// ==================== Stats ====================
const configStats = computed(() => {
  const total = configList.value.length;
  const enabled = configList.value.filter((c) => c.status === 1).length;
  const defaultCfg = configList.value.find((c) => c.isDefault);
  return { total, enabled, disabled: total - enabled, defaultName: defaultCfg?.name || '-' };
});

const bucketStats = computed(() => {
  const total = bucketList.value.length;
  const enabled = bucketList.value.filter((b) => b.status === 1).length;
  const defaultBkt = bucketList.value.find((b) => b.isDefault);
  return { total, enabled, disabled: total - enabled, defaultName: defaultBkt?.name || '-' };
});

// ==================== Helpers ====================
function getDriverInfo(driver: string) {
  return driverOptions.find((d) => d.value === driver) || { icon: '❓', label: driver, color: 'default' };
}

function getPurposeLabel(purpose: string) {
  return purposeOptions.find((p) => p.value === purpose)?.label || purpose;
}

function maskKey(key: string) {
  if (!key || key === '******') return '******';
  if (key.length <= 8) return '******';
  return `${key.slice(0, 4)}****${key.slice(-4)}`;
}

// ==================== Load Data ====================
async function loadConfigs() {
  try {
    configList.value = await getAllStorageConfigsApi();
  } catch { /* ignore */ }
}

async function loadBuckets() {
  try {
    bucketList.value = await getAllStorageBucketsApi();
  } catch { /* ignore */ }
}

async function loadData() {
  loading.value = true;
  try {
    await Promise.all([loadConfigs(), loadBuckets()]);
  } finally {
    loading.value = false;
  }
}

// ==================== Config CRUD ====================
function handleAddConfig() {
  editingId.value = null;
  Object.assign(configForm, {
    name: '', driver: 'minio', endpoint: '', accessKey: '', secretKey: '',
    bucket: '', region: '', useSsl: false, cdnDomain: '', isDefault: false,
    presignedUrlExpiry: 3600, status: 1, description: '',
  });
  modalVisible.value = true;
}

function handleEditConfig(record: StorageConfigApi.StorageConfig) {
  editingId.value = record.id;
  Object.assign(configForm, {
    name: record.name,
    driver: record.driver,
    endpoint: record.endpoint || '',
    accessKey: '******',
    secretKey: '******',
    bucket: record.bucket || '',
    region: record.region || '',
    useSsl: record.useSsl ?? false,
    cdnDomain: record.cdnDomain || '',
    isDefault: record.isDefault ?? false,
    presignedUrlExpiry: record.presignedUrlExpiry ?? 3600,
    status: record.status ?? 1,
    description: record.description || '',
  });
  modalVisible.value = true;
}

async function handleSaveConfig() {
  if (!configForm.name.trim()) { message.warning('请输入配置名称'); return; }
  saveLoading.value = true;
  try {
    const payload: any = { ...configForm };
    if (editingId.value) {
      if (payload.accessKey === '******') delete payload.accessKey;
      if (payload.secretKey === '******') delete payload.secretKey;
      await updateStorageConfigApi(editingId.value, payload);
      message.success('更新成功');
    } else {
      await createStorageConfigApi(payload);
      message.success('创建成功');
    }
    modalVisible.value = false;
    await loadConfigs();
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  } finally {
    saveLoading.value = false;
  }
}

async function handleDeleteConfig(id: number) {
  try {
    await deleteStorageConfigApi(id);
    message.success('删除成功');
    await loadConfigs();
  } catch (e: any) {
    message.error(e?.message || '删除失败');
  }
}

async function handleSetDefaultConfig(record: StorageConfigApi.StorageConfig) {
  try {
    await setDefaultStorageConfigApi(record.id);
    message.success(`已将「${record.name}」设为默认`);
    await loadConfigs();
  } catch (e: any) {
    message.error(e?.message || '设置失败');
  }
}

async function handleTestConfig(record?: StorageConfigApi.StorageConfig) {
  testLoading.value = true;
  try {
    if (record?.id) {
      await testStorageConfigConnectionApi(record.id);
    } else {
      await testStorageConfigByDataApi({ ...configForm } as any);
    }
    message.success('连接测试成功 ✓');
  } catch (e: any) {
    message.error(e?.message || '连接测试失败');
  } finally {
    testLoading.value = false;
  }
}

// ==================== Bucket CRUD ====================
function handleAddBucket() {
  editingId.value = null;
  Object.assign(bucketForm, {
    name: '', driver: 'minio', bucket: '', pathPrefix: '',
    purpose: 'file', isDefault: false, status: 1, description: '',
  });
  modalVisible.value = true;
}

function handleEditBucket(record: StorageBucketApi.StorageBucket) {
  editingId.value = record.id;
  Object.assign(bucketForm, {
    name: record.name,
    driver: record.driver,
    bucket: record.bucket || '',
    pathPrefix: record.pathPrefix || '',
    purpose: record.purpose || 'file',
    isDefault: record.isDefault ?? false,
    status: record.status ?? 1,
    description: record.description || '',
  });
  modalVisible.value = true;
}

async function handleSaveBucket() {
  if (!bucketForm.name.trim()) { message.warning('请输入桶名称'); return; }
  saveLoading.value = true;
  try {
    if (editingId.value) {
      await updateStorageBucketApi(editingId.value, { ...bucketForm } as any);
      message.success('更新成功');
    } else {
      await createStorageBucketApi({ ...bucketForm } as any);
      message.success('创建成功');
    }
    modalVisible.value = false;
    await loadBuckets();
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  } finally {
    saveLoading.value = false;
  }
}

async function handleDeleteBucket(id: number) {
  try {
    await deleteStorageBucketApi(id);
    message.success('删除成功');
    await loadBuckets();
  } catch (e: any) {
    message.error(e?.message || '删除失败');
  }
}

async function handleSetDefaultBucket(record: StorageBucketApi.StorageBucket) {
  try {
    await setDefaultStorageBucketApi(record.id);
    message.success(`已将「${record.name}」设为默认`);
    await loadBuckets();
  } catch (e: any) {
    message.error(e?.message || '设置失败');
  }
}

async function handleTestBucket(record?: StorageBucketApi.StorageBucket) {
  testLoading.value = true;
  try {
    if (record?.id) {
      await testStorageBucketConnectionApi(record.id);
    } else {
      await testStorageBucketByDriverApi({ driver: bucketForm.driver, bucket: bucketForm.bucket } as any);
    }
    message.success('连接测试成功 ✓');
  } catch (e: any) {
    message.error(e?.message || '连接测试失败');
  } finally {
    testLoading.value = false;
  }
}

// ==================== Save Handler (unified) ====================
function handleSave() {
  if (activeTab.value === 'config') handleSaveConfig();
  else handleSaveBucket();
}

function handleTest() {
  if (activeTab.value === 'config') handleTestConfig();
  else handleTestBucket();
}

// ==================== Lifecycle ====================
onMounted(() => loadData());
</script>

<template>
  <Page title="存储管理" auto-content-height>
    <!-- 统计卡片 -->
    <div class="mb-4 grid grid-cols-4 gap-4">
      <Card size="small">
        <div class="text-center">
          <div class="text-2xl font-bold text-blue-500">{{ activeTab === 'config' ? configStats.total : bucketStats.total }}</div>
          <div class="text-foreground/50 text-xs">总数</div>
        </div>
      </Card>
      <Card size="small">
        <div class="text-center">
          <div class="text-2xl font-bold text-green-500">{{ activeTab === 'config' ? configStats.enabled : bucketStats.enabled }}</div>
          <div class="text-foreground/50 text-xs">已启用</div>
        </div>
      </Card>
      <Card size="small">
        <div class="text-center">
          <div class="text-2xl font-bold text-orange-500">{{ activeTab === 'config' ? configStats.disabled : bucketStats.disabled }}</div>
          <div class="text-foreground/50 text-xs">已禁用</div>
        </div>
      </Card>
      <Card size="small">
        <div class="text-center">
          <div class="truncate text-sm font-medium text-blue-600">{{ activeTab === 'config' ? configStats.defaultName : bucketStats.defaultName }}</div>
          <div class="text-foreground/50 text-xs">当前默认</div>
        </div>
      </Card>
    </div>

    <!-- 标签页切换 + 操作栏 -->
    <Card>
      <div class="mb-4 flex items-center justify-between">
        <div class="flex gap-1">
          <Button :type="activeTab === 'config' ? 'primary' : 'default'" @click="activeTab = 'config'">
            ⚙️ 存储配置
          </Button>
          <Button :type="activeTab === 'bucket' ? 'primary' : 'default'" @click="activeTab = 'bucket'">
            📦 存储桶
          </Button>
        </div>
        <Space>
          <Button @click="loadData">🔄 刷新</Button>
          <Button v-if="canEdit" type="primary" @click="activeTab === 'config' ? handleAddConfig() : handleAddBucket()">
            <Plus class="mr-1 size-4" />
            {{ activeTab === 'config' ? '新增配置' : '新增桶' }}
          </Button>
        </Space>
      </div>

      <!-- 存储配置表格 -->
      <Table
        v-if="activeTab === 'config'"
        :columns="configColumns"
        :data-source="configList"
        :loading="loading"
        row-key="id"
        size="middle"
        :scroll="{ x: 1100 }"
        :pagination="false"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'driver'">
            <Tag :color="getDriverInfo(record.driver).color">
              {{ getDriverInfo(record.driver).icon }} {{ getDriverInfo(record.driver).label }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'connection'">
            <span v-if="record.driver === 'local'" class="text-gray-400">本地存储</span>
            <Tooltip v-else :title="record.endpoint">
              <span class="text-sm">{{ record.endpoint || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'status'">
            <Badge :status="record.status === 1 ? 'success' : 'default'" :text="record.status === 1 ? '启用' : '禁用'" />
          </template>
          <template v-else-if="column.key === 'isDefault'">
            <Tag v-if="record.isDefault" color="blue">默认</Tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button v-if="canEdit" type="link" size="small" @click="handleEditConfig(record)">编辑</Button>
              <Button type="link" size="small" :loading="testLoading" @click="handleTestConfig(record)">测试</Button>
              <Button v-if="canEdit && !record.isDefault && record.status === 1" type="link" size="small" @click="handleSetDefaultConfig(record)">设为默认</Button>
              <Popconfirm v-if="canEdit && record.driver !== 'local'" title="确定删除此配置？" @confirm="handleDeleteConfig(record.id)">
                <Button type="link" size="small" danger>删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>

      <!-- 存储桶表格 -->
      <Table
        v-if="activeTab === 'bucket'"
        :columns="bucketColumns"
        :data-source="bucketList"
        :loading="loading"
        row-key="id"
        size="middle"
        :scroll="{ x: 1100 }"
        :pagination="false"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'driver'">
            <Tag :color="getDriverInfo(record.driver).color">
              {{ getDriverInfo(record.driver).icon }} {{ getDriverInfo(record.driver).label }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'bucketPath'">
            <div>
              <span class="font-medium">{{ record.bucket }}</span>
              <span v-if="record.pathPrefix" class="ml-1 text-gray-400">/ {{ record.pathPrefix }}</span>
            </div>
          </template>
          <template v-else-if="column.key === 'purpose'">
            <Tag>{{ getPurposeLabel(record.purpose) }}</Tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <Badge :status="record.status === 1 ? 'success' : 'default'" :text="record.status === 1 ? '启用' : '禁用'" />
          </template>
          <template v-else-if="column.key === 'isDefault'">
            <Tag v-if="record.isDefault" color="blue">默认</Tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button v-if="canEdit" type="link" size="small" @click="handleEditBucket(record)">编辑</Button>
              <Button type="link" size="small" :loading="testLoading" @click="handleTestBucket(record)">测试</Button>
              <Button v-if="canEdit && !record.isDefault && record.status === 1" type="link" size="small" @click="handleSetDefaultBucket(record)">设为默认</Button>
              <Popconfirm v-if="canEdit && !record.isDefault" title="确定删除此桶？" @confirm="handleDeleteBucket(record.id)">
                <Button type="link" size="small" danger>删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 新增/编辑弹窗 -->
    <Modal
      v-model:open="modalVisible"
      :title="editingId ? '编辑' : '新增'"
      :confirm-loading="saveLoading"
      width="640px"
      @ok="handleSave"
    >
      <!-- 存储配置表单 -->
      <template v-if="activeTab === 'config'">
        <div class="mt-4 space-y-4">
          <Row :gutter="16">
            <Col :span="12">
              <div class="mb-1 text-sm font-medium">配置名称 *</div>
              <Input v-model:value="configForm.name" placeholder="如：生产环境 MinIO" />
            </Col>
            <Col :span="12">
              <div class="mb-1 text-sm font-medium">存储驱动 *</div>
              <Select v-model:value="configForm.driver" class="w-full" :disabled="!!editingId">
                <SelectOption v-for="d in driverOptions" :key="d.value" :value="d.value">
                  {{ d.icon }} {{ d.label }}
                </SelectOption>
              </Select>
            </Col>
          </Row>

          <template v-if="configForm.driver !== 'local'">
            <Row :gutter="16">
              <Col :span="configForm.driver === 'cos' ? 12 : 24">
                <div class="mb-1 text-sm font-medium">{{ configForm.driver === 'cos' ? '地域' : 'Endpoint *' }}</div>
                <Input v-model:value="configForm[configForm.driver === 'cos' ? 'region' : 'endpoint']" :placeholder="configForm.driver === 'cos' ? 'ap-guangzhou' : 'https://minio.example.com'" />
              </Col>
              <Col v-if="configForm.driver === 'cos'" :span="12">
                <div class="mb-1 text-sm font-medium">Endpoint</div>
                <Input v-model:value="configForm.endpoint" placeholder="可选，自定义域名" />
              </Col>
            </Row>
            <Row :gutter="16">
              <Col :span="12">
                <div class="mb-1 text-sm font-medium">{{ configForm.driver === 'cos' ? 'Secret ID' : 'Access Key' }}</div>
                <Input v-model:value="configForm.accessKey" placeholder="访问密钥" />
              </Col>
              <Col :span="12">
                <div class="mb-1 text-sm font-medium">Secret Key</div>
                <Input v-model:value="configForm.secretKey" type="password" placeholder="密钥" />
              </Col>
            </Row>
            <Row :gutter="16">
              <Col :span="12">
                <div class="mb-1 text-sm font-medium">桶名称 *</div>
                <Input v-model:value="configForm.bucket" placeholder="my-bucket" />
              </Col>
              <Col :span="12">
                <div class="mb-1 text-sm font-medium">CDN 域名</div>
                <Input v-model:value="configForm.cdnDomain" placeholder="可选" />
              </Col>
            </Row>
            <Row :gutter="16">
              <Col :span="8">
                <div class="mb-1 text-sm font-medium">启用 SSL</div>
                <Switch v-model:checked="configForm.useSsl" />
              </Col>
            </Row>
          </template>

          <Row :gutter="16">
            <Col :span="12">
              <div class="mb-1 text-sm font-medium">签名 URL 有效期 (秒)</div>
              <InputNumber v-model:value="configForm.presignedUrlExpiry" :min="60" :max="604800" class="w-full" />
            </Col>
            <Col :span="6">
              <div class="mb-1 text-sm font-medium">启用</div>
              <Switch v-model:checked="configForm.status" :checked-value="1" :un-checked-value="0" />
            </Col>
            <Col :span="6">
              <div class="mb-1 text-sm font-medium">设为默认</div>
              <Switch v-model:checked="configForm.isDefault" />
            </Col>
          </Row>

          <div>
            <div class="mb-1 text-sm font-medium">说明</div>
            <Input.TextArea v-model:value="configForm.description" :rows="2" placeholder="备注" />
          </div>

          <div class="flex justify-end">
            <Button :loading="testLoading" @click="handleTestConfig()">🔌 测试连接</Button>
          </div>
        </div>
      </template>

      <!-- 存储桶表单 -->
      <template v-if="activeTab === 'bucket'">
        <div class="mt-4 space-y-4">
          <Row :gutter="16">
            <Col :span="12">
              <div class="mb-1 text-sm font-medium">桶名称 *</div>
              <Input v-model:value="bucketForm.name" placeholder="如：用户文件存储" />
            </Col>
            <Col :span="12">
              <div class="mb-1 text-sm font-medium">存储驱动 *</div>
              <Select v-model:value="bucketForm.driver" class="w-full">
                <SelectOption v-for="d in driverOptions" :key="d.value" :value="d.value">
                  {{ d.icon }} {{ d.label }}
                </SelectOption>
              </Select>
            </Col>
          </Row>
          <Row :gutter="16">
            <Col :span="12">
              <div class="mb-1 text-sm font-medium">桶 *</div>
              <Input v-model:value="bucketForm.bucket" placeholder="bucket-name" />
            </Col>
            <Col :span="12">
              <div class="mb-1 text-sm font-medium">路径前缀</div>
              <Input v-model:value="bucketForm.pathPrefix" placeholder="可选，如 uploads/" />
            </Col>
          </Row>
          <Row :gutter="16">
            <Col :span="12">
              <div class="mb-1 text-sm font-medium">用途</div>
              <Select v-model:value="bucketForm.purpose" class="w-full">
                <SelectOption v-for="p in purposeOptions" :key="p.value" :value="p.value">
                  {{ p.label }}
                </SelectOption>
              </Select>
            </Col>
            <Col :span="6">
              <div class="mb-1 text-sm font-medium">启用</div>
              <Switch v-model:checked="bucketForm.status" :checked-value="1" :un-checked-value="0" />
            </Col>
            <Col :span="6">
              <div class="mb-1 text-sm font-medium">设为默认</div>
              <Switch v-model:checked="bucketForm.isDefault" />
            </Col>
          </Row>
          <div>
            <div class="mb-1 text-sm font-medium">说明</div>
            <Input.TextArea v-model:value="bucketForm.description" :rows="2" placeholder="备注" />
          </div>
          <div class="flex justify-end">
            <Button :loading="testLoading" @click="handleTestBucket()">🔌 测试连接</Button>
          </div>
        </div>
      </template>
    </Modal>
  </Page>
</template>
