<template>
  <view class="settings-page">
    <!-- 账号安全 -->
    <view class="section">
      <text class="section-title">账号安全</text>
      <view class="list-group">
        <view class="list-item" @click="goToSecurityLogs">
          <text class="item-label">安全日志</text>
          <text class="item-desc">查看登录记录</text>
          <text class="item-arrow">›</text>
        </view>
        <view class="list-item" @click="goToDevices">
          <text class="item-label">登录设备</text>
          <text class="item-desc">管理已登录设备</text>
          <text class="item-arrow">›</text>
        </view>
        <view class="list-item" @click="changePassword">
          <text class="item-label">修改密码</text>
          <text class="item-desc">定期修改更安全</text>
          <text class="item-arrow">›</text>
        </view>
      </view>
    </view>

    <!-- 通知设置 -->
    <view class="section">
      <text class="section-title">通知设置</text>
      <view class="list-group">
        <view class="list-item switch-item">
          <text class="item-label">推送通知</text>
          <switch :checked="settings.pushNotification" color="#1890ff" @change="toggleSetting('pushNotification', $event)" />
        </view>
        <view class="list-item switch-item">
          <text class="item-label">练习提醒</text>
          <switch :checked="settings.practiceReminder" color="#1890ff" @change="toggleSetting('practiceReminder', $event)" />
        </view>
      </view>
    </view>

    <!-- 学习设置 -->
    <view class="section">
      <text class="section-title">学习设置</text>
      <view class="list-group">
        <view class="list-item" @click="goToCategories">
          <text class="item-label">分类设置</text>
          <text class="item-desc">管理学习分类（最多3个）</text>
          <text class="item-arrow">›</text>
        </view>
        <view class="list-item switch-item">
          <text class="item-label">答题后显示解析</text>
          <switch :checked="settings.showAnalysis" color="#1890ff" @change="toggleSetting('showAnalysis', $event)" />
        </view>
        <view class="list-item">
          <text class="item-label">每日练习目标</text>
          <view class="target-selector" @click="changeDailyTarget">
            <text class="target-value">{{ settings.dailyTarget }} 题</text>
            <text class="item-arrow">›</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 缓存管理 -->
    <view class="section">
      <text class="section-title">缓存管理</text>
      <view class="list-group">
        <view class="list-item" @click="clearCache">
          <text class="item-label">清除缓存</text>
          <text class="item-desc">{{ cacheSize }}</text>
          <text class="item-arrow">›</text>
        </view>
      </view>
    </view>

    <!-- 系统功能 -->
    <view class="section">
      <text class="section-title">系统功能</text>
      <view class="list-group">
        <view class="list-item" @click="goToWrongBook">
          <text class="item-label">错题本</text>
          <text class="item-desc">复习做错的题目</text>
          <text class="item-arrow">›</text>
        </view>
        <view class="list-item" @click="goToSmartPractice">
          <text class="item-label">智能练习</text>
          <text class="item-desc">AI 推荐个性化练习</text>
          <text class="item-arrow">›</text>
        </view>
        <view class="list-item" @click="goToPrivacy">
          <text class="item-label">隐私设置</text>
          <text class="item-desc">管理隐私偏好</text>
          <text class="item-arrow">›</text>
        </view>
        <view class="list-item" @click="goToOAuth">
          <text class="item-label">第三方账号</text>
          <text class="item-desc">绑定微信/Google等</text>
          <text class="item-arrow">›</text>
        </view>
      </view>
    </view>

    <!-- 关于 -->
    <view class="section">
      <text class="section-title">关于</text>
      <view class="list-group">
        <view class="list-item" @click="goToAbout">
          <text class="item-label">关于我们</text>
          <text class="item-desc">v1.0.0</text>
          <text class="item-arrow">›</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';

const settings = ref({
  pushNotification: true,
  practiceReminder: true,
  showAnalysis: true,
  dailyTarget: 20,
});

const cacheSize = ref('计算中...');

onMounted(() => {
  loadSettings();
  calcCacheSize();
});

function loadSettings() {
  try {
    const saved = uni.getStorageSync('appSettings');
    if (saved) {
      settings.value = { ...settings.value, ...JSON.parse(saved) };
    }
  } catch { /* ignore */ }
}

function saveSettings() {
  try {
    uni.setStorageSync('appSettings', JSON.stringify(settings.value));
  } catch { /* ignore */ }
}

function toggleSetting(key: keyof typeof settings.value, e: any) {
  (settings.value as Record<string, any>)[key] = e.detail.value;
  saveSettings();
}

function changeDailyTarget() {
  const options = [10, 20, 30, 50, 100];
  uni.showActionSheet({
    itemList: options.map(n => `${n} 题/天`),
    success: (res) => {
      settings.value.dailyTarget = options[res.tapIndex];
      saveSettings();
    },
  });
}

async function calcCacheSize() {
  try {
    const info = await new Promise<any>((resolve) => uni.getStorageInfo({ success: resolve }));
    const sizeKB = info.currentSize || 0;
    if (sizeKB < 1024) {
      cacheSize.value = `${sizeKB} KB`;
    } else {
      cacheSize.value = `${(sizeKB / 1024).toFixed(1)} MB`;
    }
  } catch {
    cacheSize.value = '--';
  }
}

function clearCache() {
  uni.showModal({
    title: '清除缓存',
    content: '将清除本地缓存数据，不影响账号信息',
    success: (res) => {
      if (res.confirm) {
        uni.clearStorageSync();
        cacheSize.value = '0 KB';
        uni.showToast({ title: '已清除', icon: 'success' });
      }
    },
  });
}

function changePassword() {
  uni.navigateTo({ url: '/pages/login/forget' });
}

function goToSecurityLogs() {
  uni.navigateTo({ url: '/pages/profile/security-logs' });
}

function goToDevices() {
  uni.navigateTo({ url: '/pages/profile/devices' });
}

function goToCategories() {
  uni.navigateTo({ url: '/pages/profile/categories' });
}

function goToWrongBook() {
  uni.navigateTo({ url: '/pages/profile/wrong-book' });
}

function goToSmartPractice() {
  uni.navigateTo({ url: '/pages/practice/smart' });
}

function goToPrivacy() {
  uni.navigateTo({ url: '/pages/profile/privacy' });
}

function goToOAuth() {
  uni.navigateTo({ url: '/pages/profile/oauth' });
}

function goToAbout() {
  uni.showModal({
    title: '关于题小助',
    content: '版本：v1.0.0\n一款跨平台的学习应用，支持题库管理、练习模式、错题本、智能练习等功能。',
    showCancel: false,
  });
}
</script>

<style lang="scss" scoped>
.settings-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 30px;

  .section {
    margin: 12px 16px;

    .section-title {
      font-size: 14px;
      color: #999;
      margin-bottom: 8px;
      margin-top: 8px;
      display: block;
    }

    .list-group {
      background: #fff;
      border-radius: 12px;
      overflow: hidden;

      .list-item {
        display: flex;
        align-items: center;
        padding: 14px 16px;
        border-bottom: 1px solid #f5f5f5;

        &:last-child { border-bottom: none; }

        .item-label {
          font-size: 15px;
          color: #333;
          flex: 1;
        }

        .item-desc {
          font-size: 12px;
          color: #999;
          margin: 0 8px;
        }

        .item-arrow {
          font-size: 18px;
          color: #ccc;
        }

        &:active {
          background: #f9f9f9;
        }

        &.switch-item {
          justify-content: space-between;
        }
      }
    }
  }

  .target-selector {
    display: flex;
    align-items: center;
    gap: 4px;

    .target-value {
      font-size: 14px;
      color: #1890ff;
    }

    .item-arrow {
      font-size: 16px;
      color: #ccc;
    }
  }
}
</style>
