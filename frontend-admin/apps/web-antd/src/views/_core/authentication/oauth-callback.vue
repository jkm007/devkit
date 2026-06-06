<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Spin } from 'ant-design-vue';

import { handleOAuthCallback } from '#/api/core/auth';
import { useAuthStore } from '#/store';

defineOptions({ name: 'OAuthCallback' });

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const statusText = ref('正在处理第三方登录...');
const isError = ref(false);

onMounted(async () => {
  const provider = route.query.provider as string;
  const code = route.query.code as string;
  const state = route.query.state as string;

  if (!provider || !code || !state) {
    statusText.value = '缺少必要参数';
    isError.value = true;
    return;
  }

  try {
    const result = await handleOAuthCallback(provider, code, state);
    if (result?.accessToken) {
      // 使用 authStore 的 handleLoginResult 处理登录后的通用逻辑
      // 但这里直接调用 API，所以手动处理
      const { useAccessStore } = await import('@vben/stores');
      const accessStore = useAccessStore();

      accessStore.setAccessToken(result.accessToken);
      if (result.refreshToken) {
        accessStore.setRefreshToken(result.refreshToken);
      }

      // 获取用户信息
      await authStore.fetchUserInfo();

      statusText.value = '登录成功，正在跳转...';

      // 跳转到首页
      setTimeout(() => {
        router.replace('/');
      }, 500);
    } else {
      statusText.value = '登录失败，请重试';
      isError.value = true;
    }
  } catch (e: any) {
    statusText.value = e?.message || '第三方登录失败';
    isError.value = true;
    // 3秒后跳回登录页
    setTimeout(() => {
      router.replace('/auth/login');
    }, 3000);
  }
});
</script>

<template>
  <div class="flex flex-col items-center justify-center min-h-[300px]">
    <Spin v-if="!isError" size="large" />
    <div class="mt-4 text-center">
      <p :class="isError ? 'text-destructive' : 'text-muted-foreground'">
        {{ statusText }}
      </p>
      <p v-if="isError" class="mt-2 text-sm text-muted-foreground">
        3秒后自动跳转到登录页...
      </p>
    </div>
  </div>
</template>
