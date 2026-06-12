<script lang="ts" setup>
import { useRouter } from 'vue-router';

import { IconifyIcon } from '@vben/icons';
import { Page } from '@vben/common-ui';
import { preferences } from '@vben/preferences';
import { useUserStore } from '@vben/stores';

import { Card, Empty } from 'ant-design-vue';

const userStore = useUserStore();
const router = useRouter();

// 快捷入口
const quickLinks = [
  {
    color: 'bg-blue-50 text-blue-500',
    desc: '查看和管理您的文件',
    icon: 'lucide:folder',
    path: '/file/list',
    title: '文件管理',
  },
  {
    color: 'bg-purple-50 text-purple-500',
    desc: '管理您分享的文件',
    icon: 'lucide:share-2',
    path: '/file/share',
    title: '分享管理',
  },
  {
    color: 'bg-orange-50 text-orange-500',
    desc: '查看已删除的文件',
    icon: 'lucide:trash-2',
    path: '/file/recycle',
    title: '回收站',
  },
  {
    color: 'bg-green-50 text-green-500',
    desc: '管理您的账户信息',
    icon: 'lucide:user',
    path: '/account/profile',
    title: '个人中心',
  },
];

function navigateTo(path: string) {
  router.push(path).catch((err) => {
    console.error('Navigation failed:', err);
  });
}

const displayName =
  userStore.userInfo?.realName ||
  userStore.userInfo?.nickname ||
  '用户';
</script>

<template>
  <Page auto-content-height>
    <!-- 欢迎横幅 -->
    <div
      class="mb-6 rounded-xl bg-gradient-to-br from-blue-400/90 via-indigo-400/85 to-purple-400/80 p-8 text-white shadow-lg"
    >
      <div class="flex items-center gap-5">
        <img
          :src="userStore.userInfo?.avatar || preferences.app.defaultAvatar"
          alt="avatar"
          class="h-16 w-16 rounded-full border-2 border-white/40 object-cover shadow-md"
        />
        <div>
          <h1 class="text-2xl font-bold tracking-wide">
            你好，{{ displayName }}
          </h1>
          <p class="mt-2 text-sm text-white/70">
            欢迎使用 DevKit 文件管理系统，祝你工作愉快！
          </p>
        </div>
      </div>
    </div>

    <!-- 快捷入口 -->
    <Card title="快捷入口" class="mb-6" :bordered="true">
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div
          v-for="link in quickLinks"
          :key="link.path"
          class="group cursor-pointer rounded-lg border border-border p-5 transition-all duration-200 hover:-translate-y-1 hover:border-primary/30 hover:shadow-lg"
          @click="navigateTo(link.path)"
        >
          <div
            :class="link.color"
            class="mb-3 inline-flex h-11 w-11 items-center justify-center rounded-lg transition-transform duration-200 group-hover:scale-110"
          >
            <IconifyIcon :icon="link.icon" class="size-5" />
          </div>
          <h3 class="text-base font-medium text-foreground">
            {{ link.title }}
          </h3>
          <p class="mt-1 text-sm text-muted-foreground">{{ link.desc }}</p>
        </div>
      </div>
    </Card>

    <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- 系统公告 -->
      <Card title="系统公告" :bordered="true">
        <Empty description="暂无公告" />
      </Card>

      <!-- 最近访问 -->
      <Card title="最近访问" :bordered="true">
        <Empty description="暂无访问记录" />
      </Card>
    </div>

    <!-- 使用提示 -->
    <Card title="使用提示" class="mt-6" :bordered="true">
      <ul class="space-y-3 text-sm text-muted-foreground">
        <li class="flex items-start gap-2">
          <span
            class="mt-0.5 inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-blue-50 text-xs text-blue-500"
            >1</span
          >
          <span
            >您可以在<strong class="text-foreground">文件管理</strong>中上传、下载和管理您的文件。</span
          >
        </li>
        <li class="flex items-start gap-2">
          <span
            class="mt-0.5 inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-purple-50 text-xs text-purple-500"
            >2</span
          >
          <span
            >通过<strong class="text-foreground">分享管理</strong>可以创建文件分享链接，方便与他人共享文件。</span
          >
        </li>
        <li class="flex items-start gap-2">
          <span
            class="mt-0.5 inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-orange-50 text-xs text-orange-500"
            >3</span
          >
          <span
            >删除的文件会进入<strong class="text-foreground">回收站</strong>，您可以在回收站中恢复或彻底删除文件。</span
          >
        </li>
        <li class="flex items-start gap-2">
          <span
            class="mt-0.5 inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-green-50 text-xs text-green-500"
            >4</span
          >
          <span>请定期修改密码，保障您的账户安全。</span>
        </li>
      </ul>
    </Card>
  </Page>
</template>
