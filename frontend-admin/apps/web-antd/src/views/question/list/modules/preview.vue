<script lang="ts" setup>
import type { QuestionApi } from '#/api/question/question';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { VbenTiptapPreview } from '@vben/plugins/tiptap';

import { Checkbox, CheckboxGroup, Radio, RadioGroup, Tag } from 'ant-design-vue';

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

// Decode JSON-encoded string (handles double-encoding from DB)
function safeJsonParse(jsonStr: string): string {
  if (!jsonStr || jsonStr === 'null') return '';
  try {
    const parsed = JSON.parse(jsonStr);
    if (typeof parsed === 'string') {
      // Double-encoded: parse again
      try {
        const parsed2 = JSON.parse(parsed);
        return typeof parsed2 === 'string' ? parsed2 : parsed;
      } catch {
        return parsed;
      }
    }
    return jsonStr;
  } catch {
    return jsonStr;
  }
}

// Fix image URLs in HTML content (replace direct-url with view endpoint)
function fixImageUrls(html: string): string {
  if (!html) return html;
  // Only replace /files/{id}/direct-url → /files/{id}/view
  // Do NOT add /api/v1 prefix here - the browser resolves relative paths against the page origin
  return html.replace(/\/files\/(\d+)\/direct-url/g, '/files/$1/view');
}

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

// Deep parse JSON (handles double-encoding from DB JSON columns)
function deepParse(val: any): any {
  if (typeof val !== 'string') return val;
  try {
    const parsed = JSON.parse(val);
    if (typeof parsed === 'string') {
      try { return JSON.parse(parsed); } catch { return parsed; }
    }
    return parsed;
  } catch {
    return val;
  }
}

// Parse options from content JSON
function parseOptions(): Array<{ id: string; text: string }> {
  if (!questionData.value?.content) return [];
  try {
    const parsed = deepParse(questionData.value.content);
    if (Array.isArray(parsed)) return parsed;
    return [];
  } catch {
    return [];
  }
}

// Parse correct answer from answer JSON
function parseCorrectAnswer(): string[] {
  if (!questionData.value?.answer) return [];
  try {
    const parsed = deepParse(questionData.value.answer);
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
    const parsed = deepParse(questionData.value.answer);
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

// Reset answer on drawer open
const [Drawer, drawerApi] = useVbenDrawer({
  onOpenChange(isOpen) {
    if (isOpen) {
      questionData.value = drawerApi.getData<QuestionApi.Question>();
      showAnswer.value = false;
      selectedAnswer.value = isSingleChoice.value ? null : [];
    }
  },
});

// Get stem HTML (with fixed image URLs)
function getStem(): string {
  return fixImageUrls(safeJsonParse(questionData.value?.stem || ''));
}

// Get analysis HTML (with fixed image URLs)
function getAnalysis(): string {
  return fixImageUrls(safeJsonParse(questionData.value?.analysis || ''));
}

// Get essay answer HTML (with fixed image URLs)
function getEssayAnswer(): string {
  return fixImageUrls(safeJsonParse(questionData.value?.answer || ''));
}
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
        <div v-if="getStem()" class="mb-6 leading-relaxed text-gray-700">
          <VbenTiptapPreview :content="getStem()" :min-height="60" />
        </div>

        <!-- Choice question options -->
        <div v-if="isChoice && optionsList.length > 0" class="mb-6">
          <component
            :is="isSingleChoice ? RadioGroup : CheckboxGroup"
            v-model:value="selectedAnswer"
            class="flex flex-col gap-3"
          >
            <div
              v-for="opt in optionsList"
              :key="opt.id"
              class="group flex items-start gap-3 rounded-lg border border-gray-200 px-4 py-3 transition-colors hover:border-blue-300 hover:bg-blue-50"
              :class="{
                'border-green-400 bg-green-50': showAnswer && isCorrectOption(opt.id),
                'border-red-300 bg-red-50': showAnswer && selectedAnswer === opt.id && !isCorrectOption(opt.id),
              }"
            >
              <component
                :is="isSingleChoice ? Radio : Checkbox"
                :value="opt.id"
                class="mt-0.5"
              />
              <div class="flex-1">
                <span class="mr-2 font-medium text-gray-600">{{ opt.id }}.</span>
                <span>{{ opt.text }}</span>
              </div>
              <span
                v-if="showAnswer && isCorrectOption(opt.id)"
                class="ml-2 text-sm text-green-600"
              >
                ✓
              </span>
            </div>
          </component>
        </div>

        <!-- True/False options -->
        <div v-if="isTrueFalse" class="mb-6">
          <RadioGroup v-model:value="selectedAnswer" class="flex gap-4">
            <div
              class="flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-lg border border-gray-200 py-4 text-center transition-colors hover:border-blue-300 hover:bg-blue-50"
              :class="{
                'border-green-400 bg-green-50': showAnswer && parseTFAnswer() === 'true',
                'border-red-300 bg-red-50': showAnswer && selectedAnswer === 'true' && parseTFAnswer() !== 'true',
              }"
            >
              <Radio value="true" />
              <span class="text-lg">正确 ✓</span>
            </div>
            <div
              class="flex flex-1 cursor-pointer items-center justify-center gap-2 rounded-lg border border-gray-200 py-4 text-center transition-colors hover:border-blue-300 hover:bg-blue-50"
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
          <div v-if="!isChoice && !isTrueFalse && getEssayAnswer()" class="rounded-lg bg-green-50 p-4">
            <div class="mb-2 text-sm font-medium text-green-700">参考答案</div>
            <VbenTiptapPreview :content="getEssayAnswer()" :min-height="60" />
          </div>

          <!-- Analysis -->
          <div v-if="getAnalysis()" class="rounded-lg bg-blue-50 p-4">
            <div class="mb-2 text-sm font-medium text-blue-700">解析</div>
            <VbenTiptapPreview :content="getAnalysis()" :min-height="60" />
          </div>
        </div>
      </div>
    </div>
  </Drawer>
</template>
