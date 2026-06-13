<template>
  <view class="notification-page">
    <!-- 顶部操作 -->
    <view class="top-bar">
      <text class="page-title">消息通知</text>
      <view class="top-actions" v-if="notifications.length">
        <text class="mark-all-btn" @click="markAllAsRead">全部已读</text>
      </view>
    </view>

    <!-- 通知列表 -->
    <view class="list-section">
      <view v-if="loading" class="loading">
        <Skeleton type="list" :count="5" />
      </view>
      <view v-else-if="notifications.length === 0" class="empty">
        <text class="empty-icon">🔔</text>
        <text class="empty-text">暂无通知</text>
      </view>
      <view v-else>
        <view
          v-for="n in notifications"
          :key="n.id"
          class="notif-card"
          :class="{ unread: !n.isRead }"
          @click="goToDetail(n)"
        >
          <view class="notif-icon">
            <text>{{ getTypeIcon(n.type) }}</text>
          </view>
          <view class="notif-content">
            <view class="notif-header">
              <text class="notif-title">{{ n.title }}</text>
              <view v-if="!n.isRead" class="unread-dot" />
            </view>
            <text class="notif-desc">{{ n.content }}</text>
            <text class="notif-time">{{ formatTime(n.createdAt) }}</text>
          </view>
          <view class="notif-actions" v-if="!n.isRead">
            <text class="read-btn" @click.stop="markAsRead(n)">标为已读</text>
          </view>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="hasMore" class="load-more" @click="loadMore">
        <text>{{ loading ? '加载中...' : '加载更多' }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getNotifications, markRead, markAllRead, type Notification } from '@/api/notification';
import Skeleton from '@/components/Skeleton.vue';

const loading = ref(false);
const notifications = ref<Notification[]>([]);
const page = ref(1);
const pageSize = 20;
const hasMore = ref(true);

onMounted(() => {
  fetchNotifications(true);
});

async function fetchNotifications(refresh = false) {
  if (refresh) {
    page.value = 1;
    notifications.value = [];
  }
  loading.value = true;
  try {
    const res = await getNotifications({ page: page.value, pageSize });
    const items = res.items || [];
    if (refresh) notifications.value = items;
    else notifications.value = [...notifications.value, ...items];
    hasMore.value = page.value * pageSize < res.total;
  } catch {
    // Mock 数据（仅开发环境）
    if (import.meta.env.DEV) {
      const mockData: Notification[] = Array.from({ length: 5 }, (_, i) => ({
        id: i + 1,
        title: ['系统通知', '练习提醒', '错题分析', '版本更新', '活动通知'][i],
        content: ['您有新的系统消息', '今天还没练习哦，加油！', '本周错题分析已生成', '应用已更新至最新版本', '周末有优惠活动'][i],
        type: ['system', 'reminder', 'analysis', 'update', 'activity'][i],
        isRead: i > 1,
        createdAt: new Date(Date.now() - i * 86400000).toISOString(),
      }));
      if (refresh) notifications.value = mockData;
      else notifications.value = [...notifications.value, ...mockData];
      hasMore.value = false;
    } else {
      uni.showToast({ title: '加载通知失败', icon: 'none' });
    }
  } finally {
    loading.value = false;
  }
}

function loadMore() {
  page.value++;
  fetchNotifications();
}

async function markAsRead(n: Notification) {
  try {
    await markRead(n.id);
    n.isRead = true;
  } catch {
    n.isRead = true; // 乐观更新
  }
}

async function markAllAsRead() {
  try {
    await markAllRead();
    notifications.value.forEach(n => { n.isRead = true; });
    uni.showToast({ title: '已全部标为已读', icon: 'success' });
  } catch {
    // 乐观更新
    notifications.value.forEach(n => { n.isRead = true; });
  }
}

function goToDetail(n: Notification) {
  if (!n.isRead) {
    markAsRead(n);
  }
  uni.navigateTo({ url: `/pages/notification/detail?id=${n.id}` });
}

function getTypeIcon(type: string): string {
  const icons: Record<string, string> = {
    system: '📢',
    reminder: '⏰',
    analysis: '📊',
    update: '🔄',
    activity: '🎉',
  };
  return icons[type] || '🔔';
}

function formatTime(dateStr: string): string {
  const d = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  const minDiff = Math.floor(diff / 60000);
  if (minDiff < 1) return '刚刚';
  if (minDiff < 60) return `${minDiff} 分钟前`;
  const hourDiff = Math.floor(minDiff / 60);
  if (hourDiff < 24) return `${hourDiff} 小时前`;
  const dayDiff = Math.floor(hourDiff / 24);
  if (dayDiff < 7) return `${dayDiff} 天前`;
  return `${d.getMonth() + 1}/${d.getDate()}`;
}
</script>

<style lang="scss" scoped>
.notification-page {
  min-height: 100vh;
  background: #f5f5f5;

  .top-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    background: #fff;
    border-bottom: 1px solid #f0f0f0;

    .page-title {
      font-size: 18px;
      font-weight: bold;
      color: #333;
    }

    .mark-all-btn {
      font-size: 14px;
      color: #1890ff;
    }
  }

  .list-section {
    padding: 12px 16px;

    .loading { padding: 16px 0; }

    .empty {
      text-align: center;
      padding: 60px 0;

      .empty-icon {
        font-size: 48px;
        display: block;
        margin-bottom: 12px;
      }

      .empty-text {
        font-size: 14px;
        color: #999;
      }
    }

    .notif-card {
      display: flex;
      align-items: flex-start;
      background: #fff;
      border-radius: 12px;
      padding: 14px;
      margin-bottom: 10px;

      &.unread {
        background: #f0f9ff;
        border-left: 3px solid #1890ff;
      }

      .notif-icon {
        font-size: 24px;
        margin-right: 12px;
        flex-shrink: 0;
      }

      .notif-content {
        flex: 1;
        min-width: 0;

        .notif-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          margin-bottom: 4px;

          .notif-title {
            font-size: 15px;
            font-weight: 500;
            color: #333;
          }

          .unread-dot {
            width: 8px;
            height: 8px;
            background: #ff4d4f;
            border-radius: 50%;
            flex-shrink: 0;
          }
        }

        .notif-desc {
          font-size: 13px;
          color: #666;
          line-height: 1.4;
          margin-bottom: 6px;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
          overflow: hidden;
        }

        .notif-time {
          font-size: 11px;
          color: #999;
        }
      }

      .notif-actions {
        flex-shrink: 0;
        margin-left: 8px;

        .read-btn {
          font-size: 12px;
          color: #1890ff;
          padding: 4px 8px;
        }
      }
    }

    .load-more {
      text-align: center;
      padding: 16px 0;
      font-size: 13px;
      color: #999;
    }
  }
}
</style>
