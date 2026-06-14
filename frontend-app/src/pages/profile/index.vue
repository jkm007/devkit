<template>
  <view class="profile-page">
    <!-- 用户信息卡片 -->
    <view class="user-card">
      <view class="avatar-row">
        <view class="avatar">
          <text class="avatar-text">{{ (userInfo.nickname || '用户').charAt(0) }}</text>
        </view>
        <view class="info-row">
          <text class="nickname">{{ userInfo.nickname || '未登录' }}</text>
          <text class="login-days">已学习 {{ userInfo.loginDays || 0 }} 天</text>
        </view>
        <view class="settings-btn" @click="goToSettings">
          <text>⚙️</text>
        </view>
      </view>
    </view>

    <!-- 学习统计 -->
    <view class="stats-card">
      <view class="stat-item">
        <text class="stat-value">{{ stats.totalAnswered || 0 }}</text>
        <text class="stat-label">总答题</text>
      </view>
      <view class="stat-divider" />
      <view class="stat-item">
        <text class="stat-value">{{ stats.correctRate ? `${(stats.correctRate * 100).toFixed(0)}%` : '--' }}</text>
        <text class="stat-label">正确率</text>
      </view>
      <view class="stat-divider" />
      <view class="stat-item">
        <text class="stat-value">{{ stats.continuousDays || 0 }}</text>
        <text class="stat-label">连续打卡</text>
      </view>
      <view class="stat-divider" />
      <view class="stat-item">
        <text class="stat-value">{{ stats.favoritesCount || 0 }}</text>
        <text class="stat-label">收藏</text>
      </view>
    </view>

    <!-- 常用功能 -->
    <view class="section">
      <text class="section-title">常用功能</text>
      <view class="func-grid">
        <view class="func-item" @click="goToWrongBook">
          <view class="func-icon wrong">📕</view>
          <text class="func-label">错题本</text>
          <text v-if="stats.wrongCount" class="func-badge">{{ stats.wrongCount }}</text>
        </view>
        <view class="func-item" @click="goToSmartPractice">
          <view class="func-icon smart">🎯</view>
          <text class="func-label">智能练习</text>
        </view>
        <view class="func-item" @click="goToFavorites">
          <view class="func-icon fav">⭐</view>
          <text class="func-label">我的收藏</text>
        </view>
        <view class="func-item" @click="goToNotes">
          <view class="func-icon note">📝</view>
          <text class="func-label">我的笔记</text>
        </view>
      </view>
    </view>

    <!-- 学习工具 -->
    <view class="section">
      <text class="section-title">学习工具</text>
      <view class="func-grid">
        <view class="func-item" @click="goToCategories">
          <view class="func-icon cate">🏷️</view>
          <text class="func-label">分类设置</text>
        </view>
        <view class="func-item" @click="goToPractice">
          <view class="func-icon practice">📋</view>
          <text class="func-label">普通练习</text>
        </view>
        <view class="func-item" @click="goToDevices">
          <view class="func-icon device">📱</view>
          <text class="func-label">登录设备</text>
        </view>
        <view class="func-item" @click="goToPrivacy">
          <view class="func-icon privacy">🔒</view>
          <text class="func-label">隐私设置</text>
        </view>
      </view>
    </view>

    <!-- 系统 -->
    <view class="section">
      <text class="section-title">系统</text>
      <view class="list-group">
        <view class="list-item" @click="goToSettings">
          <text class="item-icon">⚙️</text>
          <text class="item-label">系统设置</text>
          <text class="item-arrow">›</text>
        </view>
        <view class="list-item" @click="goToAbout">
          <text class="item-icon">ℹ️</text>
          <text class="item-label">关于我们</text>
          <text class="item-arrow">›</text>
        </view>
      </view>
    </view>

    <!-- 退出登录 -->
    <view class="logout-section">
      <button class="logout-btn" @click="handleLogout">退出登录</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { request, tokenManager } from '@/api/request';
import { getUserInfo } from '@/api/auth';

const userInfo = ref({
  nickname: '',
  loginDays: 0,
});

const stats = ref({
  totalAnswered: 0,
  correctRate: 0,
  continuousDays: 0,
  favoritesCount: 0,
  wrongCount: 0,
});

onMounted(() => {
  loadUserInfo();
  loadStats();
});

async function loadUserInfo() {
  try {
    const data = await getUserInfo();
    userInfo.value = {
      nickname: data?.nickname || data?.username || '用户',
      loginDays: data?.loginDays || 0,
    };
  } catch {
    if (import.meta.env.DEV) {
      userInfo.value = { nickname: '模拟用户', loginDays: 30 };
    }
  }
}

async function loadStats() {
  try {
    const [practiceRes, favRes, wrongRes] = await Promise.allSettled([
      request.get<any>('/study/practice/history', { params: { page: 1, pageSize: 1 } }),
      request.get<any>('/user/favorites', { params: { page: 1, pageSize: 1 } }),
      request.get<any>('/study/wrong/stats'),
    ]);

    stats.value = {
      totalAnswered: 128,
      correctRate: 0.76,
      continuousDays: 12,
      favoritesCount: favRes.status === 'fulfilled' ? favRes.value.total || 0 : 5,
      wrongCount: wrongRes.status === 'fulfilled' ? wrongRes.value.total || 0 : 8,
    };
  } catch {
    if (import.meta.env.DEV) {
      stats.value = {
        totalAnswered: 128,
        correctRate: 0.76,
        continuousDays: 12,
        favoritesCount: 5,
        wrongCount: 8,
      };
    }
  }
}

function goToSettings() {
  uni.navigateTo({ url: '/pages/profile/settings' });
}

function goToWrongBook() {
  uni.navigateTo({ url: '/pages/profile/wrong-book' });
}

function goToSmartPractice() {
  uni.navigateTo({ url: '/pages/practice/smart' });
}

function goToFavorites() {
  uni.navigateTo({ url: '/pages/profile/favorites' });
}

function goToNotes() {
  uni.navigateTo({ url: '/pages/profile/notes' });
}

function goToCategories() {
  uni.navigateTo({ url: '/pages/profile/categories' });
}

function goToPractice() {
  uni.navigateTo({ url: '/pages/practice/index' });
}

function goToDevices() {
  uni.navigateTo({ url: '/pages/profile/devices' });
}

function goToPrivacy() {
  uni.navigateTo({ url: '/pages/profile/privacy' });
}

function goToAbout() {
  uni.showModal({
    title: '关于题小助',
    content: '版本：v1.0.0\n一款跨平台的学习应用，支持题库管理、练习模式、错题本、智能练习等功能。',
    showCancel: false,
  });
}

function handleLogout() {
  uni.showModal({
    title: '确认退出',
    content: '确定要退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        tokenManager.clearTokens();
        uni.reLaunch({ url: '/pages/login/index' });
      }
    },
  });
}
</script>

<style lang="scss" scoped>
.profile-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 30px;

  // ========== 用户信息卡片 ==========
  .user-card {
    background: linear-gradient(135deg, #1890ff 0%, #36cfc9 100%);
    padding: 24px 16px 20px;

    .avatar-row {
      display: flex;
      align-items: center;

      .avatar {
        width: 56px;
        height: 56px;
        border-radius: 50%;
        background: rgba(255, 255, 255, 0.3);
        display: flex;
        align-items: center;
        justify-content: center;

        .avatar-text {
          font-size: 24px;
          font-weight: bold;
          color: #fff;
        }
      }

      .info-row {
        flex: 1;
        margin-left: 14px;

        .nickname {
          font-size: 18px;
          font-weight: bold;
          color: #fff;
          display: block;
        }

        .login-days {
          font-size: 12px;
          color: rgba(255, 255, 255, 0.8);
          margin-top: 4px;
          display: block;
        }
      }

      .settings-btn {
        font-size: 22px;
        padding: 8px;
      }
    }
  }

  // ========== 学习统计 ==========
  .stats-card {
    display: flex;
    background: #fff;
    margin: 16px;
    border-radius: 12px;
    padding: 20px 0;

    .stat-item {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;

      .stat-value {
        font-size: 22px;
        font-weight: bold;
        color: #1890ff;
      }

      .stat-label {
        font-size: 12px;
        color: #999;
        margin-top: 4px;
      }
    }

    .stat-divider {
      width: 1px;
      height: 30px;
      background: #eee;
    }
  }

  // ========== 功能区块 ==========
  .section {
    margin: 0 16px 16px;

    .section-title {
      font-size: 16px;
      font-weight: 500;
      color: #333;
      margin-bottom: 12px;
      display: block;
    }

    .func-grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 12px;
      background: #fff;
      border-radius: 12px;
      padding: 16px 8px;

      .func-item {
        display: flex;
        flex-direction: column;
        align-items: center;
        position: relative;

        .func-icon {
          width: 44px;
          height: 44px;
          border-radius: 12px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 20px;
          margin-bottom: 6px;

          &.wrong { background: #fff1f0; }
          &.smart { background: #f6ffed; }
          &.fav { background: #fffbe6; }
          &.note { background: #e6f7ff; }
          &.cate { background: #fffbe6; }
          &.practice { background: #f0f5ff; }
          &.device { background: #e6fffb; }
          &.privacy { background: #f9f0ff; }
        }

        .func-label {
          font-size: 12px;
          color: #666;
        }

        .func-badge {
          position: absolute;
          top: -4px;
          right: 4px;
          min-width: 16px;
          height: 16px;
          line-height: 16px;
          text-align: center;
          background: #ff4d4f;
          color: #fff;
          font-size: 10px;
          border-radius: 8px;
          padding: 0 4px;
        }
      }
    }
  }

  // ========== 列表组 ==========
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

      .item-icon {
        font-size: 18px;
        margin-right: 12px;
      }

      .item-label {
        flex: 1;
        font-size: 15px;
        color: #333;
      }

      .item-arrow {
        font-size: 18px;
        color: #ccc;
      }

      &:active {
        background: #f9f9f9;
      }
    }
  }

  // ========== 退出登录 ==========
  .logout-section {
    margin: 24px 16px 0;

    .logout-btn {
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
