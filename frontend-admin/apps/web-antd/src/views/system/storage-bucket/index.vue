<script lang="ts" setup>
import { computed, onMounted, ref, reactive } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  Switch,
  message,
  Popconfirm,
  Tooltip,
} from 'ant-design-vue';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  DatabaseOutlined,
} from '@ant-design/icons-vue';

const { t } = useI18n();

// 存储桶类型定义
interface StorageBucket {
  id: number;
  driver: string;
  bucketName: string;
  usage: string;
  endpoint?: string;
  region?: string;
  isDefault: boolean;
  status: number;
  createdAt: string;
  updatedAt: string;
}

// 状态
const buckets = ref<StorageBucket[]>([]);
const loading = ref(false);
const modalVisible = ref(false);
const editing = ref<StorageBucket | null>(null);
const form = reactive({
  driver: 'local',
  bucketName: '',
  usage: '',
  endpoint: '',
  region: '',
  isDefault: false,
});

// 表格列定义
const columns = computed(() => [
  { title: '存储驱动', dataIndex: 'driver', width: 120 },
  { title: '桶名称', dataIndex: 'bucketName', width: 150 },
  { title: '用途', dataIndex: 'usage', width: 120 },
  { title: 'Endpoint', dataIndex: 'endpoint', width: 200 },
  { title: 'Region', dataIndex: 'region', width: 100 },
  { title: '默认', dataIndex: 'isDefault', width: 80 },
  { title: '状态', dataIndex: 'status', width: 80 },
  { title: '操作', key: 'action', width: 150 },
]);

// 存储驱动选项
const driverOptions = [
  { label: '本地存储', value: 'local' },
  { label: 'MinIO', value: 'minio' },
  { label: '阿里云 OSS', value: 'oss' },
  { label: '腾讯云 COS', value: 'cos' },
];

// 用途选项
const usageOptions = [
  { label: '文件管理', value: 'file_management' },
  { label: '备份存储', value: 'backup' },
  { label: '头像存储', value: 'avatar' },
  { label: '临时文件', value: 'temp' },
  { label: '其他', value: 'other' },
];

// 加载数据
const loadData = async () => {
  loading.value = true;
  try {
    // TODO: 调用 API 获取存储桶列表
    // buckets.value = await getStorageBuckets();
    buckets.value = []; // 临时空数据
  } catch {
    message.error('加载数据失败');
  } finally {
    loading.value = false;
  }
};

// 初始化
onMounted(() => {
  loadData();
});

// 打开弹窗
const openModal = (bucket?: StorageBucket) => {
  if (bucket) {
    editing.value = bucket;
    form.driver = bucket.driver;
    form.bucketName = bucket.bucketName;
    form.usage = bucket.usage;
    form.endpoint = bucket.endpoint || '';
    form.region = bucket.region || '';
    form.isDefault = bucket.isDefault;
  } else {
    editing.value = null;
    form.driver = 'local';
    form.bucketName = '';
    form.usage = '';
    form.endpoint = '';
    form.region = '';
    form.isDefault = false;
  }
  modalVisible.value = true;
};

// 提交表单
const handleSubmit = async () => {
  try {
    // TODO: 调用 API 创建/更新存储桶
    // if (editing.value) {
    //   await updateStorageBucket(editing.value.id, form);
    // } else {
    //   await createStorageBucket(form);
    // }
    message.success(editing.value ? '更新成功' : '创建成功');
    modalVisible.value = false;
    loadData();
  } catch (error: any) {
    message.error(error.message || '操作失败');
  }
};

// 删除
const handleDelete = async (id: number) => {
  try {
    // TODO: 调用 API 删除存储桶
    // await deleteStorageBucket(id);
    message.success('删除成功');
    loadData();
  } catch (error: any) {
    message.error(error.message || '删除失败');
  }
};

// 获取驱动显示名称
const getDriverLabel = (driver: string) => {
  const option = driverOptions.find(o => o.value === driver);
  return option?.label || driver;
};

// 获取用途显示名称
const getUsageLabel = (usage: string) => {
  const option = usageOptions.find(o => o.value === usage);
  return option?.label || usage;
};
</script>

<template>
  <div class="p-4">
    <Card>
      <template #title>
        <div class="flex items-center gap-2">
          <DatabaseOutlined />
          <span>存储桶管理</span>
        </div>
      </template>

      <div class="mb-4">
        <Space>
          <Button type="primary" @click="openModal()">
            <PlusOutlined />
            新增存储桶
          </Button>
          <Button @click="loadData">
            <ReloadOutlined />
            刷新
          </Button>
        </Space>
      </div>

      <Table
        :columns="columns"
        :data-source="buckets"
        :loading="loading"
        row-key="id"
        size="middle"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'driver'">
            <Tag :color="record.driver === 'local' ? 'green' : record.driver === 'minio' ? 'blue' : record.driver === 'oss' ? 'orange' : 'purple'">
              {{ getDriverLabel(record.driver) }}
            </Tag>
          </template>
          <template v-if="column.dataIndex === 'usage'">
            <Tag>{{ getUsageLabel(record.usage) }}</Tag>
          </template>
          <template v-if="column.dataIndex === 'isDefault'">
            <Tag :color="record.isDefault ? 'green' : 'default'">
              {{ record.isDefault ? '是' : '否' }}
            </Tag>
          </template>
          <template v-if="column.dataIndex === 'status'">
            <Tag :color="record.status === 1 ? 'green' : 'red'">
              {{ record.status === 1 ? '启用' : '禁用' }}
            </Tag>
          </template>
          <template v-if="column.key === 'action'">
            <Space>
              <Tooltip title="编辑">
                <Button type="link" size="small" @click="openModal(record)">
                  <EditOutlined />
                </Button>
              </Tooltip>
              <Popconfirm
                title="确定删除此存储桶?"
                @confirm="handleDelete(record.id)"
              >
                <Tooltip title="删除">
                  <Button type="link" size="small" danger>
                    <DeleteOutlined />
                  </Button>
                </Tooltip>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <!-- 编辑弹窗 -->
    <Modal
      v-model:open="modalVisible"
      :title="editing ? '编辑存储桶' : '新增存储桶'"
      @ok="handleSubmit"
      width="500px"
    >
      <Form :model="form" layout="vertical">
        <Form.Item label="存储驱动" required>
          <Select v-model:value="form.driver" :options="driverOptions" />
        </Form.Item>
        <Form.Item label="桶名称" required>
          <Input v-model:value="form.bucketName" placeholder="如: devkit-files" />
        </Form.Item>
        <Form.Item label="用途">
          <Select v-model:value="form.usage" :options="usageOptions" />
        </Form.Item>
        <Form.Item label="Endpoint" v-if="form.driver !== 'local'">
          <Input v-model:value="form.endpoint" placeholder="如: minio-server:9000" />
        </Form.Item>
        <Form.Item label="Region" v-if="form.driver === 'oss' || form.driver === 'cos'">
          <Input v-model:value="form.region" placeholder="如: oss-cn-beijing" />
        </Form.Item>
        <Form.Item label="默认存储桶">
          <Switch v-model:checked="form.isDefault" />
        </Form.Item>
      </Form>
    </Modal>
  </div>
</template>
