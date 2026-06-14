<template>
  <view class="practice-page">
    <!-- 头部 -->
    <view class="header">
      <text class="title">开始练习</text>
      <text class="subtitle">选择适合你的练习方式</text>
    </view>

    <!-- 练习模式选择 -->
    <view class="mode-section">
      <view class="mode-cards">
        <view class="mode-card random" @click="goToSession('random')">
          <view class="mode-icon-wrap">
            <text class="mode-icon">🎲</text>
          </view>
          <text class="mode-label">随机练习</text>
          <text class="mode-desc">随机抽题，巩固基础</text>
        </view>
        <view class="mode-card wrong" @click="goToWrongPractice">
          <view class="mode-icon-wrap">
            <text class="mode-icon">📕</text>
          </view>
          <view class="mode-badge" v-if="wrongCount > 0">{{ wrongCount }}</view>
          <text class="mode-label">错题练习</text>
          <text class="mode-desc">重做错题，查漏补缺</text>
        </view>
        <view class="mode-card smart" @click="goToSmartPractice">
          <view class="mode-icon-wrap">
            <text class="mode-icon">🎯</text>
          </view>
          <text class="mode-label">智能练习</text>
          <text class="mode-desc">AI推荐，精准提升</text>
        </view>
      </view>
    </view>

    <!-- 快速开始 -->
    <view class="quick-section">
      <text class="section-title">快速开始</text>
      <view class="quick-row">
        <view class="quick-btn" @click="quickStart(10)">
          <text class="quick-num">10</text>
          <text class="quick-label">快速10题</text>
        </view>
        <view class="quick-btn" @click="quickStart(20)">
          <text class="quick-num">20</text>
          <text class="quick-label">标准20题</text>
        </view>
        <view class="quick-btn" @click="quickStart(50)">
          <text class="quick-num">50</text>
          <text class="quick-label">挑战50题</text>
        </view>
      </view>
    </view>

    <!-- 练习条件筛选 -->
    <view class="filter-section">
      <view class="section-header" @click="toggleFilter">
        <text class="section-title">自定义练习</text>
        <text class="toggle-icon">{{ showFilter ? '收起' : '展开' }} {{ showFilter ? '▲' : '▼' }}</text>
      </view>

      <view v-show="showFilter" class="filter-content">
        <!-- 分类选择 -->
        <view class="filter-item">
          <text class="label">📂 分类</text>
          <scroll-view class="category-scroll" scroll-x enhanced :show-scrollbar="false">
            <view class="category-list">
              <view class="cat-btn" :class="{ active: selectedCategory === cat.id }" v-for="cat in categories" :key="cat.id" @click="selectCategory(cat.id)">
                {{ cat.name }}
              </view>
            </view>
          </scroll-view>
        </view>

        <!-- 难度 -->
        <view class="filter-item">
          <text class="label">⭐ 难度</text>
          <view class="pill-group">
            <view class="pill" :class="{ active: difficulty === 1 }" @click="difficulty = difficulty === 1 ? 0 : 1">简单</view>
            <view class="pill" :class="{ active: difficulty === 2 }" @click="difficulty = difficulty === 2 ? 0 : 2">中等</view>
            <view class="pill" :class="{ active: difficulty === 3 }" @click="difficulty = difficulty === 3 ? 0 : 3">困难</view>
          </view>
        </view>

        <!-- 题型 -->
        <view class="filter-item">
          <text class="label">📋 题型</text>
          <view class="pill-group">
            <view class="pill" v-for="t in typeOptions" :key="t.value" :class="{ active: selectedTypes.includes(t.value) }" @click="toggleType(t.value)">
              {{ t.label }}
            </view>
          </view>
        </view>

        <!-- 题目数量 -->
        <view class="filter-item">
          <text class="label">🔢 题目数量</text>
          <view class="count-selector">
            <view class="count-btn" @click="adjustCount(-5)">−</view>
            <text class="count-value">{{ questionCount }}</text>
            <view class="count-btn" @click="adjustCount(5)">+</view>
          </view>
        </view>

        <!-- 开始按钮 -->
        <button class="custom-start-btn" @click="startCustomPractice">开始自定义练习</button>
      </view>
    </view>

    <!-- 练习历史 -->
    <view class="history-section" v-if="historyList.length > 0">
      <text class="section-title">最近练习</text>
      <view class="history-list">
        <view class="history-item" v-for="(h, i) in historyList" :key="i">
          <view class="history-left">
            <text class="history-mode">{{ h.mode }}</text>
            <text class="history-time">{{ h.time }}</text>
          </view>
          <view class="history-right">
            <text class="history-score" :class="h.rate >= 0.6 ? 'good' : 'bad'">{{ (h.rate * 100).toFixed(0) }}%</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { request } from '@/api/request';

const wrongCount = ref(0);
const questionCount = ref(20);
const difficulty = ref(0);
const selectedTypes = ref<string[]>([]);
const selectedCategory = ref<number | null>(null);
const showFilter = ref(false);

const categories = ref<{ id: number | null; name: string }[]>([
  { id: null, name: '全部分类' },
]);

const typeOptions = [
  { label: '单选', value: 'single_choice' },
  { label: '多选', value: 'multiple_choice' },
  { label: '判断', value: 'true_false' },
  { label: '填空', value: 'fill_blank' },
  { label: '简答', value: 'short_answer' },
];

const historyList = ref<any[]>([]);

onMounted(() => {
  loadWrongCount();
  loadCategories();
  loadHistory();
});

async function loadWrongCount() {
  try {
    const res = await request.get<any>('/study/wrong/stats');
    wrongCount.value = res.total || 0;
  } catch {
    wrongCount.value = 8; // mock
  }
}

async function loadCategories() {
  try {
    const res = await request.get<any[]>('/exam-categories/all');
    categories.value = [{ id: null, name: '全部分类' }, ...res.map((c: any) => ({ id: c.id, name: c.name }))];
  } catch {
    categories.value = [
      { id: null, name: '全部分类' },
      { id: 1, name: '网络协议' },
      { id: 2, name: '操作系统' },
      { id: 3, name: '数据结构' },
    ];
  }
}

async function loadHistory() {
  try {
    const res = await request.get<any>('/study/practice/history', { params: { page: 1, pageSize: 5 } });
    historyList.value = (res.items || []).map((r: any) => ({
      mode: r.mode === 'random' ? '随机练习' : r.mode === 'wrong' ? '错题练习' : '智能练习',
      time: formatDate(r.createdAt),
      rate: r.total > 0 ? r.correct / r.total : 0,
    }));
  } catch {
    historyList.value = [
      { mode: '随机练习', time: '今天 14:30', rate: 0.75 },
      { mode: '错题练习', time: '昨天 09:15', rate: 0.60 },
    ];
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  const dayDiff = Math.floor(diff / 86400000);
  if (dayDiff === 0) return `今天 ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
  if (dayDiff === 1) return `昨天 ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
  return `${d.getMonth() + 1}/${d.getDate()}`;
}

function toggleFilter() {
  showFilter.value = !showFilter.value;
}

function selectCategory(id: number | null) {
  selectedCategory.value = id;
}

function adjustCount(delta: number) {
  questionCount.value = Math.max(5, Math.min(100, questionCount.value + delta));
}

function toggleType(type: string) {
  const idx = selectedTypes.value.indexOf(type);
  if (idx >= 0) selectedTypes.value.splice(idx, 1);
  else selectedTypes.value.push(type);
}

function quickStart(count: number) {
  const params = { mode: 'random', count, difficulty: 0, types: [] };
  navigateToSession(params);
}

function goToSession(mode: string) {
  const params = { mode, count: questionCount.value, difficulty: difficulty.value, types: selectedTypes.value };
  navigateToSession(params);
}

function goToWrongPractice() {
  if (wrongCount.value === 0) {
    uni.showToast({ title: '暂无错题，继续保持！', icon: 'none' });
    return;
  }
  uni.navigateTo({ url: '/pages/profile/wrong-book' });
}

function goToSmartPractice() {
  uni.navigateTo({ url: '/pages/practice/smart' });
}

function startCustomPractice() {
  const params = {
    mode: 'custom',
    count: questionCount.value,
    difficulty: difficulty.value,
    types: selectedTypes.value,
    categoryId: selectedCategory.value,
  };
  navigateToSession(params);
}

function navigateToSession(params: any) {
  uni.setStorageSync('practiceParams', JSON.stringify(params));
  uni.navigateTo({ url: '/pages/practice/session' });
}
</script>

<style lang="scss" scoped>
.practice-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 30px;

  // ========== 头部 ==========
  .header {
    padding: 24px 16px 16px;
    background: #fff;

    .title {
      font-size: 22px;
      font-weight: bold;
      color: #333;
      display: block;
    }

    .subtitle {
      font-size: 13px;
      color: #999;
      margin-top: 4px;
      display: block;
    }
  }

  // ========== 练习模式卡片 ==========
  .mode-section {
    padding: 16px;

    .mode-cards {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 12px;
    }

    .mode-card {
      background: #fff;
      border-radius: 12px;
      padding: 16px 10px;
      text-align: center;
      position: relative;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);

      &.random .mode-icon-wrap { background: linear-gradient(135deg, #e6f7ff, #bae7ff); }
      &.wrong .mode-icon-wrap { background: linear-gradient(135deg, #fff1f0, #ffccc7); }
      &.smart .mode-icon-wrap { background: linear-gradient(135deg, #f6ffed, #b7eb8f); }

      .mode-icon-wrap {
        width: 48px;
        height: 48px;
        border-radius: 14px;
        display: flex;
        align-items: center;
        justify-content: center;
        margin: 0 auto 10px;
      }

      .mode-icon {
        font-size: 24px;
      }

      .mode-badge {
        position: absolute;
        top: 8px;
        right: 8px;
        min-width: 18px;
        height: 18px;
        line-height: 18px;
        text-align: center;
        background: #ff4d4f;
        color: #fff;
        font-size: 10px;
        border-radius: 9px;
        padding: 0 4px;
      }

      .mode-label {
        display: block;
        font-size: 14px;
        font-weight: 500;
        color: #333;
        margin-bottom: 4px;
      }

      .mode-desc {
        display: block;
        font-size: 11px;
        color: #999;
      }

      &:active {
        transform: scale(0.97);
      }
    }
  }

  // ========== 快速开始 ==========
  .quick-section {
    margin: 0 16px 16px;
    background: #fff;
    border-radius: 12px;
    padding: 16px;

    .section-title {
      font-size: 14px;
      color: #666;
      margin-bottom: 12px;
      display: block;
    }

    .quick-row {
      display: flex;
      gap: 12px;

      .quick-btn {
        flex: 1;
        background: linear-gradient(135deg, #1890ff, #36cfc9);
        border-radius: 12px;
        padding: 12px 8px;
        text-align: center;

        .quick-num {
          display: block;
          font-size: 20px;
          font-weight: bold;
          color: #fff;
        }

        .quick-label {
          display: block;
          font-size: 11px;
          color: rgba(255, 255, 255, 0.85);
          margin-top: 2px;
        }

        &:active {
          transform: scale(0.96);
        }
      }
    }
  }

  // ========== 自定义练习 ==========
  .filter-section {
    margin: 0 16px 16px;
    background: #fff;
    border-radius: 12px;
    overflow: hidden;

    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 14px 16px;

      .section-title {
        font-size: 15px;
        font-weight: 500;
        color: #333;
      }

      .toggle-icon {
        font-size: 12px;
        color: #999;
      }
    }

    .filter-content {
      padding: 0 16px 16px;

      .filter-item {
        margin-bottom: 16px;

        .label {
          font-size: 13px;
          color: #666;
          margin-bottom: 8px;
          display: block;
        }
      }

      .category-scroll {
        width: 100%;
        white-space: nowrap;

        .category-list {
          display: inline-flex;
          gap: 8px;

          .cat-btn {
            padding: 6px 14px;
            background: #f5f7fa;
            border-radius: 16px;
            font-size: 13px;
            color: #666;

            &.active {
              background: #1890ff;
              color: #fff;
            }
          }
        }
      }

      .pill-group {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;

        .pill {
          padding: 6px 14px;
          background: #f5f7fa;
          border-radius: 16px;
          font-size: 13px;
          color: #666;

          &.active {
            background: #1890ff;
            color: #fff;
          }
        }
      }

      .count-selector {
        display: flex;
        align-items: center;
        gap: 16px;

        .count-btn {
          width: 32px;
          height: 32px;
          background: #f5f7fa;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 16px;
        }

        .count-value {
          font-size: 18px;
          font-weight: bold;
          min-width: 36px;
          text-align: center;
        }
      }

      .custom-start-btn {
        margin-top: 8px;
        height: 44px;
        line-height: 44px;
        background: linear-gradient(90deg, #1890ff, #36cfc9);
        color: #fff;
        border: none;
        border-radius: 22px;
        font-size: 15px;
      }
    }
  }

  // ========== 练习历史 ==========
  .history-section {
    margin: 0 16px;

    .section-title {
      font-size: 14px;
      color: #666;
      margin-bottom: 12px;
      display: block;
    }

    .history-list {
      .history-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 12px 16px;
        background: #fff;
        border-radius: 8px;
        margin-bottom: 8px;

        .history-left {
          .history-mode {
            display: block;
            font-size: 14px;
            color: #333;
            font-weight: 500;
          }

          .history-time {
            display: block;
            font-size: 12px;
            color: #999;
            margin-top: 2px;
          }
        }

        .history-right {
          .history-score {
            font-size: 18px;
            font-weight: bold;

            &.good { color: #52c41a; }
            &.bad { color: #ff4d4f; }
          }
        }
      }
    }
  }
}
</style>
