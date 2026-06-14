<template>
  <view class="question-list-page">
    <!-- 分类 Tab 横滑 -->
    <view class="category-tabs">
      <scroll-view class="tabs-scroll" scroll-x :scroll-left="scrollLeft" enhanced :show-scrollbar="false">
        <view class="tab-item" :class="{ active: selectedCategoryId === cat.id }" v-for="cat in categories" :key="cat.id" @click="selectCategory(cat)" :ref="el => setTabRef(el, cat.id)">
          <text class="tab-text">{{ cat.name }}</text>
          <view v-if="selectedCategoryId === cat.id" class="tab-underline" />
        </view>
      </scroll-view>
    </view>

    <!-- 筛选栏：题型 + 难度 + 搜索 -->
    <view class="filter-bar">
      <view class="filter-item" @click="showTypePicker = true">
        <text class="filter-icon">📋</text>
        <text>{{ selectedTypeLabel || '题型' }}</text>
      </view>
      <view class="filter-item" @click="showDifficultyPicker = true">
        <text class="filter-icon">⭐</text>
        <text>{{ selectedDifficultyLabel || '难度' }}</text>
      </view>
      <view class="filter-item search-item" @click="goToSearch">
        <text class="filter-icon">🔍</text>
        <text>搜索</text>
      </view>
    </view>

    <!-- 题目列表 -->
    <view class="list-section">
      <view v-if="loading && questions.length === 0" class="loading">
        <Skeleton type="list" :count="5" />
      </view>
      <view v-else-if="questions.length === 0" class="empty">
        <text class="empty-icon">📭</text>
        <text class="empty-text">暂无题目</text>
        <text class="empty-hint">试试切换分类或调整筛选条件</text>
      </view>
      <view v-else>
        <view
          v-for="q in questions"
          :key="q.id"
          class="question-card"
          @click="goToDetail(q.id)"
        >
          <view class="card-header">
            <view class="tags-row">
              <text class="type-tag">{{ getTypeLabel(q.questionType) }}</text>
              <text v-if="q.isNew" class="new-tag">新</text>
              <text v-if="q.hot" class="hot-tag">🔥</text>
            </view>
            <view class="difficulty">
              <text v-for="i in q.difficulty" :key="i" class="star">★</text>
            </view>
          </view>
          <text class="title">{{ q.title }}</text>
          <view class="card-footer">
            <text v-if="q.categoryName" class="category">📂 {{ q.categoryName }}</text>
            <text v-if="q.knowledgePoint" class="knowledge">💡 {{ q.knowledgePoint }}</text>
          </view>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="hasMore" class="load-more" @click="loadMore">
        <text>{{ loading ? '加载中...' : '加载更多' }}</text>
      </view>
      <view v-else-if="questions.length > 0" class="no-more">— 没有更多了 —</view>
    </view>

    <!-- 题型选择器 -->
    <up-popup v-model:show="showTypePicker" mode="bottom" round="12">
      <view class="picker">
        <view class="picker-header">
          <text class="picker-cancel" @click="showTypePicker = false">取消</text>
          <text class="picker-title">选择题型</text>
          <text class="picker-confirm" @click="confirmType">确定</text>
        </view>
        <view class="picker-options">
          <view class="picker-option" :class="{ selected: tempType === '' }" @click="tempType = ''">全部题型</view>
          <view class="picker-option" v-for="t in typeOptions" :key="t.value" :class="{ selected: tempType === t.value }" @click="tempType = t.value">
            {{ t.label }}
          </view>
        </view>
      </view>
    </up-popup>

    <!-- 难度选择器 -->
    <up-popup v-model:show="showDifficultyPicker" mode="bottom" round="12">
      <view class="picker">
        <view class="picker-header">
          <text class="picker-cancel" @click="showDifficultyPicker = false">取消</text>
          <text class="picker-title">选择难度</text>
          <text class="picker-confirm" @click="confirmDifficulty">确定</text>
        </view>
        <view class="picker-options">
          <view class="picker-option" :class="{ selected: tempDifficulty === '' }" @click="tempDifficulty = ''">全部难度</view>
          <view class="picker-option" :class="{ selected: tempDifficulty === '1' }" @click="tempDifficulty = '1'">⭐ 简单</view>
          <view class="picker-option" :class="{ selected: tempDifficulty === '2' }" @click="tempDifficulty = '2'">⭐⭐ 中等</view>
          <view class="picker-option" :class="{ selected: tempDifficulty === '3' }" @click="tempDifficulty = '3'">⭐⭐⭐ 困难</view>
        </view>
      </view>
    </up-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue';
import { request } from '@/api/request';
import type { Question } from '@/api/types';
import Skeleton from '@/components/Skeleton.vue';

const loading = ref(false);
const questions = ref<Partial<Question>[]>([]);
const page = ref(1);
const pageSize = 20;
const hasMore = ref(true);

// 分类相关
const categories = ref<{ id: number | null; name: string }[]>([
  { id: null, name: '推荐' },
]);
const selectedCategoryId = ref<number | null>(null);
const activeTabId = ref<number | null>(null);
const scrollLeft = ref(0);
const tabRefs = ref<Map<number | null, any>>(new Map());

// 筛选
const selectedType = ref('');
const selectedDifficulty = ref('');
const tempType = ref('');
const tempDifficulty = ref('');
const showTypePicker = ref(false);
const showDifficultyPicker = ref(false);

const typeOptions = [
  { label: '单选题', value: 'single_choice' },
  { label: '多选题', value: 'multiple_choice' },
  { label: '判断题', value: 'true_false' },
  { label: '填空题', value: 'fill_blank' },
  { label: '简答题', value: 'short_answer' },
];

// 计算显示文字
const selectedTypeLabel = ref('');
const selectedDifficultyLabel = ref('');

onMounted(() => {
  loadCategories();
  fetchQuestions();
});

/**
 * 加载分类列表
 */
async function loadCategories() {
  try {
    const res = await request.get<any[]>('/exam-categories/all');
    categories.value = [{ id: null, name: '推荐' }, ...res.map((c: any) => ({ id: c.id, name: c.name }))];
  } catch {
    // Mock 数据（仅开发环境）
    if (import.meta.env.DEV) {
      categories.value = [
        { id: null, name: '推荐' },
        { id: 1, name: '网络协议' },
        { id: 2, name: '操作系统' },
        { id: 3, name: '数据结构' },
        { id: 4, name: '数据库' },
        { id: 5, name: '算法' },
        { id: 6, name: '计算机网络' },
        { id: 7, name: '计算机组成原理' },
      ];
    }
  }
  // 默认选中"推荐"
  selectedCategoryId.value = null;
}

/**
 * 设置 Tab ref 用于滚动定位
 */
function setTabRef(el: any, catId: number | null) {
  if (el) tabRefs.value.set(catId, el);
}

/**
 * 选择分类
 */
async function selectCategory(cat: { id: number | null; name: string }) {
  selectedCategoryId.value = cat.id;
  activeTabId.value = cat.id;

  // 滚动到选中 tab
  await nextTick();
  const tabEl = tabRefs.value.get(cat.id);
  if (tabEl) {
    // scroll-view 会自动滚动到可见区域
  }

  // 重置筛选 + 刷新
  selectedType.value = '';
  selectedDifficulty.value = '';
  selectedTypeLabel.value = '';
  selectedDifficultyLabel.value = '';
  fetchQuestions(true);
}

/**
 * 获取题目列表
 */
async function fetchQuestions(refresh = false) {
  if (refresh) {
    page.value = 1;
    questions.value = [];
  }
  loading.value = true;
  try {
    const params: any = { page: page.value, pageSize };
    if (selectedCategoryId.value) params.categoryId = selectedCategoryId.value;
    if (selectedType.value) params.questionType = selectedType.value;
    if (selectedDifficulty.value) params.difficulty = Number(selectedDifficulty.value);

    const data = await request.get<any>('/study/questions', { params });
    const items = data.items || [];
    if (refresh) questions.value = items;
    else questions.value = [...questions.value, ...items];
    hasMore.value = page.value < (data.totalPages || 0);
  } catch {
    // 模拟数据（仅开发环境）
    if (import.meta.env.DEV) {
      const currentCatName = categories.value.find(c => c.id === selectedCategoryId.value)?.name || '推荐';
      const mockQuestions = Array.from({ length: 10 }, (_, i) => ({
        id: (page.value - 1) * 10 + i + 1,
        title: `题目 ${(page.value - 1) * 10 + i + 1}：关于${currentCatName}，以下说法正确的是？`,
        questionType: ['single_choice', 'multiple_choice', 'true_false', 'fill_blank', 'short_answer'][i % 5] as any,
        difficulty: (i % 3) + 1 as 1 | 2 | 3,
        categoryName: currentCatName,
        knowledgePoint: ['TCP/IP', 'HTTP', '进程调度', '内存管理', '二叉树'][i % 5],
        isNew: i === 0 && page.value === 1,
        hot: i % 4 === 0,
      }));
      if (refresh) questions.value = mockQuestions;
      else questions.value = [...questions.value, ...mockQuestions];
      hasMore.value = page.value < 5;
    } else {
      uni.showToast({ title: '加载题目列表失败', icon: 'none' });
    }
  } finally {
    loading.value = false;
  }
}

function loadMore() {
  page.value++;
  fetchQuestions();
}

function confirmType() {
  selectedType.value = tempType.value;
  selectedTypeLabel.value = tempType.value ? typeOptions.find(t => t.value === tempType.value)?.label || '' : '';
  showTypePicker.value = false;
  fetchQuestions(true);
}

function confirmDifficulty() {
  selectedDifficulty.value = tempDifficulty.value;
  const map: Record<string, string> = { '1': '⭐ 简单', '2': '⭐⭐ 中等', '3': '⭐⭐⭐ 困难' };
  selectedDifficultyLabel.value = map[tempDifficulty.value] || '';
  showDifficultyPicker.value = false;
  fetchQuestions(true);
}

function getTypeLabel(type: string | undefined): string {
  const map: Record<string, string> = {
    single_choice: '单选',
    multiple_choice: '多选',
    true_false: '判断',
    fill_blank: '填空',
    short_answer: '简答',
  };
  return map[type!] || type || '未知';
}

function goToSearch() {
  uni.navigateTo({ url: '/pages/question/search' });
}

function goToDetail(id: number | undefined) {
  if (!id) return;
  uni.navigateTo({ url: `/pages/question/detail?id=${id}` });
}
</script>

<style lang="scss" scoped>
.question-list-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 20px;

  // ========== 分类 Tab ==========
  .category-tabs {
    background: #fff;
    padding: 0 8px;
    border-bottom: 1px solid #f0f0f0;

    .tabs-scroll {
      width: 100%;
      white-space: nowrap;
    }

    .tab-item {
      display: inline-flex;
      flex-direction: column;
      align-items: center;
      padding: 12px 16px;
      position: relative;

      .tab-text {
        font-size: 14px;
        color: #666;
        transition: color 0.2s;
      }

      &.active .tab-text {
        color: #1890ff;
        font-weight: 500;
      }

      .tab-underline {
        position: absolute;
        bottom: 0;
        width: 20px;
        height: 3px;
        background: #1890ff;
        border-radius: 2px;
      }
    }
  }

  // ========== 筛选栏 ==========
  .filter-bar {
    display: flex;
    background: #fff;
    padding: 10px 16px;
    gap: 8px;
    margin-bottom: 8px;

    .filter-item {
      display: flex;
      align-items: center;
      gap: 4px;
      padding: 6px 14px;
      background: #f5f7fa;
      border-radius: 16px;
      font-size: 13px;
      color: #666;

      .filter-icon {
        font-size: 14px;
      }

      &.search-item {
        margin-left: auto;
        background: linear-gradient(135deg, #e6f7ff, #e6fffb);
        color: #1890ff;
      }
    }
  }

  // ========== 列表区域 ==========
  .list-section {
    padding: 0 16px;

    .loading {
      padding: 20px 0;
    }

    .empty {
      text-align: center;
      padding: 60px 0;

      .empty-icon {
        font-size: 48px;
        display: block;
        margin-bottom: 12px;
      }

      .empty-text {
        font-size: 16px;
        color: #999;
        display: block;
        margin-bottom: 4px;
      }

      .empty-hint {
        font-size: 12px;
        color: #ccc;
        display: block;
      }
    }

    // ========== 题目卡片 ==========
    .question-card {
      background: #fff;
      border-radius: 12px;
      padding: 16px;
      margin-bottom: 12px;
      box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);

      .card-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;

        .tags-row {
          display: flex;
          gap: 6px;
          align-items: center;

          .type-tag {
            font-size: 12px;
            padding: 2px 8px;
            background: #e6f7ff;
            color: #1890ff;
            border-radius: 4px;
          }

          .new-tag {
            font-size: 11px;
            padding: 2px 6px;
            background: #f6ffed;
            color: #52c41a;
            border-radius: 4px;
            font-weight: 500;
          }

          .hot-tag {
            font-size: 12px;
          }
        }

        .difficulty .star {
          color: #faad14;
          font-size: 12px;
        }
      }

      .title {
        font-size: 15px;
        color: #333;
        line-height: 1.6;
        margin-bottom: 10px;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
      }

      .card-footer {
        display: flex;
        gap: 12px;
        flex-wrap: wrap;

        .category, .knowledge {
          font-size: 12px;
          color: #999;
        }
      }
    }

    .load-more, .no-more {
      text-align: center;
      padding: 16px 0;
      font-size: 13px;
      color: #999;
    }

    .no-more {
      color: #ccc;
    }
  }

  // ========== 弹窗选择器 ==========
  .picker {
    padding-bottom: env(safe-area-inset-bottom);

    .picker-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 16px;
      border-bottom: 1px solid #f0f0f0;

      .picker-cancel {
        font-size: 14px;
        color: #999;
      }

      .picker-title {
        font-size: 16px;
        font-weight: 500;
        color: #333;
      }

      .picker-confirm {
        font-size: 14px;
        color: #1890ff;
        font-weight: 500;
      }
    }

    .picker-options {
      max-height: 300px;
      overflow-y: auto;

      .picker-option {
        padding: 14px 20px;
        text-align: center;
        font-size: 15px;
        color: #333;
        border-bottom: 1px solid #f5f5f5;

        &:last-child { border-bottom: none; }

        &:active { background: #f5f5f5; }

        &.selected {
          color: #1890ff;
          font-weight: 500;
          background: #e6f7ff;
        }
      }
    }
  }
}
</style>
