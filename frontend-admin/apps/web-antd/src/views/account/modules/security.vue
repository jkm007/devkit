<script lang="ts" setup>
import type { AccountApi } from '#/api';

import { computed, onMounted, ref } from 'vue';

import {
  Button,
  InputPassword,
  message,
  Popconfirm,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import {
  changePassword,
  getLoginDevices,
  kickAllOtherDevices,
  kickDevice,
} from '#/api';
import { $t } from '#/locales';
import { getDeviceId } from '#/utils/device-id';
import { showCaptchaVerify } from '#/utils/captcha-verify';

// ==================== 修改密码 ====================
const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
});
const changingPassword = ref(false);

async function handleChangePassword() {
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    message.warning($t('account.security.passwordMismatch'));
    return;
  }

  // 先弹验证码
  let captchaResult: { captchaId: string; captchaCode: string };
  try {
    captchaResult = await showCaptchaVerify();
  } catch {
    return; // 用户取消
  }

  changingPassword.value = true;
  try {
    await changePassword({
      ...passwordForm.value,
      captchaId: captchaResult.captchaId,
      captchaCode: captchaResult.captchaCode,
    });
    message.success($t('account.security.changePasswordSuccess'));
    passwordForm.value = { oldPassword: '', newPassword: '', confirmPassword: '' };
  } catch {
    message.error('密码修改失败');
  } finally {
    changingPassword.value = false;
  }
}

// ==================== 登录设备 ====================
const devices = ref<AccountApi.LoginDevice[]>([]);
const loadingDevices = ref(false);
const currentDeviceId = getDeviceId();

// 前端判断是否是当前设备（用 localStorage 中的设备ID比对）
function isCurrentDevice(device: AccountApi.LoginDevice): boolean {
  return device.deviceId === currentDeviceId;
}

// 排序：当前设备在最前面
const sortedDevices = computed(() => {
  return [...devices.value].sort((a, b) => {
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

function getOSIcon(os: string): string {
  const o = os.toLowerCase();
  if (o.includes('windows')) return '🪟';
  if (o.includes('mac')) return '🍎';
  if (o.includes('ios') || o.includes('iphone') || o.includes('ipad')) return '📱';
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
    devices.value = await getLoginDevices();
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
    message.success($t('account.security.kickAllOthersSuccess', [result.kickedCount]));
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
  <div class="space-y-8">
    <!-- 修改密码 -->
    <div>
      <h3 class="mb-4 text-base font-medium">{{ $t('account.security.changePassword') }}</h3>
      <div class="max-w-md">
        <div class="mb-3">
          <label class="mb-1 block text-sm">{{ $t('account.security.oldPassword') }}</label>
          <InputPassword
            v-model:value="passwordForm.oldPassword"
            :placeholder="$t('account.security.oldPassword')"
          />
        </div>
        <div class="mb-3">
          <label class="mb-1 block text-sm">{{ $t('account.security.newPassword') }}</label>
          <InputPassword
            v-model:value="passwordForm.newPassword"
            :placeholder="$t('account.security.newPassword')"
          />
        </div>
        <div class="mb-3">
          <label class="mb-1 block text-sm">{{ $t('account.security.confirmPassword') }}</label>
          <InputPassword
            v-model:value="passwordForm.confirmPassword"
            :placeholder="$t('account.security.confirmPassword')"
          />
        </div>
        <Button type="primary" :loading="changingPassword" @click="handleChangePassword">
          {{ $t('account.security.changePasswordBtn') }}
        </Button>
      </div>
    </div>

    <!-- 登录设备管理 -->
    <div>
      <div class="mb-4 flex items-center justify-between">
        <div>
          <h3 class="text-base font-medium">{{ $t('account.security.loginDevices') }}</h3>
          <p class="text-foreground/50 mt-1 text-xs">
            {{ $t('account.security.deviceCount', [devices.length]) }}
          </p>
        </div>
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

      <div v-loading="loadingDevices" class="space-y-3">
        <div
          v-for="device in sortedDevices"
          :key="device.id"
          class="border-border/60 hover:border-primary/40 rounded-lg border p-4 transition-colors"
          :class="isCurrentDevice(device) ? 'border-primary/30 bg-primary/2' : ''"
        >
          <div class="flex items-start justify-between">
            <div class="flex items-start gap-3">
              <!-- 设备图标 -->
              <div
                class="flex h-10 w-10 items-center justify-center rounded-lg text-lg"
                :class="isCurrentDevice(device) ? 'bg-primary/10' : 'bg-background-soft'"
              >
                {{ getBrowserIcon(device.browser) }}
              </div>
              <!-- 设备信息 -->
              <div>
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium">{{ device.deviceName }}</span>
                  <Tag v-if="isCurrentDevice(device)" color="processing" class="ml-0">
                    {{ $t('account.security.currentDevice') }}
                  </Tag>
                </div>
                <div class="text-foreground/60 mt-1 flex items-center gap-3 text-xs">
                  <span class="flex items-center gap-1">
                    {{ getBrowserIcon(device.browser) }} {{ device.browser }}
                  </span>
                  <span class="flex items-center gap-1">
                    {{ getOSIcon(device.os) }} {{ device.os }}
                  </span>
                  <span>🌐 {{ device.ip }}</span>
                </div>
                <div class="text-foreground/40 mt-1 flex items-center gap-3 text-xs">
                  <Tooltip v-if="device.lastActiveAt" :title="device.lastActiveAt">
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
        </div>

        <!-- 空状态 -->
        <div
          v-if="!loadingDevices && devices.length === 0"
          class="text-foreground/40 py-8 text-center text-sm"
        >
          {{ $t('account.security.noDevices') }}
        </div>
      </div>
    </div>
  </div>
</template>
