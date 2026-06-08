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
const tagColumns = computed(() => [
  { title: t('system.tag.tagKey'), dataIndex: 'tagKey', width: 120 },
  { title: t('system.tag.tagValue'), dataIndex: 'tagValue', width: 120 },
  { title: t('system.tag.tagName'), dataIndex: 'tagName', width: 120 },
  { title: t('system.tag.icon'), dataIndex: 'icon', width: 80 },
  { title: t('system.tag.color'), dataIndex: 'color', width: 100 },
  { title: t('system.tag.sortOrder'), dataIndex: 'sortOrder', width: 80 },
  { title: t('system.tag.fileCount'), dataIndex: 'fileCount', width: 80 },
  { title: t('system.tag.isSystem'), dataIndex: 'isSystem', width: 100 },
  { title: t('common.operation'), key: 'action', width: 150 },
]);

// 规则列定义
const ruleColumns = computed(() => [
  { title: t('system.tag.ruleName'), dataIndex: 'ruleName', width: 150 },
  { title: t('system.tag.priority'), dataIndex: 'priority', width: 80 },
  { title: t('system.tag.matchType'), dataIndex: 'matchType', width: 100 },
  { title: t('system.tag.targetStorage'), dataIndex: 'driver', width: 100 },
  { title: t('system.tag.bucketPath'), key: 'bucket', width: 150 },
  { title: t('system.tag.defaultRule'), dataIndex: 'isDefault', width: 100 },
  { title: t('common.status'), dataIndex: 'status', width: 80 },
  { title: t('common.operation'), key: 'action', width: 200 },
]);

// 获取文件数量
const getFileCount = (tagId: number) => {
  const stat = tagStats.value.find((s) => s.id === tagId);
  return stat?.fileCount || 0;
};

// 存储驱动选项
const driverOptions = computed(() => [
  { label: t('system.tag.storageLocal'), value: 'local' },
  { label: t('system.tag.storageMinio'), value: 'minio' },
  { label: t('system.tag.storageOss'), value: 'oss' },
  { label: t('system.tag.storageCos'), value: 'cos' },
]);

// 匹配类型选项
const matchTypeOptions = computed(() => [
  { label: t('system.tag.matchAll'), value: 'all' },
  { label: t('system.tag.matchAny'), value: 'any' },
  { label: t('system.tag.matchExact'), value: 'exact' },
]);

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
        <TabPane key="tags" :tab="t('system.tag.tagManagement')">
          <template #tab>
            <span>
              <TagsOutlined />
              {{ t('system.tag.tagManagement') }}
            </span>
          </template>

          <div class="mb-4">
            <Space>
              <Button type="primary" @click="openTagModal()">
                <PlusOutlined />
                {{ t('system.tag.addTag') }}
              </Button>
              <Button @click="loadTags">
                <ReloadOutlined />
                {{ t('common.refresh') }}
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
                  {{ record.isSystem ? t('system.tag.systemTag') : t('system.tag.customTag') }}
                </Tag>
              </template>
              <template v-if="column.key === 'action'">
                <Space>
                  <Tooltip :title="t('common.edit')">
                    <Button type="link" size="small" @click="openTagModal(record)" :disabled="record.isSystem">
                      <EditOutlined />
                    </Button>
                  </Tooltip>
                  <Popconfirm
                    :title="t('system.tag.deleteTagConfirm')"
                    @confirm="handleDeleteTag(record.id)"
                    :disabled="record.isSystem"
                  >
                    <Tooltip :title="t('common.delete')">
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

        <TabPane key="rules" :tab="t('system.tag.routingRules')">
          <template #tab>
            <span>
              <NodeIndexOutlined />
              {{ t('system.tag.routingRules') }}
            </span>
          </template>

          <div class="mb-4">
            <Space>
              <Button type="primary" @click="openRuleModal()">
                <PlusOutlined />
                {{ t('system.tag.addRule') }}
              </Button>
              <Button @click="loadRules">
                <ReloadOutlined />
                {{ t('common.refresh') }}
              </Button>
              <Button @click="openTestModal">
                {{ t('system.tag.testRoute') }}
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
                <Tag color="blue">{{ record.matchType === 'all' ? t('system.tag.matchAll') : record.matchType === 'any' ? t('system.tag.matchAny') : t('system.tag.matchExact') }}</Tag>
              </template>
              <template v-if="column.key === 'bucket'">
                {{ record.bucket || '-' }}{{ record.pathPrefix ? `/${record.pathPrefix}` : '' }}
              </template>
              <template v-if="column.dataIndex === 'isDefault'">
                <Tag :color="record.isDefault ? 'green' : 'default'">
                  {{ record.isDefault ? t('common.yes') : t('common.no') }}
                </Tag>
              </template>
              <template v-if="column.dataIndex === 'status'">
                <Switch
                  :checked="record.status === 1"
                  @change="() => handleToggleRuleStatus(record)"
                  :checked-children="t('common.enable')"
                  :un-checked-children="t('common.disable')"
                />
              </template>
              <template v-if="column.key === 'action'">
                <Space>
                  <Tooltip :title="t('common.edit')">
                    <Button type="link" size="small" @click="openRuleModal(record)">
                      <EditOutlined />
                    </Button>
                  </Tooltip>
                  <Popconfirm
                    :title="t('system.tag.deleteRuleConfirm')"
                    @confirm="handleDeleteRule(record.id)"
                    :disabled="record.isDefault"
                  >
                    <Tooltip :title="t('common.delete')">
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
      :title="tagEditing ? t('system.tag.editTag') : t('system.tag.addTag')"
      @ok="handleTagSubmit"
      width="500px"
    >
      <Form :model="tagForm" layout="vertical">
        <Form.Item :label="t('system.tag.tagKey')" required>
          <Input v-model:value="tagForm.tagKey" :placeholder="t('system.tag.tagKeyPlaceholder')" />
        </Form.Item>
        <Form.Item :label="t('system.tag.tagValue')" required>
          <Input v-model:value="tagForm.tagValue" :placeholder="t('system.tag.tagValuePlaceholder')" />
        </Form.Item>
        <Form.Item :label="t('system.tag.tagName')" required>
          <Input v-model:value="tagForm.tagName" :placeholder="t('system.tag.tagNamePlaceholder')" />
        </Form.Item>
        <Form.Item :label="t('system.tag.icon')">
          <Input v-model:value="tagForm.icon" :placeholder="t('system.tag.iconPlaceholder')" />
        </Form.Item>
        <Form.Item :label="t('system.tag.color')">
          <Input v-model:value="tagForm.color" placeholder="#1890ff" />
        </Form.Item>
        <Form.Item :label="t('system.tag.description')">
          <Input.TextArea v-model:value="tagForm.description" />
        </Form.Item>
        <Form.Item :label="t('system.tag.sortOrder')">
          <InputNumber v-model:value="tagForm.sortOrder" :min="0" />
        </Form.Item>
      </Form>
    </Modal>

    <!-- 路由规则编辑弹窗 -->
    <Modal
      v-model:open="ruleModalVisible"
      :title="ruleEditing ? t('system.tag.editRule') : t('system.tag.addRule')"
      @ok="handleRuleSubmit"
      width="600px"
    >
      <Form :model="ruleForm" layout="vertical">
        <Form.Item :label="t('system.tag.ruleName')" required>
          <Input v-model:value="ruleForm.ruleName" :placeholder="t('system.tag.ruleNamePlaceholder')" />
        </Form.Item>
        <Form.Item :label="t('system.tag.description')">
          <Input.TextArea v-model:value="ruleForm.description" />
        </Form.Item>
        <Form.Item :label="t('system.tag.priority')">
          <InputNumber v-model:value="ruleForm.priority" :min="0" />
          <span class="ml-2 text-gray-500">{{ t('system.tag.priorityHelp') }}</span>
        </Form.Item>
        <Form.Item :label="t('system.tag.matchType')">
          <Select v-model:value="ruleForm.matchType" :options="matchTypeOptions" />
        </Form.Item>
        <Form.Item :label="t('system.tag.matchConditions')">
          <div v-for="(condition, index) in ruleForm.conditions" :key="index" class="mb-2 flex gap-2">
            <Select
              v-model:value="condition.key"
              :options="tagKeyOptions"
              :placeholder="t('system.tag.tagKeyLabel')"
              style="width: 150px"
            />
            <Select
              v-model:value="condition.value"
              :options="getTagValueOptions(condition.key)"
              :placeholder="t('system.tag.tagValueLabel')"
              style="width: 200px"
            />
            <Button type="link" danger @click="removeCondition(index)">{{ t('common.delete') }}</Button>
          </div>
          <Button type="dashed" @click="addCondition" block>
            <PlusOutlined />
            {{ t('system.tag.addCondition') }}
          </Button>
        </Form.Item>
        <Form.Item :label="t('system.tag.targetStorage')" required>
          <Select v-model:value="ruleForm.driver" :options="driverOptions" />
        </Form.Item>
        <Form.Item :label="t('system.tag.bucketName')" v-if="ruleForm.driver !== 'local'">
          <Input v-model:value="ruleForm.bucket" :placeholder="t('system.tag.bucketNamePlaceholder')" />
        </Form.Item>
        <Form.Item :label="t('system.tag.pathPrefix')">
          <Input v-model:value="ruleForm.pathPrefix" :placeholder="t('system.tag.pathPrefixPlaceholder')" />
        </Form.Item>
        <Form.Item :label="t('system.tag.defaultRule')">
          <Switch v-model:checked="ruleForm.isDefault" />
          <span class="ml-2 text-gray-500">{{ t('system.tag.setDefaultRuleHelp') }}</span>
        </Form.Item>
      </Form>
    </Modal>

    <!-- 测试路由弹窗 -->
    <Modal
      v-model:open="testModalVisible"
      :title="t('system.tag.testRouteTitle')"
      @ok="handleTestRoute"
      width="500px"
    >
      <Form :model="testForm" layout="vertical">
        <Form.Item :label="t('system.tag.fileName')" required>
          <Input v-model:value="testForm.fileName" :placeholder="t('system.tag.fileNamePlaceholder')" />
        </Form.Item>
        <Form.Item :label="t('system.tag.fileType')">
          <Input v-model:value="testForm.contentType" :placeholder="t('system.tag.fileTypePlaceholder')" />
        </Form.Item>
        <Form.Item :label="t('system.tag.source')">
          <Select v-model:value="testForm.source">
            <Select.Option value="user">{{ t('system.tag.sourceUser') }}</Select.Option>
            <Select.Option value="system">{{ t('system.tag.sourceSystem') }}</Select.Option>
            <Select.Option value="import">{{ t('system.tag.sourceImport') }}</Select.Option>
          </Select>
        </Form.Item>
      </Form>

      <div v-if="testResult" class="mt-4 p-4 bg-gray-50 rounded">
        <h4 class="mb-2">{{ t('system.tag.testResult') }}:</h4>
        <div class="mb-2">
          <strong>{{ t('system.tag.matchedRule') }}:</strong> {{ testResult.result?.ruleName || t('system.tag.noMatch') }}
        </div>
        <div class="mb-2">
          <strong>{{ t('system.tag.targetStorage') }}:</strong> {{ testResult.result?.driver }}
        </div>
        <div class="mb-2">
          <strong>{{ t('system.tag.bucketName') }}:</strong> {{ testResult.result?.bucket || '-' }}
        </div>
        <div class="mb-2">
          <strong>{{ t('system.tag.pathPrefix') }}:</strong> {{ testResult.result?.pathPrefix || '-' }}
        </div>
        <div>
          <strong>{{ t('system.tag.autoTags') }}:</strong>
          <Tag v-for="tag in testResult.tags" :key="`${tag.key}-${tag.value}`" class="ml-1">
            {{ tag.key }}:{{ tag.value }}
          </Tag>
        </div>
      </div>
    </Modal>
  </div>
</template>
