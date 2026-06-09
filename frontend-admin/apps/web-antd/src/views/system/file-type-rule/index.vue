<script lang="ts" setup>
import { computed, onMounted, ref, reactive } from 'vue';

import { useAccess } from '@vben/access';

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
  InputNumber,
} from 'ant-design-vue';
import { IconifyIcon, Plus } from '@vben/icons';
import { $t } from '#/locales';

const { hasAccessByCodes } = useAccess();
import {
  getAllFileTypeRules,
  createFileTypeRule,
  updateFileTypeRule,
  deleteFileTypeRule,
  refreshAutoTagger,
} from '#/api/system/file-type-rule';
import type { FileTypeRule } from '#/api/system/file-type-rule';

const t = $t;

const rules = ref<FileTypeRule[]>([]);
const loading = ref(false);
const modalVisible = ref(false);
const editing = ref<FileTypeRule | null>(null);
const filterType = ref<string>('');
const refreshLoading = ref(false);

const form = reactive({
  extension: '',
  fileType: '' as string,
  description: '',
  status: 1,
});

// 文件类型选项
const fileTypeOptions = computed(() => [
  { label: '图片 (image)', value: 'image' },
  { label: '视频 (video)', value: 'video' },
  { label: '音频 (audio)', value: 'audio' },
  { label: '文档 (document)', value: 'document' },
  { label: '压缩包 (archive)', value: 'archive' },
  { label: '其他 (other)', value: 'other' },
]);

// 文件类型颜色映射
const fileTypeColors: Record<string, string> = {
  image: 'green',
  video: 'blue',
  audio: 'purple',
  document: 'orange',
  archive: 'cyan',
  other: 'default',
};

// 过滤后的规则
const filteredRules = computed(() => {
  if (!filterType.value) return rules.value;
  return rules.value.filter((r) => r.fileType === filterType.value);
});

// 加载规则
const loadRules = async () => {
  loading.value = true;
  try {
    rules.value = await getAllFileTypeRules();
  } catch (error) {
    message.error('加载规则失败');
  } finally {
    loading.value = false;
  }
};

// 初始化
onMounted(() => {
  loadRules();
});

// 打开编辑弹窗
const openModal = (rule?: FileTypeRule) => {
  if (rule) {
    editing.value = rule;
    form.extension = rule.extension;
    form.fileType = rule.fileType;
    form.description = rule.description;
    form.status = rule.status;
  } else {
    editing.value = null;
    form.extension = '';
    form.fileType = '';
    form.description = '';
    form.status = 1;
  }
  modalVisible.value = true;
};

// 提交表单
const handleSubmit = async () => {
  if (!form.extension) {
    message.error('请输入扩展名');
    return;
  }
  if (!form.fileType) {
    message.error('请选择文件类型');
    return;
  }

  try {
    if (editing.value) {
      await updateFileTypeRule(editing.value.id, {
        ...form,
      } as Partial<FileTypeRule>);
      message.success('更新成功');
    } else {
      await createFileTypeRule({ ...form } as Partial<FileTypeRule>);
      message.success('创建成功');
    }
    modalVisible.value = false;
    loadRules();
  } catch (error: any) {
    message.error(error.message || '操作失败');
  }
};

// 删除规则
const handleDelete = async (id: number) => {
  try {
    await deleteFileTypeRule(id);
    message.success('删除成功');
    loadRules();
  } catch (error: any) {
    message.error(error.message || '删除失败');
  }
};

// 刷新 AutoTagger
const handleRefresh = async () => {
  refreshLoading.value = true;
  try {
    const result = await refreshAutoTagger();
    message.success(`刷新成功，已加载 ${result.count} 条规则`);
  } catch (error: any) {
    message.error(error.message || '刷新失败');
  } finally {
    refreshLoading.value = false;
  }
};

// 列定义
const columns = computed(() => [
  { title: '扩展名', dataIndex: 'extension', width: 120 },
  { title: '文件类型', dataIndex: 'fileType', width: 120 },
  { title: '描述', dataIndex: 'description', width: 200 },
  { title: '状态', dataIndex: 'status', width: 100 },
  { title: '创建时间', dataIndex: 'createdAt', width: 180 },
  { title: '操作', key: 'action', width: 150 },
]);

// 统计各类型数量
const typeCounts = computed(() => {
  const counts: Record<string, number> = {};
  rules.value.forEach((r) => {
    counts[r.fileType] = (counts[r.fileType] || 0) + 1;
  });
  return counts;
});
</script>

<template>
  <div class="p-4">
    <Card>
      <!-- 统计卡片 -->
      <div class="mb-4 grid grid-cols-2 md:grid-cols-6 gap-4">
        <Card
          v-for="(count, type) in typeCounts"
          :key="type"
          size="small"
          :class="filterType === type ? 'border-blue-500' : ''"
          hoverable
          @click="filterType = filterType === type ? '' : (type as string)"
        >
          <div class="text-center">
            <Tag :color="fileTypeColors[type as string] || 'default'" class="mb-1">
              {{ type }}
            </Tag>
            <div class="text-2xl font-bold text-blue-500">{{ count }}</div>
          </div>
        </Card>
      </div>

      <!-- 工具栏 -->
      <div class="mb-4 flex justify-between items-center">
        <Space>
          <Button
            v-if="hasAccessByCodes(['system:setting:edit'])"
            type="primary"
            @click="openModal()"
          >
            <Plus class="mr-1" />
            新增规则
          </Button>
          <Button @click="loadRules">
            <IconifyIcon icon="mdi:reload" class="mr-1" />
            刷新
          </Button>
          <Button
            v-if="hasAccessByCodes(['system:setting:edit'])"
            :loading="refreshLoading"
            @click="handleRefresh"
          >
            <IconifyIcon icon="mdi:sync" class="mr-1" />
            同步到 AutoTagger
          </Button>
        </Space>
        <Space>
          <span class="text-gray-500">筛选类型：</span>
          <Select
            v-model:value="filterType"
            :options="[{ label: '全部', value: '' }, ...fileTypeOptions]"
            style="width: 180px"
            allowClear
          />
        </Space>
      </div>

      <!-- 规则表格 -->
      <Table
        :columns="columns"
        :data-source="filteredRules"
        :loading="loading"
        row-key="id"
        size="middle"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'extension'">
            <Tag>{{ record.extension }}</Tag>
          </template>
          <template v-if="column.dataIndex === 'fileType'">
            <Tag :color="fileTypeColors[record.fileType] || 'default'">
              {{ record.fileType }}
            </Tag>
          </template>
          <template v-if="column.dataIndex === 'status'">
            <Tag :color="record.status === 1 ? 'green' : 'default'">
              {{ record.status === 1 ? '启用' : '禁用' }}
            </Tag>
          </template>
          <template v-if="column.dataIndex === 'createdAt'">
            {{ record.createdAt ? new Date(record.createdAt).toLocaleString() : '-' }}
          </template>
          <template v-if="column.key === 'action'">
            <Space>
              <Tooltip
                v-if="hasAccessByCodes(['system:setting:edit'])"
                title="编辑"
              >
                <Button
                  type="link"
                  size="small"
                  @click="openModal(record)"
                >
                  <IconifyIcon icon="mdi:pencil" />
                </Button>
              </Tooltip>
              <Popconfirm
                v-if="hasAccessByCodes(['system:setting:edit'])"
                title="确定要删除此规则吗？"
                @confirm="handleDelete(record.id)"
              >
                <Tooltip title="删除">
                  <Button type="link" size="small" danger>
                    <IconifyIcon icon="mdi:delete" />
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
      :title="editing ? '编辑规则' : '新增规则'"
      @ok="handleSubmit"
      width="500px"
    >
      <Form :model="form" layout="vertical">
        <Form.Item label="扩展名" required>
          <Input
            v-model:value="form.extension"
            placeholder="例如: .jpg 或 jpg"
          />
          <div class="text-gray-400 text-xs mt-1">
            输入扩展名，不带点号会自动添加
          </div>
        </Form.Item>
        <Form.Item label="文件类型" required>
          <Select
            v-model:value="form.fileType"
            :options="fileTypeOptions"
            placeholder="请选择文件类型"
          />
        </Form.Item>
        <Form.Item label="描述">
          <Input
            v-model:value="form.description"
            placeholder="规则描述（可选）"
          />
        </Form.Item>
        <Form.Item label="状态">
          <Switch
            :checked="form.status === 1"
            @change="(checked: boolean) => (form.status = checked ? 1 : 0)"
            checked-children="启用"
            un-checked-children="禁用"
          />
        </Form.Item>
      </Form>
    </Modal>
  </div>
</template>
