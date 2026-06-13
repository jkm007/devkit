<template>
  <view class="practice-page">
    <view class="header">
      <text class="title">开始练习</text>
    </view>

    <!-- 练习模式选择 -->
    <view class="mode-section">
      <text class="section-title">选择练习模式</text>
      <view class="mode-cards">
        <view class="mode-card" :class="{ selected: mode === 'random' }" @click="mode = 'random'">
          <text class="mode-icon">🎲</text>
          <text class="mode-label">随机练习</text>
        </view>
        <view class="mode-card" :class="{ selected: mode === 'category' }" @click="mode = 'category'">
          <text class="mode-icon">📂</text>
          <text class="mode-label">分类练习</text>
        </view>
        <view class="mode-card" :class="{ selected: mode === 'knowledge' }" @click="mode = 'knowledge'">
          <text class="mode-icon">🎯</text>
          <text class="mode-label">知识点练习</text>
        </view>
      </view>
    </view>

    <!-- 条件筛选 -->
    <view class="filter-section">
      <text class="section-title">练习条件</text>

      <view class="filter-item">
        <text class="label">题目数量</text>
        <view class="count-selector">
          <view class="count-btn" @click="adjustCount(-5)">−</view>
          <text class="count-value">{{ questionCount }}</text>
          <view class="count-btn" @click="adjustCount(5)">+</view>
        </view>
      </view>

      <view class="filter-item">
        <text class="label">难度</text>
        <view class="difficulty-selector">
          <view v-for="d in [1,2,3]" :key="d" class="diff-btn" :class="{ active: difficulty === d }" @click="difficulty = d">
            {{ ['简单', '中等', '困难'][d-1] }}
          </view>
        </view>
      </view>

      <view class="filter-item">
        <text class="label">题型</text>
        <view class="type-selector">
          <view v-for="t in typeOptions" :key="t.value" class="type-btn" :class="{ active: selectedTypes.includes(t.value) }" @click="toggleType(t.value)">
            {{ t.label }}
          </view>
        </view>
      </view>
    </view>

    <!-- 开始按钮 -->
    <button class="start-btn" type="primary" @click="startPractice">开始练习</button>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const mode = ref('random');
const questionCount = ref(20);
const difficulty = ref(0); // 0=全部
const selectedTypes = ref<string[]>([]);

const typeOptions = [
  { label: '单选', value: 'single_choice' },
  { label: '多选', value: 'multiple_choice' },
  { label: '判断', value: 'true_false' },
  { label: '填空', value: 'fill_blank' },
];

function adjustCount(delta: number) {
  questionCount.value = Math.max(5, Math.min(100, questionCount.value + delta));
}

function toggleType(type: string) {
  const idx = selectedTypes.value.indexOf(type);
  if (idx >= 0) selectedTypes.value.splice(idx, 1);
  else selectedTypes.value.push(type);
}

function startPractice() {
  const params = {
    mode: mode.value,
    count: questionCount.value,
    difficulty: difficulty.value,
    types: selectedTypes.value,
  };
  uni.navigateTo({
    url: `/pages/practice/session?params=${encodeURIComponent(JSON.stringify(params))}`,
  });
}
</script>

<style lang="scss" scoped>
.practice-page { min-height: 100vh; background: #f5f5f5; padding: 16px; }
.header { margin-bottom: 20px; .title { font-size: 22px; font-weight: bold; color: #333; } }
.section-title { font-size: 16px; font-weight: 500; color: #333; margin-bottom: 12px; display: block; }
.mode-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 20px; }
.mode-card { background: #fff; border-radius: 12px; padding: 20px 12px; text-align: center; border: 2px solid transparent; }
.mode-card.selected { border-color: #1890ff; background: #e6f7ff; }
.mode-icon { font-size: 28px; display: block; margin-bottom: 8px; }
.mode-label { font-size: 13px; color: #666; }
.filter-section { background: #fff; border-radius: 12px; padding: 16px; margin-bottom: 20px; }
.filter-item { margin-bottom: 16px; .label { font-size: 14px; color: #666; margin-bottom: 8px; display: block; } }
.count-selector { display: flex; align-items: center; gap: 16px; }
.count-btn { width: 36px; height: 36px; background: #f5f7fa; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 18px; }
.count-value { font-size: 20px; font-weight: bold; min-width: 40px; text-align: center; }
.difficulty-selector, .type-selector { display: flex; gap: 8px; flex-wrap: wrap; }
.diff-btn, .type-btn { padding: 8px 16px; background: #f5f7fa; border-radius: 16px; font-size: 13px; color: #666; }
.diff-btn.active, .type-btn.active { background: #1890ff; color: #fff; }
.start-btn { height: 48px; line-height: 48px; font-size: 17px; border-radius: 24px; background: linear-gradient(90deg, #1890ff, #36cfc9); border: none; }
</style>
