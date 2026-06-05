<script lang="ts" setup>
import type { AccountApi } from '#/api';

import { onMounted, ref } from 'vue';

import {
  Button,
  List,
  ListItem,
  ListItemMeta,
  message,
  Popconfirm,
  Tag,
} from 'ant-design-vue';

import { getOAuthBindings, unbindOAuth } from '#/api';
import { $t } from '#/locales';

const bindings = ref<AccountApi.OAuthBinding[]>([]);
const loading = ref(false);

const providerLabels: Record<string, string> = {
  wechat: 'WeChat',
  github: 'GitHub',
  google: 'Google',
};

async function loadBindings() {
  loading.value = true;
  try {
    bindings.value = await getOAuthBindings();
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
  return bindings.value.some((b) => b.provider === provider);
}

onMounted(() => {
  loadBindings();
});
</script>

<template>
  <div v-loading="loading">
    <List :data-source="['wechat', 'github', 'google']" bordered>
      <template #renderItem="{ item: provider }">
        <ListItem>
          <ListItemMeta :title="providerLabels[provider]">
            <template #description>
              <Tag v-if="isProviderBound(provider)" color="success">
                {{ $t('account.security.bound') }}
              </Tag>
              <Tag v-else>{{ $t('account.security.unbound') }}</Tag>
            </template>
          </ListItemMeta>
          <template #actions>
            <Popconfirm
              v-if="isProviderBound(provider)"
              :title="$t('account.security.unbindConfirm', [providerLabels[provider]])"
              @confirm="handleUnbind(provider)"
            >
              <Button danger size="small">
                {{ $t('account.security.unbind') }}
              </Button>
            </Popconfirm>
            <Button v-else size="small" type="primary">
              {{ $t('account.security.bind') }}
            </Button>
          </template>
        </ListItem>
      </template>
    </List>
  </div>
</template>
