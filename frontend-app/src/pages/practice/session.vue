<template>
  <view class="session-page">
    <view class="progress-bar" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="progress-info">
        <text class="question-index">{{ currentIndex + 1 }}/{{ questions.length }}</text>
        <view class="progress-track"><view class="progress-fill" :style="{ width: ((currentIndex + 1) / questions.length * 100) + '%' }" /></view>
        <text class="timer">{{ formatTime(elapsed) }}</text>
      </view>
    </view>
    <view class="answer-sheet-btn" @click="showAnswerSheet = true"><text>📋</text></view>
    <view v-if="currentQuestion" class="question-content">
      <ContentBlockRenderer v-for="(block, idx) in currentQuestion.stem?.blocks" :key="idx" :block="block" />

      <!-- 选择题选项 -->
      <view v-if="isChoiceQuestion(currentQuestion.questionType) && currentQuestion.options && currentQuestion.options.length" class="options">
        <view v-for="opt in currentQuestion.options" :key="opt.label" class="option-item" :class="{ selected: answers[currentIndex] === opt.label }" @click="selectAnswer(opt.label)">
          <text class="label">{{ opt.label }}.</text>
          <view class="content">
            <ContentBlockRenderer v-for="(block, idx) in opt.content?.blocks" :key="idx" :block="block" />
          </view>
        </view>
      </view>

      <!-- 填空题输入 -->
      <view v-else-if="currentQuestion.questionType === 'fill' || currentQuestion.questionType === 'fill_blank'" class="fill-input">
        <textarea
          v-model="fillAnswers[currentIndex]"
          class="fill-textarea"
          placeholder="请输入答案"
          :maxlength="500"
          @input="updateFillAnswer"
        />
      </view>

      <!-- 简答题输入 -->
      <view v-else-if="currentQuestion.questionType === 'essay' || currentQuestion.questionType === 'short_answer'" class="essay-input">
        <textarea
          v-model="essayAnswers[currentIndex]"
          class="essay-textarea"
          placeholder="请输入你的答案"
          :maxlength="2000"
          @input="updateEssayAnswer"
        />
        <text class="word-count">{{ (essayAnswers[currentIndex] || '').length }} / 2000</text>
      </view>
    </view>
    <view class="bottom-nav">
      <button class="nav-btn" :disabled="currentIndex === 0" @click="prev">上一题</button>
      <button v-if="currentIndex === questions.length - 1" class="submit-btn" @click="submitPractice">提交</button>
      <button v-else class="next-btn" @click="next">下一题</button>
    </view>
    <up-popup v-model:show="showAnswerSheet" mode="bottom" round="12">
      <view class="answer-sheet">
        <text class="sheet-title">答题卡</text>
        <view class="sheet-grid">
          <view v-for="(q, i) in questions" :key="i" class="sheet-cell" :class="{ current: i === currentIndex, answered: answers[i] }" @click="goTo(i)">{{ i + 1 }}</view>
        </view>
      </view>
    </up-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import ContentBlockRenderer from '@/components/ContentBlockRenderer.vue';
import { request } from '@/api/request';

const statusBarHeight = ref(0);
const questions = ref<any[]>([]);
const currentIndex = ref(0);
const answers = ref<string[]>([]); // 选择题答案
const fillAnswers = ref<string[]>([]); // 填空题答案
const essayAnswers = ref<string[]>([]); // 简答题答案
const elapsed = ref(0);
const showAnswerSheet = ref(false);
let timer: ReturnType<typeof setInterval> | null = null;
const currentQuestion = computed(() => questions.value[currentIndex.value] || null);

function isChoiceQuestion(type: string): boolean {
  return ['single', 'multi', 'single_choice', 'multiple_choice', 'judge', 'true_false'].includes(type);
}

onMounted(async () => {
  const systemInfo = uni.getSystemInfoSync();
  statusBarHeight.value = systemInfo.statusBarHeight || 0;

  // 从 setStorageSync 读取参数
  const paramsStr = uni.getStorageSync('practiceParams');
  let params: any = {};
  if (paramsStr) {
    try { params = JSON.parse(paramsStr); } catch { /* ignore */ }
    uni.removeStorageSync('practiceParams');
  }

  await loadQuestions(params);
  answers.value = new Array(questions.value.length).fill('');
  fillAnswers.value = new Array(questions.value.length).fill('');
  essayAnswers.value = new Array(questions.value.length).fill('');
  timer = setInterval(() => { elapsed.value++; }, 1000);
  uni.setStorageSync('practice_answers', JSON.stringify(answers.value));
});

async function loadQuestions(params: any) {
  const mode = params.mode || 'random';
  const count = params.count || 20;

  try {
    let res: any;
    if (mode === 'wrong') {
      // 错题练习：从错题本获取随机题目
      res = await request.get<any>('/study/wrong/random', { params: { count } });
    } else if (mode === 'smart') {
      // 智能练习
      res = await request.post<any>('/study/practice/smart', { count, mode: 'mixed' });
    } else {
      // 随机/自定义练习
      res = await request.post<any>('/study/practice/questions', {
        count,
        mode: 'random',
        difficulty: params.difficulty || 0,
        types: params.types || [],
        categoryId: params.categoryId || 0,
        subjectId: params.subjectId || 0,
      });
    }

    // 处理不同的返回格式
    if (res) {
      if (Array.isArray(res)) {
        // 直接返回数组
        questions.value = res;
      } else if (res.questions) {
        // { questions: [...] }
        questions.value = res.questions;
      } else if (res.items) {
        // { items: [...] }
        questions.value = res.items;
      }
      if (questions.value.length > 0) return;
    }
  } catch (e) {
    console.error('加载题目失败:', e);
    /* fall through to mock */
  }

  // Mock 数据（仅开发环境）
  if (questions.value.length === 0 && import.meta.env.DEV) {
    const modeLabels: Record<string, string> = { random: '随机', wrong: '错题', smart: '智能', custom: '自定义' };
    const prefix = modeLabels[mode] || '练习';
    questions.value = Array.from({ length: count }, (_, i) => ({
      id: i + 1,
      questionType: 'single',
      stem: { blocks: [{ type: 'text', content: `第 ${i + 1} 题：${prefix}题目，以下说法正确的是？` }] },
      options: [
        { label: 'A', content: { blocks: [{ type: 'text', content: '选项 A 的内容' }] } },
        { label: 'B', content: { blocks: [{ type: 'text', content: '选项 B 的内容' }] } },
        { label: 'C', content: { blocks: [{ type: 'text', content: '选项 C 的内容' }] } },
        { label: 'D', content: { blocks: [{ type: 'text', content: '选项 D 的内容' }] } },
      ],
      answer: 'A',
    }));
  }
}
onUnmounted(() => { if (timer) clearInterval(timer); });
function selectAnswer(label: string) { answers.value[currentIndex.value] = label; uni.setStorageSync('practice_answers', JSON.stringify(answers.value)); }
function updateFillAnswer() { uni.setStorageSync('practice_fill_answers', JSON.stringify(fillAnswers.value)); }
function updateEssayAnswer() { uni.setStorageSync('practice_essay_answers', JSON.stringify(essayAnswers.value)); }
function prev() { if (currentIndex.value > 0) currentIndex.value--; }
function next() { if (currentIndex.value < questions.value.length - 1) currentIndex.value++; }
function goTo(i: number) { currentIndex.value = i; showAnswerSheet.value = false; }
function formatTime(s: number): string { const m = Math.floor(s / 60); const sec = s % 60; return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`; }
function submitPractice() {
  if (timer) clearInterval(timer);

  // 合并所有答案
  const allAnswers = questions.value.map((q, i) => {
    if (isChoiceQuestion(q.questionType)) {
      return answers.value[i] || '';
    } else if (q.questionType === 'fill' || q.questionType === 'fill_blank') {
      return fillAnswers.value[i] || '';
    } else if (q.questionType === 'essay' || q.questionType === 'short_answer') {
      return essayAnswers.value[i] || '';
    }
    return answers.value[i] || '';
  });

  const result = {
    total: questions.value.length,
    answered: allAnswers.filter(a => a).length,
    elapsed: elapsed.value,
    answers: allAnswers,
  };
  // 使用 setStorageSync 避免 URL 长度限制
  uni.setStorageSync('practice_result', JSON.stringify(result));
  // 保存题目信息用于判断对错
  uni.setStorageSync('practice_questions', JSON.stringify(questions.value));
  uni.navigateTo({ url: '/pages/practice/result' });
}
</script>

<style lang="scss" scoped>
.session-page { min-height: 100vh; background: #f5f5f5; padding-bottom: 70px; }
.progress-bar { background: #fff; padding: 8px 16px 12px; }
.progress-info { display: flex; align-items: center; gap: 12px; }
.question-index { font-size: 14px; font-weight: 500; }
.progress-track { flex: 1; height: 4px; background: #eee; border-radius: 2px; }
.progress-fill { height: 100%; background: #1890ff; border-radius: 2px; }
.timer { font-size: 14px; color: #1890ff; font-family: monospace; }
.answer-sheet-btn { position: fixed; top: 50px; right: 16px; z-index: 10; width: 40px; height: 40px; background: #fff; border-radius: 50%; display: flex; align-items: center; justify-content: center; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
.question-content { background: #fff; margin: 12px 16px; border-radius: 12px; padding: 20px; }
.stem { font-size: 16px; line-height: 1.6; margin-bottom: 20px; }
.option-item { display: flex; padding: 14px; margin-bottom: 10px; background: #f9f9f9; border-radius: 8px; border: 2px solid transparent; }
.option-item.selected { border-color: #1890ff; background: #e6f7ff; }
.label { font-weight: 500; margin-right: 10px; }
.content { flex: 1; font-size: 15px; line-height: 1.5; }

/* 填空题 */
.fill-input { margin-top: 16px; }
.fill-textarea {
  width: 100%;
  min-height: 80px;
  padding: 12px;
  background: #f9f9f9;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
  font-size: 15px;
  line-height: 1.5;
}

/* 简答题 */
.essay-input { margin-top: 16px; position: relative; }
.essay-textarea {
  width: 100%;
  min-height: 150px;
  padding: 12px;
  background: #f9f9f9;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
  font-size: 15px;
  line-height: 1.5;
}
.word-count {
  position: absolute;
  bottom: 8px;
  right: 12px;
  font-size: 12px;
  color: #999;
}
.bottom-nav { position: fixed; bottom: 0; left: 0; right: 0; display: flex; background: #fff; padding: 12px 16px; box-shadow: 0 -2px 10px rgba(0,0,0,0.05); }
.nav-btn, .next-btn, .submit-btn { flex: 1; height: 44px; border: none; border-radius: 22px; font-size: 15px; }
.nav-btn { background: #f5f7fa; color: #666; margin-right: 8px; }
.next-btn { background: #1890ff; color: #fff; }
.submit-btn { background: #52c41a; color: #fff; }
.answer-sheet { padding: 20px; text-align: center; }
.sheet-title { font-size: 16px; font-weight: 500; margin-bottom: 16px; display: block; }
.sheet-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 10px; }
.sheet-cell { padding: 12px 0; background: #f5f7fa; border-radius: 8px; }
.sheet-cell.current { background: #1890ff; color: #fff; }
.sheet-cell.answered { background: #e6f7ff; color: #1890ff; }
</style>
