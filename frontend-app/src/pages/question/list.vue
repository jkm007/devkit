<template>
  <view class="question-list-page">
    <!-- 搜索栏 -->
    <view class="search-bar">
      <view class="search-input-wrap">
        <uni-icons type="search" size="18" color="#999" />
        <input v-model="keyword" class="search-input" placeholder="搜索题目" confirm-type="search" @confirm="onSearch" />
        <uni-icons v-if="keyword" type="clear" size="18" color="#ccc" @click="clearSearch" />
      </view>
    </view>

    <!-- 我的分类 -->
    <view class="section" v-if="myBindings.length > 0">
      <view class="section-header">
        <text class="section-title">我的分类</text>
        <text class="section-link" @click="goToCategoryManage">管理</text>
      </view>
      <scroll-view scroll-x class="category-scroll">
        <view class="category-chips">
          <view
            v-for="b in myBindings"
            :key="b.id"
            class="chip"
            :class="{ active: selectedSubjectId === b.subjectId }"
            @click="selectMyBinding(b)"
          >
            <text class="primary-tag" v-if="b.isPrimary">主</text>
            <text class="chip-text">{{ b.subjectName }}</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- 考试分类 -->
    <view class="section">
      <view class="section-header">
        <text class="section-title">考试分类</text>
      </view>
      <!-- L1 考试大类 tabs -->
      <scroll-view scroll-x class="category-scroll">
        <view class="category-tabs">
          <view
            class="tab"
            :class="{ active: selectedExamCategoryId === null }"
            @click="selectExamCategory(null)"
          >
            <text>全部</text>
          </view>
          <view
            v-for="ec in categoryTree"
            :key="ec.id"
            class="tab"
            :class="{ active: selectedExamCategoryId === ec.id }"
            @click="selectExamCategory(ec)"
          >
            <text>{{ ec.name }}</text>
          </view>
        </view>
      </scroll-view>

      <!-- L2 考试 tabs (如果有选中的 L1) -->
      <scroll-view v-if="currentExams.length > 0" scroll-x class="category-scroll sub">
        <view class="category-tabs">
          <view
            class="tab sub"
            :class="{ active: selectedExamId === null }"
            @click="selectExam(null)"
          >
            <text>全部</text>
          </view>
          <view
            v-for="exam in currentExams"
            :key="exam.id"
            class="tab sub"
            :class="{ active: selectedExamId === exam.id }"
            @click="selectExam(exam)"
          >
            <text>{{ exam.name }}</text>
          </view>
        </view>
      </scroll-view>

      <!-- L3 科目 pills -->
      <scroll-view v-if="currentSubjects.length > 0" scroll-x class="category-scroll pills">
        <view class="category-pills">
          <view
            v-for="subject in currentSubjects"
            :key="subject.id"
            class="pill"
            :class="{ active: selectedSubjectId === subject.id }"
            @click="selectSubject(subject)"
          >
            <text>{{ subject.name }}</text>
            <text class="count" v-if="subject.questionCount">({{ subject.questionCount }})</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- 题目列表 -->
    <view class="question-list">
      <view v-if="loading && questions.length === 0" class="loading">
        <text>加载中...</text>
      </view>
      <view v-else-if="questions.length === 0" class="empty">
        <text class="empty-icon">📝</text>
        <text class="empty-text">暂无题目</text>
      </view>
      <view v-else>
        <view
          v-for="q in questions"
          :key="q.id"
          class="question-card"
          @click="goToDetail(q)"
        >
          <view class="card-header">
            <text class="type-badge" :class="q.questionType">{{ getTypeName(q.questionType) }}</text>
            <text class="diff-badge" :class="'diff-' + q.difficulty">{{ getDiffName(q.difficulty) }}</text>
          </view>
          <text class="question-title">{{ q.title }}</text>
          <view class="card-footer">
            <text class="category-name">{{ q.categoryName }}</text>
          </view>
        </view>
        <view v-if="loading" class="loading-more">
          <text>加载中...</text>
        </view>
        <view v-if="!hasMore && questions.length > 0" class="no-more">
          <text>没有更多了</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { onLoad, onShow, onReachBottom } from '@dcloudio/uni-app';
import { getQuestions } from '@/api/study';
import { getCategoryBindings, getCategoryTree } from '@/api/user';
import type { Question } from '@/api/types';

// 分类树数据结构
interface Subject {
  id: number;
  name: string;
  questionCount?: number;
}
interface Exam {
  id: number;
  name: string;
  subjects: Subject[];
}
interface ExamCategory {
  id: number;
  name: string;
  exams: Exam[];
}

const questions = ref<Question[]>([]);
const loading = ref(false);
const page = ref(1);
const hasMore = ref(true);
const keyword = ref('');
const categoryId = ref<number | undefined>(undefined);
const selectedSubjectId = ref<number | undefined>(undefined);
const pageSize = 20;

// 分类数据
const categoryTree = ref<ExamCategory[]>([]);
const myBindings = ref<any[]>([]);

// 当前选中的 L1 和 L2
const selectedExamCategoryId = ref<number | null>(null);
const selectedExamId = ref<number | null>(null);

// 当前 L2 考试列表
const currentExams = computed(() => {
  if (selectedExamCategoryId.value === null) return [];
  const ec = categoryTree.value.find(e => e.id === selectedExamCategoryId.value);
  return ec?.exams || [];
});

// 当前 L3 科目列表
const currentSubjects = computed(() => {
  if (selectedExamId.value === null) {
    // 如果只选了 L1，显示该 L1 下所有科目的合并
    const ec = categoryTree.value.find(e => e.id === selectedExamCategoryId.value);
    if (!ec) return [];
    const subjects: Subject[] = [];
    for (const exam of ec.exams) {
      for (const s of exam.subjects) {
        if (!subjects.find(x => x.id === s.id)) {
          subjects.push(s);
        }
      }
    }
    return subjects;
  }
  const exam = currentExams.value.find(e => e.id === selectedExamId.value);
  return exam?.subjects || [];
});

onLoad(() => {
  loadCategoryTree();
  loadMyBindings();
});

onShow(() => {
  loadMyBindings();
});

onReachBottom(() => {
  if (hasMore.value && !loading.value) {
    page.value++;
    fetchQuestions(false);
  }
});

async function loadCategoryTree() {
  try {
    const res = await getCategoryTree();
    categoryTree.value = Array.isArray(res) ? res : [];
  } catch {
    categoryTree.value = [];
  }
}

async function loadMyBindings() {
  try {
    const res = await getCategoryBindings();
    myBindings.value = Array.isArray(res) ? res : [];
  } catch {
    myBindings.value = [];
  }
}

async function fetchQuestions(reset = true) {
  if (reset) {
    page.value = 1;
    hasMore.value = true;
  }
  loading.value = true;
  try {
    const params: any = { page: page.value, pageSize };
    if (selectedSubjectId.value) {
      params.subjectId = selectedSubjectId.value;
    } else if (categoryId.value) {
      params.categoryId = categoryId.value;
    }
    if (keyword.value) {
      params.keyword = keyword.value;
    }
    const res = await getQuestions(params);
    if (res?.items) {
      questions.value = reset ? res.items : [...questions.value, ...res.items];
      hasMore.value = questions.value.length < res.total;
    } else {
      if (reset) questions.value = [];
      hasMore.value = false;
    }
  } catch {
    if (reset) questions.value = [];
    hasMore.value = false;
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  fetchQuestions(true);
}

function clearSearch() {
  keyword.value = '';
  fetchQuestions(true);
}

// 选择我的分类
function selectMyBinding(b: any) {
  if (selectedSubjectId.value === b.subjectId) {
    selectedSubjectId.value = undefined;
  } else {
    selectedSubjectId.value = b.subjectId;
    selectedExamCategoryId.value = null;
    selectedExamId.value = null;
  }
  fetchQuestions(true);
}

// 选择 L1 考试大类
function selectExamCategory(ec: ExamCategory | null) {
  if (ec === null) {
    selectedExamCategoryId.value = null;
    selectedExamId.value = null;
    selectedSubjectId.value = undefined;
  } else {
    selectedExamCategoryId.value = ec.id;
    selectedExamId.value = null;
    selectedSubjectId.value = undefined;
  }
  fetchQuestions(true);
}

// 选择 L2 考试
function selectExam(exam: Exam | null) {
  if (exam === null) {
    selectedExamId.value = null;
    selectedSubjectId.value = undefined;
  } else {
    selectedExamId.value = exam.id;
    selectedSubjectId.value = undefined;
  }
  fetchQuestions(true);
}

// 选择 L3 科目
function selectSubject(subject: Subject) {
  if (selectedSubjectId.value === subject.id) {
    selectedSubjectId.value = undefined;
  } else {
    selectedSubjectId.value = subject.id;
  }
  fetchQuestions(true);
}

function goToCategoryManage() {
  uni.navigateTo({ url: '/pages/profile/categories' });
}

function goToDetail(q: Question) {
  uni.navigateTo({ url: `/pages/question/detail?id=${q.id}` });
}

function getTypeName(type: string): string {
  const names: Record<string, string> = {
    single: '单选',
    multi: '多选',
    judge: '判断',
    fill: '填空',
    essay: '简答',
  };
  return names[type] || type;
}

function getDiffName(d: number): string {
  const names: Record<number, string> = { 1: '简单', 2: '中等', 3: '困难' };
  return names[d] || '未知';
}
</script>

<style lang="scss" scoped>
.question-list-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 30px;
}

/* 搜索栏 */
.search-bar {
  position: sticky;
  top: 0;
  z-index: 10;
  background: #fff;
  padding: 10px 16px;
}
.search-input-wrap {
  display: flex;
  align-items: center;
  background: #f5f7fa;
  border-radius: 20px;
  padding: 8px 16px;
  gap: 8px;
}
.search-input {
  flex: 1;
  font-size: 14px;
  background: transparent;
}

/* section */
.section {
  background: #fff;
  margin-bottom: 8px;
  padding: 12px 0;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px 8px;
}
.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #333;
}
.section-link {
  font-size: 13px;
  color: #1890ff;
}

/* 我的分类 */
.category-scroll {
  white-space: nowrap;
}
.category-chips {
  display: inline-flex;
  padding: 0 16px;
  gap: 10px;
}
.chip {
  display: inline-flex;
  align-items: center;
  padding: 6px 14px;
  background: #f5f7fa;
  border-radius: 16px;
  border: 1px solid #e8e8e8;
  gap: 4px;
}
.chip.active {
  background: #e6f7ff;
  border-color: #1890ff;
}
.primary-tag {
  font-size: 10px;
  background: #1890ff;
  color: #fff;
  padding: 1px 4px;
  border-radius: 4px;
}
.chip-text {
  font-size: 13px;
  color: #333;
}

/* 考试分类 tabs */
.category-tabs {
  display: inline-flex;
  padding: 0 16px;
  gap: 8px;
}
.tab {
  padding: 6px 14px;
  background: #f5f7fa;
  border-radius: 16px;
  font-size: 13px;
  color: #666;
  white-space: nowrap;
}
.tab.active {
  background: #1890ff;
  color: #fff;
}
.tab.sub {
  font-size: 12px;
  padding: 4px 12px;
}
.category-scroll.sub {
  padding-top: 8px;
}

/* L3 科目 pills */
.category-scroll.pills {
  padding-top: 8px;
}
.category-pills {
  display: inline-flex;
  padding: 0 16px;
  gap: 8px;
}
.pill {
  display: inline-flex;
  align-items: center;
  padding: 5px 12px;
  background: #f0f5ff;
  border-radius: 12px;
  font-size: 12px;
  color: #2f54eb;
  gap: 2px;
}
.pill.active {
  background: #2f54eb;
  color: #fff;
}
.pill .count {
  font-size: 11px;
  opacity: 0.7;
}

/* 题目列表 */
.question-list {
  padding: 8px 16px;
}
.loading, .empty {
  text-align: center;
  padding: 40px 0;
}
.empty-icon {
  font-size: 40px;
  display: block;
  margin-bottom: 8px;
}
.empty-text {
  font-size: 14px;
  color: #999;
}
.loading-more, .no-more {
  text-align: center;
  padding: 16px 0;
  font-size: 13px;
  color: #999;
}
.question-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}
.card-header {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}
.type-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  background: #f0f5ff;
  color: #2f54eb;
}
.type-badge.judge {
  background: #fff7e6;
  color: #fa8c16;
}
.diff-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
}
.diff-1 {
  background: #f6ffed;
  color: #52c41a;
}
.diff-2 {
  background: #fff7e6;
  color: #fa8c16;
}
.diff-3 {
  background: #fff2f0;
  color: #ff4d4f;
}
.question-title {
  font-size: 15px;
  color: #333;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 8px;
}
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.category-name {
  font-size: 12px;
  color: #999;
}
</style>
