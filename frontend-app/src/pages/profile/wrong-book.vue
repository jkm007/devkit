<template>
  <view class="wrong-book-page">
    <!-- 统计卡片 -->
    <view class="stats-card">
      <view class="stat-item">
        <text class="stat-value">{{ stats.total }}</text>
        <text class="stat-label">总错题</text>
      </view>
      <view class="stat-item">
        <text class="stat-value">{{ stats.thisWeek }}</text>
        <text class="stat-label">本周新增</text>
      </view>
      <view class="stat-item">
        <text class="stat-value success">{{ stats.mastered }}</text>
        <text class="stat-label">已掌握</text>
      </view>
    </view>

    <!-- 筛选栏 -->
    <view class="filter-bar">
      <view class="filter-btn" @click="showCategoryFilter = true">
        <text>{{ selectedCategory || '全部分类' }}</text>
      </view>
      <view class="filter-btn" @click="toggleMastered">
        <text>{{ showMastered ? '全部' : '未掌握' }}</text>
      </view>
    </view>

    <!-- 错题列表 -->
    <view class="list-section">
      <view v-if="loading" class="loading"><Skeleton :count="3" /></view>
      <view v-else-if="questions.length === 0" class="empty">
        <text class="empty-icon">🎉</text>
        <text class="empty-text">暂无错题，继续保持！</text>
      </view>
      <view v-else>
        <view v-for="q in questions" :key="q.id" class="wrong-card">
          <view class="card-header">
            <text class="type-tag">{{ getTypeLabel(q.questionType) }}</text>
            <view class="difficulty">
              <text v-for="i in q.difficulty" :key="i" class="star">★</text>
            </view>
          </view>
          <text class="title">{{ q.title }}</text>
          <view class="card-footer">
            <text class="wrong-count">错误 {{ q.wrongCount }} 次</text>
            <text class="time">{{ formatDate(q.lastWrongAt) }}</text>
          </view>
          <view class="card-actions">
            <view class="action-btn view" @click="goToDetail(q.questionId)">
              <text>查看</text>
            </view>
            <view class="action-btn master" :class="{ done: q.isMastered }" @click="toggleMaster(q)">
              <text>{{ q.isMastered ? '已掌握' : '标记掌握' }}</text>
            </view>
            <view class="action-btn delete" @click="removeWrong(q)">
              <text>移除</text>
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 底部操作栏 -->
    <view v-if="questions.length" class="bottom-bar">
      <button class="bar-btn" @click="startReview">开始复习（{{ questions.length }}）</button>
    </view>

    <!-- 分类选择弹窗 -->
    <up-popup v-model:show="showCategoryFilter" mode="bottom" round="12">
      <view class="category-popup">
        <text class="popup-title">选择分类</text>
        <view class="category-list">
          <view class="category-item" :class="{ active: !selectedCategory }" @click="selectCategory(undefined)">
            <text>全部分类</text>
          </view>
          <view v-for="cat in categories" :key="cat.id" class="category-item" :class="{ active: selectedCategory === cat.name }" @click="selectCategory(cat.name)">
            <text>{{ cat.name }}</text>
          </view>
        </view>
      </view>
    </up-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getWrongBooks, getWrongBookStats, markWrongMastered, deleteWrongBook } from '@/api/study';
import Skeleton from '@/components/Skeleton.vue';

const loading = ref(true);
const questions = ref<any[]>([]);
const stats = ref({ total: 0, thisWeek: 0, mastered: 0 });
const selectedCategory = ref<string | undefined>(undefined);
const showMastered = ref(false);
const showCategoryFilter = ref(false);
const categories = ref<any[]>([]);

// 从缓存获取分类列表
async function loadCategories() {
  try {
    const list = uni.getStorageSync('categoryList');
    if (list) {
      categories.value = JSON.parse(list);
      return;
    }
  } catch { /* ignore */ }
  // 降级：写死默认分类
  categories.value = [
    { id: 1, name: '网络协议' },
    { id: 2, name: '操作系统' },
    { id: 3, name: '数据结构' },
  ];
}

onMounted(() => {
  loadCategories();
  loadStats();
  loadQuestions();
});

async function loadStats() {
  try {
    stats.value = await getWrongBookStats();
  } catch { /* ignore */ }
}

async function loadQuestions() {
  loading.value = true;
  try {
    const cat = categories.value.find(c => c.name === selectedCategory.value);
    const res = await getWrongBooks({
      page: 1,
      pageSize: 50,
      categoryId: cat ? cat.id : 0,
      isMastered: showMastered.value ? undefined : false,
    });
    questions.value = res.items || [];
  } catch {
    // Mock 数据
    questions.value = [
      { id: 1, questionId: 101, title: 'TCP三次握手的目的是什么？', questionType: 'single_choice', difficulty: 2, wrongCount: 3, lastWrongAt: '2024-01-15T10:30:00Z', isMastered: false },
      { id: 2, questionId: 102, title: 'HTTP和HTTPS的区别是什么？', questionType: 'short_answer', difficulty: 3, wrongCount: 2, lastWrongAt: '2024-01-14T14:20:00Z', isMastered: false },
    ];
  } finally {
    loading.value = false;
  }
}

function toggleMastered() {
  showMastered.value = !showMastered.value;
  loadQuestions();
}

function selectCategory(name: string | undefined) {
  selectedCategory.value = name;
  showCategoryFilter.value = false;
  loadQuestions();
}

async function toggleMaster(q: any) {
  if (q.isMastered) return;
  try {
    await markWrongMastered(q.questionId);
    q.isMastered = true;
    stats.value.mastered++;
    stats.value.total--;
  } catch {
    uni.showToast({ title: '操作失败', icon: 'none' });
  }
}

async function removeWrong(q: any) {
  uni.showModal({
    title: '确认移除',
    content: `确定要移除「${q.title.substring(0, 20)}...」吗？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await deleteWrongBook(q.questionId);
          questions.value = questions.value.filter(item => item.questionId !== q.questionId);
          stats.value.total--;
        } catch {
          uni.showToast({ title: '移除失败', icon: 'none' });
        }
      }
    },
  });
}

function goToDetail(id: number) {
  uni.navigateTo({ url: `/pages/question/detail?id=${id}` });
}

function startReview() {
  const ids = questions.value.filter(q => !q.isMastered).map(q => q.questionId);
  uni.setStorageSync('wrongSessionIds', JSON.stringify(ids));
  uni.navigateTo({ url: '/pages/practice/wrong-session' });
}

function getTypeLabel(type: string): string {
  return { single_choice: '单选', multiple_choice: '多选', true_false: '判断', fill_blank: '填空', short_answer: '简答' }[type] || type;
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  return `${d.getMonth() + 1}/${d.getDate()}`;
}
</script>

<style lang="scss" scoped>
.wrong-book-page { min-height: 100vh; background: #f5f5f5; padding-bottom: 80px; }

.stats-card { display: flex; background: linear-gradient(135deg, #1890ff, #36cfc9); padding: 20px 16px; margin: 16px; border-radius: 12px; }
.stat-item { flex: 1; text-align: center; }
.stat-value { display: block; font-size: 24px; font-weight: bold; color: #fff; }
.stat-value.success { color: #52c41a; }
.stat-label { display: block; font-size: 12px; color: rgba(255,255,255,0.8); margin-top: 4px; }

.filter-bar { display: flex; gap: 8px; padding: 0 16px; margin-bottom: 12px; }
.filter-btn { padding: 8px 16px; background: #fff; border-radius: 20px; font-size: 13px; color: #333; }

.list-section { padding: 0 16px; }
.loading { padding: 16px 0; }
.empty { text-align: center; padding: 60px 0; }
.empty-icon { font-size: 48px; display: block; margin-bottom: 12px; }
.empty-text { font-size: 14px; color: #999; }

.wrong-card { background: #fff; border-radius: 12px; padding: 16px; margin-bottom: 12px; }
.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.type-tag { font-size: 12px; padding: 2px 8px; background: #e6f7ff; color: #1890ff; border-radius: 4px; }
.difficulty .star { color: #faad14; font-size: 12px; }
.title { font-size: 15px; color: #333; line-height: 1.5; margin-bottom: 8px; display: block; }
.card-footer { display: flex; justify-content: space-between; margin-bottom: 12px; }
.wrong-count { font-size: 12px; color: #ff4d4f; }
.time { font-size: 12px; color: #999; }
.card-actions { display: flex; gap: 8px; }
.action-btn { flex: 1; text-align: center; padding: 8px 0; border-radius: 8px; font-size: 13px; }
.action-btn.view { background: #f5f7fa; color: #666; }
.action-btn.master { background: #e6f7ff; color: #1890ff; }
.action-btn.master.done { background: #f6ffed; color: #52c41a; }
.action-btn.delete { background: #fff2f0; color: #ff4d4f; }

.bottom-bar { position: fixed; bottom: 0; left: 0; right: 0; padding: 12px 16px; background: #fff; box-shadow: 0 -2px 10px rgba(0,0,0,0.05); }
.bar-btn { height: 44px; line-height: 44px; background: #1890ff; color: #fff; border: none; border-radius: 22px; font-size: 15px; }

.category-popup { padding: 20px; }
.popup-title { font-size: 16px; font-weight: 500; margin-bottom: 16px; display: block; }
.category-list { max-height: 300px; overflow-y: auto; }
.category-item { padding: 12px 16px; background: #f5f7fa; border-radius: 8px; margin-bottom: 8px; font-size: 14px; }
.category-item.active { background: #e6f7ff; color: #1890ff; }
</style>
