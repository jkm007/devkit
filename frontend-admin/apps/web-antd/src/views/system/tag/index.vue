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
  InputNumber,
  Switch,
  message,
  Popconfirm,
  Tabs,
  TabPane,
  Tooltip,
  Badge,
  ColorPicker,
} from 'ant-design-vue';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  TagsOutlined,
  NodeIndexOutlined,
} from '@ant-design/icons-vue';
import {
  getAllTags,
  createTag,
  updateTag,
  deleteTag,
  getTagUsageStats,
  getAllRoutingRules,
  createRoutingRule,
  updateRoutingRule,
  deleteRoutingRule,
  updateRoutingRuleStatus,
  testRoute,
} from '#/api/system/tag';
import type { Tag as TagType, TagUsageStat, TagRouting } from '#/api/system/tag';

const { t } = useI18n();

// 标签管理相关
const tags = ref<TagType[]>([]);
const tagStats = ref<TagUsageStat[]>([]);
const tagLoading = ref(false);
const tagModalVisible = ref(false);
const tagEditing = ref<TagType | null>(null);
const tagForm = reactive({
  tagKey: '',
  tagValue: '',
  tagName: '',
  icon: '',
  color: '#1890ff',
  description: '',
  sortOrder: 0,
});

// 路由规则相关
const rules = ref<TagRouting[]>([]);
const ruleLoading = ref(false);
const ruleModalVisible = ref(false);
const ruleEditing = ref<TagRouting | null>(null);
const ruleForm = reactive({
  ruleName: '',
  description: '',
  priority: 0,
  matchType: 'all' as 'all' | 'any' | 'exact',
  conditions: [] as { key: string; value: string }[],
  driver: 'local',
  bucket: '',
  pathPrefix: '',
  isDefault: false,
});

// 测试路由相关
const testModalVisible = ref(false);
const testForm = reactive({
  fileName: '',
  contentType: '',
  source: 'user',
});
const testResult = ref<any>(null);

// 标签分组
const tagGroups = computed(() => {
  const groups: Record<string, TagType[]> = {};
  tags.value.forEach((tag) => {
    if (!groups[tag.tagKey]) {
      groups[tag.tagKey] = [];
    }
    groups[tag.tagKey].push(tag);
  });
  return groups;
});

// 加载标签
const loadTags = async () => {
  tagLoading.value = true;
  try {
    tags.value = await getAllTags();
    tagStats.value = await getTagUsageStats();
  } catch (error) {
    message.error('加载标签失败');
  } finally {
    tagLoading.value = false;
  }
};

// 加载路由规则
const loadRules = async () => {
  ruleLoading.value = true;
  try {
    rules.value = await getAllRoutingRules();
  } catch (error) {
    message.error('加载规则失败');
  } finally {
    ruleLoading.value = false;
  }
};

// 初始化
onMounted(() => {
  loadTags();
  loadRules();
});

// 标签操作
const openTagModal = (tag?: TagType) => {
  if (tag) {
    tagEditing.value = tag;
    tagForm.tagKey = tag.tagKey;
    tagForm.tagValue = tag.tagValue;
    tagForm.tagName = tag.tagName;
    tagForm.icon = tag.icon;
    tagForm.color = tag.color;
    tagForm.description = tag.description;
    tagForm.sortOrder = tag.sortOrder;
  } else {
    tagEditing.value = null;
    tagForm.tagKey = '';
    tagForm.tagValue = '';
    tagForm.tagName = '';
    tagForm.icon = '';
    tagForm.color = '#1890ff';
    tagForm.description = '';
    tagForm.sortOrder = 0;
  }
  tagModalVisible.value = true;
};

const handleTagSubmit = async () => {
  try {
    if (tagEditing.value) {
      await updateTag(tagEditing.value.id, tagForm);
      message.success('更新成功');
    } else {
      await createTag(tagForm);
      message.success('创建成功');
    }
    tagModalVisible.value = false;
    loadTags();
  } catch (error: any) {
    message.error(error.message || '操作失败');
  }
};

const handleDeleteTag = async (id: number) => {
  try {
    await deleteTag(id);
    message.success('删除成功');
    loadTags();
  } catch (error: any) {
    message.error(error.message || '删除失败');
  }
};

// 路由规则操作
const openRuleModal = (rule?: TagRouting) => {
  if (rule) {
    ruleEditing.value = rule;
    ruleForm.ruleName = rule.ruleName;
    ruleForm.description = rule.description;
    ruleForm.priority = rule.priority;
    ruleForm.matchType = rule.matchType;
    ruleForm.conditions = rule.conditions?.tags || [];
    ruleForm.driver = rule.driver;
    ruleForm.bucket = rule.bucket || '';
    ruleForm.pathPrefix = rule.pathPrefix || '';
    ruleForm.isDefault = rule.isDefault;
  } else {
    ruleEditing.value = null;
    ruleForm.ruleName = '';
    ruleForm.description = '';
    ruleForm.priority = 0;
    ruleForm.matchType = 'all';
    ruleForm.conditions = [];
    ruleForm.driver = 'local';
    ruleForm.bucket = '';
    ruleForm.pathPrefix = '';
    ruleForm.isDefault = false;
  }
  ruleModalVisible.value = true;
};

const handleRuleSubmit = async () => {
  try {
    const data = {
      ...ruleForm,
      conditions: { tags: ruleForm.conditions },
    };
    if (ruleEditing.value) {
      await updateRoutingRule(ruleEditing.value.id, data);
      message.success('更新成功');
    } else {
      await createRoutingRule(data);
      message.success('创建成功');
    }
    ruleModalVisible.value = false;
    loadRules();
  } catch (error: any) {
    message.error(error.message || '操作失败');
  }
};

const handleDeleteRule = async (id: number) => {
  try {
    await deleteRoutingRule(id);
    message.success('删除成功');
    loadRules();
  } catch (error: any) {
    message.error(error.message || '删除失败');
  }
};

const handleToggleRuleStatus = async (rule: TagRouting) => {
  try {
    await updateRoutingRuleStatus(rule.id, rule.status === 1 ? 0 : 1);
    message.success('更新成功');
    loadRules();
  } catch (error: any) {
    message.error(error.message || '操作失败');
  }
};

// 添加条件
const addCondition = () => {
  ruleForm.conditions.push({ key: '', value: '' });
};

const removeCondition = (index: number) => {
  ruleForm.conditions.splice(index, 1);
};

// 测试路由
const openTestModal = () => {
  testForm.fileName = '';
  testForm.contentType = '';
  testForm.source = 'user';
  testResult.value = null;
  testModalVisible.value = true;
};

const handleTestRoute = async () => {
  try {
    testResult.value = await testRoute(testForm.fileName, testForm.contentType, testForm.source);
  } catch (error: any) {
    message.error(error.message || '测试失败');
  }
};

// 标签列定义
const tagColumns = [
  { title: '标签键', dataIndex: 'tagKey', width: 120 },
  { title: '标签值', dataIndex: 'tagValue', width: 120 },
  { title: '显示名称', dataIndex: 'tagName', width: 120 },
  { title: '图标', dataIndex: 'icon', width: 80 },
  { title: '颜色', dataIndex: 'color', width: 100 },
  { title: '排序', dataIndex: 'sortOrder', width: 80 },
  { title: '文件数', dataIndex: 'fileCount', width: 80 },
  { title: '系统内置', dataIndex: 'isSystem', width: 100 },
  { title: '操作', key: 'action', width: 150 },
];

// 规则列定义
const ruleColumns = [
  { title: '规则名称', dataIndex: 'ruleName', width: 150 },
  { title: '优先级', dataIndex: 'priority', width: 80 },
  { title: '匹配类型', dataIndex: 'matchType', width: 100 },
  { title: '目标存储', dataIndex: 'driver', width: 100 },
  { title: '桶/路径', key: 'bucket', width: 150 },
  { title: '默认规则', dataIndex: 'isDefault', width: 100 },
  { title: '状态', dataIndex: 'status', width: 80 },
  { title: '操作', key: 'action', width: 200 },
];

// 获取文件数量
const getFileCount = (tagId: number) => {
  const stat = tagStats.value.find((s) => s.id === tagId);
  return stat?.fileCount || 0;
};

// 存储驱动选项
const driverOptions = [
  { label: '本地存储', value: 'local' },
  { label: 'MinIO', value: 'minio' },
  { label: '阿里云 OSS', value: 'oss' },
  { label: '腾讯云 COS', value: 'cos' },
];

// 匹配类型选项
const matchTypeOptions = [
  { label: '全部满足', value: 'all' },
  { label: '任一满足', value: 'any' },
  { label: '精确匹配', value: 'exact' },
];

// 标签键选项（用于条件选择）
const tagKeyOptions = computed(() => {
  const keys = new Set(tags.value.map((t) => t.tagKey));
  return Array.from(keys).map((key) => ({ label: key, value: key }));
});

// 获取指定键的标签值选项
const getTagValueOptions = (key: string) => {
  return tags.value
    .filter((t) => t.tagKey === key)
    .map((t) => ({ label: `${t.icon} ${t.tagName}`, value: t.tagValue }));
};
</script>

<template>
  <div class="p-4">
    <Card>
      <Tabs default-active-key="tags">
        <TabPane key="tags" tab="标签管理">
          <template #tab>
            <span>
              <TagsOutlined />
              标签管理
            </span>
          </template>

          <div class="mb-4">
            <Space>
              <Button type="primary" @click="openTagModal()">
                <PlusOutlined />
                新增标签
              </Button>
              <Button @click="loadTags">
                <ReloadOutlined />
                刷新
              </Button>
            </Space>
          </div>

          <Table
            :columns="tagColumns"
            :data-source="tags"
            :loading="tagLoading"
            row-key="id"
            size="middle"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.dataIndex === 'icon'">
                <span style="font-size: 20px">{{ record.icon }}</span>
              </template>
              <template v-if="column.dataIndex === 'color'">
                <Tag :color="record.color">{{ record.color }}</Tag>
              </template>
              <template v-if="column.dataIndex === 'fileCount'">
                <Badge :count="getFileCount(record.id)" :number-style="{ backgroundColor: '#52c41a' }" />
              </template>
              <template v-if="column.dataIndex === 'isSystem'">
                <Tag :color="record.isSystem ? 'blue' : 'default'">
                  {{ record.isSystem ? '系统' : '自定义' }}
                </Tag>
              </template>
              <template v-if="column.key === 'action'">
                <Space>
                  <Tooltip title="编辑">
                    <Button type="link" size="small" @click="openTagModal(record)" :disabled="record.isSystem">
                      <EditOutlined />
                    </Button>
                  </Tooltip>
                  <Popconfirm
                    title="确定删除此标签?"
                    @confirm="handleDeleteTag(record.id)"
                    :disabled="record.isSystem"
                  >
                    <Tooltip title="删除">
                      <Button type="link" size="small" danger :disabled="record.isSystem">
                        <DeleteOutlined />
                      </Button>
                    </Tooltip>
                  </Popconfirm>
                </Space>
              </template>
            </template>
          </Table>
        </TabPane>

        <TabPane key="rules" tab="路由规则">
          <template #tab>
            <span>
              <NodeIndexOutlined />
              路由规则
            </span>
          </template>

          <div class="mb-4">
            <Space>
              <Button type="primary" @click="openRuleModal()">
                <PlusOutlined />
                新增规则
              </Button>
              <Button @click="loadRules">
                <ReloadOutlined />
                刷新
              </Button>
              <Button @click="openTestModal">
                测试路由
              </Button>
            </Space>
          </div>

          <Table
            :columns="ruleColumns"
            :data-source="rules"
            :loading="ruleLoading"
            row-key="id"
            size="middle"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.dataIndex === 'matchType'">
                <Tag color="blue">{{ record.matchType === 'all' ? '全部满足' : record.matchType === 'any' ? '任一满足' : '精确匹配' }}</Tag>
              </template>
              <template v-if="column.key === 'bucket'">
                {{ record.bucket || '-' }}{{ record.pathPrefix ? `/${record.pathPrefix}` : '' }}
              </template>
              <template v-if="column.dataIndex === 'isDefault'">
                <Tag :color="record.isDefault ? 'green' : 'default'">
                  {{ record.isDefault ? '是' : '否' }}
                </Tag>
              </template>
              <template v-if="column.dataIndex === 'status'">
                <Switch
                  :checked="record.status === 1"
                  @change="() => handleToggleRuleStatus(record)"
                  checked-children="启用"
                  un-checked-children="禁用"
                />
              </template>
              <template v-if="column.key === 'action'">
                <Space>
                  <Tooltip title="编辑">
                    <Button type="link" size="small" @click="openRuleModal(record)">
                      <EditOutlined />
                    </Button>
                  </Tooltip>
                  <Popconfirm
                    title="确定删除此规则?"
                    @confirm="handleDeleteRule(record.id)"
                    :disabled="record.isDefault"
                  >
                    <Tooltip title="删除">
                      <Button type="link" size="small" danger :disabled="record.isDefault">
                        <DeleteOutlined />
                      </Button>
                    </Tooltip>
                  </Popconfirm>
                </Space>
              </template>
            </template>
          </Table>
        </TabPane>
      </Tabs>
    </Card>

    <!-- 标签编辑弹窗 -->
    <Modal
      v-model:open="tagModalVisible"
      :title="tagEditing ? '编辑标签' : '新增标签'"
      @ok="handleTagSubmit"
      width="500px"
    >
      <Form :model="tagForm" layout="vertical">
        <Form.Item label="标签键" required>
          <Input v-model:value="tagForm.tagKey" placeholder="如: type, source, sensitivity" />
        </Form.Item>
        <Form.Item label="标签值" required>
          <Input v-model:value="tagForm.tagValue" placeholder="如: image, video, user" />
        </Form.Item>
        <Form.Item label="显示名称" required>
          <Input v-model:value="tagForm.tagName" placeholder="如: 图片, 视频, 用户上传" />
        </Form.Item>
        <Form.Item label="图标">
          <Input v-model:value="tagForm.icon" placeholder="如: 🖼️, 🎬, 👤" />
        </Form.Item>
        <Form.Item label="颜色">
          <Input v-model:value="tagForm.color" placeholder="#1890ff" />
        </Form.Item>
        <Form.Item label="描述">
          <Input.TextArea v-model:value="tagForm.description" />
        </Form.Item>
        <Form.Item label="排序">
          <InputNumber v-model:value="tagForm.sortOrder" :min="0" />
        </Form.Item>
      </Form>
    </Modal>

    <!-- 路由规则编辑弹窗 -->
    <Modal
      v-model:open="ruleModalVisible"
      :title="ruleEditing ? '编辑规则' : '新增规则'"
      @ok="handleRuleSubmit"
      width="600px"
    >
      <Form :model="ruleForm" layout="vertical">
        <Form.Item label="规则名称" required>
          <Input v-model:value="ruleForm.ruleName" placeholder="如: 图片存储, 视频存储" />
        </Form.Item>
        <Form.Item label="描述">
          <Input.TextArea v-model:value="ruleForm.description" />
        </Form.Item>
        <Form.Item label="优先级">
          <InputNumber v-model:value="ruleForm.priority" :min="0" />
          <span class="ml-2 text-gray-500">数值越大优先级越高</span>
        </Form.Item>
        <Form.Item label="匹配类型">
          <Select v-model:value="ruleForm.matchType" :options="matchTypeOptions" />
        </Form.Item>
        <Form.Item label="匹配条件">
          <div v-for="(condition, index) in ruleForm.conditions" :key="index" class="mb-2 flex gap-2">
            <Select
              v-model:value="condition.key"
              :options="tagKeyOptions"
              placeholder="标签键"
              style="width: 150px"
            />
            <Select
              v-model:value="condition.value"
              :options="getTagValueOptions(condition.key)"
              placeholder="标签值"
              style="width: 200px"
            />
            <Button type="link" danger @click="removeCondition(index)">删除</Button>
          </div>
          <Button type="dashed" @click="addCondition" block>
            <PlusOutlined />
            添加条件
          </Button>
        </Form.Item>
        <Form.Item label="目标存储" required>
          <Select v-model:value="ruleForm.driver" :options="driverOptions" />
        </Form.Item>
        <Form.Item label="桶名称" v-if="ruleForm.driver !== 'local'">
          <Input v-model:value="ruleForm.bucket" placeholder="如: devkit-images" />
        </Form.Item>
        <Form.Item label="路径前缀">
          <Input v-model:value="ruleForm.pathPrefix" placeholder="如: images/, videos/" />
        </Form.Item>
        <Form.Item label="默认规则">
          <Switch v-model:checked="ruleForm.isDefault" />
          <span class="ml-2 text-gray-500">设为兜底规则（只能有一个）</span>
        </Form.Item>
      </Form>
    </Modal>

    <!-- 测试路由弹窗 -->
    <Modal
      v-model:open="testModalVisible"
      title="测试路由"
      @ok="handleTestRoute"
      width="500px"
    >
      <Form :model="testForm" layout="vertical">
        <Form.Item label="文件名" required>
          <Input v-model:value="testForm.fileName" placeholder="如: photo.jpg" />
        </Form.Item>
        <Form.Item label="文件类型">
          <Input v-model:value="testForm.contentType" placeholder="如: image/jpeg" />
        </Form.Item>
        <Form.Item label="来源">
          <Select v-model:value="testForm.source">
            <Select.Option value="user">用户上传</Select.Option>
            <Select.Option value="system">系统生成</Select.Option>
            <Select.Option value="import">批量导入</Select.Option>
          </Select>
        </Form.Item>
      </Form>

      <div v-if="testResult" class="mt-4 p-4 bg-gray-50 rounded">
        <h4 class="mb-2">匹配结果:</h4>
        <div class="mb-2">
          <strong>规则:</strong> {{ testResult.result?.ruleName || '无匹配' }}
        </div>
        <div class="mb-2">
          <strong>存储:</strong> {{ testResult.result?.driver }}
        </div>
        <div class="mb-2">
          <strong>桶:</strong> {{ testResult.result?.bucket || '-' }}
        </div>
        <div class="mb-2">
          <strong>路径前缀:</strong> {{ testResult.result?.pathPrefix || '-' }}
        </div>
        <div>
          <strong>自动生成标签:</strong>
          <Tag v-for="tag in testResult.tags" :key="`${tag.key}-${tag.value}`" class="ml-1">
            {{ tag.key }}:{{ tag.value }}
          </Tag>
        </div>
      </div>
    </Modal>
  </div>
</template>
