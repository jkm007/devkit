<template>
  <view class="devices-page">
    <view class="list-section">
      <view v-if="loading" class="loading">
        <Skeleton type="list" :count="3" />
      </view>
      <view v-else-if="devices.length === 0" class="empty">
        <text class="empty-icon">📱</text>
        <text class="empty-text">暂无设备记录</text>
      </view>
      <view v-else>
        <view v-for="d in devices" :key="d.id" class="device-card" :class="{ current: d.isCurrent }">
          <view class="device-header">
            <view class="device-info">
              <text class="device-icon">{{ getDeviceIcon(d.deviceType) }}</text>
              <view class="device-detail">
                <text class="device-name">{{ d.deviceName || d.deviceType }}</text>
                <text class="device-meta">{{ d.ipAddress }} · {{ d.location || '未知位置' }}</text>
              </view>
            </view>
            <view v-if="d.isCurrent" class="current-badge">当前</view>
          </view>
          <view class="device-footer">
            <text class="login-time">登录于 {{ formatTime(d.loginAt) }}</text>
            <text v-if="!d.isCurrent" class="kick-btn" @click="kickDevice(d)">踢出</text>
          </view>
        </view>
      </view>
    </view>

    <view v-if="devices.length > 1" class="action-section">
      <button class="kick-all-btn" @click="kickAllOther">踢出其他所有设备</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getDevices, kickDevice as kickDeviceApi, kickAllOtherDevices } from '@/api/user-feature';
import Skeleton from '@/components/Skeleton.vue';

interface Device {
  id: number;
  deviceType: string;
  deviceName: string;
  ipAddress: string;
  location: string;
  loginAt: string;
  isCurrent: boolean;
}

const loading = ref(true);
const devices = ref<Device[]>([]);

onMounted(() => {
  loadDevices();
});

async function loadDevices() {
  loading.value = true;
  try {
    const res = await getDevices();
    devices.value = res || [];
  } catch {
    // Mock data only in development
    if (import.meta.env.DEV) {
      devices.value = [
        { id: 1, deviceType: 'mobile', deviceName: 'iPhone 15', ipAddress: '192.168.1.100', location: '北京', loginAt: new Date().toISOString(), isCurrent: true },
        { id: 2, deviceType: 'web', deviceName: 'Chrome / Windows', ipAddress: '192.168.1.50', location: '北京', loginAt: new Date(Date.now() - 86400000).toISOString(), isCurrent: false },
      ];
    } else {
      uni.showToast({ title: '加载设备列表失败', icon: 'none' });
    }
  } finally {
    loading.value = false;
  }
}

async function kickDevice(d: Device) {
  uni.showModal({
    title: '确认踢出',
    content: `确定要踢出「${d.deviceName}」吗？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await kickDeviceApi(d.id);
          uni.showToast({ title: '已踢出', icon: 'success' });
          devices.value = devices.value.filter(item => item.id !== d.id);
        } catch {
          uni.showToast({ title: '操作失败', icon: 'none' });
        }
      }
    },
  });
}

async function kickAllOther() {
  uni.showModal({
    title: '确认踢出',
    content: '确定要踢出其他所有设备吗？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await kickAllOtherDevices();
          uni.showToast({ title: '已踢出所有其他设备', icon: 'success' });
          const current = devices.value.find(d => d.isCurrent);
          devices.value = current ? [current] : [];
        } catch {
          uni.showToast({ title: '操作失败', icon: 'none' });
        }
      }
    },
  });
}

function getDeviceIcon(type: string): string {
  if (type?.includes('mobile') || type?.includes('iOS') || type?.includes('Android')) return '📱';
  if (type?.includes('web') || type?.includes('Chrome') || type?.includes('Firefox')) return '💻';
  return '🖥️';
}

function formatTime(dateStr: string): string {
  const d = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  const hourDiff = Math.floor(diff / 3600000);
  if (hourDiff < 1) return '刚刚';
  if (hourDiff < 24) return `${hourDiff} 小时前`;
  const dayDiff = Math.floor(hourDiff / 24);
  if (dayDiff < 7) return `${dayDiff} 天前`;
  return `${d.getMonth() + 1}/${d.getDate()}`;
}
</script>

<style lang="scss" scoped>
.devices-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 30px;

  .list-section {
    padding: 16px;

    .loading { padding: 16px 0; }

    .empty {
      text-align: center;
      padding: 60px 0;

      .empty-icon { font-size: 48px; display: block; margin-bottom: 12px; }
      .empty-text { font-size: 14px; color: #999; }
    }

    .device-card {
      background: #fff;
      border-radius: 12px;
      padding: 14px;
      margin-bottom: 10px;

      &.current {
        border-left: 3px solid #52c41a;
      }

      .device-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 10px;

        .device-info {
          display: flex;
          align-items: center;
          gap: 10px;

          .device-icon { font-size: 28px; }

          .device-detail {
            .device-name {
              font-size: 15px;
              font-weight: 500;
              color: #333;
              display: block;
            }

            .device-meta {
              font-size: 12px;
              color: #999;
              margin-top: 2px;
              display: block;
            }
          }
        }

        .current-badge {
          padding: 2px 8px;
          background: #f6ffed;
          color: #52c41a;
          font-size: 11px;
          border-radius: 4px;
        }
      }

      .device-footer {
        display: flex;
        justify-content: space-between;
        align-items: center;

        .login-time {
          font-size: 12px;
          color: #999;
        }

        .kick-btn {
          font-size: 13px;
          color: #ff4d4f;
          padding: 4px 12px;
          background: #fff2f0;
          border-radius: 12px;
        }
      }
    }
  }

  .action-section {
    margin: 0 16px;

    .kick-all-btn {
      height: 44px;
      line-height: 44px;
      background: #fff;
      color: #ff4d4f;
      border: 1px solid #ffccc7;
      border-radius: 22px;
      font-size: 15px;
    }
  }
}
</style>
