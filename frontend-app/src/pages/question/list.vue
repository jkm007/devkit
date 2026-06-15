<template>
  <view class="question-list-page">
    <!-- 顶部分类区域 -->
    <view class="category-section">
      <!-- 我的分类 -->
      <view v-if="myBindings.length > 0" class="my-categories">
        <view class="section-header">
          <text class="section-title">📚 我的分类</text>
          <text class="manage-link" @click="goToCategories">管理 ></text>
        </view>
        <view class="category-chips">
          <view
            v-for="b in myBindings"
            :key="b.categoryId"
            class="category-chip"
            :class="{ active: selectedCategoryId === b.categoryId, primary: b.isPrimary }"
            @click="selectMyCategory(b)"
          >
            <text v-if="b.isPrimary" class="chip-badge">主</text>
            <text class="chip-text">{{ b.categoryName || '分类' }}</text>
          </view>
          <view class="category-chip" :class="{ active: selectedCategoryId === null }" @click="selectCategory(null)">
            <text class="chip-text">全部</text>
          </view>
        </view>
      </view>

      <!-- 全部分类（横滑） -->
      <view class="all-categories">
        <view class="section-header">
          <text class="section-title">📖 考试分类</text>
        </view>
        <scroll-view class="tabs-scroll" scroll-x enhanced :show-scrollbar="false">
          <view class="tab-list">
            <view
              v-for="cat in categories"
              :key="cat.id"
              class="tab-item"
              :class="{ active: selectedCategoryId === cat.id }"
              @click="selectCategory(cat.id)"
            >
              <text class="tab-text">{{ cat.name }}</text>
            </view>
          </view>
        </scroll-view>
      </view>
    </view>

    <!-- 筛选栏 -->
    <view class="filter-bar">
      <view class="filter-chip" :class="{ active: selectedType }" @click="showTypePicker = true">
        <text>{{ selectedTypeLabel || '题型' }}</text>
        <text class="arrow">▼</text>
      </view>
      <view class="filter-chip" :class="{ active: selectedDifficulty }" @click="showDifficultyPicker = true">
        <text>{{ selectedDifficultyLabel || '难度' }}</text>
        <text class="arrow">▼</text>
      </view>
      <view class="filter-chip search" @click="goToSearch">
        <text>🔍 搜索</text>
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
        <text class="empty-hint">{{ myBindings.length > 0 ? '试试切换分类或调整筛选条件' : '去「我的」绑定分类，获取个性化推荐' }}</text>
        <view v-if="myBindings.length === 0" class="empty-action" @click="goToCategories">
          <text>去绑定分类</text>
        </view>
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
              <text class="type-tag" :class="q.questionType">{{ getTypeLabel(q.questionType) }}</text>
            </view>
            <view class="difficulty">
              <text v-for="i in 3" :key="i" class="star" :class="{ active: i <= (q.difficulty || 1) }">★</text>
            </view>
          </view>
          <text class="title">{{ q.title }}</text>
          <view class="card-footer">
            <text v-if="q.categoryName" class="category">📂 {{ q.categoryName }}</text>
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
import { ref, onMounted } from 'vue';
import { request } from '@/api/request';
import { getCategoryBindings } from '@/api/user';
import Skeleton from '@/components/Skeleton.vue';

const loading = ref(false);
const questions = ref<any[]>([]);
const page = ref(1);
const pageSize = 20;
const hasMore = ref(true);

// 我的分类绑定
const myBindings = ref<any[]>([]);

// 全部分类
const categories = ref<{ id: number; name: string }[]>([]);
const selectedCategoryId = ref<number | null>(null);

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

const selectedTypeLabel = ref('');
const selectedDifficultyLabel = ref('');

onMounted(() => {
  loadMyBindings();
  loadCategories();
  fetchQuestions(true);
});

// 加载我的分类绑定
async function loadMyBindings() {
  try {
    myBindings.value = await getCategoryBindings();
    // 如果有主分类，默认选中
    const primary = myBindings.value.find((b: any) => b.isPrimary);
    if (primary) {
      selectedCategoryId.value = primary.categoryId;
    }
  } catch {
    myBindings.value = [];
  }
}

// 加载全部分类
async function loadCategories() {
  try {
    const res = await request.get<any[]>('/exam-categories/all');
    categories.value = Array.isArray(res) ? res.map((c: any) => ({ id: c.id, name: c.name })) : [];
  } catch {
    categories.value = [];
  }
}

// 选择我的分类
function selectMyCategory(b: any) {
  selectedCategoryId.value = b.categoryId;
  selectedType.value = '';
  selectedDifficulty.value = '';
  selectedTypeLabel.value = '';
  selectedDifficultyLabel.value = '';
  fetchQuestions(true);
}

// 选择分类
function selectCategory(catId: number | null) {
  selectedCategoryId.value = catId;
  selectedType.value = '';
  selectedDifficulty.value = '';
  selectedTypeLabel.value = '';
  selectedDifficultyLabel.value = '';
  fetchQuestions(true);
}

// 获取题目列表
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
    hasMore.value = items.length >= pageSize;
  } catch {
    uni.showToast({ title: '加载失败', icon: 'none' });
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
  const map: Record<string, string> = { '1': '简单', '2': '中等', '3': '困难' };
  selectedDifficultyLabel.value = map[tempDifficulty.value] || '';
  showDifficultyPicker.value = false;
  fetchQuestions(true);
}

function getTypeLabel(type: string | undefined): string {
  const map: Record<string, string> = {
    single_choice: '单选', multiple_choice: '多选', true_false: '判断',
    fill_blank: '填空', short_answer: '简答',
  };
  return map[type!] || type || '未知';
}

function goToSearch() { uni.navigateTo({ url: '/pages/question/search' }); }
function goToCategories() { uni.navigateTo({ url: '/pages/profile/categories' }); }
function goToDetail(id: number | undefined) {
  if (!id) return;
  uni.navigateTo({ url: `/pages/question/detail?id=${id}` });
}
</script>

<style lang="scss" scoped>
.question-list-page {
  min-height: 100vh;
  background: #f5f6fa;
  padding-bottom: 20px;
}

// ========== 分类区域 ==========
.category-section {
  background: #fff;
  padding-bottom: 12px;
  margin-bottom: 8px;
}

.my-categories {
  padding: 14px 16px 8px;

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .section-title {
    font-size: 15px;
    font-weight: 600;
    color: #333;
  }

  .manage-link {
    font-size: 13px;
    color: #1890ff;
  }

  .category-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .category-chip {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 14px;
    background: #f5f7fa;
    border-radius: 20px;
    border: 1.5px solid transparent;

    &.active {
      background: #e6f7ff;
      border-color: #1890ff;
    }

    &.primary .chip-text {
      font-weight: 500;
    }

    .chip-badge {
      font-size: 10px;
      padding: 1px 5px;
      background: #1890ff;
      color: #fff;
      border-radius: 3px;
    }

    .chip-text {
      font-size: 13px;
      color: #555;
    }

    &.active .chip-text {
      color: #1890ff;
    }
  }
}

.all-categories {
  padding: 0 16px;

  .section-header {
    margin-bottom: 8px;
  }

  .section-title {
    font-size: 15px;
    font-weight: 600;
    color: #333;
  }

  .tabs-scroll {
    width: 100%;
    white-space: nowrap;
  }

  .tab-list {
    display: inline-flex;
    gap: 8px;
    padding-right: 16px;
  }

  .tab-item {
    padding: 6px 14px;
    background: #f5f7fa;
    border-radius: 16px;

    &.active {
      background: linear-gradient(135deg, #1890ff, #36cfc9);
    }

    .tab-text {
      font-size: 13px;
      color: #666;
    }

    &.active .tab-text {
      color: #fff;
      font-weight: 500;
    }
  }
}

// ========== 筛选栏 ==========
.filter-bar {
  display: flex;
  padding: 8px 16px;
  gap: 8px;
  background: #fff;
  margin-bottom: 8px;

  .filter-chip {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 12px;
    background: #f5f7fa;
    border-radius: 16px;
    font-size: 13px;
    color: #666;

    &.active {
      background: #e6f7ff;
      color: #1890ff;
    }

    &.search {
      margin-left: auto;
      background: linear-gradient(135deg, #e6f7ff, #e6fffb);
      color: #1890ff;
    }

    .arrow {
      font-size: 10px;
      color: #999;
    }
  }
}

// ========== 列表区域 ==========
.list-section {
  padding: 0 16px;
}

.loading { padding: 20px 0; }

.empty {
  text-align: center;
  padding: 60px 0;

  .empty-icon { font-size: 48px; display: block; margin-bottom: 12px; }
  .empty-text { font-size: 16px; color: #999; display: block; margin-bottom: 4px; }
  .empty-hint { font-size: 13px; color: #bbb; display: block; margin-bottom: 16px; }
  .empty-action {
    display: inline-block;
    padding: 8px 24px;
    background: #1890ff;
    color: #fff;
    border-radius: 20px;
    font-size: 14px;
  }
}

.question-card {
  background: #fff;
  border-radius: 12px;
  padding: 14px 16px;
  margin-bottom: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;

    .tags-row { display: flex; gap: 6px; }

    .type-tag {
      font-size: 11px;
      padding: 2px 8px;
      background: #e6f7ff;
      color: #1890ff;
      border-radius: 4px;
      font-weight: 500;

      &.single_choice { background: #e6f7ff; color: #1890ff; }
      &.multiple_choice { background: #fff7e6; color: #fa8c16; }
      &.true_false { background: #f6ffed; color: #52c41a; }
      &.fill_blank { background: #f9f0ff; color: #722ed1; }
      &.short_answer { background: #fff1f0; color: #f5222d; }
    }

    .difficulty .star {
      font-size: 12px;
      color: #ddd;
      &.active { color: #faad14; }
    }
  }

  .title {
    font-size: 14px;
    color: #333;
    line-height: 1.6;
    margin-bottom: 8px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .card-footer {
    .category { font-size: 12px; color: #999; }
  }
}

.load-more, .no-more {
  text-align: center;
  padding: 16px 0;
  font-size: 13px;
  color: #999;
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

    .picker-cancel { font-size: 14px; color: #999; }
    .picker-title { font-size: 16px; font-weight: 500; color: #333; }
    .picker-confirm { font-size: 14px; color: #1890ff; font-weight: 500; }
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
      &.selected { color: #1890ff; font-weight: 500; background: #e6f7ff; }
    }
  }
}
</style>
