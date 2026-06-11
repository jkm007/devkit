<script lang="ts" setup>
import type { SystemSettingsApi } from '#/api/system/settings';

import { computed, onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Alert,
  Button,
  Col,
  Divider,
  Input,
  message,
  Modal,
  Row,
  Spin,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import { getAllSettings, updateSettingsByGroup } from '#/api/system/settings';
import { $t } from '#/locales';

// ==================== State ====================
const loading = ref(false);
const saving = ref(false);
const allSettings = ref<SystemSettingsApi.SettingsGroup>({});

const formValues = reactive<Record<string, any>>({});

// ==================== Computed ====================
const storageItems = computed(() => {
  return (allSettings.value.storage || []).filter(
    (i) =>
      i.key === 'storage_local_path' || i.key === 'storage_local_url_prefix',
  );
});

// ==================== Methods ====================
async function loadSettings() {
  loading.value = true;
  try {
    const res = await getAllSettings();
    allSettings.value = res;

    // Populate form values for storage group only
    const items = res.storage || [];
    for (const item of items) {
      formValues[item.key] = item.value;
    }
  } catch {
    message.error($t('system.settings.loadError'));
  } finally {
    loading.value = false;
  }
}

async function handleSave() {
  saving.value = true;
  try {
    const settings: Record<string, any> = {};
    for (const item of storageItems.value) {
      settings[item.key] = formValues[item.key];
    }
    const result = await updateSettingsByGroup('storage', settings);
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
    // Reload to get updated values
    await loadSettings();
  } catch {
    message.error($t('system.settings.saveError'));
  } finally {
    saving.value = false;
  }
}

// ==================== Lifecycle ====================
onMounted(() => {
  loadSettings();
});
</script>

<template>
  <Page :title="$t('system.settings.groups.storage')" auto-content-height>
    <Spin :spinning="loading">
      <div class="storage-settings-content">
        <Alert
          message="存储管理已独立"
          description="存储连接配置和存储桶管理已拆分为独立页面，支持更灵活的多配置管理。"
          type="info"
          show-icon
          class="mb-4"
        />

        <!-- 存储管理入口 -->
        <div class="settings-section">
          <h3 class="settings-section-title">⚙️ 存储管理入口</h3>
          <div class="flex flex-wrap gap-3">
            <Button
              type="primary"
              @click="$router.push('/storage/storage-manage')"
            >
              📡 存储管理
            </Button>
            <Button @click="$router.push('/storage/tag-routing')">
              🏷️ 标签路由
            </Button>
            <Button @click="$router.push('/storage/file-type-rule')">
              📋 文件类型规则
            </Button>
          </div>
          <div class="mt-4 text-sm text-gray-500">
            <ul class="list-disc pl-5 space-y-1">
              <li>
                <b>存储配置</b>：管理存储连接信息（MinIO/OSS/COS 的
                endpoint、密钥等），支持多个同类型配置
              </li>
              <li>
                <b>存储桶</b>：管理具体的存储桶，自动使用存储配置中的连接信息
              </li>
              <li><b>标签路由</b>：根据文件类型自动选择存储到哪个桶</li>
            </ul>
          </div>
        </div>

        <Divider />

        <!-- 本地存储基本配置 -->
        <div class="settings-section">
          <h3 class="settings-section-title">
            💻 本地存储基本设置
            <Tag color="green" class="ml-2">始终启用</Tag>
          </h3>
          <Row :gutter="[16, 16]">
            <Col v-for="item in storageItems" :key="item.key" :span="12">
              <div class="setting-item">
                <label class="setting-label">
                  {{ item.label }}
                  <Tooltip v-if="item.tip" :title="item.tip">
                    <span class="text-foreground/40 ml-1 cursor-help">ⓘ</span>
                  </Tooltip>
                </label>
                <Input
                  v-model:value="formValues[item.key]"
                  :placeholder="item.tip"
                />
              </div>
            </Col>
          </Row>
        </div>

        <Alert
          message="存储优先级"
          description="本地存储始终启用。默认存储由「存储配置」页面中的默认设置决定。通过标签路由可以实现按文件类型自动选择存储。"
          type="info"
          show-icon
          class="mt-4"
        />

        <!-- 保存按钮 -->
        <Divider />
        <div class="flex justify-end">
          <Button type="primary" :loading="saving" @click="handleSave">
            {{ $t('system.settings.saveGroup') }}
          </Button>
        </div>
      </div>
    </Spin>
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
