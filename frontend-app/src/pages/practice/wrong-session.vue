<template>
  <view class="wrong-session-page">
    <view class="progress-bar">
      <view class="progress-info">
        <text class="question-index">{{ currentIndex + 1 }}/{{ questions.length }}</text>
        <view class="progress-track"><view class="progress-fill" :style="{ width: ((currentIndex + 1) / questions.length * 100) + '%' }" /></view>
      </view>
    </view>

    <view v-if="currentQuestion" class="question-content">
      <view class="stem"><text>{{ currentQuestion.title }}</text></view>
      <view v-if="currentQuestion.options && currentQuestion.options.length" class="options">
        <view v-for="opt in currentQuestion.options" :key="opt.label" class="option-item" :class="{ selected: answers[currentIndex] === opt.label }" @click="selectAnswer(opt.label)">
          <text class="label">{{ opt.label }}.</text>
          <text class="content">{{ opt.content }}</text>
        </view>
      </view>
    </view>

    <view class="bottom-nav">
      <button class="nav-btn" :disabled="currentIndex === 0" @click="prev">上一题</button>
      <button v-if="currentIndex === questions.length - 1" class="submit-btn" @click="submitReview">完成复习</button>
      <button v-else class="next-btn" @click="next">下一题</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getWrongBookRandom } from '@/api/study';

const questions = ref<any[]>([]);
const currentIndex = ref(0);
const answers = ref<string[]>([]);
const questionIds = ref<number[]>([]);

onMounted(() => {
  // 从 setStorageSync 读取错题 IDs
  const idsStr = uni.getStorageSync('wrongSessionIds');
  if (idsStr) {
    try { questionIds.value = JSON.parse(idsStr); } catch { /* ignore */ }
  }
  loadQuestions();
});

async function loadQuestions() {
  try {
    const res = await getWrongBookRandom(questionIds.value.length || 20);
    questions.value = (res || []) as any[];
    answers.value = new Array(questions.value.length).fill('');
  } catch {
    if (import.meta.env.DEV) {
      questions.value = Array.from({ length: 10 }, (_, i) => ({
        id: i + 1, title: `错题 ${i + 1}`, options: [
          { label: 'A', content: '选项 A' }, { label: 'B', content: '选项 B' },
          { label: 'C', content: '选项 C' }, { label: 'D', content: '选项 D' },
        ],
      }));
      answers.value = new Array(questions.value.length).fill('');
    } else {
      uni.showToast({ title: '加载错题失败', icon: 'none' });
    }
  }
}

const currentQuestion = computed(() => questions.value[currentIndex.value] || null);
function selectAnswer(label: string) { answers.value[currentIndex.value] = label; }
function prev() { if (currentIndex.value > 0) currentIndex.value--; }
function next() { if (currentIndex.value < questions.value.length - 1) currentIndex.value++; }

function submitReview() {
  const answered = answers.value.filter(a => a).length;
  uni.showModal({
    title: '完成复习',
    content: `已完成 ${answered}/${questions.value.length} 题，确定结束吗？`,
    success: (res) => {
      if (res.confirm) {
        uni.navigateBack();
      }
    },
  });
}
</script>

<style lang="scss" scoped>
.wrong-session-page { min-height: 100vh; background: #f5f5f5; padding-bottom: 70px; }
.progress-bar { background: #fff; padding: 12px 16px; }
.progress-info { display: flex; align-items: center; gap: 12px; }
.question-index { font-size: 14px; font-weight: 500; }
.progress-track { flex: 1; height: 4px; background: #eee; border-radius: 2px; }
.progress-fill { height: 100%; background: #1890ff; border-radius: 2px; }
.question-content { background: #fff; margin: 12px 16px; border-radius: 12px; padding: 20px; }
.stem { font-size: 16px; line-height: 1.6; margin-bottom: 20px; }
.option-item { display: flex; padding: 14px; margin-bottom: 10px; background: #f9f9f9; border-radius: 8px; border: 2px solid transparent; }
.option-item.selected { border-color: #1890ff; background: #e6f7ff; }
.label { font-weight: 500; margin-right: 10px; }
.content { flex: 1; font-size: 15px; }
.bottom-nav { position: fixed; bottom: 0; left: 0; right: 0; display: flex; background: #fff; padding: 12px 16px; box-shadow: 0 -2px 10px rgba(0,0,0,0.05); }
.nav-btn { flex: 1; height: 44px; border: none; border-radius: 22px; font-size: 15px; background: #f5f7fa; color: #666; margin-right: 8px; }
.next-btn { flex: 1; height: 44px; border: none; border-radius: 22px; font-size: 15px; background: #1890ff; color: #fff; }
.submit-btn { flex: 1; height: 44px; border: none; border-radius: 22px; font-size: 15px; background: #52c41a; color: #fff; }
</style>
