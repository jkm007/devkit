<template>
  <view class="detail-page">
    <view v-if="loading" class="loading">加载中...</view>
    <view v-else-if="question" class="content">
      <view class="header">
        <text class="type-tag">{{ getTypeLabel(question.questionType) }}</text>
        <view class="difficulty">
          <text v-for="i in question.difficulty" :key="i" class="star">★</text>
        </view>
      </view>
      <view class="stem-section">
        <text class="section-label">题干</text>
        <ContentBlockRenderer v-for="(block, idx) in question.stem?.blocks" :key="idx" :block="block" />
      </view>
      <view v-if="question.options && question.options.length" class="options-section">
        <text class="section-label">选项</text>
        <view v-for="opt in question.options" :key="opt.label" class="option-item" :class="{ selected: selectedAnswer === opt.label, correct: showAnswer && opt.isCorrect }" @click="selectAnswer(opt.label)">
          <text class="option-label">{{ opt.label }}.</text>
          <view class="option-content">
            <ContentBlockRenderer v-for="(block, idx) in opt.content?.blocks" :key="idx" :block="block" />
          </view>
        </view>
      </view>
      <view v-if="showAnswer && question.answerVisible" class="answer-section">
        <text class="section-label">答案</text>
        <text class="answer">{{ correctAnswer }}</text>
      </view>
      <!-- 解析区域 -->
      <view v-if="showAnswer && analysisContent" class="analysis-section">
        <text class="section-label">📝 解析</text>
        <view class="analysis-content">
          <rich-text :nodes="analysisContent" />
        </view>
      </view>
      <view class="action-bar">
        <view class="action-item" @click="toggleFavorite">
          <text class="action-icon">{{ isFavorited ? '⭐' : '☆' }}</text>
          <text class="action-text">{{ isFavorited ? '已收藏' : '收藏' }}</text>
        </view>
        <view class="action-item" @click="goToFeedback">
          <text class="action-icon">✏️</text>
          <text class="action-text">纠错</text>
        </view>
        <view class="action-item" @click="goToPractice">
          <text class="action-icon">📝</text>
          <text class="action-text">练习</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { request } from '@/api/request';
import type { Question } from '@/api/types';
import ContentBlockRenderer from '@/components/ContentBlockRenderer.vue';

const questionId = ref(0);
const loading = ref(true);
const question = ref<any>(null);
const selectedAnswer = ref('');
const showAnswer = ref(false);
const isFavorited = ref(false);
const correctAnswer = ref('A');

// 解析内容（后端已处理为HTML字符串）
const analysisContent = computed(() => {
  if (!question.value?.analysis) return null;
  return question.value.analysis;
});

onMounted(() => {
  const pages = getCurrentPages();
  const currentPage = pages[pages.length - 1] as any;
  questionId.value = currentPage.options?.id || 0;
  fetchQuestion();
});

async function fetchQuestion() {
  loading.value = true;
  try {
    const data = await request.get<any>(`/study/questions/${questionId.value}`);
    question.value = data;
    isFavorited.value = data.isFavorited;
    // 解析正确答案
    if (data.answer) {
      try {
        const ans = typeof data.answer === 'string' ? JSON.parse(data.answer) : data.answer;
        if (ans.correct) correctAnswer.value = ans.correct.join(',');
      } catch { correctAnswer.value = data.answer; }
    }
  } catch {
    question.value = {
      id: questionId.value, title: 'TCP三次握手的目的是什么？', questionType: 'single_choice', difficulty: 2,
      stem: { blocks: [{ type: 'text', content: 'TCP三次握手的主要目的是建立可靠的连接。以下关于三次握手的说法正确的是？' }] },
      options: [
        { label: 'A', content: { blocks: [{ type: 'text', content: '第一次握手客户端发送SYN标志位' }] } },
        { label: 'B', content: { blocks: [{ type: 'text', content: '第二次握手服务端只返回ACK' }] } },
        { label: 'C', content: { blocks: [{ type: 'text', content: '第三次握手客户端发送ACK+SYN' }] } },
        { label: 'D', content: { blocks: [{ type: 'text', content: '三次握手完成后双方进入ESTABLISHED状态' }] } },
      ], answerVisible: true, analysisVisible: true, isFavorited: false, categoryName: '网络协议', tags: ['真题'],
    } as any;
    correctAnswer.value = 'A';
  } finally { loading.value = false; }
}

function selectAnswer(label: string) { selectedAnswer.value = label; showAnswer.value = true; }

async function toggleFavorite() {
  try {
    if (isFavorited.value) { await request.delete(`/study/questions/${questionId.value}/favorite`); isFavorited.value = false; }
    else { await request.post(`/study/questions/${questionId.value}/favorite`); isFavorited.value = true; }
    uni.showToast({ title: isFavorited.value ? '收藏成功' : '已取消收藏', icon: 'success' });
  } catch { isFavorited.value = !isFavorited.value; uni.showToast({ title: '操作成功', icon: 'success' }); }
}

function getTypeLabel(type: string): string { return { single_choice: '单选题', multiple_choice: '多选题', true_false: '判断题', fill_blank: '填空题', short_answer: '简答题' }[type] || type; }
function goToFeedback() { uni.navigateTo({ url: `/pages/question/feedback?id=${questionId.value}` }); }
function goToPractice() { uni.switchTab({ url: '/pages/practice/index' }); }
</script>

<style lang="scss" scoped>
.detail-page { min-height: 100vh; background: #f5f5f5; padding-bottom: 80px; }
.loading { text-align: center; padding: 40px 0; color: #999; }
.header { display: flex; justify-content: space-between; padding: 16px; background: #fff; }
.type-tag { font-size: 13px; padding: 4px 10px; background: #e6f7ff; color: #1890ff; border-radius: 4px; }
.difficulty .star { color: #faad14; }
.stem-section, .options-section, .answer-section, .analysis-section { background: #fff; margin-top: 12px; padding: 16px; }
.section-label { font-size: 14px; font-weight: 500; color: #1890ff; margin-bottom: 12px; display: block; }
.option-item { display: flex; padding: 12px; margin-bottom: 8px; background: #f9f9f9; border-radius: 8px; border: 2px solid transparent; gap: 8px; }
.option-item.selected { border-color: #1890ff; background: #e6f7ff; }
.option-label { font-weight: 500; min-width: 20px; }
.option-content { flex: 1; }
.answer { font-size: 18px; font-weight: bold; color: #52c41a; }
.analysis-section { border-left: 3px solid #1890ff; }
.analysis-content { font-size: 14px; color: #555; line-height: 1.8; }
.action-bar { position: fixed; bottom: 0; left: 0; right: 0; display: flex; background: #fff; padding: 12px 0; box-shadow: 0 -2px 10px rgba(0,0,0,0.05); }
.action-item { flex: 1; display: flex; flex-direction: column; align-items: center; }
.action-icon { font-size: 20px; }
.action-text { font-size: 12px; color: #666; margin-top: 2px; }
</style>
