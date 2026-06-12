<script lang="ts" setup>
import type { NotificationItem } from '@vben/layouts';

import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { AuthenticationLoginExpiredModal } from '@vben/common-ui';
import { VBEN_DOC_URL, VBEN_GITHUB_URL } from '@vben/constants';
import { useWatermark } from '@vben/hooks';
import { BookOpenText, CircleHelp, SvgGithubIcon } from '@vben/icons';
import {
  BasicLayout,
  LockScreen,
  Notification,
  UserDropdown,
} from '@vben/layouts';
import { preferences, usePreferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { openWindow } from '@vben/utils';

import {
  deleteNotification,
  getNotifications,
  getUnreadCount,
  markAllRead,
  markRead,
} from '#/api/notification';
import { $t } from '#/locales';
import { useAuthStore } from '#/store';
import LoginForm from '#/views/_core/authentication/login.vue';

const notifications = ref<NotificationItem[]>([]);
const unreadCount = ref(0);
let pollTimer: ReturnType<typeof setInterval> | null = null;

// 通知类型图标映射
const typeAvatarMap: Record<string, string> = {
  login_alert: 'https://avatar.vercel.sh/login?text=🔐',
  upload_done: 'https://avatar.vercel.sh/upload?text=📁',
  role_change: 'https://avatar.vercel.sh/role?text=👤',
  role_approved: 'https://avatar.vercel.sh/approved?text=✅',
  role_rejected: 'https://avatar.vercel.sh/rejected?text=❌',
  announcement: 'https://avatar.vercel.sh/announce?text=📢',
  storage_warn: 'https://avatar.vercel.sh/storage?text=💾',
  security_warn: 'https://avatar.vercel.sh/security?text=⚠️',
};

// 时间格式化
function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60_000);
  if (minutes < 1) return '刚刚';
  if (minutes < 60) return `${minutes}分钟前`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}小时前`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}天前`;
  return date.toLocaleDateString('zh-CN');
}

// 加载通知列表
async function loadNotifications() {
  try {
    const res = await getNotifications({ page: 1, pageSize: 20 });
    notifications.value = (res.items || []).map((n) => ({
      id: n.id,
      avatar: typeAvatarMap[n.type] || 'https://avatar.vercel.sh/default?text=🔔',
      date: formatRelativeTime(n.createdAt),
      isRead: n.isRead,
      message: n.content,
      title: n.title,
      link: n.link || undefined,
    }));
  } catch {
    // 静默失败
  }
}

// 加载未读数量
async function loadUnreadCount() {
  try {
    const res = await getUnreadCount();
    unreadCount.value = res.count || 0;
  } catch {
    // 静默失败
  }
}

// 初始化
onMounted(async () => {
  await Promise.all([loadNotifications(), loadUnreadCount()]);
  // 每30秒轮询未读数量
  pollTimer = setInterval(loadUnreadCount, 30_000);
});

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer);
});

const router = useRouter();
const userStore = useUserStore();
const authStore = useAuthStore();
const accessStore = useAccessStore();
const { destroyWatermark, updateWatermark } = useWatermark();
const { isDark } = usePreferences();
const showDot = computed(() => unreadCount.value > 0);

const menus = computed(() => [
  {
    handler: () => {
      router.push({ name: 'AccountIndex' });
    },
    icon: 'lucide:user',
    text: $t('page.auth.profile'),
  },
  {
    handler: () => {
      openWindow(VBEN_DOC_URL, {
        target: '_blank',
      });
    },
    icon: BookOpenText,
    text: $t('ui.widgets.document'),
  },
  {
    handler: () => {
      openWindow(VBEN_GITHUB_URL, {
        target: '_blank',
      });
    },
    icon: SvgGithubIcon,
    text: 'GitHub',
  },
  {
    handler: () => {
      openWindow(`${VBEN_GITHUB_URL}/issues`, {
        target: '_blank',
      });
    },
    icon: CircleHelp,
    text: $t('ui.widgets.qa'),
  },
]);

const avatar = computed(() => {
  return userStore.userInfo?.avatar ?? preferences.app.defaultAvatar;
});

async function handleLogout() {
  await authStore.logout(false);
}

async function handleNoticeClear() {
  try {
    await markAllRead();
  } catch {
    // 静默失败
  }
  notifications.value.forEach((item) => (item.isRead = true));
  unreadCount.value = 0;
}

// 查看全部 - 跳转到个人中心安全日志页
function handleViewAll() {
  router.push({ path: '/account/index', query: { tab: 'security' } });
}

async function markReadItem(id: number | string) {
  const numId = Number(id);
  try {
    await markRead(numId);
    const item = notifications.value.find((item) => item.id === id);
    if (item) {
      item.isRead = true;
      unreadCount.value = Math.max(0, unreadCount.value - 1);
    }
  } catch {
    // 静默失败
  }
}

async function removeItem(id: number | string) {
  const numId = Number(id);
  try {
    await deleteNotification(numId);
    const item = notifications.value.find((item) => item.id === id);
    notifications.value = notifications.value.filter((item) => item.id !== id);
    if (item && !item.isRead) {
      unreadCount.value = Math.max(0, unreadCount.value - 1);
    }
  } catch {
    // 静默失败
  }
}

async function handleMakeAll() {
  try {
    await markAllRead();
    notifications.value.forEach((item) => (item.isRead = true));
    unreadCount.value = 0;
  } catch {
    // 静默失败
  }
}

const handleClick = (item: NotificationItem) => {
  if (item.link) {
    navigateTo(item.link, item.query, item.state);
  }
};

function navigateTo(
  link: string,
  query?: Record<string, any>,
  state?: Record<string, any>,
) {
  if (link.startsWith('http://') || link.startsWith('https://')) {
    window.open(link, '_blank', 'noopener,noreferrer');
  } else {
    router.push({
      path: link,
      query: query || {},
      state,
    });
  }
}

// WebSocket 实时通知
let wsInstance: WebSocket | null = null;
let wsRetryCount = 0;
const WS_MAX_RETRY = 5;

function connectWebSocket() {
  const accessStore = useAccessStore();
  if (!accessStore.accessToken) return;
  if (wsRetryCount >= WS_MAX_RETRY) return;

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  // token 通过 query 参数传递（浏览器 WebSocket 不支持自定义 Header）
  const wsUrl = `${protocol}//${host}/api/v1/ws?token=${encodeURIComponent(accessStore.accessToken)}`;

  try {
    wsInstance = new WebSocket(wsUrl);
    wsInstance.onopen = () => {
      wsRetryCount = 0;
    };
    wsInstance.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'notification') {
          loadNotifications();
          loadUnreadCount();
        }
      } catch {
        // 忽略非 JSON 消息
      }
    };
    wsInstance.onerror = () => {
      // 静默处理，不输出到控制台
    };
    wsInstance.onclose = () => {
      wsInstance = null;
      wsRetryCount++;
      // 指数退避重连：5s, 10s, 20s, 40s, 80s
      const delay = Math.min(5000 * 2 ** (wsRetryCount - 1), 80_000);
      setTimeout(connectWebSocket, delay);
    };
  } catch {
    // WebSocket 连接失败，静默处理
  }
}

onMounted(() => {
  connectWebSocket();
});

onUnmounted(() => {
  if (wsInstance) {
    wsInstance.close();
    wsInstance = null;
  }
});

watch(
  () => ({
    enable: preferences.app.watermark,
    content: preferences.app.watermarkContent,
    opacity: preferences.app.watermarkOpacity,
    isDark: isDark.value,
  }),
  async ({ enable, content, opacity, isDark: isDarkValue }) => {
    if (enable) {
      const opacityValue = opacity || 0.12;
      const watermarkColor = isDarkValue
        ? `rgba(255, 255, 255, ${opacityValue})`
        : `rgba(0, 0, 0, ${opacityValue})`;

      await updateWatermark({
        advancedStyle: {
          colorStops: [
            {
              color: watermarkColor,
              offset: 0,
            },
            {
              color: watermarkColor,
              offset: 1,
            },
          ],
          type: 'linear',
        },
        content:
          content ||
          `${userStore.userInfo?.username} - ${userStore.userInfo?.realName}`,
      });
    } else {
      destroyWatermark();
    }
  },
  {
    immediate: true,
  },
);
</script>

<template>
  <BasicLayout @clear-preferences-and-logout="handleLogout">
    <template #user-dropdown>
      <UserDropdown
        :avatar
        :menus
        :text="userStore.userInfo?.realName"
        description="ann.vben@gmail.com"
        tag-text="Pro"
        @logout="handleLogout"
        @clear-preferences-and-logout="handleLogout"
      />
    </template>
    <template #notification>
      <Notification
        :dot="showDot"
        :notifications="notifications"
        @clear="handleNoticeClear"
        @read="(item: any) => item.id && markReadItem(item.id)"
        @remove="(item: any) => item.id && removeItem(item.id)"
        @makeAll="handleMakeAll"
        @onClick="handleClick"
        @viewAll="handleViewAll"
      />
    </template>
    <template #extra>
      <AuthenticationLoginExpiredModal
        v-model:open="accessStore.loginExpired"
        :avatar
      >
        <LoginForm />
      </AuthenticationLoginExpiredModal>
    </template>
    <template #lock-screen>
      <LockScreen :avatar @to-login="handleLogout" />
    </template>
  </BasicLayout>
</template>
