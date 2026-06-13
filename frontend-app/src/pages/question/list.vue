<template>
  <view class="question-list-page">
    <!-- 筛选条件 -->
    <view class="filter-bar">
      <view class="filter-item" @click="showTypePicker = true">
        <text>{{ selectedType || '题型' }}</text>
      </view>
      <view class="filter-item" @click="showDifficultyPicker = true">
        <text>{{ selectedDifficulty || '难度' }}</text>
      </view>
    </view>

    <!-- 题目列表 -->
    <view class="list-section">
      <view v-if="loading && questions.length === 0" class="loading">加载中...</view>
      <view v-else-if="questions.length === 0" class="empty">暂无题目</view>
      <view v-else>
        <view
          v-for="q in questions"
          :key="q.id"
          class="question-card"
          @click="goToDetail(q.id)"
        >
          <view class="card-header">
            <text class="type-tag">{{ getTypeLabel(q.questionType) }}</text>
            <view class="difficulty">
              <text v-for="i in q.difficulty" :key="i" class="star">★</text>
            </view>
          </view>
          <text class="title">{{ q.title }}</text>
          <view class="card-footer">
            <text v-if="q.categoryName" class="category">{{ q.categoryName }}</text>
            <text v-if="q.tags && q.tags.length" class="tags">{{ q.tags.slice(0, 2).join(' · ') }}</text>
          </view>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="hasMore" class="load-more" @click="loadMore">
        <text>{{ loading ? '加载中...' : '加载更多' }}</text>
      </view>
      <view v-else-if="questions.length > 0" class="no-more">没有更多了</view>
    </view>

    <!-- 题型选择器 -->
    <up-popup v-model:show="showTypePicker" mode="bottom">
      <view class="picker">
        <view class="picker-title">选择题型</view>
        <view class="picker-options">
          <view class="picker-option" @click="selectType('')">全部</view>
          <view class="picker-option" v-for="t in typeOptions" :key="t.value" @click="selectType(t.value)">
            {{ t.label }}
          </view>
        </view>
      </view>
    </up-popup>

    <!-- 难度选择器 -->
    <up-popup v-model:show="showDifficultyPicker" mode="bottom">
      <view class="picker">
        <view class="picker-title">选择难度</view>
        <view class="picker-options">
          <view class="picker-option" @click="selectDifficulty('')">全部</view>
          <view class="picker-option" @click="selectDifficulty(1)">简单</view>
          <view class="picker-option" @click="selectDifficulty(2)">中等</view>
          <view class="picker-option" @click="selectDifficulty(3)">困难</view>
        </view>
      </view>
    </up-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { request } from '@/api/request';
import type { Question } from '@/api/types';

const loading = ref(false);
const questions = ref<Partial<Question>[]>([]);
const page = ref(1);
const pageSize = 20;
const hasMore = ref(true);

const selectedType = ref('');
const selectedDifficulty = ref('');
const showTypePicker = ref(false);
const showDifficultyPicker = ref(false);

const typeOptions = [
  { label: '单选题', value: 'single_choice' },
  { label: '多选题', value: 'multiple_choice' },
  { label: '判断题', value: 'true_false' },
  { label: '填空题', value: 'fill_blank' },
  { label: '简答题', value: 'short_answer' },
];

onMounted(() => {
  fetchQuestions();
});

async function fetchQuestions(refresh = false) {
  if (refresh) {
    page.value = 1;
    questions.value = [];
  }
  loading.value = true;
  try {
    const params: any = { page: page.value, pageSize };
    if (selectedType.value) params.question_type = selectedType.value;
    if (selectedDifficulty.value) params.difficulty = selectedDifficulty.value;

    const data = await request.get<any>('/api/v1/study/questions', { params });
    if (refresh) questions.value = data.items;
    else questions.value = [...questions.value, ...data.items];
    hasMore.value = page.value < data.totalPages;
  } catch {
    // 模拟数据
    const mockQuestions = Array.from({ length: 10 }, (_, i) => ({
      id: (page.value - 1) * 10 + i + 1,
      title: `题目 ${(page.value - 1) * 10 + i + 1}：以下说法正确的是？`,
      questionType: ['single_choice', 'multiple_choice', 'true_false'][i % 3] as any,
      difficulty: (i % 3) + 1 as 1 | 2 | 3,
      categoryName: '计算机网络',
      tags: ['真题', '高频'].slice(0, (i % 2) + 1),
    }));
    if (refresh) questions.value = mockQuestions;
    else questions.value = [...questions.value, ...mockQuestions];
    hasMore.value = page.value < 5;
  } finally {
    loading.value = false;
  }
}

function loadMore() {
  page.value++;
  fetchQuestions();
}

function selectType(type: string) {
  selectedType.value = type;
  showTypePicker.value = false;
  fetchQuestions(true);
}

function selectDifficulty(diff: string | number) {
  selectedDifficulty.value = String(diff);
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

function goToDetail(id: number | undefined) {
  if (!id) return;
  uni.navigateTo({ url: `/pages/question/detail?id=${id}` });
}
</script>

<style lang="scss" scoped>
.question-list-page {
  min-height: 100vh;
  background: #f5f5f5;

  .filter-bar {
    display: flex;
    background: #fff;
    padding: 12px 16px;
    gap: 16px;

    .filter-item {
      padding: 6px 16px;
      background: #f5f7fa;
      border-radius: 16px;
      font-size: 14px;
      color: #666;
    }
  }

  .list-section {
    padding: 12px 16px;

    .loading, .empty {
      text-align: center;
      padding: 40px 0;
      color: #999;
    }

    .question-card {
      background: #fff;
      border-radius: 12px;
      padding: 16px;
      margin-bottom: 12px;

      .card-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;

        .type-tag {
          font-size: 12px;
          padding: 2px 8px;
          background: #e6f7ff;
          color: #1890ff;
          border-radius: 4px;
        }

        .difficulty .star {
          color: #faad14;
          font-size: 12px;
        }
      }

      .title {
        font-size: 15px;
        color: #333;
        line-height: 1.5;
        margin-bottom: 8px;
      }

      .card-footer {
        display: flex;
        gap: 12px;

        .category, .tags {
          font-size: 12px;
          color: #999;
        }
      }
    }

    .load-more, .no-more {
      text-align: center;
      padding: 16px 0;
      font-size: 14px;
      color: #999;
    }
  }

  .picker {
    padding: 20px;

    .picker-title {
      font-size: 16px;
      font-weight: 500;
      margin-bottom: 16px;
      text-align: center;
    }

    .picker-options {
      .picker-option {
        padding: 14px 0;
        text-align: center;
        font-size: 15px;
        color: #333;
        border-bottom: 1px solid #f0f0f0;

        &:last-child { border-bottom: none; }

        &:active { background: #f5f5f5; }
      }
    }
  }
}
</style>
