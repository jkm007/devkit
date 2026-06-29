<script lang="ts" setup>
import type { SystemUserApi } from '#/api/system/user';

import { computed, ref } from 'vue';

import { useVbenDrawer, VbenDescriptions } from '@vben/common-ui';

import { Progress } from 'ant-design-vue';

import { $t } from '#/locales';

import { useDescriptionItems } from '../data';

const detailData = ref<SystemUserApi.SystemUser>();

const items = computed(() => useDescriptionItems(detailData.value));

const storageInfo = computed(() => {
  const row = detailData.value;
  if (!row) return null;
  const used = row.storageUsed || 0;
  const quota =
    row.storageQuota > 0 ? row.storageQuota : row.roleStorageQuota || 0;
  const formatSize = (bytes: number) => {
    if (bytes <= 0) return '0 B';
    if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(2)} GB`;
    if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(1)} MB`;
    if (bytes >= 1024) return `${(bytes / 1024).toFixed(0)} KB`;
    return `${bytes} B`;
  };
  const percent = quota > 0 ? Math.min(Math.round((used / quota) * 100), 100) : 0;
  const status: 'exception' | 'normal' | 'active' = percent >= 90 ? 'exception' : percent >= 70 ? 'normal' : 'active';
  return { used, quota, percent, status, formatSize };
});

const [Drawer, drawerApi] = useVbenDrawer({
  onOpenChange(isOpen) {
    if (isOpen) {
      detailData.value = drawerApi.getData<SystemUserApi.SystemUser>();
    }
  },
});
</script>
<template>
  <Drawer :footer="false" :title="$t('common.detail')">
    <VbenDescriptions bordered :column="1" :items="items" />
    <div v-if="storageInfo" class="mt-4 rounded border p-4">
      <div class="mb-2 font-medium">{{ $t('system.user.storage') }}</div>
      <div class="flex items-center gap-3">
        <Progress
          :percent="storageInfo.percent"
          :status="storageInfo.status"
          size="small"
          style="width: 200px; margin-bottom: 0"
        />
        <span class="text-sm">
          {{ storageInfo.formatSize(storageInfo.used) }}
          <template v-if="storageInfo.quota > 0">
            / {{ storageInfo.formatSize(storageInfo.quota) }}
          </template>
          <template v-else> (不限) </template>
        </span>
      </div>
    </div>
  </Drawer>
</template>
