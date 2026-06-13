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
import { request } from '@/api/request';
import type { Question } from '@/api/types';

const statusBarHeight = ref(0);
const unreadCount = ref(0);
const loading = ref(false);

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

// 获取系统状态栏高度
onMounted(() => {
  const systemInfo = uni.getSystemInfoSync();
  statusBarHeight.value = systemInfo.statusBarHeight || 0;
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
    }>('/api/v1/user/home');
    stats.value = data.stats;
    recommended.value = data.recommended;
  } catch {
    // 使用模拟数据
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
  } finally {
    loading.value = false;
  }
}

/**
 * 获取题型标签
 */
function getQuestionTypeLabel(type: string): string {
  const map: Record<string, string> = {
    single_choice: '单选',
    multiple_choice: '多选',
    true_false: '判断',
    fill_blank: '填空',
    short_answer: '简答',
  };
  return map[type] || type;
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
          &.question { background: #f6ffed; }
          &.fav { background: #fffbe6; }
          &.notif { background: #fff1f0; }
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
