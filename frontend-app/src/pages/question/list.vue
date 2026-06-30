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
    <view class="section" v-if="!isClassMode && myBindings.length > 0">
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
            :class="{ active: activeBindingId === b.id }"
            @click="selectMyBinding(b)"
          >
            <text class="primary-tag" v-if="b.isPrimary">主</text>
            <text class="chip-text">{{ b.subjectName }}</text>
            <text class="chip-sub" v-if="b.categoryName">·{{ b.categoryName }}</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- 班级模式标题 -->
    <view v-if="isClassMode" class="class-header">
      <text class="class-title">{{ className || '班级题目' }}</text>
    </view>

    <!-- 考试分类 -->
    <view class="section" v-if="!isClassMode">
      <view class="section-header">
        <text class="section-title">考试分类</text>
        <text class="section-path" v-if="currentPath">{{ currentPath }}</text>
      </view>
      <!-- L1 考试大类 tabs -->
      <scroll-view scroll-x class="category-scroll">
        <view class="category-tabs">
          <view
            class="tab"
            :class="{ active: selectedL1Id === null }"
            @click="selectL1(null)"
          >
            <text>全部</text>
          </view>
          <view
            v-for="ec in categoryTree"
            :key="ec.id"
            class="tab"
            :class="{ active: selectedL1Id === ec.id }"
            @click="selectL1(ec)"
          >
            <text>{{ ec.name }}</text>
          </view>
        </view>
      </scroll-view>

      <!-- L2 考试 tabs -->
      <scroll-view v-if="currentL2List.length > 0" scroll-x class="category-scroll sub">
        <view class="category-tabs">
          <view
            class="tab sub"
            :class="{ active: selectedL2Id === null }"
            @click="selectL2(null)"
          >
            <text>全部</text>
          </view>
          <view
            v-for="exam in currentL2List"
            :key="exam.id"
            class="tab sub"
            :class="{ active: selectedL2Id === exam.id }"
            @click="selectL2(exam)"
          >
            <text>{{ exam.name }}</text>
          </view>
        </view>
      </scroll-view>

      <!-- L3 科目 pills -->
      <scroll-view v-if="currentL3List.length > 0" scroll-x class="category-scroll pills">
        <view class="category-pills">
          <view
            v-for="subject in currentL3List"
            :key="subject.id"
            class="pill"
            :class="{ active: selectedL3Id === subject.id, expand: subject.children && subject.children.length > 0 }"
            @click="selectL3(subject)"
          >
            <text>{{ subject.name }}</text>
            <text class="count" v-if="subject.questionCount">({{ subject.questionCount }})</text>
            <text class="arrow" v-if="subject.children && subject.children.length > 0">▼</text>
          </view>
        </view>
      </scroll-view>

      <!-- L4 章节 pills -->
      <scroll-view v-if="currentL4List.length > 0" scroll-x class="category-scroll pills l4">
        <view class="category-pills">
          <view
            v-for="cat in currentL4List"
            :key="cat.id"
            class="pill l4"
            :class="{ active: selectedL4Id === cat.id }"
            @click="selectL4(cat)"
          >
            <text>{{ cat.name }}</text>
            <text class="count" v-if="cat.questionCount">({{ cat.questionCount }})</text>
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
import { QUESTION_TYPE_LABELS, type Question } from '@/api/types';

// 分类树数据结构
interface CategoryItem {
  id: number;
  name: string;
  questionCount?: number;
  children?: CategoryItem[];
}
interface CategoryTreeNode {
  id: number;
  name: string;
  exams: CategoryItem[]; // L2 考试，包含 children (L3 科目)
}

const questions = ref<Question[]>([]);
const loading = ref(false);
const page = ref(1);
const hasMore = ref(true);
const keyword = ref('');
const pageSize = 20;

// 分类数据
const categoryTree = ref<CategoryTreeNode[]>([]);
const myBindings = ref<any[]>([]);
const activeBindingId = ref<number | null>(null);

// 当前选中的层级 ID
const selectedL1Id = ref<number | null>(null);
const selectedL2Id = ref<number | null>(null);
const selectedL3Id = ref<number | null>(null);
const selectedL4Id = ref<number | null>(null);

// 班级模式
const classId = ref(0);
const className = ref('');
const isClassMode = computed(() => classId.value > 0);

// 当前 L2 考试列表
const currentL2List = computed(() => {
  if (selectedL1Id.value === null) return [];
  const ec = categoryTree.value.find(e => e.id === selectedL1Id.value);
  return ec?.exams || [];
});

// 当前 L3 科目列表
const currentL3List = computed(() => {
  if (selectedL2Id.value === null) {
    // 只选了 L1，显示该 L1 下所有科目（去重）
    const ec = categoryTree.value.find(e => e.id === selectedL1Id.value);
    if (!ec) return [];
    const subjects: CategoryItem[] = [];
    for (const exam of ec.exams) {
      for (const s of (exam.children || [])) {
        if (!subjects.find(x => x.id === s.id)) {
          subjects.push(s);
        }
      }
    }
    return subjects;
  }
  const exam = currentL2List.value.find(e => e.id === selectedL2Id.value);
  return exam?.children || [];
});

// 当前 L4 章节列表
const currentL4List = computed(() => {
  if (selectedL3Id.value === null) return [];
  const subject = currentL3List.value.find(s => s.id === selectedL3Id.value);
  return subject?.children || [];
});

// 当前路径显示
const currentPath = computed(() => {
  const parts: string[] = [];
  if (selectedL1Id.value) {
    const ec = categoryTree.value.find(e => e.id === selectedL1Id.value);
    if (ec) parts.push(ec.name);
  }
  if (selectedL2Id.value) {
    const exam = currentL2List.value.find(e => e.id === selectedL2Id.value);
    if (exam) parts.push(exam.name);
  }
  if (selectedL3Id.value) {
    const subject = currentL3List.value.find(s => s.id === selectedL3Id.value);
    if (subject) parts.push(subject.name);
  }
  if (selectedL4Id.value) {
    const cat = currentL4List.value.find(c => c.id === selectedL4Id.value);
    if (cat) parts.push(cat.name);
  }
  return parts.length > 0 ? parts.join(' > ') : '';
});

onLoad((options: any) => {
  loadCategoryTree();
  loadMyBindings();

  // 支持从分类收藏等页面跳转过来时直接按指定分类筛选
  if (options) {
    const categoryId = options.categoryId ? Number(options.categoryId) : 0;
    const subjectId = options.subjectId ? Number(options.subjectId) : 0;
    const examId = options.examId ? Number(options.examId) : 0;
    const examCategoryId = options.examCategoryId ? Number(options.examCategoryId) : 0;
    const cid = options.classId ? Number(options.classId) : 0;
    const cname = options.className || '';

    if (cid > 0) {
      classId.value = cid;
      className.value = cname;
    } else if (categoryId > 0) {
      selectedL4Id.value = categoryId;
    } else if (subjectId > 0) {
      selectedL3Id.value = subjectId;
    } else if (examId > 0) {
      selectedL2Id.value = examId;
    } else if (examCategoryId > 0) {
      selectedL1Id.value = examCategoryId;
    }
  }

  // 首次进入加载题目列表
  fetchQuestions(true);
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
    if (classId.value > 0) {
      params.classId = classId.value;
    } else if (selectedL4Id.value) {
      params.categoryId = selectedL4Id.value;
    } else if (selectedL3Id.value) {
      params.subjectId = selectedL3Id.value;
    } else if (selectedL2Id.value) {
      params.examId = selectedL2Id.value;
    } else if (selectedL1Id.value) {
      params.examCategoryId = selectedL1Id.value;
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

// 选择我的分类 - 联动到考试分类
function selectMyBinding(b: any) {
  if (activeBindingId.value === b.id) {
    // 取消选中
    activeBindingId.value = null;
    resetSelection();
  } else {
    activeBindingId.value = b.id;
    // 联动：自动展开到对应的分类路径
    selectedL1Id.value = b.examCategoryId || null;
    selectedL2Id.value = b.examId || null;
    selectedL3Id.value = b.subjectId || null;
    selectedL4Id.value = b.level === 'category' ? b.categoryId : null;
  }
  fetchQuestions(true);
}

// 选择 L1 考试大类
function selectL1(ec: CategoryTreeNode | null) {
  activeBindingId.value = null;
  if (ec === null) {
    resetSelection();
  } else {
    selectedL1Id.value = ec.id;
    selectedL2Id.value = null;
    selectedL3Id.value = null;
    selectedL4Id.value = null;
  }
  fetchQuestions(true);
}

// 选择 L2 考试
function selectL2(exam: CategoryItem | null) {
  activeBindingId.value = null;
  if (exam === null) {
    selectedL2Id.value = null;
    selectedL3Id.value = null;
    selectedL4Id.value = null;
  } else {
    selectedL2Id.value = exam.id;
    selectedL3Id.value = null;
    selectedL4Id.value = null;
  }
  fetchQuestions(true);
}

// 选择 L3 科目
function selectL3(subject: CategoryItem) {
  activeBindingId.value = null;
  if (selectedL3Id.value === subject.id) {
    // 取消选中
    selectedL3Id.value = null;
    selectedL4Id.value = null;
  } else {
    selectedL3Id.value = subject.id;
    selectedL4Id.value = null;
  }
  fetchQuestions(true);
}

// 选择 L4 章节
function selectL4(cat: CategoryItem) {
  activeBindingId.value = null;
  if (selectedL4Id.value === cat.id) {
    selectedL4Id.value = null;
  } else {
    selectedL4Id.value = cat.id;
  }
  fetchQuestions(true);
}

function resetSelection() {
  selectedL1Id.value = null;
  selectedL2Id.value = null;
  selectedL3Id.value = null;
  selectedL4Id.value = null;
}

function goToCategoryManage() {
  uni.navigateTo({ url: '/pages/profile/categories' });
}

function goToDetail(q: Question) {
  uni.navigateTo({ url: `/pages/question/detail?id=${q.id}` });
}

function getTypeName(type: string): string {
  return QUESTION_TYPE_LABELS[type as keyof typeof QUESTION_TYPE_LABELS] || type;
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

/* 班级模式标题 */
.class-header {
  background: linear-gradient(135deg, #1890ff 0%, #36cfc9 100%);
  padding: 16px;
  .class-title {
    font-size: 16px;
    font-weight: bold;
    color: #fff;
  }
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
.section-path {
  font-size: 12px;
  color: #999;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
.chip-sub {
  font-size: 12px;
  color: #999;
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

/* L3/L4 pills */
.category-scroll.pills {
  padding-top: 8px;
}
.category-scroll.pills.l4 {
  padding-left: 16px;
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
.pill.expand {
  padding-right: 8px;
}
.pill .count {
  font-size: 11px;
  opacity: 0.7;
}
.pill .arrow {
  font-size: 10px;
  margin-left: 2px;
}
.pill.l4 {
  background: #f6ffed;
  color: #52c41a;
}
.pill.l4.active {
  background: #52c41a;
  color: #fff;
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
