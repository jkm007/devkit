<script lang="ts" setup>
import type { QuestionApi } from '#/api/question/question';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { VbenTiptapPreview } from '@vben/plugins/tiptap';

import { Checkbox, CheckboxGroup, Radio, RadioGroup, Tag } from 'ant-design-vue';

import {
  processMediaHtml,
  safeJsonParse,
} from '#/utils/media-url';

import {
  DIFFICULTY_OPTIONS,
  QUESTION_TYPE_OPTIONS,
  STATUS_OPTIONS,
} from '../data';

const questionData = ref<QuestionApi.Question | null>(null);
const showAnswer = ref(false);
const selectedAnswer = ref<any>(null);

// Question type categories
const CHOICE_TYPES = ['single_choice', 'multiple_choice', 'indefinite_choice'];

function getLabel(options: any[], value: any): string {
  return options.find((o) => o.value === value)?.label || String(value);
}

function getStatusColor(value: string): string {
  const map: Record<string, string> = {
    draft: 'default',
    pending: 'processing',
    rejected: 'error',
    published: 'success',
    archived: 'warning',
  };
  return map[value] || 'default';
}

function getDifficultyColor(value: number): string {
  const map: Record<number, string> = { 1: 'green', 2: 'orange', 3: 'red' };
  return map[value] || 'default';
}

function getDifficultyLabel(value: number): string {
  return getLabel(DIFFICULTY_OPTIONS, value);
}

// Parse options from content JSON - robust handling
function parseOptions(): Array<{ id: string; text: string }> {
  if (!questionData.value?.content) return [];
  try {
    let parsed: any = safeJsonParse(questionData.value.content);
    if (typeof parsed === 'string') {
      try { parsed = JSON.parse(parsed); } catch { /* ignore */ }
    }
    if (Array.isArray(parsed) && parsed.length > 0) {
      return parsed.map((o: any, i: number) => ({
        id: o.id || String.fromCharCode(65 + i),
        text: o.text || o.label || o.content || o.value || '',
      }));
    }
    return [];
  } catch {
    return [];
  }
}

// Parse correct answer from answer JSON
function parseCorrectAnswer(): string[] {
  if (!questionData.value?.answer) return [];
  try {
    const parsed = safeJsonParse(questionData.value.answer);
    if (parsed && parsed.correct) return Array.isArray(parsed.correct) ? parsed.correct : [parsed.correct];
    if (parsed && parsed.blanks) return parsed.blanks;
    return [];
  } catch {
    return [];
  }
}

// Parse true/false answer
function parseTFAnswer(): string {
  if (!questionData.value?.answer) return '';
  try {
    const parsed = safeJsonParse(questionData.value.answer);
    return (parsed && parsed.correct) || '';
  } catch {
    return '';
  }
}

// Get question type info
const questionType = computed(() => questionData.value?.questionType || '');
const isChoice = computed(() => CHOICE_TYPES.includes(questionType.value));
const isSingleChoice = computed(() =>
  ['single_choice', 'true_false'].includes(questionType.value),
);
const isTrueFalse = computed(() => questionType.value === 'true_false');
const isFillBlank = computed(() => questionType.value === 'fill_blank');

// Options list
const optionsList = computed(() => parseOptions());
const correctList = computed(() => parseCorrectAnswer());

// Check if an option is correct
function isCorrectOption(optionId: string): boolean {
  return correctList.value.includes(optionId);
}

// Processed HTML content (with token for direct display)
const processedStem = ref('');
const processedAnalysisText = ref('');
const processedAnalysisMedia = ref('');
const processedEssayAnswer = ref('');

// Reset answer on drawer open
const [Drawer, drawerApi] = useVbenDrawer({
  onOpenChange(isOpen) {
    if (isOpen) {
      questionData.value = drawerApi.getData<QuestionApi.Question>();
      showAnswer.value = false;
      selectedAnswer.value = isSingleChoice.value ? null : [];

      if (questionData.value) {
        // Parse and process all HTML fields (synchronous!)
        const rawStem = safeJsonParse(questionData.value.stem || '') || '';
        const rawAnswer = safeJsonParse(questionData.value.answer || '') || '';

        // Parse analysis: handle {text, media} format and legacy plain HTML
        let rawAnalysisText = '';
        let rawAnalysisMedia = '';
        try {
          const parsed = safeJsonParse(questionData.value.analysis || '');
          if (parsed && typeof parsed === 'object' && ('text' in parsed || 'media' in parsed)) {
            rawAnalysisText = parsed.text || '';
            rawAnalysisMedia = parsed.media || '';
          } else if (typeof parsed === 'string') {
            rawAnalysisText = parsed;
          }
        } catch {
          rawAnalysisText = String(questionData.value.analysis || '');
        }

        // Process media URLs: normalize + add token (synchronous, no async fetch!)
        processedStem.value = processMediaHtml(rawStem);
        processedAnalysisText.value = processMediaHtml(rawAnalysisText);
        processedAnalysisMedia.value = processMediaHtml(rawAnalysisMedia);
        processedEssayAnswer.value = processMediaHtml(rawAnswer);
      }
    }
  },
});
</script>

<template>
  <Drawer :footer="false" title="题目预览" class="w-[800px]">
    <div v-if="questionData" class="bg-white">
      <!-- Question header bar -->
      <div class="border-b border-gray-100 px-6 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <Tag :color="getDifficultyColor(questionData.difficulty)">
              {{ getDifficultyLabel(questionData.difficulty) }}
            </Tag>
            <Tag color="blue">
              {{ getLabel(QUESTION_TYPE_OPTIONS, questionData.questionType) }}
            </Tag>
          </div>
          <Tag :color="getStatusColor(questionData.status)">
            {{ getLabel(STATUS_OPTIONS, questionData.status) }}
          </Tag>
        </div>
      </div>

      <!-- Exam-style question body -->
      <div class="px-6 py-6">
        <!-- Question title -->
        <div class="mb-6 text-base font-medium text-gray-800">
          {{ questionData.title }}
        </div>

        <!-- Stem content -->
        <div v-if="processedStem" class="mb-6 leading-relaxed text-gray-700">
          <VbenTiptapPreview :content="processedStem" :min-height="60" />
        </div>

        <!-- Choice question options -->
        <div v-if="isChoice && optionsList.length > 0" class="mb-6">
          <div class="mb-3 text-sm font-medium text-gray-600">
            {{ isSingleChoice ? '请选择一个选项：' : '请选择所有正确选项：' }}
          </div>
          <component
            :is="isSingleChoice ? RadioGroup : CheckboxGroup"
            v-model:value="selectedAnswer"
            class="flex flex-col gap-3"
          >
            <div
              v-for="opt in optionsList"
              :key="opt.id"
              class="group flex items-start gap-3 rounded-lg border border-gray-200 px-4 py-3 transition-all hover:border-blue-300 hover:bg-blue-50"
              :class="{
                'border-green-400 bg-green-50': showAnswer && isCorrectOption(opt.id),
                'border-red-300 bg-red-50': showAnswer && (Array.isArray(selectedAnswer) ? selectedAnswer.includes(opt.id) : selectedAnswer === opt.id) && !isCorrectOption(opt.id),
              }"
            >
              <component
                :is="isSingleChoice ? Radio : Checkbox"
                :value="opt.id"
                class="mt-0.5"
              />
              <div class="flex-1 text-base">
                <span class="mr-2 inline-flex size-6 items-center justify-center rounded-full bg-gray-100 text-sm font-bold text-gray-600">
                  {{ opt.id }}
                </span>
                <span class="text-gray-800">{{ opt.text || '(未填写选项内容)' }}</span>
              </div>
              <span
                v-if="showAnswer && isCorrectOption(opt.id)"
                class="ml-2 text-sm font-medium text-green-600"
              >
                ✓ 正确
              </span>
            </div>
          </component>
        </div>

        <!-- No options warning -->
        <div v-if="isChoice && optionsList.length === 0" class="mb-6 rounded-lg bg-yellow-50 p-4 text-sm text-yellow-700">
          ⚠ 该题目没有选项数据
        </div>

        <!-- True/False options -->
        <div v-if="isTrueFalse" class="mb-6">
          <div class="mb-3 text-sm font-medium text-gray-600">请判断对错：</div>
          <RadioGroup v-model:value="selectedAnswer" class="flex gap-4">
            <div
              class="flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-lg border border-gray-200 py-4 text-center transition-all hover:border-blue-300 hover:bg-blue-50"
              :class="{
                'border-green-400 bg-green-50': showAnswer && parseTFAnswer() === 'true',
                'border-red-300 bg-red-50': showAnswer && selectedAnswer === 'true' && parseTFAnswer() !== 'true',
              }"
            >
              <Radio value="true" />
              <span class="text-lg">正确 ✓</span>
            </div>
            <div
              class="flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-lg border border-gray-200 py-4 text-center transition-all hover:border-blue-300 hover:bg-blue-50"
              :class="{
                'border-green-400 bg-green-50': showAnswer && parseTFAnswer() === 'false',
                'border-red-300 bg-red-50': showAnswer && selectedAnswer === 'false' && parseTFAnswer() !== 'false',
              }"
            >
              <Radio value="false" />
              <span class="text-lg">错误 ✗</span>
            </div>
          </RadioGroup>
        </div>

        <!-- Submit button (for interactive preview) -->
        <div v-if="!showAnswer" class="mb-6">
          <button
            class="rounded-lg bg-blue-500 px-8 py-2.5 text-white transition-colors hover:bg-blue-600"
            @click="showAnswer = true"
          >
            提交并查看答案
          </button>
        </div>

        <!-- Answer section (shown after "submit") -->
        <div v-if="showAnswer" class="mt-6 space-y-4">
          <!-- Divider -->
          <div class="flex items-center gap-3">
            <div class="h-px flex-1 bg-gray-200" />
            <span class="text-sm text-gray-400">答案与解析</span>
            <div class="h-px flex-1 bg-gray-200" />
          </div>

          <!-- Correct answer -->
          <div v-if="isChoice || isTrueFalse" class="rounded-lg bg-green-50 p-4">
            <div class="mb-1 text-sm font-medium text-green-700">正确答案</div>
            <div class="text-lg font-bold text-green-600">
              <template v-if="isChoice">
                {{ correctList.join(', ') }}
              </template>
              <template v-else-if="isTrueFalse">
                {{ parseTFAnswer() === 'true' ? '正确 ✓' : '错误 ✗' }}
              </template>
            </div>
          </div>

          <!-- Essay/fill-blank answer -->
          <div v-if="!isChoice && !isTrueFalse && processedEssayAnswer" class="rounded-lg bg-green-50 p-4">
            <div class="mb-2 text-sm font-medium text-green-700">参考答案</div>
            <VbenTiptapPreview :content="processedEssayAnswer" :min-height="60" />
          </div>

          <!-- Analysis: text -->
          <div v-if="processedAnalysisText" class="rounded-lg bg-blue-50 p-4">
            <div class="mb-2 text-sm font-medium text-blue-700">文字解析</div>
            <VbenTiptapPreview :content="processedAnalysisText" :min-height="60" />
          </div>

          <!-- Analysis: media -->
          <div v-if="processedAnalysisMedia" class="rounded-lg bg-blue-50 p-4">
            <div class="mb-2 text-sm font-medium text-blue-700">图片/视频解析</div>
            <VbenTiptapPreview :content="processedAnalysisMedia" :min-height="60" />
          </div>
        </div>
      </div>
    </div>
  </Drawer>
</template>
