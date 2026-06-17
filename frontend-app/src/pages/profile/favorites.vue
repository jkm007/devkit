<template>
  <view class="favorites-page">
    <!-- 头部 -->
    <view class="header">
      <text class="title">我的收藏</text>
      <text class="count">共 {{ total }} 题</text>
    </view>

    <!-- 列表 -->
    <view v-if="loading" class="loading">
      <text>加载中...</text>
    </view>
    <view v-else-if="list.length === 0" class="empty">
      <text class="empty-icon">⭐</text>
      <text class="empty-text">暂无收藏题目</text>
      <text class="empty-hint">在题目详情页点击收藏按钮添加</text>
    </view>
    <view v-else class="list">
      <view
        v-for="item in list"
        :key="item.id"
        class="item"
        @click="goToDetail(item.questionId || item.id)"
      >
        <view class="item-header">
          <text class="type-tag" :class="item.questionType">{{ getTypeLabel(item.questionType) }}</text>
          <view class="difficulty">
            <text v-for="i in (item.difficulty || 1)" :key="i" class="star">★</text>
          </view>
        </view>
        <text class="item-title">{{ item.title }}</text>
        <view class="item-footer">
          <text v-if="item.categoryName" class="category">{{ item.categoryName }}</text>
          <text class="time">{{ formatTime(item.createdAt || item.createTime) }}</text>
        </view>
        <view class="item-actions">
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
import { getFavorites, removeFavorite } from '@/api/study';

const loading = ref(false);
const loadingMore = ref(false);
const list = ref<any[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = 20;

onMounted(() => {
  loadList();
});

async function loadList() {
  loading.value = true;
  page.value = 1;
  try {
    const res = await getFavorites({ page: page.value, pageSize });
    list.value = res.items || [];
    total.value = res.total || 0;
  } catch (e) {
    console.error('加载收藏失败:', e);
    list.value = [];
  } finally {
    loading.value = false;
  }
}

async function loadMore() {
  if (loadingMore.value) return;
  loadingMore.value = true;
  page.value++;
  try {
    const res = await getFavorites({ page: page.value, pageSize });
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
    content: '确定取消收藏该题目？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await removeFavorite(item.questionId || item.id);
          list.value = list.value.filter(i => i.id !== item.id);
          total.value--;
          uni.showToast({ title: '已取消收藏', icon: 'success' });
        } catch (e) {
          uni.showToast({ title: '操作失败', icon: 'none' });
        }
      }
    }
  });
}

function goToDetail(id: number) {
  uni.navigateTo({ url: `/pages/question/detail?id=${id}` });
}

function getTypeLabel(type: string): string {
  const map: Record<string, string> = {
    single_choice: '单选', multiple_choice: '多选', indefinite_choice: '不定项',
    true_false: '判断', fill_blank: '填空', cloze: '完形',
    short_answer: '简答', essay_question: '论述', composition: '作文',
    term_explanation: '名词解释', material: '材料', case_analysis: '案例',
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
.favorites-page {
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
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);

  .item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .type-tag {
    font-size: 12px;
    padding: 2px 8px;
    background: #e6f7ff;
    color: #1890ff;
    border-radius: 4px;
  }

  .difficulty .star { font-size: 12px; color: #faad14; }

  .item-title {
    font-size: 15px;
    color: #333;
    line-height: 1.5;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .item-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 12px;
  }

  .category { font-size: 12px; color: #1890ff; background: #f0f8ff; padding: 2px 8px; border-radius: 4px; }
  .time { font-size: 12px; color: #999; }

  .item-actions {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid #f0f0f0;
    text-align: right;
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
