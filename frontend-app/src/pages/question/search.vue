<template>
  <view class="search-page">
    <view class="search-bar">
      <input v-model="keyword" class="input" placeholder="搜索题目、知识点..." confirm-type="search" @confirm="handleSearch" />
      <button class="search-btn" @click="handleSearch">搜索</button>
    </view>
    <view v-if="!hasSearched && history.length" class="history-section">
      <text class="section-title">搜索历史</text>
      <view class="history-tags">
        <view v-for="(h, i) in history" :key="i" class="history-tag" @click="searchFromHistory(h)">{{ h }}</view>
      </view>
    </view>
    <view v-if="hasSearched" class="result-section">
      <view v-if="loading" class="loading">搜索中...</view>
      <view v-else-if="results.length === 0" class="empty">没有找到相关题目</view>
      <view v-else>
        <view v-for="q in results" :key="q.id" class="result-item" @click="goToDetail(q.id)">
          <text class="title">{{ q.title }}</text>
          <view class="footer">
            <text class="type">{{ getTypeLabel(q.questionType) }}</text>
            <text class="category">{{ q.categoryName }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { request } from '@/api/request';
const keyword = ref('');
const loading = ref(false);
const hasSearched = ref(false);
const results = ref<any[]>([]);
const history = ref<string[]>([]);
onMounted(() => { history.value = uni.getStorageSync('search_history') || []; });
async function handleSearch() {
  if (!keyword.value.trim()) return;
  saveHistory(keyword.value.trim());
  loading.value = true; hasSearched.value = true;
  try {
    const data = await request.get<any>('/questions/search', { params: { keyword: keyword.value, page: 1, pageSize: 20 } });
    results.value = data.items || [];
  } catch (e) {
    console.error('搜索失败:', e);
    results.value = [];
  } finally { loading.value = false; }
}
function searchFromHistory(kw: string) { keyword.value = kw; handleSearch(); }
function saveHistory(kw: string) { history.value = [kw, ...history.value.filter(h => h !== kw)].slice(0, 10); uni.setStorageSync('search_history', history.value); }
function getTypeLabel(type: string): string { return { single_choice: '单选', multiple_choice: '多选', true_false: '判断', fill_blank: '填空', short_answer: '简答' }[type] || type; }
function goToDetail(id: number) { uni.navigateTo({ url: `/pages/question/detail?id=${id}` }); }
</script>

<style lang="scss" scoped>
.search-page { min-height: 100vh; background: #f5f5f5; }
.search-bar { display: flex; padding: 12px 16px; background: #fff; gap: 8px; }
.input { flex: 1; background: #f5f7fa; border-radius: 20px; padding: 8px 16px; font-size: 14px; }
.search-btn { background: #1890ff; color: #fff; border: none; border-radius: 20px; padding: 0 20px; font-size: 14px; }
.history-section { background: #fff; margin: 12px 16px; border-radius: 8px; padding: 16px; }
.section-title { font-size: 14px; font-weight: 500; color: #333; margin-bottom: 12px; display: block; }
.history-tags { display: flex; flex-wrap: wrap; gap: 8px; }
.history-tag { padding: 6px 12px; background: #f5f7fa; border-radius: 16px; font-size: 13px; color: #666; }
.result-section { padding: 0 16px; }
.loading, .empty { text-align: center; padding: 40px 0; color: #999; }
.result-item { background: #fff; border-radius: 8px; padding: 16px; margin-bottom: 8px; }
.title { font-size: 15px; color: #333; line-height: 1.5; margin-bottom: 8px; }
.footer { display: flex; gap: 12px; }
.type, .category { font-size: 12px; color: #999; }
</style>
