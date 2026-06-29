<template>
  <view class="webview-page">
    <web-view v-if="url" :src="url" />
    <view v-else class="empty">
      <text class="icon">💬</text>
      <text class="text">暂无客服链接</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { onLoad } from '@dcloudio/uni-app';
import { request } from '@/api/request';

const url = ref('');

onLoad((options: any) => {
  if (options?.url) {
    url.value = decodeURIComponent(options.url);
    return;
  }
});

onMounted(async () => {
  if (!url.value) {
    try {
      const data = await request.get<any>('/mobile/settings');
      if (data?.customerServiceUrl) {
        url.value = data.customerServiceUrl;
      }
    } catch {
      // ignore
    }
  }
});
</script>

<style lang="scss" scoped>
.webview-page {
  width: 100%;
  height: 100vh;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #f5f5f5;

  .icon { font-size: 48px; margin-bottom: 12px; }
  .text { font-size: 14px; color: #999; }
}
</style>
