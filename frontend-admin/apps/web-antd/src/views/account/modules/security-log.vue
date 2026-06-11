<script lang="ts" setup>
import type { AccountApi } from '#/api';

import { onMounted, ref } from 'vue';

import { List, ListItem, ListItemMeta, Tag } from 'ant-design-vue';

import { getMySecurityLogs } from '#/api';
import { $t } from '#/locales';

const securityLogs = ref<AccountApi.SecurityLog[]>([]);
const loading = ref(false);

async function loadSecurityLogs() {
  loading.value = true;
  try {
    const res = await getMySecurityLogs({ page: 1, pageSize: 20 });
    securityLogs.value = res.items || [];
  } catch {
    securityLogs.value = [];
  } finally {
    loading.value = false;
  }
}

function getEventTypeLabel(type: string) {
  return $t(`account.eventType.${type}`);
}

function getStatusTag(status: number) {
  return status === 1
    ? { color: 'success', text: $t('account.security.statusSuccess') }
    : { color: 'error', text: $t('account.security.statusFail') };
}

onMounted(() => {
  loadSecurityLogs();
});
</script>

<template>
  <div v-loading="loading">
    <List :data-source="securityLogs" bordered size="small">
      <template #renderItem="{ item: log }">
        <ListItem>
          <ListItemMeta>
            <template #title>
              <Tag :color="getStatusTag(log.status).color">
                {{ getStatusTag(log.status).text }}
              </Tag>
              <span class="ml-2">{{ getEventTypeLabel(log.eventType) }}</span>
            </template>
            <template #description>
              <span class="text-foreground/50">{{ log.eventDetail }}</span>
              <span class="text-foreground/40 ml-4 text-xs">
                {{ log.ip }} · {{ log.createdAt }}
              </span>
            </template>
          </ListItemMeta>
        </ListItem>
      </template>
    </List>
  </div>
</template>
