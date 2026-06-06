<script lang="ts" setup>
import type { AccountApi } from '#/api';

import { onMounted, ref } from 'vue';

import { Button, Card, message, Popconfirm, Spin, Tag } from 'ant-design-vue';

import { getOAuthBindings, unbindOAuth } from '#/api';
import { $t } from '#/locales';

const bindings = ref<AccountApi.OAuthBinding[]>([]);
const loading = ref(false);

const providers = [
  { key: 'wechat', label: 'WeChat' },
  { key: 'github', label: 'GitHub' },
  { key: 'google', label: 'Google' },
];

async function loadBindings() {
  loading.value = true;
  try {
    const data = await getOAuthBindings();
    bindings.value = Array.isArray(data) ? data : [];
  } catch {
    bindings.value = [];
  } finally {
    loading.value = false;
  }
}

async function handleUnbind(provider: string) {
  try {
    await unbindOAuth({ provider });
    message.success($t('account.security.unbindSuccess'));
    await loadBindings();
  } catch (error: any) {
    if (error?.message?.includes('only login method')) {
      message.error($t('account.security.unbindOnlyMethod'));
    }
  }
}

function isProviderBound(provider: string) {
  return (bindings.value || []).some((b) => b.provider === provider);
}

onMounted(() => {
  loadBindings();
});
</script>

<template>
  <Spin :spinning="loading">
    <div class="space-y-4">
      <Card v-for="provider in providers" :key="provider.key" size="small">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <span class="font-medium">{{ provider.label }}</span>
            <Tag v-if="isProviderBound(provider.key)" color="success">
              {{ $t('account.security.bound') }}
            </Tag>
            <Tag v-else>{{ $t('account.security.unbound') }}</Tag>
          </div>
          <Popconfirm
            v-if="isProviderBound(provider.key)"
            :title="$t('account.security.unbindConfirm', [provider.label])"
            @confirm="handleUnbind(provider.key)"
          >
            <Button danger size="small">
              {{ $t('account.security.unbind') }}
            </Button>
          </Popconfirm>
          <Button v-else size="small" type="primary">
            {{ $t('account.security.bind') }}
          </Button>
        </div>
      </Card>
    </div>
  </Spin>
</template>