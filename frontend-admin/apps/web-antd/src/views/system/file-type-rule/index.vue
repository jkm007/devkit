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
  AutoComplete,
  Switch,
  message,
  Popconfirm,
  Tooltip,
} from 'ant-design-vue';
import { IconifyIcon, Plus } from '@vben/icons';

const { hasAccessByCodes } = useAccess();
import {
  getAllFileTypeRules,
  createFileTypeRule,
  updateFileTypeRule,
  deleteFileTypeRule,
  refreshAutoTagger,
} from '#/api/system/file-type-rule';
import type { FileTypeRule } from '#/api/system/file-type-rule';

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

// 默认文件类型
const defaultTypes = [
  { label: '图片 (image)', value: 'image', color: 'green' },
  { label: '视频 (video)', value: 'video', color: 'blue' },
  { label: '音频 (audio)', value: 'audio', color: 'purple' },
  { label: '文档 (document)', value: 'document', color: 'orange' },
  { label: '压缩包 (archive)', value: 'archive', color: 'cyan' },
  { label: '其他 (other)', value: 'other', color: 'default' },
];

// 文件类型选项（默认 + 从规则中动态提取的自定义类型）
const fileTypeOptions = computed(() => {
  const defaultValues = new Set(defaultTypes.map((t) => t.value));
  const customTypes = [
    ...new Set(
      rules.value.map((r) => r.fileType).filter((t) => !defaultValues.has(t)),
    ),
  ];
  return [
    ...defaultTypes,
    ...customTypes.map((t) => ({ label: t, value: t })),
  ];
});

// 文件类型颜色映射（默认类型 + 动态扩展）
const fileTypeColors = computed(() => {
  const colors: Record<string, string> = {
    image: 'green',
    video: 'blue',
    audio: 'purple',
    document: 'orange',
    archive: 'cyan',
    other: 'default',
  };
  const customColors = [
    'magenta',
    'red',
    'lime',
    'gold',
    'geekblue',
    'volcano',
  ];
  let index = 0;
  rules.value.forEach((r) => {
    if (!colors[r.fileType]) {
      colors[r.fileType] = customColors[index % customColors.length]!;
      index++;
    }
  });
  return colors;
});

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
  } catch {
    message.error('加载规则失败');
  } finally {
    loading.value = false;
  }
};

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
    message.error('请选择或输入文件类型');
    return;
  }

  try {
    if (editing.value) {
      await updateFileTypeRule(editing.value.id, { ...form });
      message.success('更新成功');
    } else {
      await createFileTypeRule({ ...form });
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
const columns = [
  { title: '扩展名', dataIndex: 'extension', width: 120 },
  { title: '文件类型', dataIndex: 'fileType', width: 120 },
  { title: '描述', dataIndex: 'description', ellipsis: true },
  { title: '状态', dataIndex: 'status', width: 80, align: 'center' as const },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    width: 170,
    align: 'center' as const,
  },
  { title: '操作', key: 'action', width: 120, align: 'center' as const },
];

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
    <!-- 统计卡片 -->
    <div class="mb-4 grid grid-cols-2 md:grid-cols-6 gap-3">
      <Card
        v-for="(count, type) in typeCounts"
        :key="type"
        size="small"
        hoverable
        :class="filterType === type ? 'border-blue-500 shadow' : 'cursor-pointer'"
        @click="filterType = filterType === type ? '' : (type as string)"
      >
        <div class="text-center py-1">
          <Tag :color="fileTypeColors[type as string] || 'default'" class="mb-1">
            {{ type }}
          </Tag>
          <div class="text-xl font-bold text-blue-500">{{ count }}</div>
        </div>
      </Card>
    </div>

    <Card>
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
          <span class="text-gray-500">筛选：</span>
          <Select
            v-model:value="filterType"
            :options="[{ label: '全部', value: '' }, ...fileTypeOptions]"
            style="width: 180px"
            allow-clear
            placeholder="按类型筛选"
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
        :pagination="{ pageSize: 20, showSizeChanger: true, showTotal: (t: number) => `共 ${t} 条` }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'extension'">
            <Tag color="blue">{{ record.extension }}</Tag>
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
                <Button type="link" size="small" @click="openModal(record)">
                  <IconifyIcon icon="mdi:pencil" />
                </Button>
              </Tooltip>
              <Popconfirm
                v-if="hasAccessByCodes(['system:setting:edit'])"
                title="确定删除此规则？"
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
      width="480px"
      :ok-text="editing ? '保存' : '添加'"
      cancel-text="取消"
    >
      <Form :model="form" layout="vertical" class="mt-4">
        <Form.Item label="扩展名" required>
          <Input
            v-model:value="form.extension"
            placeholder="例如: .jpg 或 jpg"
          />
          <div class="text-gray-400 text-xs mt-1">
            不带点号会自动添加
          </div>
        </Form.Item>
        <Form.Item label="文件类型" required>
          <AutoComplete
            v-model:value="form.fileType"
            :options="fileTypeOptions"
            placeholder="输入自定义类型，如 design、cad、font"
            :filter-option="(input: string, option: any) =>
              option.value.toLowerCase().includes(input.toLowerCase())"
          />
          <div class="mt-2">
            <span class="text-gray-400 text-xs">快速选择：</span>
            <Tag
              v-for="t in defaultTypes"
              :key="t.value"
              :color="t.color"
              class="cursor-pointer ml-1 mb-1"
              @click="form.fileType = t.value"
            >
              {{ t.value }}
            </Tag>
          </div>
        </Form.Item>
        <Form.Item label="描述">
          <Input
            v-model:value="form.description"
            placeholder="可选，例如：JPEG 图片"
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
