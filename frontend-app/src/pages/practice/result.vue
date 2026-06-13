<template>
  <view class="result-page">
    <view class="header">
      <text class="title">练习完成 🎉</text>
    </view>

    <view class="stats-card">
      <view class="stat-row">
        <text class="stat-label">总题数</text>
        <text class="stat-value">{{ result.total }}</text>
      </view>
      <view class="stat-row">
        <text class="stat-label">已答题</text>
        <text class="stat-value">{{ result.answered }}</text>
      </view>
      <view class="stat-row">
        <text class="stat-label">用时</text>
        <text class="stat-value">{{ formatTime(result.elapsed) }}</text>
      </view>
    </view>

    <view class="action-buttons">
      <button class="action-btn" @click="goToPractice">再来一次</button>
      <button class="action-btn primary" @click="goHome">返回首页</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';

interface PracticeResult { total: number; answered: number; elapsed: number; answers: string[]; }
const result = ref<PracticeResult>({ total: 0, answered: 0, elapsed: 0, answers: [] });

onMounted(() => {
  const pages = getCurrentPages();
  const currentPage = pages[pages.length - 1] as any;
  const resultStr = currentPage.options?.result;
  if (resultStr) {
    try { result.value = JSON.parse(decodeURIComponent(resultStr)); } catch { /* ignore */ }
  }
  uni.removeStorageSync('practice_answers');
});

function formatTime(s: number): string { const m = Math.floor(s / 60); const sec = s % 60; return `${m}分${sec}秒`; }
function goToPractice() { uni.switchTab({ url: '/pages/practice/index' }); }
function goHome() { uni.switchTab({ url: '/pages/index/index' }); }
</script>

<style lang="scss" scoped>
.result-page { min-height: 100vh; background: #f5f5f5; padding: 16px; }
.header { text-align: center; padding: 30px 0; .title { font-size: 24px; font-weight: bold; } }
.stats-card { background: #fff; border-radius: 12px; padding: 24px; margin-bottom: 20px; }
.stat-row { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid #f0f0f0; &:last-child { border-bottom: none; } }
.stat-label { font-size: 15px; color: #666; }
.stat-value { font-size: 18px; font-weight: bold; color: #333; }
.action-buttons { display: flex; gap: 12px; }
.action-btn { flex: 1; height: 48px; line-height: 48px; border: none; border-radius: 24px; font-size: 16px; background: #f5f7fa; color: #333; }
.action-btn.primary { background: #1890ff; color: #fff; }
</style>
