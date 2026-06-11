<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { preferences } from '@vben/preferences';
import { useUserStore } from '@vben/stores';

const userStore = useUserStore();
const router = useRouter();

// 快捷入口
const quickLinks = ref([
  {
    icon: '📁',
    title: '文件管理',
    desc: '查看和管理您的文件',
    path: '/file/list',
  },
  {
    icon: '🔗',
    title: '分享管理',
    desc: '管理您分享的文件',
    path: '/file/share',
  },
  {
    icon: '🗑️',
    title: '回收站',
    desc: '查看已删除的文件',
    path: '/file/recycle',
  },
  {
    icon: '👤',
    title: '个人中心',
    desc: '管理您的账户信息',
    path: '/account/profile',
  },
]);

function navigateTo(path: string) {
  router.push(path).catch((err) => {
    console.error('Navigation failed:', err);
  });
}
</script>

<template>
  <div class="p-5">
    <!-- 欢迎区域 -->
    <div
      class="mb-6 rounded-lg bg-gradient-to-r from-blue-500 to-indigo-600 p-6 text-white shadow-md"
    >
      <div class="flex items-center gap-4">
        <img
          :src="userStore.userInfo?.avatar || preferences.app.defaultAvatar"
          alt="avatar"
          class="h-16 w-16 rounded-full border-2 border-white/30 object-cover"
        />
        <div>
          <h1 class="text-2xl font-bold">
            你好,
            {{
              userStore.userInfo?.realName ||
              userStore.userInfo?.nickname ||
              '用户'
            }}
          </h1>
          <p class="mt-1 text-white/80">欢迎使用 DevKit 文件管理系统</p>
        </div>
      </div>
    </div>

    <!-- 快捷入口 -->
    <div class="mb-6">
      <h2 class="mb-4 text-lg font-semibold text-foreground">快捷入口</h2>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div
          v-for="link in quickLinks"
          :key="link.path"
          class="cursor-pointer rounded-lg border border-border bg-card p-5 shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md"
          @click="navigateTo(link.path)"
        >
          <div class="mb-3 text-3xl">{{ link.icon }}</div>
          <h3 class="text-base font-medium text-foreground">
            {{ link.title }}
          </h3>
          <p class="mt-1 text-sm text-muted-foreground">{{ link.desc }}</p>
        </div>
      </div>
    </div>

    <!-- 使用提示 -->
    <div class="rounded-lg border border-border bg-card p-5 shadow-sm">
      <h2 class="mb-3 text-lg font-semibold text-foreground">使用提示</h2>
      <ul class="space-y-2 text-sm text-muted-foreground">
        <li class="flex items-start gap-2">
          <span class="mt-0.5 text-blue-500">💡</span>
          <span
            >您可以在<strong class="text-foreground">文件管理</strong
            >中上传、下载和管理您的文件。</span
          >
        </li>
        <li class="flex items-start gap-2">
          <span class="mt-0.5 text-blue-500">💡</span>
          <span
            >通过<strong class="text-foreground">分享管理</strong
            >可以创建文件分享链接，方便与他人共享文件。</span
          >
        </li>
        <li class="flex items-start gap-2">
          <span class="mt-0.5 text-blue-500">💡</span>
          <span
            >删除的文件会进入<strong class="text-foreground">回收站</strong
            >，您可以在回收站中恢复或彻底删除文件。</span
          >
        </li>
        <li class="flex items-start gap-2">
          <span class="mt-0.5 text-blue-500">💡</span>
          <span>请定期修改密码，保障您的账户安全。</span>
        </li>
      </ul>
    </div>
  </div>
</template>
