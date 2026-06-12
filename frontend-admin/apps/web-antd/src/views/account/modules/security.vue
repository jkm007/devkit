<script lang="ts" setup>
import type { AccountApi } from '#/api';

import { computed, onMounted, ref } from 'vue';

import {
  Button,
  Card,
  Empty,
  message,
  Popconfirm,
  Select,
  SelectOption,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import {
  getLoginDevices,
  kickAllOtherDevices,
  kickDevice,
} from '#/api';
import { $t } from '#/locales';
import { getDeviceId } from '#/utils/device-id';

// ==================== 登录设备 ====================
const devices = ref<AccountApi.LoginDevice[]>([]);
const loadingDevices = ref(false);
const currentDeviceId = getDeviceId();
const deviceTypeFilter = ref<AccountApi.DeviceType | undefined>();

function isCurrentDevice(device: AccountApi.LoginDevice): boolean {
  return device.deviceId === currentDeviceId;
}

// 排序：当前设备在最前面
const sortedDevices = computed(() => {
  return [...devices.value].toSorted((a, b) => {
    const aCurrent = isCurrentDevice(a) ? 1 : 0;
    const bCurrent = isCurrentDevice(b) ? 1 : 0;
    return bCurrent - aCurrent;
  });
});

function getBrowserIcon(browser: string): string {
  const b = browser.toLowerCase();
  if (b.includes('chrome')) return '🌐';
  if (b.includes('firefox')) return '🦊';
  if (b.includes('safari')) return '🧭';
  if (b.includes('edge')) return '📐';
  if (b.includes('opera')) return '🔴';
  return '🌍';
}

function getDeviceTypeColor(type: string): string {
  const colorMap: Record<string, string> = {
    app: 'purple',
    h5: 'cyan',
    miniapp: 'green',
    web: 'blue',
  };
  return colorMap[type] || 'default';
}

function getDeviceTypeText(type: string): string {
  const textMap: Record<string, string> = {
    app: 'App',
    h5: 'H5',
    miniapp: '小程序',
    web: 'Web',
  };
  return textMap[type] || type || '未知';
}

function getOSIcon(os: string): string {
  const o = os.toLowerCase();
  if (o.includes('windows')) return '🪟';
  if (o.includes('mac')) return '🍎';
  if (o.includes('ios') || o.includes('iphone') || o.includes('ipad'))
    return '📱';
  if (o.includes('android')) return '🤖';
  if (o.includes('linux')) return '🐧';
  return '💻';
}

function formatTime(timeStr: string | null): string {
  if (!timeStr) return $t('account.security.unknown');
  try {
    const date = new Date(timeStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return $t('account.security.justNow');
    if (minutes < 60) return $t('account.security.minutesAgo', [minutes]);
    if (hours < 24) return $t('account.security.hoursAgo', [hours]);
    if (days < 30) return $t('account.security.daysAgo', [days]);
    return date.toLocaleDateString();
  } catch {
    return timeStr;
  }
}

async function loadDevices() {
  loadingDevices.value = true;
  try {
    devices.value = await getLoginDevices({
      deviceType: deviceTypeFilter.value,
    });
  } catch {
    devices.value = [];
  } finally {
    loadingDevices.value = false;
  }
}

async function handleKickDevice(id: number) {
  try {
    await kickDevice(id);
    message.success($t('account.security.kickDeviceSuccess'));
    await loadDevices();
  } catch {
    message.error('操作失败');
  }
}

async function handleKickAllOthers() {
  try {
    const result = await kickAllOtherDevices();
    message.success(
      $t('account.security.kickAllOthersSuccess', [result.kickedCount]),
    );
    await loadDevices();
  } catch {
    message.error('操作失败');
  }
}

onMounted(() => {
  loadDevices();
});
</script>

<template>
  <div>
    <!-- 标题 & 操作栏 -->
    <div class="mb-4 flex items-center justify-between">
      <div>
        <h3 class="text-base font-medium">
          {{ $t('account.security.loginDevices') }}
        </h3>
        <p class="text-foreground/50 mt-1 text-xs">
          {{ $t('account.security.deviceCount', [devices.length]) }}
        </p>
      </div>
      <div class="flex items-center gap-2">
        <Select
          v-model:value="deviceTypeFilter"
          allow-clear
          placeholder="全部设备"
          size="small"
          style="width: 120px"
          @change="loadDevices"
        >
          <SelectOption value="web">Web</SelectOption>
          <SelectOption value="h5">H5</SelectOption>
          <SelectOption value="app">App</SelectOption>
          <SelectOption value="miniapp">小程序</SelectOption>
        </Select>
        <Popconfirm
          v-if="devices.length > 1"
          :title="$t('account.security.kickAllOthersConfirm')"
          @confirm="handleKickAllOthers"
        >
          <Button danger size="small">
            {{ $t('account.security.kickAllOthers') }}
          </Button>
        </Popconfirm>
      </div>
    </div>

    <!-- 设备列表 -->
    <div v-loading="loadingDevices">
      <!-- 空状态 -->
      <Empty
        v-if="!loadingDevices && devices.length === 0"
        :description="$t('account.security.noDevices')"
        class="py-12"
      />

      <!-- 设备卡片网格 -->
      <div
        v-else
        class="grid grid-cols-1 gap-4 md:grid-cols-2"
      >
        <Card
          v-for="device in sortedDevices"
          :key="device.id"
          size="small"
          :bordered="true"
          :class="
            isCurrentDevice(device)
              ? 'border-primary shadow-sm'
              : 'border-border/60 hover:border-primary/40'
          "
          class="transition-colors"
        >
          <div class="flex items-start justify-between">
            <div class="flex items-start gap-3">
              <!-- 设备图标 -->
              <div
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-lg"
                :class="
                  isCurrentDevice(device)
                    ? 'bg-primary/10'
                    : 'bg-background-soft'
                "
              >
                {{ getBrowserIcon(device.browser) }}
              </div>
              <!-- 设备信息 -->
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-1.5">
                  <span class="text-sm font-medium">{{
                    device.deviceName
                  }}</span>
                  <Tag :color="getDeviceTypeColor(device.deviceType)" class="ml-0">
                    {{ getDeviceTypeText(device.deviceType) }}
                  </Tag>
                  <Tag
                    v-if="isCurrentDevice(device)"
                    color="processing"
                    class="ml-0"
                  >
                    {{ $t('account.security.currentDevice') }}
                  </Tag>
                </div>
                <div
                  class="text-foreground/60 mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs"
                >
                  <span class="flex items-center gap-1">
                    {{ getBrowserIcon(device.browser) }} {{ device.browser }}
                  </span>
                  <span class="flex items-center gap-1">
                    {{ getOSIcon(device.os) }} {{ device.os }}
                  </span>
                  <span>🌐 {{ device.ip }}</span>
                </div>
                <div
                  v-if="
                    device.deviceModel ||
                    device.systemVersion ||
                    device.appVersion ||
                    device.channel
                  "
                  class="text-foreground/50 mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs"
                >
                  <span v-if="device.deviceModel">📱 {{ device.deviceModel }}</span>
                  <span v-if="device.systemVersion">⚙️ {{ device.systemVersion }}</span>
                  <span v-if="device.appVersion">🏷️ v{{ device.appVersion }}</span>
                  <span v-if="device.channel">📦 {{ device.channel }}</span>
                </div>
                <div
                  class="text-foreground/40 mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs"
                >
                  <Tooltip
                    v-if="device.lastActiveAt"
                    :title="device.lastActiveAt"
                  >
                    <span>{{ formatTime(device.lastActiveAt) }}</span>
                  </Tooltip>
                  <span v-else>{{ $t('account.security.unknown') }}</span>
                  <span v-if="device.location">· {{ device.location }}</span>
                </div>
              </div>
            </div>
            <!-- 操作按钮 -->
            <Popconfirm
              v-if="!isCurrentDevice(device)"
              :title="$t('account.security.kickDeviceConfirm')"
              @confirm="handleKickDevice(device.id)"
            >
              <Button danger size="small">
                {{ $t('account.security.kickDevice') }}
              </Button>
            </Popconfirm>
          </div>
        </Card>
      </div>
    </div>
  </div>
</template>
