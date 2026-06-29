<template>
  <view class="index-page">
    <!-- 顶部导航 -->
    <view class="header" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="header-content">
        <text class="title">题小助</text>
        <view class="notif-icon" @click="goToNotif">
          <text class="icon">🔔</text>
          <view v-if="unreadCount > 0" class="badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</view>
        </view>
      </view>
    </view>

    <!-- 搜索入口 -->
    <view class="search-bar" @click="goToSearch">
      <text class="search-icon">🔍</text>
      <text class="search-placeholder">搜索题目、知识点...</text>
    </view>

    <!-- 轮播图 -->
    <view class="carousel-section">
      <swiper
        class="carousel-swiper"
        :indicator-dots="banners.length > 1"
        indicator-color="rgba(255,255,255,0.5)"
        indicator-active-color="#fff"
        :autoplay="true"
        :interval="4000"
        :duration="500"
        :circular="true"
        @change="onBannerChange"
      >
        <swiper-item v-for="(banner, index) in banners" :key="index">
          <view class="banner-item" @click="onBannerTap(banner)">
            <image
              class="banner-image"
              :src="banner.image"
              mode="aspectFill"
              lazy-load
            />
            <view class="banner-overlay">
              <text class="banner-title">{{ banner.title }}</text>
            </view>
          </view>
        </swiper-item>
      </swiper>
      <!-- 自定义指示器 -->
      <view v-if="banners.length > 1" class="custom-indicator">
        <text class="indicator-text">{{ currentBannerIndex + 1 }}/{{ banners.length }}</text>
      </view>
    </view>

    <!-- 数据概览 -->
    <view class="stats-section">
      <view class="stat-item">
        <text class="stat-value">{{ stats.totalQuestions || '--' }}</text>
        <text class="stat-label">总题量</text>
      </view>
      <view class="stat-divider" />
      <view class="stat-item">
        <text class="stat-value">{{ stats.todayPracticeCount || 0 }}</text>
        <text class="stat-label">今日练习</text>
      </view>
      <view class="stat-divider" />
      <view class="stat-item">
        <text class="stat-value">{{ stats.todayCorrectRate ? `${(stats.todayCorrectRate * 100).toFixed(0)}%` : '--' }}</text>
        <text class="stat-label">正确率</text>
      </view>
      <view class="stat-divider" />
      <view class="stat-item">
        <text class="stat-value">{{ stats.continuousDays || 0 }}</text>
        <text class="stat-label">连续打卡</text>
      </view>
    </view>

    <!-- 快捷入口 -->
    <view class="quick-entry">
      <text class="section-title">快捷入口</text>
      <view class="entry-grid">
        <view class="entry-item" @click="goToPractice">
          <view class="entry-icon daily">📝</view>
          <text class="entry-label">每日练习</text>
        </view>
        <view class="entry-item" @click="goToWrongBook">
          <view class="entry-icon wrong">📕</view>
          <text class="entry-label">错题本</text>
        </view>
        <view class="entry-item" @click="goToSmartPractice">
          <view class="entry-icon smart">🎯</view>
          <text class="entry-label">智能练习</text>
        </view>
        <view class="entry-item" @click="goToCategories">
          <view class="entry-icon cate">🏷️</view>
          <text class="entry-label">我的分类</text>
        </view>
      </view>
      <view class="entry-grid" style="margin-top: 12px;">
        <view class="entry-item" @click="goToQuestion">
          <view class="entry-icon question">📚</view>
          <text class="entry-label">我的题库</text>
        </view>
        <view class="entry-item" @click="goToFavorites">
          <view class="entry-icon fav">⭐</view>
          <text class="entry-label">我的收藏</text>
        </view>
        <view class="entry-item" @click="goToNotif">
          <view class="entry-icon notif">💬</view>
          <text class="entry-label">消息通知</text>
        </view>
        <view class="entry-item" @click="goToSearch">
          <view class="entry-icon search">🔍</view>
          <text class="entry-label">搜索题目</text>
        </view>
      </view>
    </view>

    <!-- 推荐题目 -->
    <view class="recommend-section">
      <view class="section-header">
        <text class="section-title">推荐题目</text>
        <text class="more-link" @click="goToQuestion">更多 ›</text>
      </view>
      <view v-if="loading" class="loading-state">
        <text>加载中...</text>
      </view>
      <view v-else-if="recommended.length === 0" class="empty-state">
        <text>暂无推荐题目</text>
      </view>
      <view v-else class="question-list">
        <view
          v-for="q in recommended"
          :key="q.id"
          class="question-item"
          @click="goToDetail(q.id)"
        >
          <view class="question-header">
            <text class="question-type" :class="q.questionType">{{ getQuestionTypeLabel(q.questionType) }}</text>
            <view class="difficulty" :class="`d${q.difficulty}`">
              <text v-for="i in q.difficulty" :key="i" class="star">★</text>
            </view>
          </view>
          <text class="question-title">{{ q.title }}</text>
          <text v-if="q.categoryName" class="question-category">{{ q.categoryName }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { onShow } from '@dcloudio/uni-app';
import { request } from '@/api/request';
import { QUESTION_TYPE_LABELS, type Question } from '@/api/types';
import { getBanners, batchGetPublicURLs, type BannerItem } from '@/api/banner';

const statusBarHeight = ref(0);
const unreadCount = ref(0);
const loading = ref(false);
const currentBannerIndex = ref(0);

const banners = ref<BannerItem[]>([]);

interface HomeStats {
  totalQuestions: number;
  todayPracticeCount: number;
  todayCorrectRate: number;
  continuousDays: number;
}

const stats = ref<HomeStats>({
  totalQuestions: 0,
  todayPracticeCount: 0,
  todayCorrectRate: 0,
  continuousDays: 0,
});

interface RecommendedQuestion extends Pick<Question, 'id' | 'title' | 'questionType' | 'difficulty' | 'categoryName'> {}

const recommended = ref<RecommendedQuestion[]>([]);

// 获取系统状态栏高度（仅首次加载）
onMounted(() => {
  const systemInfo = uni.getSystemInfoSync();
  statusBarHeight.value = systemInfo.statusBarHeight || 0;
});

// 每次页面显示时刷新数据（支持从登录页跳转）
onShow(() => {
  fetchHomeData();
});

/**
 * 获取首页数据
 */
async function fetchHomeData() {
  loading.value = true;
  try {
    const data = await request.get<{
      stats: HomeStats;
      recommended: RecommendedQuestion[];
    }>('/user/home');
    stats.value = data.stats;
    recommended.value = data.recommended || [];
  } catch {
    // 使用模拟数据（仅开发环境）
    if (import.meta.env.DEV) {
      stats.value = {
        totalQuestions: 15234,
        todayPracticeCount: 45,
        todayCorrectRate: 0.78,
        continuousDays: 12,
      };
      recommended.value = [
        { id: 1, title: 'TCP三次握手的目的是什么？', questionType: 'single_choice', difficulty: 2, categoryName: '网络协议' },
        { id: 2, title: '以下哪些是HTTP请求方法？', questionType: 'multiple_choice', difficulty: 1, categoryName: 'HTTP协议' },
        { id: 3, title: '请简述进程和线程的区别', questionType: 'short_answer', difficulty: 3, categoryName: '操作系统' },
      ];
    }
  }

  // 轮播图单独请求
  try {
    const bannerList = await getBanners();

    // 批量获取轮播图的可访问URL
    const fileIds = bannerList.filter(b => b.fileId).map(b => b.fileId!);
    if (fileIds.length > 0) {
      try {
        const { urls } = await batchGetPublicURLs(fileIds);
        // 将获取到的URL设置到对应的banner
        bannerList.forEach(banner => {
          if (banner.fileId && urls[banner.fileId]) {
            banner.image = urls[banner.fileId];
          }
        });
      } catch (err) {
        console.warn('获取轮播图URL失败，使用原始URL:', err);
      }
    }

    banners.value = bannerList;
  } catch {
    // 降级模拟数据（仅开发环境）
    if (import.meta.env.DEV) {
      banners.value = [
        { id: 1, title: '📚 题库更新：新增 500 道网络协议真题', image: 'https://picsum.photos/750/300?random=1', link: '/pages/question/list', linkType: 'page' },
        { id: 2, title: '🎯 智能练习上线，根据你的薄弱点推荐题目', image: 'https://picsum.photos/750/300?random=2', link: '/pages/practice/smart', linkType: 'page' },
        { id: 3, title: '📕 错题本复习功能，艾宾浩斯记忆曲线助力', image: 'https://picsum.photos/750/300?random=3', link: '/pages/profile/wrong-book', linkType: 'page' },
      ];
    }
  }

  loading.value = false;
}

/**
 * 获取题型标签
 */
function getQuestionTypeLabel(type: string): string {
  return QUESTION_TYPE_LABELS[type as keyof typeof QUESTION_TYPE_LABELS] || type;
}

/**
 * 导航
 */
function goToSearch() {
  uni.navigateTo({ url: '/pages/question/search' });
}

function goToPractice() {
  uni.switchTab({ url: '/pages/practice/index' });
}

function goToWrongBook() {
  uni.navigateTo({ url: '/pages/profile/wrong-book' });
}

function goToSmartPractice() {
  uni.navigateTo({ url: '/pages/practice/smart' });
}

function goToCategories() {
  uni.navigateTo({ url: '/pages/profile/categories' });
}

function goToQuestion() {
  uni.switchTab({ url: '/pages/question/list' });
}

function goToFavorites() {
  uni.navigateTo({ url: '/pages/profile/favorites' });
}

function goToNotif() {
  uni.navigateTo({ url: '/pages/notification/list' });
}

function goToDetail(id: number) {
  uni.navigateTo({ url: `/pages/question/detail?id=${id}` });
}

/**
 * 轮播图事件
 */
function onBannerChange(e: any) {
  currentBannerIndex.value = e.detail.current;
}

function onBannerTap(banner: BannerItem) {
  if (banner.link) {
    if (banner.linkType === 'external') {
      uni.navigateTo({
        url: `/pages/webview/index?url=${encodeURIComponent(banner.link)}`,
      });
    } else {
      uni.navigateTo({ url: banner.link });
    }
  }
}
</script>

<style lang="scss" scoped>
.index-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 60px;

  .header {
    background: linear-gradient(135deg, #1890ff 0%, #36cfc9 100%);
    padding: 12px 16px 16px;

    .header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .title {
        font-size: 22px;
        font-weight: bold;
        color: #fff;
      }

      .notif-icon {
        position: relative;
        font-size: 22px;

        .badge {
          position: absolute;
          top: -6px;
          right: -10px;
          background: #ff4d4f;
          color: #fff;
          font-size: 10px;
          padding: 1px 5px;
          border-radius: 10px;
          min-width: 16px;
          text-align: center;
        }
      }
    }
  }

  .search-bar {
    margin: 16px;
    background: #fff;
    border-radius: 20px;
    padding: 10px 16px;
    display: flex;
    align-items: center;
    gap: 8px;

    .search-icon {
      font-size: 16px;
    }

    .search-placeholder {
      color: #999;
      font-size: 14px;
    }
  }

  .carousel-section {
    margin: 0 16px 16px;
    border-radius: 12px;
    overflow: hidden;
    position: relative;

    .carousel-swiper {
      width: 100%;
      height: 150px;
    }

    .banner-item {
      width: 100%;
      height: 100%;
      position: relative;

      .banner-image {
        width: 100%;
        height: 100%;
      }

      .banner-overlay {
        position: absolute;
        bottom: 0;
        left: 0;
        right: 0;
        padding: 12px 16px;
        background: linear-gradient(transparent, rgba(0,0,0,0.6));

        .banner-title {
          color: #fff;
          font-size: 14px;
          font-weight: 500;
        }
      }
    }

    .custom-indicator {
      position: absolute;
      bottom: 8px;
      right: 12px;
      background: rgba(0,0,0,0.3);
      border-radius: 10px;
      padding: 2px 8px;

      .indicator-text {
        color: #fff;
        font-size: 10px;
      }
    }
  }

  .stats-section {
    margin: 0 16px 16px;
    background: #fff;
    border-radius: 12px;
    padding: 20px 0;
    display: flex;
    align-items: center;
    justify-content: space-around;

    .stat-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      flex: 1;

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

  .quick-entry {
    margin: 0 16px 16px;
    background: #fff;
    border-radius: 12px;
    padding: 16px;

    .section-title {
      font-size: 16px;
      font-weight: 500;
      color: #333;
      margin-bottom: 16px;
    }

    .entry-grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 16px;

      .entry-item {
        display: flex;
        flex-direction: column;
        align-items: center;

        .entry-icon {
          width: 48px;
          height: 48px;
          border-radius: 12px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 22px;
          margin-bottom: 8px;

          &.daily { background: #e6f7ff; }
          &.wrong { background: #fff1f0; }
          &.smart { background: #f6ffed; }
          &.cate { background: #fffbe6; }
          &.question { background: #f0f5ff; }
          &.fav { background: #fffbe6; }
          &.notif { background: #fff1f0; }
          &.search { background: #e6fffb; }
        }

        .entry-label {
          font-size: 12px;
          color: #666;
        }
      }
    }
  }

  .recommend-section {
    margin: 0 16px;
    background: #fff;
    border-radius: 12px;
    padding: 16px;

    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;

      .section-title {
        font-size: 16px;
        font-weight: 500;
        color: #333;
      }

      .more-link {
        font-size: 13px;
        color: #1890ff;
      }
    }

    .loading-state, .empty-state {
      text-align: center;
      padding: 30px 0;
      color: #999;
      font-size: 14px;
    }

    .question-list {
      .question-item {
        padding: 12px 0;
        border-bottom: 1px solid #f0f0f0;

        &:last-child {
          border-bottom: none;
        }

        .question-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 6px;

          .question-type {
            font-size: 12px;
            padding: 2px 8px;
            border-radius: 4px;
            background: #e6f7ff;
            color: #1890ff;
          }

          .difficulty {
            .star {
              color: #faad14;
              font-size: 12px;
            }
          }
        }

        .question-title {
          font-size: 14px;
          color: #333;
          line-height: 1.5;
          margin-bottom: 4px;
        }

        .question-category {
          font-size: 12px;
          color: #999;
        }
      }
    }
  }
}
</style>
