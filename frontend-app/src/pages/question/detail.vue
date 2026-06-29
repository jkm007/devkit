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

      <!-- 题干 -->
      <view class="stem-section">
        <ContentBlockRenderer v-for="(block, idx) in question.stem?.blocks" :key="idx" :block="block" />
      </view>

      <!-- 选择题选项 -->
      <view v-if="isChoiceQuestion(question.questionType) && question.options && question.options.length" class="options-section">
        <view v-for="opt in question.options" :key="opt.label" class="option-item" :class="getOptionClass(opt)" @click="selectAnswer(opt.label)">
          <text class="option-label">{{ opt.label }}.</text>
          <view class="option-content">
            <ContentBlockRenderer v-for="(block, idx) in opt.content?.blocks" :key="idx" :block="block" />
          </view>
          <text v-if="showAnswer && isCorrectOption(opt)" class="option-icon correct">✓</text>
          <text v-else-if="showAnswer && selectedAnswer === opt.label && !isCorrectOption(opt)" class="option-icon wrong">✗</text>
        </view>
      </view>

      <!-- 填空题输入 -->
      <view v-else-if="isFillQuestion(question.questionType)" class="fill-section">
        <textarea v-model="fillAnswer" class="answer-textarea" placeholder="请输入答案" :maxlength="500" :disabled="showAnswer" />
        <button v-if="!showAnswer" class="submit-btn" @click="submitFillAnswer">提交答案</button>
      </view>

      <!-- 简答/论述等题型输入 -->
      <view v-else class="essay-section">
        <textarea v-model="essayAnswer" class="answer-textarea essay" placeholder="请输入你的答案" :maxlength="2000" :disabled="showAnswer" />
        <text class="word-count">{{ (essayAnswer || '').length }} / 2000</text>
        <button v-if="!showAnswer" class="submit-btn" @click="submitEssayAnswer">提交答案</button>
      </view>

      <!-- 答题结果 -->
      <view v-if="showAnswer" class="result-section" :class="isCorrect ? 'correct' : 'wrong'">
        <text class="result-icon">{{ isCorrect ? '🎉' : '😔' }}</text>
        <text class="result-text">{{ isCorrect ? '回答正确！' : '回答错误' }}</text>
      </view>

      <!-- 正确答案 -->
      <view v-if="showAnswer" class="answer-section">
        <text class="section-label">正确答案</text>
        <text class="answer">{{ correctAnswer }}</text>
      </view>

      <!-- 解析区域 -->
      <view v-if="showAnswer && analysisContent" class="analysis-section">
        <text class="section-label">📝 解析</text>
        <view class="analysis-content">
          <rich-text v-if="analysisText" :nodes="analysisText" />
          <view v-for="(img, idx) in analysisImages" :key="'img-'+idx" class="media-item">
            <image :src="img" mode="widthFix" class="analysis-image" @click="previewImage(img)" />
          </view>
          <view v-for="(video, idx) in analysisVideos" :key="'video-'+idx" class="media-item">
            <video :src="video" controls class="analysis-video" />
          </view>
        </view>
      </view>

      <!-- 底部操作栏 -->
      <view class="action-bar">
        <view class="action-item" @click="toggleFavorite">
          <text class="action-icon">{{ isFavorited ? '⭐' : '☆' }}</text>
          <text class="action-text">{{ isFavorited ? '已收藏' : '收藏' }}</text>
        </view>
        <view class="action-item" @click="loadNextQuestion">
          <text class="action-icon">➡️</text>
          <text class="action-text">下一题</text>
        </view>
        <view class="action-item" @click="goToPractice">
          <text class="action-icon">📝</text>
          <text class="action-text">练习模式</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { onLoad } from '@dcloudio/uni-app';
import { request, tokenManager } from '@/api/request';
import { QUESTION_TYPE_LABELS, CHOICE_TYPES, FILL_TYPES, type Question } from '@/api/types';
import ContentBlockRenderer from '@/components/ContentBlockRenderer.vue';

// 为需要认证的文件 URL 添加 token（小程序 <image>/<video> 不支持 header/cookie 认证）
function addTokenToUrl(url: string): string {
  if (!url) return url;
  // 只给本后端 /api/v1/files/ 的代理链接加 token；外部 presigned URL 不加
  const isBackendFileUrl =
    url.startsWith('/api/v1/files/') ||
    (url.startsWith('http') && url.includes('/api/v1/files/'));
  if (!isBackendFileUrl) return url;
  const token = tokenManager.getAccessToken();
  if (!token) return url;
  const separator = url.includes('?') ? '&' : '?';
  return `${url}${separator}token=${encodeURIComponent(token)}`;
}

const questionId = ref(0);
const loading = ref(true);
const question = ref<any>(null);
const selectedAnswer = ref('');
const showAnswer = ref(false);
const isFavorited = ref(false);
const isCorrect = ref(false);
const correctAnswer = ref('A');
const fillAnswer = ref('');
const essayAnswer = ref('');
const startTime = ref(0);

// 选择题类型、填空题类型从 @/api/types 引入

function isChoiceQuestion(type: string): boolean {
  return CHOICE_TYPES.includes(type as any);
}

function isFillQuestion(type: string): boolean {
  return FILL_TYPES.includes(type as any);
}

// API基础URL
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

// 解析内容处理
const analysisContent = computed(() => {
  return question.value?.analysis || null;
});

const analysisText = computed(() => {
  if (!analysisContent.value) return '';
  let html = analysisContent.value;
  html = html.replace(/<video[^>]*>.*?<\/video>/gi, '');
  html = html.replace(/<img[^>]*>/gi, '');
  return html.trim();
});

const analysisImages = computed(() => {
  if (!analysisContent.value) return [];
  const images: string[] = [];
  const imgRegex = /<img[^>]+src="([^"]+)"/gi;
  let match;
  while ((match = imgRegex.exec(analysisContent.value)) !== null) {
    let src = match[1];
    if (src.startsWith('/')) {
      src = apiBaseUrl.replace('/api/v1', '') + src;
    }
    images.push(addTokenToUrl(src));
  }
  return images;
});

const analysisVideos = computed(() => {
  if (!analysisContent.value) return [];
  const videos: string[] = [];
  const videoRegex = /<video[^>]+src="([^"]+)"/gi;
  let match;
  while ((match = videoRegex.exec(analysisContent.value)) !== null) {
    let src = match[1];
    if (src.startsWith('/')) {
      src = apiBaseUrl.replace('/api/v1', '') + src;
    }
    videos.push(addTokenToUrl(src));
  }
  return videos;
});

onLoad((options: any) => {
  questionId.value = options?.id || 0;
  fetchQuestion();
});

async function fetchQuestion() {
  loading.value = true;
  showAnswer.value = false;
  selectedAnswer.value = '';
  fillAnswer.value = '';
  essayAnswer.value = '';
  isCorrect.value = false;
  startTime.value = Date.now();

  try {
    const data = await request.get<any>(`/study/questions/${questionId.value}`);
    question.value = data;
    isFavorited.value = data.isFavorited;
    // 解析正确答案
    if (data.answer) {
      try {
        const ans = typeof data.answer === 'string' ? JSON.parse(data.answer) : data.answer;
        // 选择题格式: {"correct": ["A"]} 或 {"correct": ["A","B"]}
        if (ans.correct && Array.isArray(ans.correct)) {
          correctAnswer.value = ans.correct.join(',');
        }
        // 简答题/论述题等格式: {"text": "答案内容"} 或直接是文本
        else if (ans.text) {
          correctAnswer.value = ans.text;
        }
        // 其他格式
        else {
          correctAnswer.value = typeof data.answer === 'string' ? data.answer : JSON.stringify(ans);
        }
      } catch {
        // 解析失败，直接显示原始内容
        correctAnswer.value = data.answer;
      }
    }
  } catch (e) {
    console.error('获取题目失败:', e);
    question.value = null;
    uni.showToast({ title: '获取题目失败', icon: 'none' });
  } finally { loading.value = false; }
}

function selectAnswer(label: string) {
  if (showAnswer.value) return; // 已答题不能再选
  selectedAnswer.value = label;
  showAnswer.value = true;
  isCorrect.value = checkAnswer(label, correctAnswer.value);
  submitAnswerRecord();
}

function submitFillAnswer() {
  if (!fillAnswer.value.trim()) {
    uni.showToast({ title: '请输入答案', icon: 'none' });
    return;
  }
  showAnswer.value = true;
  // 填空题暂不自动判对错
  isCorrect.value = false;
  submitAnswerRecord();
}

function submitEssayAnswer() {
  if (!essayAnswer.value.trim()) {
    uni.showToast({ title: '请输入答案', icon: 'none' });
    return;
  }
  showAnswer.value = true;
  // 简答题暂不自动判对错
  isCorrect.value = false;
  submitAnswerRecord();
}

function checkAnswer(userAnswer: string, correct: string): boolean {
  if (!correct || !userAnswer) return false;
  // 标准化答案
  const normalize = (ans: string) => ans.trim().toUpperCase();
  return normalize(userAnswer) === normalize(correct);
}

function isCorrectOption(opt: any): boolean {
  // 判断选项是否是正确答案
  const correctLabels = correctAnswer.value.split(',').map(l => l.trim().toUpperCase());
  return correctLabels.includes(opt.label.toUpperCase());
}

function getOptionClass(opt: any): string {
  if (!showAnswer.value) {
    return selectedAnswer.value === opt.label ? 'selected' : '';
  }
  if (isCorrectOption(opt)) return 'correct';
  if (selectedAnswer.value === opt.label && !isCorrectOption(opt)) return 'wrong';
  return '';
}

async function submitAnswerRecord() {
  try {
    const elapsed = startTime.value > 0 ? Math.max(0, Math.floor((Date.now() - startTime.value) / 1000)) : 0;
    await request.post('/study/practice/submit', {
      total: 1,
      answered: 1,
      correct: isCorrect.value ? 1 : 0,
      elapsed,
      answers: [selectedAnswer.value || fillAnswer.value || essayAnswer.value],
    });
  } catch (e) {
    console.error('提交答题记录失败:', e);
  }
}

async function loadNextQuestion() {
  loading.value = true;
  try {
    // 获取随机题目
    const res = await request.post<any>('/study/practice/questions', {
      count: 1,
      mode: 'random',
    });
    let nextId = 0;
    if (Array.isArray(res) && res.length > 0) {
      nextId = res[0].id;
    } else if (res?.questions && res.questions.length > 0) {
      nextId = res.questions[0].id;
    }
    if (nextId && nextId !== questionId.value) {
      questionId.value = nextId;
      await fetchQuestion();
    } else {
      uni.showToast({ title: '没有更多题目了', icon: 'none' });
    }
  } catch (e) {
    console.error('获取下一题失败:', e);
    uni.showToast({ title: '获取题目失败', icon: 'none' });
  } finally {
    loading.value = false;
  }
}

async function toggleFavorite() {
  try {
    if (isFavorited.value) {
      await request.delete(`/study/questions/${questionId.value}/favorite`);
      isFavorited.value = false;
    } else {
      await request.post(`/study/questions/${questionId.value}/favorite`);
      isFavorited.value = true;
    }
    uni.showToast({ title: isFavorited.value ? '收藏成功' : '已取消收藏', icon: 'success' });
  } catch {
    isFavorited.value = !isFavorited.value;
  }
}

function getTypeLabel(type: string): string {
  return QUESTION_TYPE_LABELS[type as keyof typeof QUESTION_TYPE_LABELS] || type;
}

function goToPractice() {
  uni.switchTab({ url: '/pages/practice/index' });
}

function previewImage(url: string) {
  uni.previewImage({
    urls: analysisImages.value,
    current: url,
  });
}
</script>

<style lang="scss" scoped>
.detail-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 80px;
}

.loading {
  text-align: center;
  padding: 40px 0;
  color: #999;
}

.header {
  display: flex;
  justify-content: space-between;
  padding: 16px;
  background: #fff;
}

.type-tag {
  font-size: 13px;
  padding: 4px 10px;
  background: #e6f7ff;
  color: #1890ff;
  border-radius: 4px;
}

.difficulty .star {
  color: #faad14;
}

.stem-section {
  background: #fff;
  padding: 16px;
  font-size: 16px;
  line-height: 1.8;
}

.options-section {
  background: #fff;
  margin-top: 12px;
  padding: 16px;
}

.option-item {
  display: flex;
  padding: 14px;
  margin-bottom: 10px;
  background: #f9f9f9;
  border-radius: 10px;
  border: 2px solid transparent;
  gap: 10px;
  align-items: center;
  transition: all 0.2s;

  &:active {
    transform: scale(0.98);
  }

  &.selected {
    border-color: #1890ff;
    background: #e6f7ff;
  }

  &.correct {
    border-color: #52c41a;
    background: #f6ffed;
  }

  &.wrong {
    border-color: #ff4d4f;
    background: #fff2f0;
  }
}

.option-label {
  font-weight: 600;
  min-width: 24px;
  font-size: 15px;
}

.option-content {
  flex: 1;
  font-size: 15px;
  line-height: 1.5;
}

.option-icon {
  font-size: 18px;
  font-weight: bold;
  &.correct { color: #52c41a; }
  &.wrong { color: #ff4d4f; }
}

/* 填空题和简答题 */
.fill-section, .essay-section {
  background: #fff;
  margin-top: 12px;
  padding: 16px;
}

.answer-textarea {
  width: 100%;
  min-height: 100px;
  padding: 14px;
  background: #f9f9f9;
  border-radius: 10px;
  border: 1px solid #e8e8e8;
  font-size: 15px;
  line-height: 1.6;

  &:disabled {
    background: #f5f5f5;
    color: #666;
  }
}

.answer-textarea.essay {
  min-height: 180px;
}

.word-count {
  display: block;
  text-align: right;
  font-size: 12px;
  color: #999;
  margin-top: 8px;
}

.submit-btn {
  margin-top: 16px;
  height: 48px;
  line-height: 48px;
  background: linear-gradient(135deg, #1890ff, #36cfc9);
  color: #fff;
  border: none;
  border-radius: 24px;
  font-size: 16px;
  font-weight: 500;
}

/* 答题结果 */
.result-section {
  margin-top: 12px;
  padding: 20px;
  text-align: center;
  border-radius: 12px;

  &.correct {
    background: linear-gradient(135deg, #f6ffed, #d9f7be);
  }

  &.wrong {
    background: linear-gradient(135deg, #fff2f0, #ffccc7);
  }
}

.result-icon {
  font-size: 36px;
  display: block;
  margin-bottom: 8px;
}

.result-text {
  font-size: 18px;
  font-weight: bold;
  &.correct { color: #52c41a; }
  &.wrong { color: #ff4d4f; }
}

/* 正确答案 */
.answer-section {
  background: #fff;
  margin-top: 12px;
  padding: 16px;
}

.section-label {
  font-size: 14px;
  font-weight: 500;
  color: #1890ff;
  margin-bottom: 12px;
  display: block;
}

.answer {
  font-size: 20px;
  font-weight: bold;
  color: #52c41a;
}

/* 解析 */
.analysis-section {
  background: #fff;
  margin-top: 12px;
  padding: 16px;
  border-left: 4px solid #1890ff;
}

.analysis-content {
  font-size: 14px;
  color: #555;
  line-height: 1.8;
}

.media-item {
  margin-top: 12px;
}

.analysis-image {
  width: 100%;
  border-radius: 8px;
}

.analysis-video {
  width: 100%;
  border-radius: 8px;
}

/* 底部操作栏 */
.action-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  background: #fff;
  padding: 12px 0;
  box-shadow: 0 -2px 10px rgba(0,0,0,0.08);
}

.action-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;

  &:active {
    opacity: 0.7;
  }
}

.action-icon {
  font-size: 22px;
}

.action-text {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
}
</style>
