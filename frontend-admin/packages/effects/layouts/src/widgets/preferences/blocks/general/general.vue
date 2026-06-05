<script setup lang="ts">
import { onMounted, ref, unref } from 'vue';

import { SUPPORT_LANGUAGES } from '@vben/constants';
import { $t } from '@vben/locales';
import { useTimezoneStore } from '@vben/stores';

import SelectItem from '../select-item.vue';
import SwitchItem from '../switch-item.vue';

defineOptions({
  name: 'PreferenceGeneralConfig',
});

const appLocale = defineModel<string>('appLocale');
const appTimezone = defineModel<string>('appTimezone');
const appDynamicTitle = defineModel<boolean>('appDynamicTitle');
const appEnableCheckUpdates = defineModel<boolean>('appEnableCheckUpdates');
const appEnableCopyPreferences = defineModel<boolean>(
  'appEnableCopyPreferences',
);
const timezoneStore = useTimezoneStore();

const timezoneOptionsRef = ref<
  {
    label: string;
    value: string;
  }[]
>([]);

onMounted(async () => {
  timezoneOptionsRef.value = await timezoneStore.getTimezoneOptions();
  // 获取当前时区，例如：Asia/Shanghai
  const timezoneValue = unref(timezoneStore.timezone);
  if (timezoneValue) {
    appTimezone.value = timezoneValue;
  }
});
</script>

<template>
  <SelectItem v-model="appLocale" :items="SUPPORT_LANGUAGES">
    {{ $t('preferences.language') }}
  </SelectItem>
  <SelectItem v-model="appTimezone" :items="timezoneOptionsRef">
    {{ $t('preferences.timezone') }}
  </SelectItem>
  <SwitchItem v-model="appDynamicTitle">
    {{ $t('preferences.dynamicTitle') }}
  </SwitchItem>
  <SwitchItem v-model="appEnableCheckUpdates">
    {{ $t('preferences.checkUpdates') }}
  </SwitchItem>
  <SwitchItem v-model="appEnableCopyPreferences">
    {{ $t('preferences.enableCopyPreferences') }}
  </SwitchItem>
</template>
