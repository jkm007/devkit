<script lang="ts" setup>
import { computed } from 'vue';

import { Fallback } from '@vben/common-ui';
import { useUserStore } from '@vben/stores';

import { preferences } from '@vben/preferences';

defineOptions({ name: 'Fallback404Demo' });

const userStore = useUserStore();
const homePath = computed(
  () => userStore.userInfo?.homePath || preferences.app.defaultHomePath,
);
</script>

<template>
  <Fallback status="404" :home-path="homePath" />
</template>
