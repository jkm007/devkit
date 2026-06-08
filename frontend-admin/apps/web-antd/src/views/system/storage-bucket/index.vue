<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import { message, Modal } from 'ant-design-vue';

import {
  createStorageBucketApi,
  deleteStorageBucketApi,
  getAllStorageBucketsApi,
  setDefaultStorageBucketApi,
  updateStorageBucketApi,
  type StorageBucketApi,
} from '#/api/system/storage-bucket';
import { useI18n } from '#/locales';

const { t } = useI18n();

// 数据列表
const bucketList = ref<StorageBucketApi.StorageBucket[]>([]);
const loading = ref(false);

// 弹窗控制
const modalVisible = ref(false);
const modalTitle = ref('');
const isEditing = ref(false);
const currentBucket = ref<StorageBucketApi.CreateStorageBucket>({
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

// 驱动选项
const driverOptions = [
  { value: 'local', label: 'Local 本地存储' },
  { value: 'minio', label: 'MinIO 对象存储' },
  { value: 'oss', label: '阿里云 OSS' },
  { value: 'cos', label: '腾讯云 COS' },
];

// 用途选项
const purposeOptions = [
  { value: 'file', label: '文件管理' },
  { value: 'backup', label: '系统备份' },
  { value: 'avatar', label: '用户头像' },
  { value: 'temp', label: '临时文件' },
];

// 表格列
const columns = [
  { title: 'ID', dataIndex: 'id', width: 60 },
  { title: '名称', dataIndex: 'name', width: 150 },
  { title: '驱动', dataIndex: 'driver', width: 100 },
  { title: '桶/路径', dataIndex: 'bucket', width: 150 },
  { title: '用途', dataIndex: 'purpose', width: 100 },
  { title: '状态', key: 'status', width: 100 },
  { title: '默认', key: 'isDefault', width: 80 },
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

onMounted(() => {
  fetchBuckets();
});

// 驱动文本
const getDriverText = (driver: string) => {
  const option = driverOptions.find((item) => item.value === driver);
  return option ? option.label : driver;
};

// 用途文本
const getPurposeText = (purpose: string) => {
  const option = purposeOptions.find((item) => item.value === purpose);
  return option ? option.label : purpose;
};

// 是否显示密钥字段
const showSecretFields = computed(() => {
  return (
    currentBucket.value.driver === 'minio' ||
    currentBucket.value.driver === 'oss' ||
    currentBucket.value.driver === 'cos'
  );
});

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
const handleEdit = (record: StorageBucketApi.StorageBucket) => {
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

// 删除
const handleDelete = (record: StorageBucketApi.StorageBucket) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除存储桶 "${record.name}" 吗？`,
    onOk: async () => {
      try {
        await deleteStorageBucketApi(record.id);
        message.success('删除成功');
        fetchBuckets();
      } catch (error: any) {
        message.error(error?.message || '删除失败');
      }
    },
  });
};

// 设置默认
const handleSetDefault = async (record: StorageBucketApi.StorageBucket) => {
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
    <div class="mb-4">
      <h2 class="text-xl font-bold">存储桶管理</h2>
      <p class="text-gray-500 mt-1">管理不同存储服务的配置，每个存储桶可绑定特定用途</p>
    </div>

    <!-- 操作栏 -->
    <div class="mb-4 flex justify-between items-center">
      <Button type="primary" @click="handleAdd">
        <template #icon><PlusOutlined /></template>
        添加存储桶
      </Button>
      <Button @click="fetchBuckets">
        <template #icon><ReloadOutlined /></template>
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
        <template v-if="column.dataIndex === 'driver'">
          <Tag>{{ getDriverText(record.driver) }}</Tag>
        </template>
        <template v-else-if="column.dataIndex === 'purpose'">
          <Tag color="blue">{{ getPurposeText(record.purpose) }}</Tag>
        </template>
        <template v-else-if="column.key === 'status'">
          <Badge :status="record.status === 1 ? 'success' : 'default'" :text="record.status === 1 ? '启用' : '禁用'" />
        </template>
        <template v-else-if="column.key === 'isDefault'">
          <Tag v-if="record.isDefault" color="gold">默认</Tag>
          <Button v-else type="link" size="small" @click="handleSetDefault(record)">设为默认</Button>
        </template>
        <template v-else-if="column.key === 'action'">
          <Space>
            <Button type="link" size="small" @click="handleEdit(record)">编辑</Button>
            <Popconfirm
              v-if="!record.isDefault"
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
          <Input v-model:value="currentBucket.name" placeholder="输入存储桶名称" />
        </FormItem>
        <FormItem label="存储驱动" required>
          <Select v-model:value="currentBucket.driver" :options="driverOptions" placeholder="选择驱动" />
        </FormItem>

        <template v-if="currentBucket.driver !== 'local'">
          <FormItem label="Endpoint" required>
            <Input v-model:value="currentBucket.endpoint" placeholder="服务端点" />
          </FormItem>
          <FormItem label="Bucket" required>
            <Input v-model:value="currentBucket.bucket" placeholder="桶名称" />
          </FormItem>
        </template>

        <template v-if="showSecretFields">
          <FormItem label="Access Key" required>
            <Input v-model:value="currentBucket.accessKey" placeholder="访问密钥ID" />
          </FormItem>
          <FormItem label="Secret Key" required>
            <InputPassword v-model:value="currentBucket.secretKey" placeholder="访问密钥Secret" />
          </FormItem>
        </template>

        <FormItem v-if="currentBucket.driver === 'oss' || currentBucket.driver === 'cos'" label="Region">
          <Input v-model:value="currentBucket.region" placeholder="区域" />
        </FormItem>

        <FormItem v-if="currentBucket.driver !== 'local'" label="使用SSL">
          <Switch v-model:checked="currentBucket.useSsl" />
        </FormItem>

        <FormItem label="CDN域名">
          <Input v-model:value="currentBucket.cdnDomain" placeholder="CDN域名（可选）" />
        </FormItem>

        <FormItem label="路径前缀">
          <Input v-model:value="currentBucket.pathPrefix" placeholder="路径前缀（可选）" />
        </FormItem>

        <FormItem label="用途">
          <Select v-model:value="currentBucket.purpose" :options="purposeOptions" placeholder="选择用途" allow-clear />
        </FormItem>

        <FormItem label="设为默认">
          <Switch v-model:checked="currentBucket.isDefault" />
        </FormItem>

        <FormItem label="状态">
          <Switch v-model:checked="currentBucket.status" :checked-value="1" :un-checked-value="0" />
        </FormItem>

        <FormItem label="描述">
          <Input.TextArea v-model:value="currentBucket.description" placeholder="描述信息" :rows="2" />
        </FormItem>
      </Form>

      <template #footer>
        <Button @click="handleCancel">取消</Button>
        <Button type="primary" @click="handleSave">保存</Button>
      </template>
    </Modal>
  </div>
</template>
