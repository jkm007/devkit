<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import { useAccess } from '@vben/access';

import {
  Badge,
  Button,
  Col,
  Divider,
  Form,
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
} from 'ant-design-vue';
import { IconifyIcon, Plus } from '@vben/icons';

import {
  createStorageConfigApi,
  deleteStorageConfigApi,
  getAllStorageConfigsApi,
  getStorageConfigEnabledDriversApi,
  setDefaultStorageConfigApi,
  testStorageConfigByDataApi,
  testStorageConfigConnectionApi,
  updateStorageConfigApi,
} from '#/api/system/storage-config';
import type { StorageConfigApi } from '#/api/system/storage-config';

const { hasAccessByCodes } = useAccess();

// 数据列表
const configList = ref<StorageConfigApi.StorageConfig[]>([]);
const loading = ref(false);
const testLoading = ref(false);
const saveLoading = ref(false);
const enabledDrivers = ref<
  Array<{ value: string; label: string; icon: string; enabled: boolean }>
>([]);

// 弹窗控制
const modalVisible = ref(false);
const modalTitle = ref('');
const isEditing = ref(false);
const currentConfig = reactive<
  StorageConfigApi.CreateStorageConfig & { _editId?: number }
>({
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

// 驱动选项（含颜色）
const driverColors: Record<string, string> = {
  local: '#52c41a',
  minio: '#1677ff',
  oss: '#ff6a00',
  cos: '#006eff',
};

const driverIcons: Record<string, string> = {
  local: '💻',
  minio: '📦',
  oss: '☁️',
  cos: '🌊',
};

const driverLabels: Record<string, string> = {
  local: '本地存储',
  minio: 'MinIO',
  oss: '阿里云 OSS',
  cos: '腾讯云 COS',
};

// 过滤后的驱动选项（只显示已启用的 + local）
const driverOptions = computed(() => {
  return enabledDrivers.value
    .filter((d) => d.enabled || d.value === 'local')
    .map((d) => ({
      value: d.value,
      label: `${d.icon} ${d.label}`,
      color: driverColors[d.value] || '#8c8c8c',
    }));
});

// 统计数据
const stats = computed(() => {
  const total = configList.value.length;
  const enabled = configList.value.filter((c) => c.status === 1).length;
  const disabled = total - enabled;
  const defaultConfig = configList.value.find((c) => c.isDefault);

  // 按驱动统计
  const byDriver: Record<string, number> = {};
  for (const c of configList.value) {
    byDriver[c.driver] = (byDriver[c.driver] || 0) + 1;
  }

  return { total, enabled, disabled, defaultConfig, byDriver };
});

// 表格列
const columns = [
  { title: '配置名称', dataIndex: 'name', width: 180 },
  { title: '存储驱动', key: 'driver', width: 130 },
  { title: '连接信息', key: 'connection', width: 250 },
  { title: 'Bucket', dataIndex: 'bucket', width: 150 },
  {
    title: 'URL过期(秒)',
    dataIndex: 'presignedUrlExpiry',
    width: 120,
    align: 'center' as const,
  },
  { title: '状态', key: 'status', width: 100 },
  { title: '默认', key: 'isDefault', width: 100 },
  { title: '描述', dataIndex: 'description', ellipsis: true },
  { title: '操作', key: 'action', width: 250, fixed: 'right' as const },
];

// 加载数据
const fetchConfigs = async () => {
  loading.value = true;
  try {
    configList.value = await getAllStorageConfigsApi();
  } catch (error) {
    console.error('获取存储配置列表失败:', error);
    message.error('获取存储配置列表失败');
  } finally {
    loading.value = false;
  }
};

// 加载已启用的驱动
const loadEnabledDrivers = async () => {
  try {
    enabledDrivers.value = await getStorageConfigEnabledDriversApi();
  } catch (error) {
    console.error('获取启用驱动失败:', error);
  }
};

onMounted(() => {
  fetchConfigs();
  loadEnabledDrivers();
});

// 获取连接信息显示文本
const getConnectionInfo = (
  record: StorageConfigApi.StorageConfig | Record<string, any>,
) => {
  if (record.driver === 'local') return '本地文件系统';
  if (record.driver === 'cos') return record.region || '-';
  return record.endpoint || '-';
};

// 重置表单
const resetForm = () => {
  Object.assign(currentConfig, {
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
    _editId: undefined,
  });
};

// 添加
const handleAdd = () => {
  resetForm();
  isEditing.value = false;
  modalTitle.value = '添加存储配置';
  modalVisible.value = true;
};

// 编辑
const handleEdit = (
  record: StorageConfigApi.StorageConfig | Record<string, any>,
) => {
  isEditing.value = true;
  modalTitle.value = '编辑存储配置';
  Object.assign(currentConfig, {
    name: record.name,
    driver: record.driver,
    endpoint: record.endpoint,
    accessKey: '******',
    secretKey: '******',
    bucket: record.bucket,
    region: record.region,
    useSsl: record.useSsl,
    cdnDomain: record.cdnDomain,
    isDefault: record.isDefault,
    presignedUrlExpiry: record.presignedUrlExpiry || 3600,
    status: record.status,
    description: record.description,
  });
  currentConfig._editId = record.id;
  modalVisible.value = true;
};

// 保存
const handleSave = async () => {
  if (!currentConfig.name) {
    message.warning('请输入配置名称');
    return;
  }

  // 外部存储需要 bucket
  if (currentConfig.driver !== 'local' && !currentConfig.bucket) {
    message.warning('请输入 Bucket 名称');
    return;
  }

  saveLoading.value = true;
  try {
    if (isEditing.value) {
      // 编辑模式下，如果凭证字段未修改（仍为脱敏值），则不发送给后端，避免覆盖原值
      const updateData: Record<string, any> = { ...currentConfig };
      if (updateData.accessKey === '******') {
        delete updateData.accessKey;
      }
      if (updateData.secretKey === '******') {
        delete updateData.secretKey;
      }
      delete updateData._editId;
      await updateStorageConfigApi(currentConfig._editId!, updateData as any);
      message.success('更新成功');
    } else {
      await createStorageConfigApi(currentConfig as any);
      message.success('创建成功');
    }
    modalVisible.value = false;
    fetchConfigs();
    loadEnabledDrivers();
  } catch (error: any) {
    message.error(error.message || '操作失败');
  } finally {
    saveLoading.value = false;
  }
};

// 删除
const handleDelete = async (
  record: StorageConfigApi.StorageConfig | Record<string, any>,
) => {
  try {
    await deleteStorageConfigApi(record.id);
    message.success('删除成功');
    fetchConfigs();
    loadEnabledDrivers();
  } catch (error: any) {
    message.error(error.message || '删除失败');
  }
};

// 设为默认
const handleSetDefault = async (
  record: StorageConfigApi.StorageConfig | Record<string, any>,
) => {
  try {
    await setDefaultStorageConfigApi(record.id);
    message.success('已设为默认');
    fetchConfigs();
  } catch (error: any) {
    message.error(error.message || '设置失败');
  }
};

// 测试连接
const handleTestConnection = async () => {
  testLoading.value = true;
  try {
    if (isEditing.value && currentConfig._editId) {
      await testStorageConfigConnectionApi(currentConfig._editId);
    } else {
      await testStorageConfigByDataApi({
        driver: currentConfig.driver,
        endpoint: currentConfig.endpoint,
        accessKey: currentConfig.accessKey,
        secretKey: currentConfig.secretKey,
        bucket: currentConfig.bucket,
        region: currentConfig.region,
        useSsl: currentConfig.useSsl,
      });
    }
    message.success('连接成功！');
  } catch (error: any) {
    message.error(error.message || '连接失败');
  } finally {
    testLoading.value = false;
  }
};

// 取消
const handleCancel = () => {
  modalVisible.value = false;
  resetForm();
};

// 是否显示外部存储字段
const isExternalDriver = computed(() => currentConfig.driver !== 'local');
// 是否是 COS（使用 region 而不是 endpoint）
const isCOS = computed(() => currentConfig.driver === 'cos');
</script>

<template>
  <div class="p-4">
    <!-- 统计卡片 -->
    <div class="mb-4 grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="rounded-lg border border-gray-200 bg-white p-3">
        <div class="text-xs text-gray-500">总配置数</div>
        <div class="mt-1 text-2xl font-bold text-blue-600">
          {{ stats.total }}
        </div>
      </div>
      <div class="rounded-lg border border-gray-200 bg-white p-3">
        <div class="text-xs text-gray-500">已启用</div>
        <div class="mt-1 text-2xl font-bold text-green-600">
          {{ stats.enabled }}
        </div>
      </div>
      <div class="rounded-lg border border-gray-200 bg-white p-3">
        <div class="text-xs text-gray-500">已禁用</div>
        <div class="mt-1 text-2xl font-bold text-gray-400">
          {{ stats.disabled }}
        </div>
      </div>
      <div class="rounded-lg border border-gray-200 bg-white p-3">
        <div class="text-xs text-gray-500">默认配置</div>
        <div class="mt-1 text-sm font-bold text-amber-600 truncate">
          {{ stats.defaultConfig ? stats.defaultConfig.name : '未设置' }}
        </div>
      </div>
    </div>

    <!-- 使用说明 -->
    <div
      class="mb-4 rounded-lg border border-blue-200 bg-blue-50 p-3 text-sm text-blue-700"
    >
      <div class="font-medium mb-1">💡 存储配置说明</div>
      <ul class="list-disc pl-5 space-y-0.5">
        <li>
          <b>存储配置</b> = 一组存储连接信息（如 MinIO 的 endpoint + 密钥）
        </li>
        <li>
          可以为同一个驱动类型创建多个配置（如：生产环境 MinIO、测试环境 MinIO）
        </li>
        <li><b>存储桶管理</b>页面中的桶会自动使用这里配置的连接信息</li>
        <li>设为「默认」的配置将作为新文件上传的首选存储</li>
      </ul>
    </div>

    <!-- 操作栏 -->
    <div class="mb-4 flex justify-between items-center">
      <Space>
        <Button
          v-if="hasAccessByCodes(['storage:config:edit'])"
          type="primary"
          @click="handleAdd"
        >
          <Plus class="mr-1" />
          添加存储配置
        </Button>
      </Space>
      <Button @click="fetchConfigs">
        <IconifyIcon icon="mdi:reload" class="mr-1" />
        刷新
      </Button>
    </div>

    <!-- 数据表格 -->
    <Table
      :columns="columns"
      :data-source="configList"
      :loading="loading"
      row-key="id"
      :scroll="{ x: 1200 }"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'driver'">
          <Tag :color="driverColors[record.driver] || '#8c8c8c'">
            {{ driverIcons[record.driver] || '📁' }}
            {{ driverLabels[record.driver] || record.driver }}
          </Tag>
        </template>
        <template v-else-if="column.key === 'connection'">
          <span class="text-sm text-gray-600">{{
            getConnectionInfo(record)
          }}</span>
        </template>
        <template v-else-if="column.key === 'status'">
          <Badge
            :status="record.status === 1 ? 'success' : 'default'"
            :text="record.status === 1 ? '启用' : '禁用'"
          />
        </template>
        <template v-else-if="column.key === 'isDefault'">
          <Tag v-if="record.isDefault" color="gold">⭐ 默认</Tag>
          <Button
            v-else-if="
              hasAccessByCodes(['storage:config:edit']) && record.status === 1
            "
            type="link"
            size="small"
            @click="handleSetDefault(record)"
          >
            设为默认
          </Button>
        </template>
        <template v-else-if="column.key === 'action'">
          <Space>
            <Button
              v-if="hasAccessByCodes(['storage:config:edit'])"
              type="link"
              size="small"
              @click="handleEdit(record)"
            >
              编辑
            </Button>
            <Button
              v-if="
                hasAccessByCodes(['storage:config:edit']) &&
                record.driver !== 'local'
              "
              type="link"
              size="small"
              @click="
                async () => {
                  try {
                    await testStorageConfigConnectionApi(record.id);
                    message.success('连接正常');
                  } catch (e: any) {
                    message.error(e.message || '连接失败');
                  }
                }
              "
            >
              测试
            </Button>
            <Popconfirm
              v-if="
                hasAccessByCodes(['storage:config:delete']) &&
                record.driver !== 'local'
              "
              title="确定删除此配置?"
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
      :width="650"
      @cancel="handleCancel"
    >
      <Form layout="vertical" class="mt-4">
        <Row :gutter="16">
          <Col :span="12">
            <Form.Item label="配置名称" required>
              <Input
                v-model:value="currentConfig.name"
                placeholder="如：生产环境 MinIO"
              />
            </Form.Item>
          </Col>
          <Col :span="12">
            <Form.Item label="存储驱动" required>
              <Select
                v-model:value="currentConfig.driver"
                placeholder="选择驱动"
                :disabled="isEditing"
              >
                <SelectOption
                  v-for="opt in driverOptions"
                  :key="opt.value"
                  :value="opt.value"
                >
                  {{ opt.label }}
                </SelectOption>
              </Select>
            </Form.Item>
          </Col>
        </Row>

        <template v-if="isExternalDriver">
          <template v-if="!isCOS">
            <Form.Item label="Endpoint" required>
              <Input
                v-model:value="currentConfig.endpoint"
                placeholder="如：minio.example.com:9000"
              />
            </Form.Item>
          </template>

          <template v-if="isCOS">
            <Form.Item label="Region" required>
              <Input
                v-model:value="currentConfig.region"
                placeholder="如：ap-guangzhou"
              />
            </Form.Item>
          </template>

          <Row :gutter="16">
            <Col :span="12">
              <Form.Item :label="isCOS ? 'Secret ID' : 'Access Key'" required>
                <Input
                  v-model:value="currentConfig.accessKey"
                  :placeholder="isCOS ? '腾讯云 Secret ID' : 'Access Key'"
                />
              </Form.Item>
            </Col>
            <Col :span="12">
              <Form.Item label="Secret Key" required>
                <Input.Password
                  v-model:value="currentConfig.secretKey"
                  :placeholder="isCOS ? '腾讯云 Secret Key' : 'Secret Key'"
                />
              </Form.Item>
            </Col>
          </Row>

          <Row :gutter="16">
            <Col :span="12">
              <Form.Item label="Bucket 名称" required>
                <Input
                  v-model:value="currentConfig.bucket"
                  placeholder="如：my-bucket"
                />
              </Form.Item>
            </Col>
            <Col :span="12">
              <Form.Item label="CDN 域名">
                <Input
                  v-model:value="currentConfig.cdnDomain"
                  placeholder="可选，如：cdn.example.com"
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item v-if="!isCOS" label="使用 SSL">
            <Switch v-model:checked="currentConfig.useSsl" />
          </Form.Item>
        </template>

        <Divider />

        <Row :gutter="16">
          <Col :span="8">
            <Form.Item label="设为默认">
              <Switch v-model:checked="currentConfig.isDefault" />
            </Form.Item>
          </Col>
          <Col :span="8">
            <Form.Item label="状态">
              <Switch
                v-model:checked="currentConfig.status"
                :checked-value="1"
                :un-checked-value="0"
              />
            </Form.Item>
          </Col>
          <Col :span="8">
            <Form.Item
              label="URL过期(秒)"
              tooltip="预签名URL默认过期时间，预览=300s，头像=604800s"
            >
              <InputNumber
                v-model:value="currentConfig.presignedUrlExpiry"
                :min="60"
                :max="604800"
                style="width: 100%"
              />
            </Form.Item>
          </Col>
        </Row>

        <Form.Item label="描述">
          <Input.TextArea
            v-model:value="currentConfig.description"
            placeholder="备注信息"
            :rows="2"
          />
        </Form.Item>
      </Form>

      <template #footer>
        <Space>
          <Button @click="handleCancel">取消</Button>
          <Button
            v-if="isExternalDriver"
            :loading="testLoading"
            @click="handleTestConnection"
          >
            测试连接
          </Button>
          <Button type="primary" :loading="saveLoading" @click="handleSave">
            保存
          </Button>
        </Space>
      </template>
    </Modal>
  </div>
</template>
