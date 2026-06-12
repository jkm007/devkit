<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { IconifyIcon } from '@vben/icons';
import { Page } from '@vben/common-ui';
import { preferences } from '@vben/preferences';
import { useUserStore } from '@vben/stores';

import { Badge, Card, Empty, Progress, Tag, Tooltip } from 'ant-design-vue';

import type { HomeData } from '#/api/user-home';
import { getHomeData } from '#/api/user-home';

const userStore = useUserStore();
const router = useRouter();

const homeData = ref<HomeData | null>(null);
const loading = ref(true);

onMounted(async () => {
  try {
    homeData.value = await getHomeData();
  } catch {
    // 静默失败
  } finally {
    loading.value = false;
  }
});

// 格式化文件大小
function formatSize(bytes: number): string {
  if (bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / 1024 ** i).toFixed(1)} ${units[i]}`;
}

// 存储进度条颜色
function getStorageColor(percent: number): string {
  if (percent >= 90) return '#ff4d4f';
  if (percent >= 70) return '#faad14';
  return '#52c41a';
}

// 设备类型图标
const deviceIconMap: Record<string, string> = {
  web: 'lucide:monitor',
  h5: 'lucide:smartphone',
  app: 'lucide:smartphone',
  miniapp: 'lucide:layout-grid',
};

// 设备类型中文
const deviceTypeMap: Record<string, string> = {
  web: 'Web浏览器',
  h5: 'H5移动端',
  app: 'APP',
  miniapp: '小程序',
};

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

    <!-- 存储使用 + 角色信息 -->
    <div class="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- 存储使用 -->
      <Card title="存储空间" :bordered="true">
        <div v-if="homeData?.storage" class="space-y-4">
          <div class="flex items-center justify-between">
            <span class="text-sm text-muted-foreground">已使用</span>
            <span class="text-lg font-semibold">
              {{ formatSize(homeData.storage.used) }}
              <span v-if="homeData.storage.quota > 0" class="text-sm font-normal text-muted-foreground">
                / {{ formatSize(homeData.storage.quota) }}
              </span>
              <span v-else class="text-sm font-normal text-muted-foreground">
                （不限制）
              </span>
            </span>
          </div>
          <Progress
            v-if="homeData.storage.quota > 0"
            :percent="Math.min(homeData.storage.usedPercent, 100)"
            :stroke-color="getStorageColor(homeData.storage.usedPercent)"
            :show-text="true"
            :format="(p: number) => `${p.toFixed(1)}%`"
          />
          <div v-else class="text-sm text-muted-foreground">
            您的存储空间没有容量限制
          </div>
          <div v-if="homeData.storage.quota > 0 && homeData.storage.usedPercent >= 80" class="rounded-lg bg-orange-50 p-3 text-sm text-orange-600">
            ⚠️ 存储空间即将用尽，请及时清理文件或联系管理员扩容
          </div>
        </div>
        <Empty v-else-if="!loading" description="暂无数据" />
      </Card>

      <!-- 角色信息 -->
      <Card title="我的角色" :bordered="true">
        <div v-if="homeData?.roles && homeData.roles.length > 0" class="space-y-3">
          <div v-for="role in homeData.roles" :key="role.id" class="flex items-center gap-3 rounded-lg border border-border p-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-full bg-blue-50 text-blue-500">
              <IconifyIcon icon="lucide:shield-check" class="size-5" />
            </div>
            <div>
              <div class="font-medium text-foreground">{{ role.name }}</div>
              <div class="text-xs text-muted-foreground">{{ role.code }}</div>
            </div>
          </div>
        </div>
        <Empty v-else-if="!loading" description="暂无角色信息" />
      </Card>
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

    <!-- 登录设备 -->
    <Card title="登录设备" :bordered="true">
      <div v-if="homeData?.devices && homeData.devices.length > 0" class="space-y-3">
        <div
          v-for="device in homeData.devices"
          :key="device.id"
          class="flex items-center gap-4 rounded-lg border border-border p-4 transition-colors hover:bg-muted/50"
        >
          <div class="flex h-10 w-10 items-center justify-center rounded-full bg-muted">
            <IconifyIcon :icon="deviceIconMap[device.deviceType] || 'lucide:monitor'" class="size-5 text-muted-foreground" />
          </div>
          <div class="flex-1">
            <div class="flex items-center gap-2">
              <span class="font-medium text-foreground">
                {{ device.deviceName || deviceTypeMap[device.deviceType] || '未知设备' }}
              </span>
              <Badge v-if="device.isCurrent" status="success" text="当前设备" />
            </div>
            <div class="mt-1 text-xs text-muted-foreground">
              <span v-if="device.browser">{{ device.browser }}</span>
              <span v-if="device.os"> · {{ device.os }}</span>
              <span v-if="device.ip"> · {{ device.ip }}</span>
            </div>
          </div>
          <div class="text-right text-xs text-muted-foreground">
            <div v-if="device.lastActiveAt">{{ device.lastActiveAt }}</div>
            <Tag v-if="device.platform" size="small">{{ device.platform }}</Tag>
          </div>
        </div>
      </div>
      <Empty v-else-if="!loading" description="暂无登录设备" />
    </Card>
  </Page>
</template>
