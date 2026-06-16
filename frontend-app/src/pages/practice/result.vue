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
        <text class="stat-label">答对</text>
        <text class="stat-value correct">{{ correctCount }}</text>
      </view>
      <view class="stat-row">
        <text class="stat-label">答错</text>
        <text class="stat-value wrong">{{ wrongCount }}</text>
      </view>
      <view class="stat-row">
        <text class="stat-label">正确率</text>
        <text class="stat-value" :class="correctRate >= 0.6 ? 'correct' : 'wrong'">{{ (correctRate * 100).toFixed(0) }}%</text>
      </view>
      <view class="stat-row">
        <text class="stat-label">用时</text>
        <text class="stat-value">{{ formatTime(result.elapsed) }}</text>
      </view>
    </view>

    <!-- 错题列表 -->
    <view v-if="wrongQuestions.length > 0" class="wrong-section">
      <view class="section-header">
        <text class="section-title">错题详情</text>
        <text class="save-hint" v-if="!savedToWrongBook">正在保存到错题本...</text>
        <text class="save-success" v-else>✅ 已保存到错题本</text>
      </view>
      <view v-for="(q, idx) in wrongQuestions" :key="idx" class="wrong-item">
        <text class="wrong-index">{{ idx + 1 }}.</text>
        <text class="wrong-title">{{ q.title || '题目 ' + q.id }}</text>
        <view class="wrong-answers">
          <text class="your-answer">你的答案: {{ q.userAnswer || '未答' }}</text>
          <text class="correct-answer">正确答案: {{ q.correctAnswer }}</text>
        </view>
      </view>
    </view>

    <view class="action-buttons">
      <button class="action-btn" @click="goToPractice">再来一次</button>
      <button class="action-btn primary" @click="goHome">返回首页</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { request } from '@/api/request';

interface PracticeResult {
  total: number;
  answered: number;
  elapsed: number;
  answers: string[];
  questions?: any[]; // 题目信息
}

const result = ref<PracticeResult>({ total: 0, answered: 0, elapsed: 0, answers: [] });
const correctCount = ref(0);
const wrongCount = ref(0);
const correctRate = computed(() => {
  const total = correctCount.value + wrongCount.value;
  return total > 0 ? correctCount.value / total : 0;
});
const wrongQuestions = ref<any[]>([]);
const savedToWrongBook = ref(false);

onMounted(async () => {
  // 从 storage 读取结果
  const resultStr = uni.getStorageSync('practice_result');
  if (resultStr) {
    try {
      result.value = JSON.parse(resultStr);
      uni.removeStorageSync('practice_result');
    } catch {
      uni.showToast({ title: '数据加载失败', icon: 'none' });
      setTimeout(() => uni.navigateBack(), 1500);
      return;
    }
  }

  // 从 storage 读取题目信息（如果有）
  const questionsStr = uni.getStorageSync('practice_questions');
  if (questionsStr) {
    try {
      result.value.questions = JSON.parse(questionsStr);
      uni.removeStorageSync('practice_questions');
    } catch { /* ignore */ }
  }

  // 计算对错
  calculateResults();

  // 保存错题到错题本
  if (wrongQuestions.value.length > 0) {
    await saveWrongQuestions();
  }
});

function calculateResults() {
  const questions = result.value.questions || [];
  const answers = result.value.answers || [];
  const corrects: any[] = [];
  const wrongs: any[] = [];

  for (let i = 0; i < questions.length; i++) {
    const q = questions[i];
    const userAnswer = answers[i];
    const correctAnswer = q.answer || q.correctAnswer;

    if (!userAnswer) {
      // 未答
      wrongs.push({
        ...q,
        userAnswer: '',
        correctAnswer: correctAnswer,
      });
    } else if (isAnswerCorrect(userAnswer, correctAnswer, q.questionType)) {
      corrects.push(q);
    } else {
      wrongs.push({
        ...q,
        userAnswer: userAnswer,
        correctAnswer: correctAnswer,
      });
    }
  }

  correctCount.value = corrects.length;
  wrongCount.value = wrongs.length;
  wrongQuestions.value = wrongs;
}

function isAnswerCorrect(userAnswer: string, correctAnswer: string, questionType: string): boolean {
  if (!correctAnswer) return false;

  // 标准化答案
  const normalize = (ans: string) => ans.trim().toUpperCase();

  if (questionType === 'multi' || questionType === 'multiple_choice') {
    // 多选题：需要完全匹配
    const userSet = new Set(normalize(userAnswer).split('').sort());
    const correctSet = new Set(normalize(correctAnswer).split('').sort());
    if (userSet.size !== correctSet.size) return false;
    for (const item of userSet) {
      if (!correctSet.has(item)) return false;
    }
    return true;
  }

  // 单选/判断题
  return normalize(userAnswer) === normalize(correctAnswer);
}

async function saveWrongQuestions() {
  try {
    const questionIds = wrongQuestions.value.map(q => q.id).filter(Boolean);
    if (questionIds.length === 0) {
      savedToWrongBook.value = true;
      return;
    }

    await request.post('/study/wrong/batch', {
      questionIds: questionIds,
      categoryId: wrongQuestions.value[0]?.categoryId || 0,
    });
    savedToWrongBook.value = true;
  } catch (e) {
    console.error('保存错题失败:', e);
    savedToWrongBook.value = true; // 即使失败也标记为已处理
  }
}

function formatTime(s: number): string {
  const m = Math.floor(s / 60);
  const sec = s % 60;
  return `${m}分${sec}秒`;
}

function goToPractice() {
  uni.switchTab({ url: '/pages/practice/index' });
}

function goHome() {
  uni.switchTab({ url: '/pages/index/index' });
}
</script>

<style lang="scss" scoped>
.result-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 16px;
}

.header {
  text-align: center;
  padding: 30px 0;

  .title {
    font-size: 24px;
    font-weight: bold;
  }
}

.stats-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
}

.stat-row {
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;

  &:last-child {
    border-bottom: none;
  }
}

.stat-label {
  font-size: 15px;
  color: #666;
}

.stat-value {
  font-size: 18px;
  font-weight: bold;
  color: #333;

  &.correct {
    color: #52c41a;
  }

  &.wrong {
    color: #ff4d4f;
  }
}

.wrong-section {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

.save-hint {
  font-size: 12px;
  color: #999;
}

.save-success {
  font-size: 12px;
  color: #52c41a;
}

.wrong-item {
  padding: 12px;
  background: #fff2f0;
  border-radius: 8px;
  margin-bottom: 8px;
}

.wrong-index {
  font-weight: bold;
  color: #ff4d4f;
  margin-right: 8px;
}

.wrong-title {
  font-size: 14px;
  color: #333;
  display: block;
  margin-bottom: 8px;
}

.wrong-answers {
  display: flex;
  gap: 16px;
  font-size: 13px;
}

.your-answer {
  color: #ff4d4f;
}

.correct-answer {
  color: #52c41a;
}

.action-buttons {
  display: flex;
  gap: 12px;
}

.action-btn {
  flex: 1;
  height: 48px;
  line-height: 48px;
  border: none;
  border-radius: 24px;
  font-size: 16px;
  background: #f5f7fa;
  color: #333;
}

.action-btn.primary {
  background: #1890ff;
  color: #fff;
}
</style>
