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
      <!-- 题干 -->
      <view class="stem">
        <template v-if="currentQuestion.stem && typeof currentQuestion.stem === 'object' && currentQuestion.stem.blocks">
          <view v-for="(block, idx) in currentQuestion.stem.blocks" :key="idx">
            <text v-if="block.type === 'text'">{{ block.content }}</text>
            <image v-else-if="block.type === 'image'" :src="block.content" mode="widthFix" class="stem-image" />
          </view>
        </template>
        <text v-else>{{ currentQuestion.title || '题目加载中...' }}</text>
      </view>

      <!-- 选择题选项 -->
      <view v-if="isChoiceQuestion(currentQuestion) && currentQuestion.options" class="options">
        <view v-for="opt in currentQuestion.options" :key="opt.label" class="option-item" :class="{ selected: answers[currentIndex] === opt.label }" @click="selectAnswer(opt.label)">
          <text class="label">{{ opt.label }}.</text>
          <view class="content">
            <template v-if="opt.content && typeof opt.content === 'object' && opt.content.blocks">
              <text v-for="(block, idx) in opt.content.blocks" :key="idx">{{ block.content }}</text>
            </template>
            <text v-else>{{ typeof opt.content === 'string' ? opt.content : '' }}</text>
          </view>
        </view>
      </view>

      <!-- 填空题 -->
      <view v-else-if="(currentQuestion.questionType || '').includes('fill')" class="fill-section">
        <view class="fill-item">
          <text class="fill-label">请填写答案：</text>
          <input class="fill-input" v-model="fillAnswers[currentIndex]" placeholder="请输入答案" />
        </view>
      </view>

      <!-- 简答题 -->
      <view v-else-if="(currentQuestion.questionType || '').includes('essay')" class="essay-section">
        <view class="essay-item">
          <text class="essay-label">请作答：</text>
          <textarea class="essay-input" v-model="essayAnswers[currentIndex]" placeholder="请输入你的答案" :maxlength="-1" />
          <text class="word-count">{{ (essayAnswers[currentIndex] || '').length }} 字</text>
        </view>
      </view>

      <!-- 其他题型 -->
      <view v-else-if="currentQuestion.options" class="options">
        <view v-for="opt in currentQuestion.options" :key="opt.label" class="option-item" :class="{ selected: answers[currentIndex] === opt.label }" @click="selectAnswer(opt.label)">
          <text class="label">{{ opt.label }}.</text>
          <text class="content">{{ typeof opt.content === 'string' ? opt.content : '' }}</text>
        </view>
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
const fillAnswers = ref<string[]>([]);
const essayAnswers = ref<string[]>([]);
const elapsed = ref(0);
let timer: ReturnType<typeof setInterval> | null = null;

// 判断是否为选择题
function isChoiceQuestion(q: any): boolean {
  const type = (q.questionType || q.question_type || '').toLowerCase();
  return type.includes('choice') || type.includes('single') || type.includes('multiple') || type === '';
}

onMounted(() => {
  // 从 storage 读取数据，避免 URL 长度限制
  const dataStr = uni.getStorageSync('smartPracticeData');
  if (dataStr) {
    try {
      const parsed = JSON.parse(dataStr);
      questions.value = Array.isArray(parsed) ? parsed : (parsed.questions || parsed.items || []);
      uni.removeStorageSync('smartPracticeData');
    } catch { /* ignore */ }
  }
  if (questions.value.length === 0) {
    uni.showToast({ title: '暂无题目', icon: 'none' });
    setTimeout(() => uni.navigateBack(), 1500);
    return;
  }
  answers.value = new Array(questions.value.length).fill('');
  fillAnswers.value = new Array(questions.value.length).fill('');
  essayAnswers.value = new Array(questions.value.length).fill('');
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
  const answered = answers.value.filter(a => a).length + fillAnswers.value.filter(a => a).length + essayAnswers.value.filter(a => a).length;

  // 保存练习结果
  const result = {
    questions: questions.value,
    answers: answers.value,
    fillAnswers: fillAnswers.value,
    essayAnswers: essayAnswers.value,
    elapsed: elapsed.value
  };
  uni.setStorageSync('practiceResult', result);

  uni.showModal({
    title: '练习完成',
    content: `已完成 ${answered}/${questions.value.length} 题，用时 ${formatTime(elapsed.value)}`,
    success: (res) => {
      if (res.confirm) {
        uni.navigateTo({ url: '/pages/practice/result' });
      }
    },
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

.fill-section { margin-top: 16px; }
.fill-item { background: #f9f9f9; border-radius: 8px; padding: 16px; }
.fill-label { font-size: 14px; font-weight: 500; color: #333; margin-bottom: 8px; display: block; }
.fill-input { width: 100%; height: 40px; border: 1px solid #ddd; border-radius: 6px; padding: 0 12px; font-size: 15px; background: #fff; }

.essay-section { margin-top: 16px; }
.essay-item { background: #f9f9f9; border-radius: 8px; padding: 16px; }
.essay-label { font-size: 14px; font-weight: 500; color: #333; margin-bottom: 8px; display: block; }
.essay-input { width: 100%; min-height: 120px; border: 1px solid #ddd; border-radius: 6px; padding: 12px; font-size: 15px; background: #fff; }
.word-count { font-size: 12px; color: #999; text-align: right; display: block; margin-top: 4px; }
.bottom-nav { position: fixed; bottom: 0; left: 0; right: 0; display: flex; background: #fff; padding: 12px 16px; box-shadow: 0 -2px 10px rgba(0,0,0,0.05); }
.nav-btn { flex: 1; height: 44px; border: none; border-radius: 22px; font-size: 15px; background: #f5f7fa; color: #666; margin-right: 8px; }
.next-btn { flex: 1; height: 44px; border: none; border-radius: 22px; font-size: 15px; background: #52c41a; color: #fff; }
.submit-btn { flex: 1; height: 44px; border: none; border-radius: 22px; font-size: 15px; background: #1890ff; color: #fff; }
</style>
