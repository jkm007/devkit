<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import { Page, VbenAvatar } from '@vben/common-ui';

import { Card, Menu, MenuItem } from 'ant-design-vue';

import { getUserInfo } from '#/api';
import type { AccountApi } from '#/api';
import { $t } from '#/locales';

import OauthTab from './modules/oauth.vue';
import ProfileTab from './modules/profile.vue';
import SecurityTab from './modules/security.vue';
import PrivacyTab from './modules/privacy.vue';

const activeKey = ref<string[]>(['profile']);
const userInfo = ref<AccountApi.UserInfo>({} as AccountApi.UserInfo);

const menuItems = computed(() => [
  { key: 'profile', label: $t('account.profile.title'), icon: 'lucide:user' },
  {
    key: 'security',
    label: $t('account.security.changePassword'),
    icon: 'lucide:shield',
  },
  {
    key: 'oauth',
    label: $t('account.security.oauthBindings'),
    icon: 'lucide:link',
  },
  { key: 'privacy', label: $t('account.privacy.title'), icon: 'lucide:lock' },
]);

async function loadUserInfo() {
  try {
    userInfo.value = await getUserInfo();
  } catch {
    // 静默处理
  }
}

onMounted(() => {
  loadUserInfo();
});
</script>

<template>
  <Page auto-content-height>
    <div class="flex h-full gap-4">
      <!-- 左侧用户卡片 -->
      <Card class="w-72 shrink-0">
        <div class="flex flex-col items-center py-6">
          <VbenAvatar
            :src="userInfo.avatar"
            :alt="userInfo.username || 'avatar'"
            :size="80"
            class="mb-4"
          />
          <div class="text-lg font-medium">
            {{ userInfo.nickname || userInfo.username }}
          </div>
          <div class="text-foreground/50 mt-1 text-sm">
            {{ userInfo.username }}
          </div>
          <div v-if="userInfo.email" class="text-foreground/40 mt-2 text-xs">
            {{ userInfo.email }}
          </div>
          <div
            v-if="userInfo.roles?.length"
            class="mt-3 flex flex-wrap justify-center gap-1"
          >
            <span
              v-for="role in userInfo.roles"
              :key="role"
              class="bg-primary/10 text-primary rounded-full px-2 py-0.5 text-xs"
            >
              {{ role }}
            </span>
          </div>
        </div>

        <Menu
          v-model:selectedKeys="activeKey"
          mode="inline"
          :bordered="false"
          class="bg-transparent"
        >
          <MenuItem v-for="item in menuItems" :key="item.key">
            <template #icon>
              <span :class="item.icon" class="size-4" />
            </template>
            {{ item.label }}
          </MenuItem>
        </Menu>
      </Card>

      <!-- 右侧内容区 -->
      <Card class="min-w-0 flex-1">
        <ProfileTab v-if="activeKey[0] === 'profile'" />
        <SecurityTab v-else-if="activeKey[0] === 'security'" />
        <OauthTab v-else-if="activeKey[0] === 'oauth'" />
        <PrivacyTab v-else-if="activeKey[0] === 'privacy'" />
      </Card>
    </div>
  </Page>
</template>
