<template>
  <view class="categories-page">
    <view class="header">
      <text class="title">我的分类</text>
      <text class="subtitle">最多绑定 3 个分类</text>
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
            <text class="level-tag" :class="b.level">{{ b.level === 'subject' ? '科目' : '章节' }}</text>
            <text class="category-name">{{ b.categoryName || b.subjectName }}</text>
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
      <text class="tip-text">绑定科目或章节后，题库将默认显示对应的题目内容</text>
    </view>

    <!-- 分类选择弹窗 -->
    <up-popup v-model:show="showPicker" mode="bottom" round="12">
      <view class="picker-panel">
        <!-- 面包屑导航 -->
        <view class="picker-header">
          <text class="picker-title">选择分类</text>
          <view class="breadcrumb" v-if="selectedPath.length > 0">
            <text
              v-for="(item, idx) in selectedPath"
              :key="idx"
              class="breadcrumb-item"
              @click="goToLevel(idx)"
            >
              {{ item.name }}
              <text v-if="idx < selectedPath.length - 1" class="breadcrumb-sep"> > </text>
            </text>
          </view>
        </view>

        <!-- 返回按钮 -->
        <view v-if="currentLevel > 0" class="back-btn" @click="goBack">
          <text>← 返回</text>
        </view>

        <!-- 当前层级列表 -->
        <view class="picker-list">
          <view
            v-for="item in currentList"
            :key="item.id"
            class="picker-item"
            :class="{ bound: isBound(item.id), selectable: !item.children || item.children.length === 0 }"
            @click="selectItem(item)"
          >
            <view class="item-info">
              <text class="item-name">{{ item.name }}</text>
              <text class="item-count" v-if="item.questionCount">{{ item.questionCount }} 题</text>
            </view>
            <text v-if="item.children && item.children.length > 0" class="item-arrow">›</text>
            <text v-else-if="isBound(item.id)" class="item-bound">已绑定</text>
            <text v-else class="item-bind">绑定</text>
          </view>
        </view>
      </view>
    </up-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { onShow } from '@dcloudio/uni-app';
import { getCategoryBindings, bindSubject, unbindCategory, setPrimaryCategory, getCategoryTree } from '@/api/user';

// 分类树数据结构
interface CategoryItem {
  id: number;
  name: string;
  questionCount?: number;
  children?: CategoryItem[];
}

const bindings = ref<any[]>([]);
const showPicker = ref(false);
const categoryTree = ref<CategoryItem[]>([]);

// 级联选择状态
const selectedPath = ref<CategoryItem[]>([]); // 选中路径
const currentLevel = ref(0); // 当前层级 0=L1, 1=L2, 2=L3, 3=L4
const currentList = ref<CategoryItem[]>([]); // 当前显示的列表

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
    // API 返回结构：L1 { id, name, exams: [L2 { id, name, children: [L3 { id, name, children: [L4] }] }] }
    categoryTree.value = (Array.isArray(res) ? res : []).map((ec: any) => ({
      id: ec.id,
      name: ec.name,
      children: (ec.exams || []).map((exam: any) => ({
        id: exam.id,
        name: exam.name,
        questionCount: exam.questionCount,
        children: (exam.children || []).map((subj: any) => ({
          id: subj.id,
          name: subj.name,
          questionCount: subj.questionCount,
          children: (subj.children || []).map((cat: any) => ({
            id: cat.id,
            name: cat.name,
            questionCount: cat.questionCount,
          })),
        })),
      })),
    }));
  } catch {
    categoryTree.value = [];
    uni.showToast({ title: '加载分类失败', icon: 'none' });
  }
}

async function openPicker() {
  showPicker.value = true;
  selectedPath.value = [];
  currentLevel.value = 0;
  await loadCategoryTree();
  currentList.value = categoryTree.value;
}

function selectItem(item: CategoryItem) {
  if (isBound(item.id)) {
    uni.showToast({ title: '已绑定该分类', icon: 'none' });
    return;
  }

  // 如果有子节点，进入下一级
  if (item.children && item.children.length > 0) {
    selectedPath.value.push(item);
    currentLevel.value++;
    currentList.value = item.children;
    return;
  }

  // 没有子节点，执行绑定
  doBind(item);
}

async function doBind(item: CategoryItem) {
  showPicker.value = false;
  try {
    // 根据层级决定绑定类型
    // L3=科目, L4=章节
    const isSubject = currentLevel.value === 2; // 第3层是科目
    const data = isSubject
      ? { subjectId: item.id, isPrimary: bindings.value.length === 0 }
      : { categoryId: item.id, isPrimary: bindings.value.length === 0 };

    await bindSubject(data);
    uni.showToast({ title: '绑定成功', icon: 'success' });
    await loadBindings();
  } catch (e: any) {
    uni.showToast({ title: e.message || '绑定失败', icon: 'none' });
  }
}

function goBack() {
  if (selectedPath.value.length > 0) {
    selectedPath.value.pop();
    currentLevel.value--;
    if (selectedPath.value.length === 0) {
      currentList.value = categoryTree.value;
    } else {
      currentList.value = selectedPath.value[selectedPath.value.length - 1].children || [];
    }
  }
}

function goToLevel(idx: number) {
  selectedPath.value = selectedPath.value.slice(0, idx);
  currentLevel.value = idx;
  if (idx === 0) {
    currentList.value = categoryTree.value;
  } else {
    currentList.value = selectedPath.value[idx - 1].children || [];
  }
}

function isBound(id: number): boolean {
  return bindings.value.some((b: any) => b.categoryId === id || b.subjectId === id);
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
  const name = b.categoryName || b.subjectName;
  uni.showModal({
    title: '确认解绑',
    content: `确定要解绑「${name}」吗？`,
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
.level-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}
.level-tag.subject {
  background: #f0f5ff;
  color: #2f54eb;
}
.level-tag.category {
  background: #f6ffed;
  color: #52c41a;
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
  max-height: 70vh;
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
}
.breadcrumb-item {
  color: #1890ff;
}
.breadcrumb-sep {
  color: #999;
  margin: 0 2px;
}

.back-btn {
  padding: 8px 0;
  margin-bottom: 12px;
  font-size: 14px;
  color: #1890ff;
}

.picker-list {
  max-height: 50vh;
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
.picker-item.bound {
  opacity: 0.6;
}
.item-info {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}
.item-name {
  font-size: 15px;
  color: #333;
}
.item-count {
  font-size: 12px;
  color: #999;
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
.item-bind {
  font-size: 12px;
  color: #1890ff;
  background: #e6f7ff;
  padding: 2px 8px;
  border-radius: 4px;
}
</style>
