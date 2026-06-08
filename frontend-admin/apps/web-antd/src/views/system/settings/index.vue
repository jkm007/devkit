<script lang="ts" setup>
import type { SystemSettingsApi } from '#/api/system/settings';

import { computed, onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';
import { updatePreferences } from '@vben/preferences';

import {
  Alert,
  Button,
  Card,
  Col,
  Divider,
  Input,
  InputNumber,
  Menu,
  MenuItem,
  message,
  Modal,
  Row,
  Select,
  SelectOption,
  Slider,
  Spin,
  Switch,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import {
  getAllSettings,
  testCaptcha,
  testEmail,
  testSms,
  updateSettingsByGroup,
  verifyCaptcha,
} from '#/api/system/settings';
import {
  NumericCaptcha,
  PointSelectionCaptcha,
} from '@vben/common-ui';
import BackendCaptcha from '#/components/captcha/backend-captcha.vue';
import BackendRotateCaptcha from '#/components/captcha/backend-rotate-captcha.vue';
import { $t } from '#/locales';

// ==================== Group Config ====================
const groupConfig: Record<
  string,
  { icon: string; title: string; description: string }
> = {
  basic: {
    icon: '🏠',
    title: $t('system.settings.groups.basic'),
    description: $t('system.settings.groupDesc.basic'),
  },
  auth: {
    icon: '🔐',
    title: '认证设置',
    description: '管理登录方式的启用与配置',
  },
  email: {
    icon: '📧',
    title: $t('system.settings.groups.email'),
    description: $t('system.settings.groupDesc.email'),
  },
  sms: {
    icon: '💬',
    title: $t('system.settings.groups.sms'),
    description: $t('system.settings.groupDesc.sms'),
  },
  captcha: {
    icon: '🔒',
    title: $t('system.settings.groups.captcha'),
    description: $t('system.settings.groupDesc.captcha'),
  },
  risk_score: {
    icon: '⚠️',
    title: $t('system.settings.groups.riskScore'),
    description: $t('system.settings.groupDesc.riskScore'),
  },
  storage: {
    icon: '📁',
    title: $t('system.settings.groups.storage'),
    description: $t('system.settings.groupDesc.storage'),
  },
  wechat: {
    icon: '💚',
    title: $t('system.settings.groups.wechat'),
    description: $t('system.settings.groupDesc.wechat'),
  },
  security: {
    icon: '🛡️',
    title: $t('system.settings.groups.security'),
    description: $t('system.settings.groupDesc.security'),
  },
};

const ALL_GROUPS = [
  'basic',
  'auth',
  'email',
  'sms',
  'captcha',
  'risk_score',
  'storage',
  'wechat',
  'security',
];

// 验证码类型中文映射
const CAPTCHA_TYPE_LABELS: Record<string, string> = {
  numeric: '数字',
  slider: '滑块',
  puzzle: '拼图',
  rotation: '旋转',
  point: '点选',
};

// ==================== State ====================
const loading = ref(false);
const saving = ref(false);
const selectedKeys = ref<string[]>(['basic']);
const activeGroup = computed(() => selectedKeys.value[0] || 'basic');
const allSettings = ref<SystemSettingsApi.SettingsGroup>({});

// Pre-initialize form values for all groups
const formValues = reactive<Record<string, Record<string, any>>>(
  Object.fromEntries(ALL_GROUPS.map((g) => [g, {}])),
);

// Test email/sms
const testEmailAddress = ref('');
const testPhoneNumber = ref('');
const testEmailLoading = ref(false);
const testSmsLoading = ref(false);

// ==================== Captcha Test ====================
const captchaTestVisible = ref(false);
const captchaTestLoading = ref(false);
const captchaTestVerifying = ref(false);
const captchaTestId = ref('');
const captchaTestImage = ref('');
const captchaTestThumb = ref('');
const captchaTestThumbY = ref(0);  // 缩略图初始 Y 位置
const captchaTestType = ref('numeric');
const captchaTestStartTime = ref(0);
const captchaTestHintText = ref('');
const captchaTestChars = ref<string[]>([]);
const captchaTestLength = ref(4);  // 数字验证码长度
const captchaTestResult = ref<{ valid: boolean; message: string } | null>(null);

// 滑块/拼图组件引用 & 状态
const captchaSliderPassing = ref(false);

// 旋转组件引用 & 状态
const captchaRotatePassing = ref(false);

// 点选状态
const captchaPointPassing = ref(false);

const CAPTCHA_IMG_W = 320;
const CAPTCHA_IMG_H_POINT = 220;  // point/click (go-captcha 使用 220)

// ==================== Computed ====================
const groupKeys = computed(() => {
  return Object.keys(allSettings.value).filter((key) => groupConfig[key]);
});

const currentGroupItems = computed(() => {
  return allSettings.value[activeGroup.value] || [];
});

const menuItems = computed(() => {
  return groupKeys.value.map((key) => ({
    key,
    ...groupConfig[key],
  }));
});

// Basic settings sub-groups for organized display
const basicSiteItems = computed(() =>
  currentGroupItems.value.filter((i) =>
    ['site_name', 'site_logo', 'site_description'].includes(i.key),
  ),
);

const basicThemeItems = computed(() =>
  currentGroupItems.value.filter((i) =>
    ['default_theme', 'default_lang'].includes(i.key),
  ),
);

// Captcha settings sub-groups for organized display
const captchaGeneralItems = computed(() =>
  currentGroupItems.value.filter((i) =>
    ['captcha_enabled', 'captcha_type', 'captcha_expire', 'captcha_max_fail', 'captcha_login_trigger', 'captcha_min_duration'].includes(i.key),
  ),
);

const captchaNumericItems = computed(() =>
  currentGroupItems.value.filter((i) => i.key.startsWith('numeric_')),
);

const captchaSliderItems = computed(() =>
  currentGroupItems.value.filter((i) => i.key.startsWith('slider_')),
);

const captchaPuzzleItems = computed(() =>
  currentGroupItems.value.filter((i) => i.key.startsWith('puzzle_')),
);

const captchaRotationItems = computed(() =>
  currentGroupItems.value.filter((i) => i.key.startsWith('rotation_')),
);

const captchaPointItems = computed(() =>
  currentGroupItems.value.filter((i) => i.key.startsWith('point_')),
);

// 当前验证码类型（去除可能的引号）
const currentCaptchaType = computed(() => {
  const type = formValues.captcha?.captcha_type || 'slider';
  return String(type).replace(/"/g, '');
});

// ==================== Methods ====================
async function loadSettings() {
  loading.value = true;
  try {
    const res = await getAllSettings();
    allSettings.value = res;

    // Populate form values
    for (const [group, items] of Object.entries(res)) {
      if (!formValues[group]) {
        formValues[group] = {};
      }
      for (const item of items) {
        formValues[group][item.key] = item.value;
      }
    }
  } catch {
    message.error($t('system.settings.loadError'));
  } finally {
    loading.value = false;
  }
}

async function handleSaveGroup(group: string) {
  saving.value = true;
  try {
    const settings = formValues[group] || {};
    const result = await updateSettingsByGroup(group, settings);
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
    // 基础设置保存后立即应用到前端
    if (group === 'basic') {
      applyBasicSettings(settings);
    }
    // Reload to get updated values
    await loadSettings();
  } catch {
    message.error($t('system.settings.saveError'));
  } finally {
    saving.value = false;
  }
}

/**
 * 将基础设置立即应用到前端偏好
 * updatePreferences 要求嵌套对象格式
 */
function applyBasicSettings(settings: Record<string, any>) {
  const updates: Record<string, any> = {};

  // 站点名称
  if (settings.site_name !== undefined) {
    updates.app = updates.app || {};
    updates.app.name = settings.site_name;
  }
  // 站点 Logo
  if (settings.site_logo !== undefined) {
    updates.logo = updates.logo || {};
    updates.logo.source = settings.site_logo;
  }
  // 版权信息
  if (settings.copyright !== undefined) {
    updates.copyright = updates.copyright || {};
    updates.copyright.companyName = settings.copyright;
  }
  // 版权开关
  if (settings.copyright_enabled !== undefined) {
    updates.copyright = updates.copyright || {};
    updates.copyright.enable = settings.copyright_enabled;
  }
  // 公司网站
  if (settings.copyright_company_site !== undefined) {
    updates.copyright = updates.copyright || {};
    updates.copyright.companySiteLink = settings.copyright_company_site;
  }
  // 版权年份
  if (settings.copyright_date !== undefined) {
    updates.copyright = updates.copyright || {};
    updates.copyright.date = settings.copyright_date;
  }
  // ICP 备案
  if (settings.copyright_icp !== undefined) {
    updates.copyright = updates.copyright || {};
    updates.copyright.icp = settings.copyright_icp;
  }
  // ICP 链接
  if (settings.copyright_icp_link !== undefined) {
    updates.copyright = updates.copyright || {};
    updates.copyright.icpLink = settings.copyright_icp_link;
  }
  // 水印开关
  if (settings.watermark_enabled !== undefined) {
    updates.app = updates.app || {};
    updates.app.watermark = settings.watermark_enabled;
  }
  // 水印内容
  if (settings.watermark_content !== undefined) {
    updates.app = updates.app || {};
    updates.app.watermarkContent = settings.watermark_content;
  }
  // 水印透明度
  if (settings.watermark_opacity !== undefined) {
    updates.app = updates.app || {};
    updates.app.watermarkOpacity = settings.watermark_opacity;
  }
  // 页脚开关
  if (settings.footer_enabled !== undefined) {
    updates.footer = updates.footer || {};
    updates.footer.enable = settings.footer_enabled;
  }
  // 固定页脚
  if (settings.footer_fixed !== undefined) {
    updates.footer = updates.footer || {};
    updates.footer.fixed = settings.footer_fixed;
  }
  // 默认主题
  if (settings.default_theme !== undefined) {
    const mode = settings.default_theme;
    if (mode === 'auto' || mode === 'light' || mode === 'dark') {
      updates.theme = updates.theme || {};
      updates.theme.mode = mode;
    }
  }
  // 默认语言
  if (settings.default_lang !== undefined) {
    updates.app = updates.app || {};
    updates.app.locale = settings.default_lang;
  }

  if (Object.keys(updates).length > 0) {
    updatePreferences(updates);
  }
}

async function handleTestEmail() {
  if (!testEmailAddress.value) {
    message.warning($t('system.settings.testEmailPlaceholder'));
    return;
  }
  testEmailLoading.value = true;
  try {
    await testEmail(testEmailAddress.value);
    message.success($t('system.settings.testEmailSuccess'));
  } catch {
    message.error($t('system.settings.testEmailError'));
  } finally {
    testEmailLoading.value = false;
  }
}

async function handleTestSms() {
  if (!testPhoneNumber.value) {
    message.warning($t('system.settings.testSmsPlaceholder'));
    return;
  }
  testSmsLoading.value = true;
  try {
    await testSms(testPhoneNumber.value);
    message.success($t('system.settings.testSmsSuccess'));
  } catch {
    message.error($t('system.settings.testSmsError'));
  } finally {
    testSmsLoading.value = false;
  }
}

// ==================== Captcha Test Methods ====================
function resetCaptchaTestInput() {
  captchaTestResult.value = null;
  captchaTestHintText.value = '';
  captchaTestChars.value = [];
  captchaSliderPassing.value = false;
  captchaRotatePassing.value = false;
  captchaPointPassing.value = false;
  captchaTestStartTime.value = 0; // 重置开始时间，避免"操作过于迅速"错误
  captchaTestThumbY.value = 0;    // 重置缩略图 Y 位置
}

async function openCaptchaTest() {
  captchaTestType.value =
    formValues.captcha?.captcha_type?.replace(/"/g, '') || 'numeric';
  captchaTestVisible.value = true;
  await loadCaptchaForTest();
}

async function loadCaptchaForTest() {
  captchaTestLoading.value = true;
  resetCaptchaTestInput();

  try {
    const data = await testCaptcha(captchaTestType.value);
    if (data && data.captcha_id && data.image) {
      captchaTestId.value = data.captcha_id;
      captchaTestImage.value = data.image;
      captchaTestThumb.value = data.thumb || '';
      captchaTestThumbY.value = (data as any).thumb_y || 0;  // 缩略图 Y 位置
      captchaTestHintText.value = (data as any).hint_text || '';
      captchaTestChars.value = (data as any).chars || [];
      captchaTestLength.value = (data as any).length || 4;  // 数字验证码长度
      // startTime 在用户开始操作时设置，不在这里设置
    } else {
      message.error('获取验证码失败');
    }
  } catch (e: any) {
    message.error('获取验证码失败：' + (e?.message || '未知错误'));
  } finally {
    captchaTestLoading.value = false;
  }
}

async function handleCaptchaTestVerify(payload: {
  captchaCode: string;
  points?: Array<{ x: number; y: number }>;
}) {
  captchaTestVerifying.value = true;
  try {
    const result = await verifyCaptcha({
      captchaId: captchaTestId.value,
      startTime: captchaTestStartTime.value,
      ...payload,
    });
    captchaTestResult.value = result;
    if (result.valid) {
      message.success('验证通过！');
    } else {
      message.error('验证失败：' + result.message);
    }
  } catch (e: any) {
    message.error('验证请求失败：' + (e?.message || '未知错误'));
  } finally {
    captchaTestVerifying.value = false;
  }
}

// ========== 点选 - 使用 PointSelectionCaptcha 组件 ==========
function onPointClick(_point: { x: number; y: number; t: number; i: number }) {
  // 第一次点击时记录开始时间
  if (captchaTestStartTime.value === 0) {
    captchaTestStartTime.value = Date.now();
  }
}

function onPointConfirm(points: Array<{ x: number; y: number }>, clear: () => void) {
  const serverPoints = points.map((p) => ({ x: p.x, y: p.y }));
  captchaTestVerifying.value = true;
  verifyCaptcha({
    captchaId: captchaTestId.value,
    startTime: captchaTestStartTime.value,
    captchaCode: JSON.stringify(serverPoints),
    points: serverPoints,
  }).then((result) => {
    captchaTestResult.value = result;
    if (result.valid) {
      captchaPointPassing.value = true;
      message.success('验证通过！');
    } else {
      clear();
      message.error('验证失败：' + result.message);
    }
  }).catch((e: any) => {
    clear();
    message.error('验证请求失败：' + (e?.message || '未知错误'));
  }).finally(() => {
    captchaTestVerifying.value = false;
  });
}

function onPointRefresh() {
  loadCaptchaForTest();
}

function getSelectOptions(
  item: SystemSettingsApi.SettingItem,
): Array<{ label: string; value: string }> {
  if (item.options && Array.isArray(item.options)) {
    return item.options;
  }
  return [];
}

function getStorageSubGroup(item: SystemSettingsApi.SettingItem): string {
  const key = item.key;
  // 支持带前缀和不带前缀的 key
  if (key.startsWith('storage_minio_') || key.startsWith('minio_')) return 'minio';
  if (key.startsWith('storage_oss_') || key.startsWith('oss_')) return 'oss';
  if (key.startsWith('storage_cos_') || key.startsWith('cos_')) return 'cos';
  return 'general';
}

// ==================== Lifecycle ====================
onMounted(() => {
  loadSettings();
});
</script>

<template>
  <Page :title="$t('system.settings.title')" auto-content-height>
    <Spin :spinning="loading">
      <div class="flex h-full gap-4">
        <!-- 左侧分组菜单 -->
        <Card class="w-56 shrink-0">
          <Menu
            v-model:selected-keys="selectedKeys"
            mode="inline"
            :bordered="false"
            class="bg-transparent"
          >
            <MenuItem v-for="item in menuItems" :key="item.key">
              <span class="mr-2">{{ item.icon }}</span>
              {{ item.title }}
            </MenuItem>
          </Menu>
        </Card>

        <!-- 右侧配置内容 -->
        <Card class="flex-1 overflow-y-auto">
          <div v-if="groupConfig[activeGroup]" class="mb-4">
            <h2 class="text-lg font-medium">
              {{ groupConfig[activeGroup]?.icon }}
              {{ groupConfig[activeGroup]?.title }}
            </h2>
            <p class="text-foreground/50 mt-1 text-sm">
              {{ groupConfig[activeGroup]?.description }}
            </p>
            <Divider class="my-3" />
          </div>

          <!-- ==================== 基础设置 ==================== -->
          <div v-if="activeGroup === 'basic' && formValues.basic">
            <!-- 站点信息 -->
            <div class="settings-section">
              <h3 class="settings-section-title">🌐 {{ $t('system.settings.basic.siteInfo') }}</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in basicSiteItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <Input
                      v-if="item.type === 'string'"
                      v-model:value="formValues.basic[item.key]"
                      :placeholder="item.tip"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <Divider />

            <!-- 水印设置 -->
            <div class="settings-section">
              <h3 class="settings-section-title">💧 {{ $t('system.settings.basic.watermark') }}</h3>
              <Row :gutter="[16, 16]">
                <Col :span="12">
                  <div class="setting-item">
                    <label class="setting-label">{{ $t('system.settings.basic.watermarkEnabled') }}</label>
                    <Switch v-model:checked="formValues.basic.watermark_enabled" />
                  </div>
                </Col>
              </Row>
              <template v-if="formValues.basic.watermark_enabled">
                <Row :gutter="[16, 16]" class="mt-3">
                  <Col :span="12">
                    <div class="setting-item">
                      <label class="setting-label">{{ $t('system.settings.basic.watermarkContent') }}</label>
                      <Input
                        v-model:value="formValues.basic.watermark_content"
                        :placeholder="$t('system.settings.basic.watermarkContentPlaceholder')"
                      />
                    </div>
                  </Col>
                  <Col :span="12">
                    <div class="setting-item">
                      <label class="setting-label">{{ $t('system.settings.basic.watermarkOpacity') }}</label>
                      <div class="flex items-center gap-3">
                        <Slider
                          v-model:value="formValues.basic.watermark_opacity"
                          :min="0.05"
                          :max="1"
                          :step="0.01"
                          class="flex-1"
                        />
                        <span class="w-12 text-right text-sm">
                          {{ formValues.basic.watermark_opacity }}
                        </span>
                      </div>
                    </div>
                  </Col>
                </Row>
              </template>
            </div>

            <Divider />

            <!-- 页脚设置 -->
            <div class="settings-section">
              <h3 class="settings-section-title">📎 {{ $t('system.settings.basic.footer') }}</h3>
              <Row :gutter="[16, 16]">
                <Col :span="12">
                  <div class="setting-item">
                    <label class="setting-label">{{ $t('system.settings.basic.footerEnabled') }}</label>
                    <Switch v-model:checked="formValues.basic.footer_enabled" />
                  </div>
                </Col>
                <Col :span="12">
                  <div class="setting-item">
                    <label class="setting-label">{{ $t('system.settings.basic.footerFixed') }}</label>
                    <Switch
                      v-model:checked="formValues.basic.footer_fixed"
                      :disabled="!formValues.basic.footer_enabled"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <Divider />

            <!-- 版权设置 -->
            <div class="settings-section">
              <h3 class="settings-section-title">©️ {{ $t('system.settings.basic.copyright') }}</h3>
              <Row :gutter="[16, 16]">
                <Col :span="12">
                  <div class="setting-item">
                    <label class="setting-label">{{ $t('system.settings.basic.copyrightEnabled') }}</label>
                    <Switch
                      v-model:checked="formValues.basic.copyright_enabled"
                      :disabled="!formValues.basic.footer_enabled"
                    />
                  </div>
                </Col>
              </Row>
              <template v-if="formValues.basic.copyright_enabled && formValues.basic.footer_enabled">
                <Row :gutter="[16, 16]" class="mt-3">
                  <Col :span="12">
                    <div class="setting-item">
                      <label class="setting-label">{{ $t('system.settings.basic.copyrightName') }}</label>
                      <Input
                        v-model:value="formValues.basic.copyright"
                        :placeholder="$t('system.settings.basic.copyrightNamePlaceholder')"
                      />
                    </div>
                  </Col>
                  <Col :span="12">
                    <div class="setting-item">
                      <label class="setting-label">{{ $t('system.settings.basic.copyrightSite') }}</label>
                      <Input
                        v-model:value="formValues.basic.copyright_company_site"
                        placeholder="https://example.com"
                      />
                    </div>
                  </Col>
                </Row>
                <Row :gutter="[16, 16]" class="mt-3">
                  <Col :span="12">
                    <div class="setting-item">
                      <label class="setting-label">{{ $t('system.settings.basic.copyrightDate') }}</label>
                      <Input
                        v-model:value="formValues.basic.copyright_date"
                        placeholder="2026"
                      />
                    </div>
                  </Col>
                </Row>
                <Row :gutter="[16, 16]" class="mt-3">
                  <Col :span="12">
                    <div class="setting-item">
                      <label class="setting-label">{{ $t('system.settings.basic.copyrightIcp') }}</label>
                      <Input
                        v-model:value="formValues.basic.copyright_icp"
                        :placeholder="$t('system.settings.basic.copyrightIcpPlaceholder')"
                      />
                    </div>
                  </Col>
                  <Col :span="12">
                    <div class="setting-item">
                      <label class="setting-label">{{ $t('system.settings.basic.copyrightIcpLink') }}</label>
                      <Input
                        v-model:value="formValues.basic.copyright_icp_link"
                        placeholder="https://beian.miit.gov.cn/"
                      />
                    </div>
                  </Col>
                </Row>
              </template>
            </div>

            <Divider />

            <!-- 主题与语言 -->
            <div class="settings-section">
              <h3 class="settings-section-title">🎨 {{ $t('system.settings.basic.themeAndLang') }}</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in basicThemeItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <Select
                      v-if="item.type === 'select'"
                      v-model:value="formValues.basic[item.key]"
                      class="w-full"
                    >
                      <SelectOption
                        v-for="opt in getSelectOptions(item)"
                        :key="opt.value"
                        :value="opt.value"
                      >
                        {{ opt.label }}
                      </SelectOption>
                    </Select>
                  </div>
                </Col>
              </Row>
            </div>
          </div>

          <!-- ==================== 邮箱设置 ==================== -->
          <div v-else-if="activeGroup === 'email' && formValues.email">
            <Row :gutter="[16, 16]">
              <Col
                v-for="item in currentGroupItems"
                :key="item.key"
                :span="12"
              >
                <div class="setting-item">
                  <label class="setting-label">
                    {{ item.label }}
                    <Tooltip v-if="item.tip" :title="item.tip">
                      <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                    </Tooltip>
                    <Tag v-if="item.isSensitive" color="orange" class="ml-2">
                      {{ $t('system.settings.sensitive') }}
                    </Tag>
                  </label>
                  <Input
                    v-if="item.type === 'string'"
                    v-model:value="formValues.email[item.key]"
                    :placeholder="item.tip"
                    :type="item.isSensitive ? 'password' : 'text'"
                  />
                  <Switch
                    v-else-if="item.type === 'boolean'"
                    v-model:checked="formValues.email[item.key]"
                  />
                  <InputNumber
                    v-else-if="item.type === 'number'"
                    v-model:value="formValues.email[item.key]"
                    class="w-full"
                    :min="0"
                  />
                </div>
              </Col>
            </Row>
            <Divider />
            <div class="flex items-center gap-3">
              <Input
                v-model:value="testEmailAddress"
                :placeholder="$t('system.settings.testEmailPlaceholder')"
                class="w-64"
              />
              <Button :loading="testEmailLoading" @click="handleTestEmail">
                {{ $t('system.settings.testEmail') }}
              </Button>
            </div>
          </div>

          <!-- ==================== 短信设置 ==================== -->
          <div v-else-if="activeGroup === 'sms' && formValues.sms">
            <Row :gutter="[16, 16]">
              <Col
                v-for="item in currentGroupItems"
                :key="item.key"
                :span="12"
              >
                <div class="setting-item">
                  <label class="setting-label">
                    {{ item.label }}
                    <Tooltip v-if="item.tip" :title="item.tip">
                      <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                    </Tooltip>
                    <Tag v-if="item.isSensitive" color="orange" class="ml-2">
                      {{ $t('system.settings.sensitive') }}
                    </Tag>
                  </label>
                  <Input
                    v-if="item.type === 'string'"
                    v-model:value="formValues.sms[item.key]"
                    :placeholder="item.tip"
                    :type="item.isSensitive ? 'password' : 'text'"
                  />
                  <Switch
                    v-else-if="item.type === 'boolean'"
                    v-model:checked="formValues.sms[item.key]"
                  />
                  <Select
                    v-else-if="item.type === 'select'"
                    v-model:value="formValues.sms[item.key]"
                    class="w-full"
                  >
                    <SelectOption
                      v-for="opt in getSelectOptions(item)"
                      :key="opt.value"
                      :value="opt.value"
                    >
                      {{ opt.label }}
                    </SelectOption>
                  </Select>
                </div>
              </Col>
            </Row>
            <Divider />
            <div class="flex items-center gap-3">
              <Input
                v-model:value="testPhoneNumber"
                :placeholder="$t('system.settings.testSmsPlaceholder')"
                class="w-64"
              />
              <Button :loading="testSmsLoading" @click="handleTestSms">
                {{ $t('system.settings.testSms') }}
              </Button>
            </div>
          </div>

          <!-- ==================== 验证码设置 ==================== -->
          <div
            v-else-if="activeGroup === 'captcha' && formValues.captcha"
          >
            <!-- 通用配置 -->
            <div class="settings-section">
              <h3 class="settings-section-title">⚙️ 通用配置</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in captchaGeneralItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="formValues.captcha[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="formValues.captcha[item.key]"
                      class="w-full"
                      :min="0"
                    />
                    <Select
                      v-else-if="item.type === 'select'"
                      v-model:value="formValues.captcha[item.key]"
                      class="w-full"
                    >
                      <SelectOption
                        v-for="opt in getSelectOptions(item)"
                        :key="opt.value"
                        :value="opt.value"
                      >
                        {{ opt.label }}
                      </SelectOption>
                    </Select>
                  </div>
                </Col>
              </Row>
            </div>

            <Divider />

            <!-- 数字验证码配置 - 仅当类型为 numeric 时显示 -->
            <div v-if="currentCaptchaType === 'numeric'" class="settings-section">
              <h3 class="settings-section-title">🔢 数字验证码配置</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in captchaNumericItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <InputNumber
                      v-model:value="formValues.captcha[item.key]"
                      class="w-full"
                      :min="0"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <!-- 滑块验证码配置 - 仅当类型为 slider 时显示 -->
            <div v-if="currentCaptchaType === 'slider'" class="settings-section">
              <h3 class="settings-section-title">📊 滑块验证码配置</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in captchaSliderItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <InputNumber
                      v-model:value="formValues.captcha[item.key]"
                      class="w-full"
                      :min="0"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <!-- 拼图验证码配置 - 仅当类型为 puzzle 时显示 -->
            <div v-if="currentCaptchaType === 'puzzle'" class="settings-section">
              <h3 class="settings-section-title">🧩 拼图验证码配置</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in captchaPuzzleItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <InputNumber
                      v-if="item.type === 'number'"
                      v-model:value="formValues.captcha[item.key]"
                      class="w-full"
                      :min="0"
                    />
                    <Switch
                      v-else-if="item.type === 'boolean'"
                      v-model:checked="formValues.captcha[item.key]"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <!-- 旋转验证码配置 - 仅当类型为 rotation 时显示 -->
            <div v-if="currentCaptchaType === 'rotation'" class="settings-section">
              <h3 class="settings-section-title">🔄 旋转验证码配置</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in captchaRotationItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <InputNumber
                      v-model:value="formValues.captcha[item.key]"
                      class="w-full"
                      :min="0"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <!-- 点选验证码配置 - 仅当类型为 point 时显示 -->
            <div v-if="currentCaptchaType === 'point'" class="settings-section">
              <h3 class="settings-section-title">🎯 点选验证码配置</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in captchaPointItems"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <InputNumber
                      v-model:value="formValues.captcha[item.key]"
                      class="w-full"
                      :min="0"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <Divider />

            <!-- 测试验证码 -->
            <div class="flex items-center gap-3">
              <Button type="primary" @click="openCaptchaTest">
                {{ $t('system.settings.testCaptcha') }}
              </Button>
              <span class="text-sm text-foreground/40">
                {{ $t('system.settings.testCaptchaTip') }}
              </span>
            </div>

            <!-- 验证码测试弹窗 -->
            <Modal
              v-model:open="captchaTestVisible"
              title="验证码测试"
              :footer="null"
              width="480px"
              @cancel="resetCaptchaTestInput"
            >
              <Spin :spinning="captchaTestLoading">
                <div class="flex flex-col items-center py-2">
                  <!-- 类型切换 -->
                  <div class="mb-3 flex w-full items-center justify-center gap-2">
                    <Tag
                      v-for="t in ['numeric', 'slider', 'puzzle', 'rotation', 'point']"
                      :key="t"
                      :color="captchaTestType === t ? 'blue' : 'default'"
                      class="cursor-pointer"
                      @click="captchaTestType = t; loadCaptchaForTest()"
                    >
                      {{ CAPTCHA_TYPE_LABELS[t] || t }}
                    </Tag>
                  </div>

                  <!-- ========== 数字验证码 (NumericCaptcha 组件) ========== -->
                  <template v-if="captchaTestType === 'numeric'">
                    <NumericCaptcha
                      ref="captchaNumericRef"
                      :server-image="captchaTestImage"
                      :server-captcha-id="captchaTestId"
                      :width="200"
                      :height="60"
                      :char-length="captchaTestLength"
                      class="mb-3"
                      @success="(data: any) => {
                        handleCaptchaTestVerify({ captchaCode: data.code });
                      }"
                    />
                    <div v-if="captchaTestResult" class="mt-2 w-full">
                      <Alert
                        :type="captchaTestResult.valid ? 'success' : 'error'"
                        :message="captchaTestResult.valid ? '验证通过 ✓' : '验证失败 ✗'"
                        :description="captchaTestResult.message"
                        show-icon
                      />
                    </div>
                  </template>

                  <!-- ========== 滑块验证码 (BackendCaptcha 组件) ========== -->
                  <template v-if="captchaTestType === 'slider'">
                    <BackendCaptcha
                      ref="captchaSliderRef"
                      v-model="captchaSliderPassing"
                      captcha-type="slider"
                      :server-image="captchaTestImage"
                      :server-thumb="captchaTestThumb"
                      :server-thumb-y="captchaTestThumbY"
                      :server-captcha-id="captchaTestId"
                      @refresh="loadCaptchaForTest"
                      @success="(data: any) => handleCaptchaTestVerify({ captchaCode: data.captchaCode })"
                      class="mb-3"
                    />
                    <div v-if="captchaTestResult" class="mt-2 w-full">
                      <Alert
                        :type="captchaTestResult.valid ? 'success' : 'error'"
                        :message="captchaTestResult.valid ? '验证通过 ✓' : '验证失败 ✗'"
                        :description="captchaTestResult.message"
                        show-icon
                      />
                    </div>
                  </template>

                  <!-- ========== 拼图验证码 (BackendCaptcha 组件) ========== -->
                  <template v-if="captchaTestType === 'puzzle'">
                    <BackendCaptcha
                      ref="captchaSliderRef"
                      v-model="captchaSliderPassing"
                      captcha-type="puzzle"
                      :server-image="captchaTestImage"
                      :server-thumb="captchaTestThumb"
                      :server-thumb-y="captchaTestThumbY"
                      :server-captcha-id="captchaTestId"
                      @refresh="loadCaptchaForTest"
                      @success="(data: any) => handleCaptchaTestVerify({ captchaCode: data.captchaCode })"
                      class="mb-3"
                    />
                    <div v-if="captchaTestResult" class="mt-2 w-full">
                      <Alert
                        :type="captchaTestResult.valid ? 'success' : 'error'"
                        :message="captchaTestResult.valid ? '验证通过 ✓' : '验证失败 ✗'"
                        :description="captchaTestResult.message"
                        show-icon
                      />
                    </div>
                  </template>

                  <!-- ========== 旋转验证码 (BackendRotateCaptcha 组件) ========== -->
                  <template v-if="captchaTestType === 'rotation'">
                    <BackendRotateCaptcha
                      ref="captchaRotateRef"
                      v-model="captchaRotatePassing"
                      :server-image="captchaTestImage"
                      :server-thumb="captchaTestThumb"
                      :server-captcha-id="captchaTestId"
                      :image-size="220"
                      @refresh="loadCaptchaForTest"
                      @success="(data: any) => handleCaptchaTestVerify({ captchaCode: data.captchaCode })"
                      class="mb-3"
                    />
                    <div v-if="captchaTestResult" class="mt-2 w-full">
                      <Alert
                        :type="captchaTestResult.valid ? 'success' : 'error'"
                        :message="captchaTestResult.valid ? '验证通过 ✓' : '验证失败 ✗'"
                        :description="captchaTestResult.message"
                        show-icon
                      />
                    </div>
                  </template>

                  <!-- ========== 点选验证码 (PointSelectionCaptcha 组件) ========== -->
                  <template v-if="captchaTestType === 'point'">
                    <PointSelectionCaptcha
                      :captcha-image="captchaTestImage"
                      :hint-text="captchaTestHintText || '请依次点击图片中的文字'"
                      :show-confirm="true"
                      :width="CAPTCHA_IMG_W"
                      :height="CAPTCHA_IMG_H_POINT"
                      class="mb-3"
                      @click="onPointClick"
                      @confirm="onPointConfirm"
                      @refresh="onPointRefresh"
                    />
                    <div v-if="captchaTestResult" class="mt-2 w-full">
                      <Alert
                        :type="captchaTestResult.valid ? 'success' : 'error'"
                        :message="captchaTestResult.valid ? '验证通过 ✓' : '验证失败 ✗'"
                        :description="captchaTestResult.message"
                        show-icon
                      />
                    </div>
                  </template>
                </div>
              </Spin>
            </Modal>
          </div>

          <!-- ==================== 风险评分设置 ==================== -->
          <div
            v-else-if="activeGroup === 'risk_score' && formValues.risk_score"
          >
            <!-- 通用配置 -->
            <div class="settings-section">
              <h3 class="settings-section-title">⚙️ 通用配置</h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in currentGroupItems.filter((i) => i.key.startsWith('risk_') && !i.key.startsWith('rule_'))"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="formValues.risk_score[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="formValues.risk_score[item.key]"
                      class="w-full"
                      :min="0"
                    />
                    <Input
                      v-else-if="item.type === 'string'"
                      v-model:value="formValues.risk_score[item.key]"
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
                  v-for="item in currentGroupItems.filter((i) => i.key.startsWith('rule_frequency'))"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="formValues.risk_score[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="formValues.risk_score[item.key]"
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
                  v-for="item in currentGroupItems.filter((i) => i.key.startsWith('rule_no_referer'))"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="formValues.risk_score[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="formValues.risk_score[item.key]"
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
                <Col
                  v-for="item in currentGroupItems.filter((i) => i.key.startsWith('rule_no_lang'))"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="formValues.risk_score[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="formValues.risk_score[item.key]"
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
                <Col
                  v-for="item in currentGroupItems.filter((i) => i.key.startsWith('rule_ua'))"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="formValues.risk_score[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="formValues.risk_score[item.key]"
                      class="w-full"
                      :min="0"
                    />
                    <Input
                      v-else-if="item.type === 'string'"
                      v-model:value="formValues.risk_score[item.key]"
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
                  v-for="item in currentGroupItems.filter((i) => i.key.startsWith('rule_interval'))"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <Switch
                      v-if="item.type === 'boolean'"
                      v-model:checked="formValues.risk_score[item.key]"
                    />
                    <InputNumber
                      v-else-if="item.type === 'number'"
                      v-model:value="formValues.risk_score[item.key]"
                      class="w-full"
                      :min="0"
                    />
                  </div>
                </Col>
              </Row>
            </div>
          </div>

          <!-- ==================== 存储设置 ==================== -->
          <div
            v-else-if="activeGroup === 'storage' && formValues.storage"
          >
            <!-- 本地存储（始终启用） -->
            <div class="settings-section">
              <h3 class="settings-section-title">
                💻 本地存储
                <Tag color="green" class="ml-2">始终启用</Tag>
              </h3>
              <Row :gutter="[16, 16]">
                <Col
                  v-for="item in currentGroupItems.filter((i) => i.key === 'storage_local_path' || i.key === 'storage_local_url_prefix')"
                  :key="item.key"
                  :span="12"
                >
                  <div class="setting-item">
                    <label class="setting-label">
                      {{ item.label }}
                      <Tooltip v-if="item.tip" :title="item.tip">
                        <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                      </Tooltip>
                    </label>
                    <Input
                      v-model:value="formValues.storage[item.key]"
                      :placeholder="item.tip"
                    />
                  </div>
                </Col>
              </Row>
            </div>

            <Divider />

            <!-- MinIO 存储 -->
            <div class="settings-section">
              <h3 class="settings-section-title">
                📦 MinIO 存储
              </h3>
              <Row :gutter="[16, 16]">
                <Col :span="12">
                  <div class="setting-item">
                    <label class="setting-label">启用 MinIO</label>
                    <Switch v-model:checked="formValues.storage.storage_minio_enabled" />
                  </div>
                </Col>
              </Row>
              <template v-if="formValues.storage.storage_minio_enabled">
                <Row :gutter="[16, 16]" class="mt-3">
                  <Col
                    v-for="item in currentGroupItems.filter((i) => getStorageSubGroup(i) === 'minio' && i.key !== 'storage_minio_enabled')"
                    :key="item.key"
                    :span="12"
                  >
                    <div class="setting-item">
                      <label class="setting-label">
                        {{ item.label }}
                        <Tag v-if="item.isSensitive" color="orange" class="ml-2">{{ $t('system.settings.sensitive') }}</Tag>
                      </label>
                      <Input
                        v-if="item.type === 'string'"
                        v-model:value="formValues.storage[item.key]"
                        :placeholder="item.tip"
                        :type="item.isSensitive ? 'password' : 'text'"
                      />
                      <Switch
                        v-else-if="item.type === 'boolean'"
                        v-model:checked="formValues.storage[item.key]"
                      />
                    </div>
                  </Col>
                </Row>
              </template>
            </div>

            <Divider />

            <!-- OSS 存储 -->
            <div class="settings-section">
              <h3 class="settings-section-title">
                ☁️ 阿里云 OSS
              </h3>
              <Row :gutter="[16, 16]">
                <Col :span="12">
                  <div class="setting-item">
                    <label class="setting-label">启用 OSS</label>
                    <Switch v-model:checked="formValues.storage.storage_oss_enabled" />
                  </div>
                </Col>
              </Row>
              <template v-if="formValues.storage.storage_oss_enabled">
                <Row :gutter="[16, 16]" class="mt-3">
                  <Col
                    v-for="item in currentGroupItems.filter((i) => getStorageSubGroup(i) === 'oss' && i.key !== 'storage_oss_enabled')"
                    :key="item.key"
                    :span="12"
                  >
                    <div class="setting-item">
                      <label class="setting-label">
                        {{ item.label }}
                        <Tag v-if="item.isSensitive" color="orange" class="ml-2">{{ $t('system.settings.sensitive') }}</Tag>
                      </label>
                      <Input
                        v-model:value="formValues.storage[item.key]"
                        :placeholder="item.tip"
                        :type="item.isSensitive ? 'password' : 'text'"
                      />
                    </div>
                  </Col>
                </Row>
              </template>
            </div>

            <Divider />

            <!-- COS 存储 -->
            <div class="settings-section">
              <h3 class="settings-section-title">
                🌊 腾讯云 COS
              </h3>
              <Row :gutter="[16, 16]">
                <Col :span="12">
                  <div class="setting-item">
                    <label class="setting-label">启用 COS</label>
                    <Switch v-model:checked="formValues.storage.storage_cos_enabled" />
                  </div>
                </Col>
              </Row>
              <template v-if="formValues.storage.storage_cos_enabled">
                <Row :gutter="[16, 16]" class="mt-3">
                  <Col
                    v-for="item in currentGroupItems.filter((i) => getStorageSubGroup(i) === 'cos' && i.key !== 'storage_cos_enabled')"
                    :key="item.key"
                    :span="12"
                  >
                    <div class="setting-item">
                      <label class="setting-label">
                        {{ item.label }}
                        <Tag v-if="item.isSensitive" color="orange" class="ml-2">{{ $t('system.settings.sensitive') }}</Tag>
                      </label>
                      <Input
                        v-model:value="formValues.storage[item.key]"
                        :placeholder="item.tip"
                        :type="item.isSensitive ? 'password' : 'text'"
                      />
                    </div>
                  </Col>
                </Row>
              </template>
            </div>

            <Alert
              message="存储优先级"
              description="本地存储始终启用。如果同时启用了多个外部存储，优先级为：COS > OSS > MinIO。启用外部存储后，新上传的文件将存储到外部存储，但历史文件仍可通过本地存储访问。"
              type="info"
              show-icon
              class="mt-4"
            />

            <Divider />

            <!-- 标签路由管理入口 -->
            <div class="settings-section">
              <h3 class="settings-section-title">
                🏷️ {{ $t('system.tag.routeManagementEntry') }}
              </h3>
              <p class="text-gray-500 mb-3">
                {{ $t('system.tag.routeManagementDesc') }}
              </p>
              <Button type="primary" @click="$router.push('/system/tag')">
                {{ $t('system.tag.enterRouteManagement') }}
              </Button>
            </div>
          </div>

          <!-- ==================== 微信设置 ==================== -->
          <div v-else-if="activeGroup === 'wechat' && formValues.wechat">
            <Row :gutter="[16, 16]">
              <Col
                v-for="item in currentGroupItems"
                :key="item.key"
                :span="12"
              >
                <div class="setting-item">
                  <label class="setting-label">
                    {{ item.label }}
                    <Tooltip v-if="item.tip" :title="item.tip">
                      <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                    </Tooltip>
                    <Tag v-if="item.isSensitive" color="orange" class="ml-2">
                      {{ $t('system.settings.sensitive') }}
                    </Tag>
                  </label>
                  <Input
                    v-if="item.type === 'string'"
                    v-model:value="formValues.wechat[item.key]"
                    :placeholder="item.tip"
                    :type="item.isSensitive ? 'password' : 'text'"
                  />
                  <Switch
                    v-else-if="item.type === 'boolean'"
                    v-model:checked="formValues.wechat[item.key]"
                  />
                </div>
              </Col>
            </Row>
          </div>

          <!-- ==================== 认证设置 ==================== -->
          <div
            v-else-if="activeGroup === 'auth' && formValues.auth"
          >
            <Alert
              message="配置登录方式的开启与关闭"
              type="info"
              show-icon
              class="mb-4"
            />
            <Row :gutter="[16, 16]">
              <Col
                v-for="item in currentGroupItems"
                :key="item.key"
                :span="12"
              >
                <div class="setting-item">
                  <label class="setting-label">
                    {{ item.label }}
                    <Tooltip v-if="item.tip" :title="item.tip">
                      <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                    </Tooltip>
                  </label>
                  <Switch
                    v-if="item.type === 'boolean'"
                    v-model:checked="formValues.auth[item.key]"
                  />
                  <Select
                    v-else-if="item.type === 'array'"
                    v-model:value="formValues.auth[item.key]"
                    mode="tags"
                    class="w-full"
                    :placeholder="item.tip || '选择启用的提供商'"
                  >
                    <SelectOption value="github">GitHub</SelectOption>
                    <SelectOption value="wechat">微信</SelectOption>
                    <SelectOption value="google">Google</SelectOption>
                  </Select>
                </div>
              </Col>
            </Row>
          </div>

          <!-- ==================== 安全设置 ==================== -->
          <div
            v-else-if="activeGroup === 'security' && formValues.security"
          >
            <Alert
              :message="$t('system.settings.securityTip')"
              type="info"
              show-icon
              class="mb-4"
            />
            <Row :gutter="[16, 16]">
              <Col
                v-for="item in currentGroupItems"
                :key="item.key"
                :span="12"
              >
                <div class="setting-item">
                  <label class="setting-label">
                    {{ item.label }}
                    <Tooltip v-if="item.tip" :title="item.tip">
                      <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                    </Tooltip>
                  </label>
                  <Input
                    v-if="item.type === 'string'"
                    v-model:value="formValues.security[item.key]"
                    :placeholder="item.tip"
                  />
                  <Switch
                    v-else-if="item.type === 'boolean'"
                    v-model:checked="formValues.security[item.key]"
                  />
                  <InputNumber
                    v-else-if="item.type === 'number'"
                    v-model:value="formValues.security[item.key]"
                    class="w-full"
                    :min="0"
                  />
                </div>
              </Col>
            </Row>
          </div>

          <!-- 保存按钮 -->
          <Divider />
          <div class="flex justify-end">
            <Button
              type="primary"
              :loading="saving"
              @click="handleSaveGroup(activeGroup)"
            >
              {{ $t('system.settings.saveGroup') }}
            </Button>
          </div>
        </Card>
      </div>
    </Spin>
  </Page>
</template>

<style scoped>
.settings-section {
  margin-bottom: 8px;
}
.settings-section-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
  color: rgba(0, 0, 0, 0.85);
}
.setting-item {
  margin-bottom: 8px;
}
.setting-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
}
</style>
