<template>
  <view class="categories-page">
    <view class="header">
      <text class="title">我的分类</text>
      <text class="subtitle">最多绑定 3 个科目（1 个主分类 + 2 个副分类）</text>
    </view>

    <!-- 已绑定列表 -->
    <view class="bound-section">
      <view v-if="bindings.length === 0" class="empty">
        <text class="empty-text">暂未绑定分类</text>
      </view>
      <view v-else>
        <view v-for="b in bindings" :key="b.id" class="binding-card">
          <view class="card-header">
            <text class="primary-tag" v-if="b.isPrimary">主</text>
            <text class="category-name">{{ b.subjectName || '科目 ' + b.subjectId }}</text>
          </view>
          <text class="path-text" v-if="b.path">{{ b.path }}</text>
          <text class="bound-time">绑定于 {{ formatDate(b.boundAt) }}</text>
          <view class="card-actions">
            <view v-if="!b.isPrimary" class="action-btn" @click="setPrimary(b)">设为主分类</view>
            <view class="action-btn delete" @click="unbind(b)">✕ 解绑</view>
          </view>
        </view>
      </view>
    </view>

    <!-- 添加按钮 -->
    <button class="add-btn" :disabled="bindings.length >= 3" @click="openPicker">
      <text>{{ bindings.length >= 3 ? '已满 3 个分类' : '+ 添加分类' }}</text>
    </button>

    <!-- 提示 -->
    <view class="tip-section">
      <text class="tip-text">绑定科目后，题库将默认显示该科目下的题目内容</text>
    </view>

    <!-- 分类选择弹窗 -->
    <up-popup v-model:show="showPicker" mode="bottom" round="12">
      <view class="picker-panel">
        <!-- 面包屑导航 -->
        <view class="picker-header">
          <text class="picker-title">选择科目</text>
          <view class="breadcrumb" v-if="selectedL1 || selectedL2">
            <text class="breadcrumb-item" @click="goBackToL1" v-if="selectedL1">{{ selectedL1.name }}</text>
            <text class="breadcrumb-sep" v-if="selectedL1 && selectedL2"> > </text>
            <text class="breadcrumb-item" v-if="selectedL2">{{ selectedL2.name }}</text>
          </view>
        </view>

        <!-- 返回按钮 -->
        <view v-if="selectedL1 || selectedL2" class="back-btn" @click="goBack">
          <text>← 返回</text>
        </view>

        <!-- Step 1: 选择考试大类 -->
        <view v-if="!selectedL1" class="picker-list">
          <view v-for="ec in categoryTree" :key="ec.id" class="picker-item" @click="selectL1(ec)">
            <text class="item-name">{{ ec.name }}</text>
            <text class="item-arrow">›</text>
          </view>
        </view>

        <!-- Step 2: 选择考试 -->
        <view v-else-if="selectedL1 && !selectedL2" class="picker-list">
          <view v-for="exam in selectedL1.exams" :key="exam.id" class="picker-item" @click="selectL2(exam)">
            <text class="item-name">{{ exam.name }}</text>
            <text class="item-count">{{ exam.subjects?.length || 0 }} 个科目</text>
            <text class="item-arrow">›</text>
          </view>
        </view>

        <!-- Step 3: 选择科目 -->
        <view v-else-if="selectedL2" class="picker-list">
          <view
            v-for="subject in selectedL2.subjects"
            :key="subject.id"
            class="picker-item subject"
            @click="selectSubject(subject)"
          >
            <text class="item-name">{{ subject.name }}</text>
            <text class="item-count" v-if="subject.questionCount">{{ subject.questionCount }} 题</text>
            <text class="item-bound" v-if="isBound(subject.id)">已绑定</text>
          </view>
        </view>
      </view>
    </up-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { onShow } from '@dcloudio/uni-app';
import { getCategoryBindings, bindSubject, unbindCategory, setPrimaryCategory, getCategoryTree } from '@/api/user';

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

const bindings = ref<any[]>([]);
const showPicker = ref(false);
const categoryTree = ref<ExamCategory[]>([]);

// 级联选择状态
const selectedL1 = ref<ExamCategory | null>(null);
const selectedL2 = ref<Exam | null>(null);

onMounted(() => {
  loadBindings();
});

onShow(() => {
  loadBindings();
});

async function loadBindings() {
  try {
    bindings.value = await getCategoryBindings();
  } catch {
    bindings.value = [];
  }
}

async function loadCategoryTree() {
  try {
    const res = await getCategoryTree();
    categoryTree.value = Array.isArray(res) ? res : [];
  } catch {
    categoryTree.value = [];
    uni.showToast({ title: '加载分类失败', icon: 'none' });
  }
}

async function openPicker() {
  showPicker.value = true;
  selectedL1.value = null;
  selectedL2.value = null;
  await loadCategoryTree();
}

function selectL1(ec: ExamCategory) {
  selectedL1.value = ec;
  selectedL2.value = null;
}

function selectL2(exam: Exam) {
  selectedL2.value = exam;
}

async function selectSubject(subject: Subject) {
  if (isBound(subject.id)) {
    uni.showToast({ title: '已绑定该科目', icon: 'none' });
    return;
  }
  showPicker.value = false;
  try {
    await bindSubject({ subjectId: subject.id, isPrimary: bindings.value.length === 0 });
    uni.showToast({ title: '绑定成功', icon: 'success' });
    await loadBindings();
  } catch (e: any) {
    uni.showToast({ title: e.message || '绑定失败', icon: 'none' });
  }
}

function goBack() {
  if (selectedL2.value) {
    selectedL2.value = null;
  } else if (selectedL1.value) {
    selectedL1.value = null;
  }
}

function goBackToL1() {
  selectedL1.value = null;
  selectedL2.value = null;
}

function isBound(subjectId: number): boolean {
  return bindings.value.some((b: any) => b.subjectId === subjectId);
}

async function setPrimary(b: any) {
  try {
    await setPrimaryCategory(b.id);
    uni.showToast({ title: '已设为主分类', icon: 'success' });
    loadBindings();
  } catch {
    uni.showToast({ title: '操作失败', icon: 'none' });
  }
}

async function unbind(b: any) {
  uni.showModal({
    title: '确认解绑',
    content: `确定要解绑「${b.subjectName}」吗？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await unbindCategory(b.id);
          uni.showToast({ title: '已解绑', icon: 'success' });
          await loadBindings();
        } catch {
          uni.showToast({ title: '解绑失败', icon: 'none' });
        }
      }
    },
  });
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}
</script>

<style lang="scss" scoped>
.categories-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 16px;
}

.header {
  margin-bottom: 20px;
}
.title {
  font-size: 20px;
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

.bound-section {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
}
.empty {
  text-align: center;
  padding: 20px 0;
}
.empty-text {
  font-size: 14px;
  color: #999;
}

.binding-card {
  padding: 16px;
  background: #f9f9f9;
  border-radius: 8px;
  margin-bottom: 12px;
}
.binding-card:last-child {
  margin-bottom: 0;
}
.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}
.primary-tag {
  padding: 2px 8px;
  background: #1890ff;
  color: #fff;
  border-radius: 4px;
  font-size: 12px;
}
.category-name {
  font-size: 16px;
  font-weight: 500;
  color: #333;
}
.path-text {
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
  display: block;
}
.bound-time {
  font-size: 12px;
  color: #999;
  margin-bottom: 12px;
  display: block;
}
.card-actions {
  display: flex;
  gap: 8px;
}
.action-btn {
  padding: 6px 16px;
  background: #f5f7fa;
  border-radius: 16px;
  font-size: 13px;
  color: #666;
}
.action-btn.delete {
  color: #ff4d4f;
}

.add-btn {
  height: 48px;
  line-height: 48px;
  border: 2px dashed #d9d9d9;
  border-radius: 12px;
  background: transparent;
  color: #1890ff;
  font-size: 15px;
}
.add-btn[disabled] {
  color: #ccc;
  border-color: #eee;
}

.tip-section {
  margin-top: 16px;
  padding: 12px;
  background: #fffbe6;
  border-radius: 8px;
}
.tip-text {
  font-size: 12px;
  color: #d48806;
}

/* 选择器弹窗 */
.picker-panel {
  padding: 20px;
  max-height: 60vh;
  overflow-y: auto;
}
.picker-header {
  margin-bottom: 16px;
}
.picker-title {
  font-size: 16px;
  font-weight: 500;
  display: block;
}
.breadcrumb {
  margin-top: 8px;
  font-size: 13px;
  color: #1890ff;
}
.breadcrumb-item {
  color: #1890ff;
}
.breadcrumb-sep {
  color: #999;
  margin: 0 4px;
}

.back-btn {
  padding: 8px 0;
  margin-bottom: 12px;
  font-size: 14px;
  color: #1890ff;
}

.picker-list {
  max-height: 45vh;
  overflow-y: auto;
}
.picker-item {
  display: flex;
  align-items: center;
  padding: 14px 16px;
  background: #f5f7fa;
  border-radius: 8px;
  margin-bottom: 8px;
}
.picker-item:active {
  background: #e6f7ff;
}
.item-name {
  flex: 1;
  font-size: 15px;
  color: #333;
}
.item-count {
  font-size: 12px;
  color: #999;
  margin-right: 8px;
}
.item-arrow {
  font-size: 18px;
  color: #ccc;
}
.item-bound {
  font-size: 12px;
  color: #52c41a;
  background: #f6ffed;
  padding: 2px 8px;
  border-radius: 4px;
}
.picker-item.subject {
  background: #f0f5ff;
}
</style>
