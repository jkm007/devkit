<template>
  <view class="smart-session-page">
    <view class="progress-bar">
      <view class="progress-info">
        <text class="question-index">{{ currentIndex + 1 }}/{{ questions.length }}</text>
        <view class="progress-track"><view class="progress-fill" :style="{ width: ((currentIndex + 1) / questions.length * 100) + '%' }" /></view>
        <text class="timer">{{ formatTime(elapsed) }}</text>
      </view>
    </view>

    <view v-if="currentQuestion" class="question-content">
      <view class="stem">
        <ContentBlockRenderer v-if="currentQuestion.stemBlocks?.length" :blocks="currentQuestion.stemBlocks" />
        <text v-else>{{ currentQuestion.title }}</text>
      </view>
      <view v-if="currentQuestion.options" class="options">
        <view v-for="opt in currentQuestion.options" :key="opt.label" class="option-item" :class="{ selected: answers[currentIndex] === opt.label }" @click="selectAnswer(opt.label)">
          <text class="label">{{ opt.label }}.</text>
          <ContentBlockRenderer v-if="opt.contentBlocks?.length" :blocks="opt.contentBlocks" />
          <text v-else class="content">{{ opt.content }}</text>
        </view>
      </view>
      <view v-if="currentQuestion.analysis" class="analysis-section">
        <text class="analysis-title">解析</text>
        <ContentBlockRenderer v-if="typeof currentQuestion.analysis === 'object'" :blocks="currentQuestion.analysis" />
        <text v-else>{{ currentQuestion.analysis }}</text>
      </view>
    </view>

    <view class="bottom-nav">
      <button class="nav-btn" :disabled="currentIndex === 0" @click="prev">上一题</button>
      <button v-if="currentIndex === questions.length - 1" class="submit-btn" @click="submitPractice">完成</button>
      <button v-else class="next-btn" @click="next">下一题</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import ContentBlockRenderer from '@/components/ContentBlockRenderer.vue';

const questions = ref<any[]>([]);
const currentIndex = ref(0);
const answers = ref<string[]>([]);
const elapsed = ref(0);
let timer: ReturnType<typeof setInterval> | null = null;

onMounted(() => {
  const pages = getCurrentPages();
  const currentPage = pages[pages.length - 1] as any;
  const dataStr = currentPage.options?.data;
  if (dataStr) {
    try { questions.value = JSON.parse(decodeURIComponent(dataStr)); } catch { /* ignore */ }
  }
  if (questions.value.length === 0) {
    if (import.meta.env.DEV) {
      questions.value = Array.from({ length: 10 }, (_, i) => ({
        id: i + 1, title: `智能练习题 ${i + 1}`,
        options: [
          { label: 'A', content: '选项 A' }, { label: 'B', content: '选项 B' },
          { label: 'C', content: '选项 C' }, { label: 'D', content: '选项 D' },
        ],
      }));
    } else {
      uni.showToast({ title: '加载练习失败', icon: 'none' });
    }
  }
  answers.value = new Array(questions.value.length).fill('');
  timer = setInterval(() => { elapsed.value++; }, 1000);
});

onUnmounted(() => { if (timer) clearInterval(timer); });

const currentQuestion = computed(() => questions.value[currentIndex.value] || null);
function selectAnswer(label: string) { answers.value[currentIndex.value] = label; }
function prev() { if (currentIndex.value > 0) currentIndex.value--; }
function next() { if (currentIndex.value < questions.value.length - 1) currentIndex.value++; }
function formatTime(s: number): string { const m = Math.floor(s / 60); const sec = s % 60; return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`; }

function submitPractice() {
  if (timer) clearInterval(timer);
  const answered = answers.value.filter(a => a).length;
  uni.showModal({
    title: '练习完成',
    content: `已完成 ${answered}/${questions.value.length} 题，用时 ${formatTime(elapsed.value)}`,
    success: (res) => { if (res.confirm) uni.navigateBack(); },
  });
}
</script>

<style lang="scss" scoped>
.smart-session-page { min-height: 100vh; background: #f5f5f5; padding-bottom: 70px; }
.progress-bar { background: #fff; padding: 8px 16px; }
.progress-info { display: flex; align-items: center; gap: 12px; }
.question-index { font-size: 14px; font-weight: 500; }
.progress-track { flex: 1; height: 4px; background: #eee; border-radius: 2px; }
.progress-fill { height: 100%; background: #52c41a; border-radius: 2px; }
.timer { font-size: 14px; color: #52c41a; font-family: monospace; }
.question-content { background: #fff; margin: 12px 16px; border-radius: 12px; padding: 20px; }
.stem { font-size: 16px; line-height: 1.6; margin-bottom: 20px; }
.option-item { display: flex; padding: 14px; margin-bottom: 10px; background: #f9f9f9; border-radius: 8px; border: 2px solid transparent; }
.option-item.selected { border-color: #52c41a; background: #f6ffed; }
.label { font-weight: 500; margin-right: 10px; }
.content { flex: 1; font-size: 15px; }
.analysis-section { margin-top: 20px; padding-top: 16px; border-top: 1px solid #eee; }
.analysis-title { font-size: 14px; font-weight: 500; color: #52c41a; margin-bottom: 8px; display: block; }
.bottom-nav { position: fixed; bottom: 0; left: 0; right: 0; display: flex; background: #fff; padding: 12px 16px; box-shadow: 0 -2px 10px rgba(0,0,0,0.05); }
.nav-btn { flex: 1; height: 44px; border: none; border-radius: 22px; font-size: 15px; background: #f5f7fa; color: #666; margin-right: 8px; }
.next-btn { flex: 1; height: 44px; border: none; border-radius: 22px; font-size: 15px; background: #52c41a; color: #fff; }
.submit-btn { flex: 1; height: 44px; border: none; border-radius: 22px; font-size: 15px; background: #1890ff; color: #fff; }
</style>
