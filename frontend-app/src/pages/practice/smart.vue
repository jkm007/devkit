<template>
  <view class="smart-practice-page">
    <view class="header">
      <text class="title">智能练习</text>
    </view>

    <!-- 学习分析卡片 -->
    <view class="analysis-card">
      <view class="analysis-header">
        <text class="analysis-title">学习分析</text>
      </view>
      <view class="stats-row">
        <view class="stat-item">
          <text class="stat-value">{{ analysis.totalPractice }}</text>
          <text class="stat-label">已练题数</text>
        </view>
        <view class="stat-item">
          <text class="stat-value">{{ analysis.totalWrong }}</text>
          <text class="stat-label">错题数量</text>
        </view>
        <view class="stat-item">
          <text class="stat-value">{{ Math.round(analysis.accuracy * 100) }}%</text>
          <text class="stat-label">正确率</text>
        </view>
      </view>
      <view class="stats-row">
        <view class="stat-item">
          <text class="stat-value">{{ ['简单', '中等', '困难'][analysis.suggestedDiff - 1] }}</text>
          <text class="stat-label">建议难度</text>
        </view>
      </view>
      <view v-if="analysis.weakKnowledge.length" class="weak-section">
        <text class="weak-title">薄弱知识点</text>
        <view class="weak-tags">
          <text v-for="(kp, i) in analysis.weakKnowledge" :key="i" class="weak-tag">{{ kp }}</text>
        </view>
      </view>
      <view v-else class="weak-section">
        <text class="weak-title">暂无薄弱知识点数据</text>
        <text class="weak-hint">多做练习后会自动分析</text>
      </view>
    </view>

    <!-- 练习模式选择 -->
    <view class="mode-section">
      <text class="section-title">练习模式</text>
      <view class="mode-cards">
        <view class="mode-card" :class="{ selected: mode === 'mixed' }" @click="mode = 'mixed'">
          <text class="mode-icon">🎯</text>
          <text class="mode-label">混合模式</text>
          <text class="mode-desc">错题 + 随机推荐</text>
        </view>
        <view class="mode-card" :class="{ selected: mode === 'review' }" @click="mode = 'review'">
          <text class="mode-icon">📖</text>
          <text class="mode-label">复习模式</text>
          <text class="mode-desc">只练错题</text>
        </view>
        <view class="mode-card" :class="{ selected: mode === 'weak' }" @click="mode = 'weak'">
          <text class="mode-icon">💪</text>
          <text class="mode-label">薄弱模式</text>
          <text class="mode-desc">针对薄弱知识点</text>
        </view>
      </view>
    </view>

    <!-- 题目数量 -->
    <view class="count-section">
      <text class="section-title">题目数量</text>
      <view class="count-selector">
        <view class="count-btn" @click="adjustCount(-5)">−</view>
        <text class="count-value">{{ count }}</text>
        <view class="count-btn" @click="adjustCount(5)">+</view>
      </view>
    </view>

    <!-- 开始按钮 -->
    <button class="start-btn" type="primary" :loading="loading" @click="startPractice">
      {{ loading ? '加载中...' : '开始智能练习' }}
    </button>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getPracticeAnalysis, getSmartPractice } from '@/api/study';

const mode = ref<'review' | 'weak' | 'mixed'>('mixed');
const count = ref(20);
const loading = ref(false);
const analysis = ref({
  accuracy: 0,
  suggestedDiff: 2,
  weakKnowledge: [] as string[],
  totalWrong: 0,
  totalPractice: 0
});

onMounted(() => {
  loadAnalysis();
});

async function loadAnalysis() {
  try {
    const res = await getPracticeAnalysis();
    if (res) {
      analysis.value = {
        accuracy: res.accuracy || 0,
        suggestedDiff: res.suggestedDiff || 2,
        weakKnowledge: res.weakKnowledge || [],
        totalWrong: res.totalWrong || 0,
        totalPractice: res.totalPractice || 0
      };
    }
  } catch (e) {
    console.error('加载分析失败:', e);
  }
}

function adjustCount(delta: number) {
  count.value = Math.max(5, Math.min(100, count.value + delta));
}

async function startPractice() {
  loading.value = true;
  try {
    console.log('[SmartPractice] 开始, mode:', mode.value, 'count:', count.value);
    const res = await getSmartPractice({
      count: count.value,
      mode: mode.value,
    });

    console.log('[SmartPractice] API响应类型:', typeof res);
    console.log('[SmartPractice] API响应:', JSON.stringify(res).substring(0, 200));

    // 处理不同的响应格式
    let questions: any[] = [];
    if (Array.isArray(res)) {
      questions = res;
      console.log('[SmartPractice] 格式1: 直接数组, 长度:', questions.length);
    } else if (res && typeof res === 'object') {
      if (Array.isArray((res as any).questions)) {
        questions = (res as any).questions;
        console.log('[SmartPractice] 格式2: res.questions, 长度:', questions.length);
      } else if (Array.isArray((res as any).items)) {
        questions = (res as any).items;
        console.log('[SmartPractice] 格式3: res.items, 长度:', questions.length);
      } else if (Array.isArray((res as any).data)) {
        questions = (res as any).data;
        console.log('[SmartPractice] 格式4: res.data数组, 长度:', questions.length);
      } else {
        console.log('[SmartPractice] 未知格式, keys:', Object.keys(res as any));
      }
    }

    if (questions.length === 0) {
      uni.showToast({ title: '暂无题目', icon: 'none' });
      return;
    }

    // 使用 storage 传递数据，避免 URL 长度限制
    uni.setStorageSync('smartPracticeData', JSON.stringify(questions));
    uni.navigateTo({
      url: '/pages/practice/smart-session',
    });
  } catch (e) {
    console.error('[SmartPractice] 获取失败:', e);
    uni.showToast({ title: '获取题目失败', icon: 'none' });
  } finally {
    loading.value = false;
  }
}
</script>

<style lang="scss" scoped>
.smart-practice-page { min-height: 100vh; background: #f5f5f5; padding: 16px; }

.header { margin-bottom: 16px; }
.title { font-size: 22px; font-weight: bold; color: #333; }

.analysis-card { background: #fff; border-radius: 12px; padding: 20px; margin-bottom: 16px; }
.analysis-title { font-size: 16px; font-weight: 500; color: #333; margin-bottom: 16px; display: block; }
.stats-row { display: flex; gap: 16px; margin-bottom: 16px; }
.stat-item { flex: 1; text-align: center; padding: 12px; background: #f9f9f9; border-radius: 8px; }
.stat-value { display: block; font-size: 20px; font-weight: bold; color: #1890ff; }
.stat-label { display: block; font-size: 12px; color: #999; margin-top: 4px; }
.weak-title { font-size: 14px; font-weight: 500; color: #333; margin-bottom: 8px; display: block; }
.weak-hint { font-size: 12px; color: #999; display: block; }
.weak-tags { display: flex; flex-wrap: wrap; gap: 8px; }
.weak-tag { padding: 4px 12px; background: #fff3e0; color: #f57c00; border-radius: 12px; font-size: 12px; }

.section-title { font-size: 16px; font-weight: 500; color: #333; margin-bottom: 12px; display: block; }
.mode-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 20px; }
.mode-card { background: #fff; border-radius: 12px; padding: 16px 8px; text-align: center; border: 2px solid transparent; }
.mode-card.selected { border-color: #1890ff; background: #e6f7ff; }
.mode-icon { font-size: 24px; display: block; margin-bottom: 8px; }
.mode-label { font-size: 13px; font-weight: 500; color: #333; display: block; margin-bottom: 4px; }
.mode-desc { font-size: 11px; color: #999; }

.count-section { background: #fff; border-radius: 12px; padding: 16px; margin-bottom: 20px; }
.count-selector { display: flex; align-items: center; gap: 16px; }
.count-btn { width: 36px; height: 36px; background: #f5f7fa; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 18px; }
.count-value { font-size: 20px; font-weight: bold; min-width: 40px; text-align: center; }

.start-btn { height: 48px; line-height: 48px; font-size: 17px; border-radius: 24px; background: linear-gradient(90deg, #1890ff, #36cfc9); border: none; }
</style>
