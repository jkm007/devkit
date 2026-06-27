<template>
  <view class="category-favorites-page">
    <!-- 头部 -->
    <view class="header">
      <text class="title">分类收藏</text>
      <text class="count">共 {{ total }} 个</text>
    </view>

    <!-- 列表 -->
    <view v-if="loading" class="loading">
      <text>加载中...</text>
    </view>
    <view v-else-if="list.length === 0" class="empty">
      <text class="empty-icon">🏷️</text>
      <text class="empty-text">暂无收藏分类</text>
      <text class="empty-hint">在浏览分类时长按或点击收藏按钮添加</text>
    </view>
    <view v-else class="list">
      <view
        v-for="item in list"
        :key="item.id"
        class="item"
        @click="goToQuestions(item)"
      >
        <view class="item-header">
          <text class="level-tag" :class="item.targetType">{{ getLevelLabel(item.targetType) }}</text>
          <text class="item-name">{{ item.targetName }}</text>
        </view>
        <text class="path-text">{{ item.path }}</text>
        <view class="item-footer">
          <text class="time">{{ formatTime(item.createdAt) }}</text>
          <text class="action-btn cancel" @click.stop="removeFav(item)">取消收藏</text>
        </view>
      </view>
    </view>

    <!-- 加载更多 -->
    <view v-if="list.length > 0 && list.length < total" class="load-more" @click="loadMore">
      <text>{{ loadingMore ? '加载中...' : '加载更多' }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { onShow } from '@dcloudio/uni-app';
import { getCategoryFavorites, removeCategoryFavorite } from '@/api/study';

const loading = ref(false);
const loadingMore = ref(false);
const list = ref<any[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = 20;

onMounted(() => {
  loadList();
});

onShow(() => {
  loadList();
});

async function loadList() {
  loading.value = true;
  page.value = 1;
  try {
    const res = await getCategoryFavorites({ page: page.value, pageSize });
    list.value = res.items || [];
    total.value = res.total || 0;
  } catch (e) {
    console.error('加载分类收藏失败:', e);
    list.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
  }
}

async function loadMore() {
  if (loadingMore.value) return;
  loadingMore.value = true;
  page.value++;
  try {
    const res = await getCategoryFavorites({ page: page.value, pageSize });
    list.value = [...list.value, ...(res.items || [])];
  } catch (e) {
    console.error('加载更多失败:', e);
  } finally {
    loadingMore.value = false;
  }
}

async function removeFav(item: any) {
  uni.showModal({
    title: '提示',
    content: '确定取消收藏该分类？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await removeCategoryFavorite(item.id);
          list.value = list.value.filter(i => i.id !== item.id);
          total.value--;
          uni.showToast({ title: '已取消收藏', icon: 'success' });
        } catch (e) {
          uni.showToast({ title: '操作失败', icon: 'none' });
        }
      }
    },
  });
}

function goToQuestions(item: any) {
  const params: Record<string, number> = {};
  switch (item.targetType) {
    case 'exam_category':
      params.examCategoryId = item.targetId;
      break;
    case 'exam':
      params.examId = item.targetId;
      break;
    case 'subject':
      params.subjectId = item.targetId;
      break;
    case 'category':
      params.categoryId = item.targetId;
      break;
  }
  const query = Object.entries(params)
    .map(([k, v]) => `${k}=${v}`)
    .join('&');
  uni.navigateTo({ url: `/pages/question/list?${query}` });
}

function getLevelLabel(type: string): string {
  const map: Record<string, string> = {
    exam_category: '考试大类',
    exam: '考试',
    subject: '科目',
    category: '章节',
  };
  return map[type] || type;
}

function formatTime(time: string): string {
  if (!time) return '';
  const d = new Date(time);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  const days = Math.floor(diff / 86400000);
  if (days === 0) return '今天';
  if (days === 1) return '昨天';
  if (days < 7) return `${days}天前`;
  return `${d.getMonth() + 1}/${d.getDate()}`;
}
</script>

<style lang="scss" scoped>
.category-favorites-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #fff;
  .title { font-size: 18px; font-weight: bold; color: #333; }
  .count { font-size: 13px; color: #999; }
}

.loading, .empty {
  text-align: center;
  padding: 60px 20px;
}
.empty-icon { font-size: 48px; display: block; margin-bottom: 16px; }
.empty-text { font-size: 16px; color: #666; display: block; }
.empty-hint { font-size: 13px; color: #999; display: block; margin-top: 8px; }

.list { padding: 12px 16px; }

.item {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);

  &:active {
    opacity: 0.7;
  }

  .item-header {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 8px;
  }

  .level-tag {
    font-size: 11px;
    padding: 2px 8px;
    border-radius: 4px;
    color: #fff;
    flex-shrink: 0;

    &.exam_category { background: #1890ff; }
    &.exam { background: #2f54eb; }
    &.subject { background: #52c41a; }
    &.category { background: #faad14; }
  }

  .item-name {
    font-size: 16px;
    font-weight: 500;
    color: #333;
    flex: 1;
  }

  .path-text {
    font-size: 13px;
    color: #999;
    line-height: 1.5;
    margin-bottom: 12px;
    display: block;
  }

  .item-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .time {
    font-size: 12px;
    color: #999;
  }

  .action-btn {
    font-size: 13px;
    padding: 4px 12px;
    border-radius: 16px;
    &.cancel { color: #ff4d4f; background: #fff1f0; }
  }
}

.load-more {
  text-align: center;
  padding: 16px;
  color: #1890ff;
  font-size: 14px;
}
</style>
