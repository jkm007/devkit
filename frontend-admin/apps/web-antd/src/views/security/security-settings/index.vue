<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Alert,
  Button,
  Card,
  Col,
  Divider,
  Input,
  InputNumber,
  message,
  Modal,
  Popconfirm,
  Row,
  Select,
  Spin,
  Switch,
  Table,
  Tabs,
  TabPane,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import {
  createRateLimitRule,
  deleteRateLimitRule,
  getRateLimitRules,
  updateRateLimitRule,
  updateRateLimitRuleStatus,
} from '#/api/security/rate-limit';
import type { RateLimitRuleApi } from '#/api/security/rate-limit';
import { getAllSettings, updateSettingsByGroup } from '#/api/system/settings';
import { $t } from '#/locales';

// ==================== State ====================
const loading = ref(false);
const saving = ref(false);
const allSettings = ref<Record<string, any[]>>({});

const riskForm = reactive<Record<string, any>>({});
const securityForm = reactive<Record<string, any>>({});

// ==================== 限流规则 State ====================
const rateLimitRules = ref<RateLimitRuleApi.Rule[]>([]);
const rateLimitLoading = ref(false);
const rateLimitModalVisible = ref(false);
const rateLimitEditing = ref<RateLimitRuleApi.Rule | null>(null);
const rateLimitForm = reactive<RateLimitRuleApi.RuleForm>({
  pathPattern: '',
  method: '*',
  rate: 10,
  burst: 20,
  cooldown: 0,
  blockDuration: 0,
  maxViolations: 0,
  violationScore: 0,
  description: '',
  enabled: true,
  priority: 0,
});

const methodOptions = [
  { label: '所有方法 (*)', value: '*' },
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'DELETE', value: 'DELETE' },
];

const rateLimitColumns = [
  {
    title: '路径模式',
    dataIndex: 'pathPattern',
    key: 'pathPattern',
    width: 160,
  },
  { title: '方法', dataIndex: 'method', key: 'method', width: 60 },
  {
    title: '速率',
    dataIndex: 'rate',
    key: 'rate',
    width: 70,
    customRender: ({ text }: { text: number }) => `${text}/s`,
  },
  { title: '突发', dataIndex: 'burst', key: 'burst', width: 60 },
  { title: '冷却(s)', dataIndex: 'cooldown', key: 'cooldown', width: 80 },
  {
    title: '封禁(s)',
    dataIndex: 'blockDuration',
    key: 'blockDuration',
    width: 80,
  },
  {
    title: '违规阈值',
    dataIndex: 'maxViolations',
    key: 'maxViolations',
    width: 80,
  },
  {
    title: '风险分',
    dataIndex: 'violationScore',
    key: 'violationScore',
    width: 80,
  },
  {
    title: '描述',
    dataIndex: 'description',
    key: 'description',
    ellipsis: true,
  },
  { title: '状态', dataIndex: 'enabled', key: 'enabled', width: 60 },
  { title: '操作', key: 'actions', width: 150, fixed: 'right' as const },
];

// ==================== Computed ====================
const riskItems = computed(() => allSettings.value.risk_score || []);
const securityItems = computed(() => allSettings.value.security || []);

const riskGeneralItems = computed(() =>
  riskItems.value.filter(
    (i: any) => i.key.startsWith('risk_') && !i.key.startsWith('rule_'),
  ),
);
const riskFrequencyItems = computed(() =>
  riskItems.value.filter((i: any) => i.key.startsWith('rule_frequency')),
);
const riskNoRefererItems = computed(() =>
  riskItems.value.filter((i: any) => i.key.startsWith('rule_no_referer')),
);
const riskNoLangItems = computed(() =>
  riskItems.value.filter((i: any) => i.key.startsWith('rule_no_lang')),
);
const riskUaItems = computed(() =>
  riskItems.value.filter((i: any) => i.key.startsWith('rule_ua')),
);
const riskIntervalItems = computed(() =>
  riskItems.value.filter((i: any) => i.key.startsWith('rule_interval')),
);

// ==================== Methods ====================
async function loadSettings() {
  loading.value = true;
  try {
    const res = await getAllSettings();
    allSettings.value = res;

    // Populate risk_score form
    if (res.risk_score) {
      for (const item of res.risk_score) {
        riskForm[item.key] = item.value;
      }
    }
    // Populate security form
    if (res.security) {
      for (const item of res.security) {
        securityForm[item.key] = item.value;
      }
    }
  } catch {
    message.error($t('system.settings.loadError'));
  } finally {
    loading.value = false;
  }
}

async function handleSaveRisk() {
  saving.value = true;
  try {
    const result = await updateSettingsByGroup('risk_score', { ...riskForm });
    message.success(
      $t('system.settings.saveSuccess', { count: result.updated }),
    );
    if (result.restartRequired) {
      Modal.warning({
        title: $t('system.settings.restartRequired'),
        content: $t('system.settings.restartItems', {
          items: result.restartItems.join(', '),
        }),
      });
    }
    await loadSettings();
  } catch {
    message.error($t('system.settings.saveError'));
  } finally {
    saving.value = false;
  }
}

async function handleSaveSecurity() {
  saving.value = true;
  try {
    const result = await updateSettingsByGroup('security', { ...securityForm });
    message.success(
      $t('system.settings.saveSuccess', { count: result.updated }),
    );
    if (result.restartRequired) {
      Modal.warning({
        title: $t('system.settings.restartRequired'),
        content: $t('system.settings.restartItems', {
          items: result.restartItems.join(', '),
        }),
      });
    }
    await loadSettings();
  } catch {
    message.error($t('system.settings.saveError'));
  } finally {
    saving.value = false;
  }
}

// ==================== 限流规则 Methods ====================
async function loadRateLimitRules() {
  rateLimitLoading.value = true;
  try {
    rateLimitRules.value = await getRateLimitRules();
  } catch {
    message.error('获取限流规则失败');
  } finally {
    rateLimitLoading.value = false;
  }
}

function openRateLimitModal(record?: RateLimitRuleApi.Rule) {
  rateLimitEditing.value = record || null;
  if (record) {
    rateLimitForm.pathPattern = record.pathPattern;
    rateLimitForm.method = record.method;
    rateLimitForm.rate = record.rate;
    rateLimitForm.burst = record.burst;
    rateLimitForm.cooldown = record.cooldown;
    rateLimitForm.blockDuration = record.blockDuration;
    rateLimitForm.maxViolations = record.maxViolations;
    rateLimitForm.violationScore = record.violationScore;
    rateLimitForm.description = record.description;
    rateLimitForm.enabled = record.enabled;
    rateLimitForm.priority = record.priority;
  } else {
    rateLimitForm.pathPattern = '';
    rateLimitForm.method = '*';
    rateLimitForm.rate = 10;
    rateLimitForm.burst = 20;
    rateLimitForm.cooldown = 0;
    rateLimitForm.blockDuration = 0;
    rateLimitForm.maxViolations = 0;
    rateLimitForm.violationScore = 0;
    rateLimitForm.description = '';
    rateLimitForm.enabled = true;
    rateLimitForm.priority = 0;
  }
  rateLimitModalVisible.value = true;
}

async function handleRateLimitSubmit() {
  try {
    if (rateLimitEditing.value) {
      await updateRateLimitRule(rateLimitEditing.value.id, {
        ...rateLimitForm,
      });
      message.success('规则更新成功');
    } else {
      await createRateLimitRule({ ...rateLimitForm });
      message.success('规则创建成功');
    }
    rateLimitModalVisible.value = false;
    await loadRateLimitRules();
  } catch {
    message.error('操作失败');
  }
}

async function handleRateLimitDelete(id: number) {
  try {
    await deleteRateLimitRule(id);
    message.success('规则删除成功');
    await loadRateLimitRules();
  } catch {
    message.error('删除失败');
  }
}

async function handleRateLimitStatusChange(rule: RateLimitRuleApi.Rule) {
  try {
    await updateRateLimitRuleStatus(rule.id, !rule.enabled);
    message.success(rule.enabled ? '已禁用' : '已启用');
    await loadRateLimitRules();
  } catch {
    message.error('状态更新失败');
  }
}

// ==================== Lifecycle ====================
onMounted(() => {
  loadSettings();
  loadRateLimitRules();
});
</script>

<template>
  <Page title="安全设置" auto-content-height>
    <Spin :spinning="loading">
      <Card>
        <Tabs default-active-key="risk_score">
          <!-- ==================== 风险评分 ==================== -->
          <TabPane key="risk_score" tab="⚠️ 风险评分">
            <!-- 通用配置 -->
            <div class="settings-section">
              <h3 class="settings-section-title">⚙️ 通用配置</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in riskGeneralItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help"
                          >ⓘ</span
                        >
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="riskForm[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="riskForm[item.key]"
                      class="w-full"
                      :min="0"
                    />
                    <Input
                      v-else-if="item.type === 'string'"
                      v-model:value="riskForm[item.key]"
                      :placeholder="item.tip"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <Divider />

            <!-- 频率检测规则 -->
            <div class="settings-section">
              <h3 class="settings-section-title">📊 频率检测规则</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in riskFrequencyItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help"
                          >ⓘ</span
                        >
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="riskForm[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="riskForm[item.key]"
                      class="w-full"
                      :min="0"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <Divider />

            <!-- 无 Referer 规则 -->
            <div class="settings-section">
              <h3 class="settings-section-title">🔗 无 Referer 规则</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in riskNoRefererItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help"
                          >ⓘ</span
                        >
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="riskForm[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="riskForm[item.key]"
                      class="w-full"
                      :min="0"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <Divider />

            <!-- 无 Accept-Language 规则 -->
            <div class="settings-section">
              <h3 class="settings-section-title">🌐 无 Accept-Language 规则</h3>
              <Row :gutter="[16, 16]">
                <Col v-for="item in riskNoLangItems" :key="item.key" :span="12">
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help"
                          >ⓘ</span
                        >
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="riskForm[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="riskForm[item.key]"
                      class="w-full"
                      :min="0"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <Divider />

            <!-- UA 异常规则 -->
            <div class="settings-section">
              <h3 class="settings-section-title">🤖 UA 异常规则</h3>
              <Row :gutter="[16, 16]">
                <Col v-for="item in riskUaItems" :key="item.key" :span="12">
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help"
                          >ⓘ</span
                        >
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="riskForm[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="riskForm[item.key]"
                      class="w-full"
                      :min="0"
                    />
                    <Input
                      v-else-if="item.type === 'string'"
                      v-model:value="riskForm[item.key]"
                      :placeholder="item.tip"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <Divider />

            <!-- 请求间隔规则 -->
            <div class="settings-section">
              <h3 class="settings-section-title">⏱️ 请求间隔规则</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in riskIntervalItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help"
                          >ⓘ</span
                        >
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="riskForm[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="riskForm[item.key]"
                      class="w-full"
                      :min="0"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <Divider />

            <div class="flex justify-end">
              <Button type="primary" :loading="saving" @click="handleSaveRisk">
                {{ $t('system.settings.saveGroup') }}
              </Button>
            </div>
          </TabPane>

          <!-- ==================== 安全设置 ==================== -->
          <TabPane key="security" tab="🛡️ 安全设置">
            <Alert
              :message="$t('system.settings.securityTip')"
              type="info"
              show-icon
              class="mb-4"
            />
            <Row :gutter="[16, 16]">
              <Col v-for="item in securityItems" :key="item.key" :span="12">
                <div class="setting-item">
                  <label class="setting-label">
                    {{ item.label }}
                    <Tooltip v-if="item.tip" :title="item.tip">
                      <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                    </Tooltip>
                  </label>
                  <Input
                    v-if="item.type === 'string'"
                    v-model:value="securityForm[item.key]"
                    :placeholder="item.tip"
                  />
                  <Switch
                    v-else-if="item.type === 'boolean'"
                    v-model:checked="securityForm[item.key]"
                  />
                  <InputNumber
                    v-else-if="item.type === 'number'"
                    v-model:value="securityForm[item.key]"
                    class="w-full"
                    :min="0"
                  />
                </div>
              </Col>
            </Row>

            <Divider />

            <div class="flex justify-end">
              <Button
                type="primary"
                :loading="saving"
                @click="handleSaveSecurity"
              >
                {{ $t('system.settings.saveGroup') }}
              </Button>
            </div>
          </TabPane>

          <!-- ==================== 限流规则 ==================== -->
          <TabPane key="rate_limit" tab="🚦 限流规则">
            <Alert
              message="接口限流规则管理"
              description="配置各 API 接口的请求频率限制，支持路径通配符匹配。规则按优先级排序，数值越大越先匹配。修改后自动生效，无需重启服务。"
              type="info"
              show-icon
              class="mb-4"
            />

            <div class="mb-4 flex justify-end">
              <Button type="primary" @click="openRateLimitModal()">
                + 添加规则
              </Button>
            </div>

            <Table
              :columns="rateLimitColumns"
              :data-source="rateLimitRules"
              :loading="rateLimitLoading"
              row-key="id"
              :pagination="false"
              size="middle"
              :scroll="{ x: 900 }"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'method'">
                  <Tag :color="record.method === '*' ? 'blue' : 'green'">
                    {{ record.method }}
                  </Tag>
                </template>
                <template v-else-if="column.key === 'enabled'">
                  <Tag :color="record.enabled ? 'success' : 'default'">
                    {{ record.enabled ? '启用' : '禁用' }}
                  </Tag>
                </template>
                <template v-else-if="column.key === 'actions'">
                  <div class="flex gap-2">
                    <Button
                      type="link"
                      size="small"
                      @click="
                        openRateLimitModal(record as RateLimitRuleApi.Rule)
                      "
                    >
                      编辑
                    </Button>
                    <Button
                      type="link"
                      size="small"
                      @click="
                        handleRateLimitStatusChange(
                          record as RateLimitRuleApi.Rule,
                        )
                      "
                    >
                      {{
                        (record as RateLimitRuleApi.Rule).enabled
                          ? '禁用'
                          : '启用'
                      }}
                    </Button>
                    <Popconfirm
                      title="确定删除此规则？"
                      @confirm="
                        handleRateLimitDelete(
                          (record as RateLimitRuleApi.Rule).id,
                        )
                      "
                    >
                      <Button type="link" size="small" danger> 删除 </Button>
                    </Popconfirm>
                  </div>
                </template>
              </template>
            </Table>
          </TabPane>
        </Tabs>
      </Card>
    </Spin>

    <!-- 限流规则编辑弹窗 -->
    <Modal
      v-model:open="rateLimitModalVisible"
      :title="rateLimitEditing ? '编辑限流规则' : '添加限流规则'"
      @ok="handleRateLimitSubmit"
    >
      <div class="space-y-4 py-4">
        <div>
          <label class="mb-1 block text-sm font-medium"
            >路径模式 <span class="text-red-500">*</span></label
          >
          <Input
            v-model:value="rateLimitForm.pathPattern"
            placeholder="例如: /auth/login 或 /share/*"
          />
          <p class="mt-1 text-xs text-gray-400">
            支持 * 通配符，如 /auth/* 匹配所有认证接口
          </p>
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">HTTP 方法</label>
          <Select
            v-model:value="rateLimitForm.method"
            :options="methodOptions"
            class="w-full"
          />
        </div>
        <Row :gutter="16">
          <Col :span="12">
            <label class="mb-1 block text-sm font-medium"
              >速率 (req/s) <span class="text-red-500">*</span></label
            >
            <InputNumber
              v-model:value="rateLimitForm.rate"
              :min="0.1"
              :step="1"
              class="w-full"
              placeholder="每秒请求数"
            />
          </Col>
          <Col :span="12">
            <label class="mb-1 block text-sm font-medium"
              >突发容量 <span class="text-red-500">*</span></label
            >
            <InputNumber
              v-model:value="rateLimitForm.burst"
              :min="1"
              :step="1"
              class="w-full"
              placeholder="突发请求数"
            />
          </Col>
        </Row>
        <Divider orientation="left" class="text-sm">缓冲配置</Divider>
        <Row :gutter="16">
          <Col :span="8">
            <label class="mb-1 block text-sm font-medium">
              冷却时间(秒)
              <Tooltip title="触发限流后需要等待多久才能恢复请求">
                <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
              </Tooltip>
            </label>
            <InputNumber
              v-model:value="rateLimitForm.cooldown"
              :min="0"
              class="w-full"
              placeholder="0=无冷却"
            />
          </Col>
          <Col :span="8">
            <label class="mb-1 block text-sm font-medium">
              封禁时长(秒)
              <Tooltip title="超过违规次数后封禁 IP 的时长">
                <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
              </Tooltip>
            </label>
            <InputNumber
              v-model:value="rateLimitForm.blockDuration"
              :min="0"
              class="w-full"
              placeholder="0=不封禁"
            />
          </Col>
          <Col :span="6">
            <label class="mb-1 block text-sm font-medium">
              违规阈值
              <Tooltip title="触发封禁前允许的违规次数">
                <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
              </Tooltip>
            </label>
            <InputNumber
              v-model:value="rateLimitForm.maxViolations"
              :min="0"
              class="w-full"
              placeholder="0=不限制"
            />
          </Col>
          <Col :span="6">
            <label class="mb-1 block text-sm font-medium">
              风险分
              <Tooltip title="触发限流时累加到风险评分系统的分数，0=不累加">
                <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
              </Tooltip>
            </label>
            <InputNumber
              v-model:value="rateLimitForm.violationScore"
              :min="0"
              class="w-full"
              placeholder="0=不累加"
            />
          </Col>
        </Row>
        <Divider orientation="left" class="text-sm">其他配置</Divider>
        <div>
          <label class="mb-1 block text-sm font-medium">优先级</label>
          <InputNumber
            v-model:value="rateLimitForm.priority"
            :min="0"
            class="w-full"
            placeholder="数值越大越先匹配"
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">描述</label>
          <Input
            v-model:value="rateLimitForm.description"
            placeholder="规则用途说明"
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">启用</label>
          <Switch v-model:checked="rateLimitForm.enabled" />
        </div>
      </div>
    </Modal>
  </Page>
</template>

<style scoped>
.settings-section {
  margin-bottom: 8px;
}

.settings-section-title {
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 600;
  color: rgb(0 0 0 / 85%);
}

.setting-item {
  margin-bottom: 8px;
}

.setting-label {
  display: block;
  margin-bottom: 4px;
  font-size: 13px;
  font-weight: 500;
}
</style>
