<script lang="ts" setup>
import type { QuestionApi } from '#/api/question/question';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { Plus } from '@vben/icons';
import { VbenTiptap } from '@vben/plugins/tiptap';

import { Button, Input, message, Radio, RadioGroup, Tooltip } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { simpleUpload } from '#/api/file';
import {
  getExamAll,
  getQuestionCategoryAll,
  getSubjectAll,
} from '#/api/question/category';
import { createQuestion, updateQuestion } from '#/api/question/question';
import { useAccessStore } from '@vben/stores';

import {
  DIFFICULTY_OPTIONS,
  QUESTION_TYPE_OPTIONS,
  RESOURCE_TYPE_OPTIONS,
} from '../data';

const emits = defineEmits(['success']);

const formData = ref<QuestionApi.Question>();
const id = ref<number>();

// Access store for token
const accessStore = useAccessStore();
function getAuthToken(): string {
  return accessStore.accessToken || '';
}

// API base URL for image paths
const apiBase = import.meta.env.VITE_GLOB_API_URL || '/api/v1';

// Rich text field values
const stemHtml = ref('');
const analysisTextHtml = ref(''); // 文字解析
const analysisMediaHtml = ref(''); // 图片/视频解析

// Choice question options
interface OptionItem {
  id: string;
  text: string;
}
const options = ref<OptionItem[]>([
  { id: 'A', text: '' },
  { id: 'B', text: '' },
]);
const correctAnswers = ref<string[]>([]);

// True/false answer
const tfAnswer = ref<string>('true');

// Fill-blank answer (essay answer for non-choice types)
const essayAnswerHtml = ref('');

// Question type categories
const CHOICE_TYPES = ['single_choice', 'multiple_choice', 'indefinite_choice'];
const isChoiceType = computed(() => CHOICE_TYPES.includes(currentType.value));
const isTrueFalse = computed(() => currentType.value === 'true_false');
const isFillBlank = computed(() => currentType.value === 'fill_blank');

// Current question type for template - synced from form select
const currentType = ref('');

// Exam/Subject/Category cascading options
const examOptions = ref<any[]>([]);
const subjectOptions = ref<any[]>([]);
const categoryOptions = ref<any[]>([]);

async function loadExamOptions() {
  try {
    const res = await getExamAll();
    examOptions.value = (res || []).map((item: any) => ({
      label: item.name,
      value: item.id,
    }));
  } catch {
    // ignore
  }
}

async function loadSubjectOptions(examId: number) {
  try {
    const res = await getSubjectAll(examId);
    subjectOptions.value = (res || []).map((item: any) => ({
      label: item.name,
      value: item.id,
    }));
  } catch {
    subjectOptions.value = [];
  }
}

async function loadCategoryOptions() {
  try {
    const res = await getQuestionCategoryAll();
    categoryOptions.value = (res || []).map((item: any) => ({
      label: item.name,
      value: item.id,
    }));
  } catch {
    categoryOptions.value = [];
  }
}

loadExamOptions();
loadCategoryOptions();

// Blob URL → real URL mapping (for converting back before save)
const blobToRealUrl = new Map<string, string>();

// Decode JSON-encoded string (handles double-encoding from DB)
function safeJsonParse(jsonStr: string): string {
  if (!jsonStr || jsonStr === 'null') return '';
  try {
    const parsed = JSON.parse(jsonStr);
    if (typeof parsed === 'string') {
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

// Fix image/video URLs in HTML content (normalize to /api/v1/files/{id}/view)
function fixImageUrls(html: string): string {
  if (!html) return html;
  // Replace /files/{id}/direct-url → /files/{id}/view (without adding prefix)
  let fixed = html.replace(/\/files\/(\d+)\/direct-url/g, '/files/$1/view');
  // Add /api/v1 prefix to bare /files/ paths (in src, href, or poster attributes)
  fixed = fixed.replace(/((?:src|href|poster)=")(\/files\/\d+\/view)/g, `$1${apiBase}$2`);
  return fixed;
}

// Encode HTML string to JSON for backend
function htmlToJson(html: string): string {
  return JSON.stringify(html || '');
}

// Fetch a media file with auth header and return a blob URL
async function fetchMediaAsBlob(url: string): Promise<string> {
  try {
    const token = getAuthToken();
    const response = await fetch(url, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    if (response.ok) {
      const blob = await response.blob();
      const blobUrl = URL.createObjectURL(blob);
      blobToRealUrl.set(blobUrl, url);
      return blobUrl;
    }
  } catch {
    // ignore
  }
  return url;
}

// Convert real media URLs in HTML to blob URLs (for display in editor)
// Handles <img>, <video>, <source>, and poster attributes
async function convertMediaToBlobUrls(html: string): Promise<string> {
  if (!html) return html;
  // Match src="..." on img/video/source tags, and poster="..." on video tags
  const urlRegex = /(<(?:img|video|source)[^>]+(?:src|poster)=")([^"]+)(")/g;
  const matches = [...html.matchAll(urlRegex)];
  if (matches.length === 0) return html;

  // Deduplicate URLs to avoid fetching the same URL multiple times
  const uniqueUrls = [...new Set(matches.map((m) => m[2]).filter((u) => !u.startsWith('blob:') && !u.startsWith('data:')))];
  if (uniqueUrls.length === 0) return html;

  // Fetch all unique URLs in parallel
  const urlMap = new Map<string, string>();
  await Promise.all(
    uniqueUrls.map(async (url) => {
      const blobUrl = await fetchMediaAsBlob(url);
      if (blobUrl !== url) {
        urlMap.set(url, blobUrl);
      }
    }),
  );

  // Replace each URL precisely using the full match to avoid over-replacing
  if (urlMap.size === 0) return html;
  let result = html;
  // Process in reverse order to preserve match indices
  for (let i = matches.length - 1; i >= 0; i--) {
    const match = matches[i];
    const originalUrl = match[2];
    const blobUrl = urlMap.get(originalUrl);
    if (!blobUrl) continue;
    // Replace only this specific occurrence using the full match
    const fullMatch = match[0];
    const newMatch = fullMatch.replace(originalUrl, blobUrl);
    result = result.substring(0, match.index) + newMatch + result.substring(match.index + fullMatch.length);
  }
  return result;
}

// Convert blob URLs in HTML back to real URLs (for saving to DB)
function convertBlobUrlsToReal(html: string): string {
  if (!html) return html;
  let result = html;
  for (const [blobUrl, realUrl] of blobToRealUrl.entries()) {
    result = result.replaceAll(blobUrl, realUrl);
  }
  // Also handle via regex in case mapping was lost
  result = result.replace(/blob:http[^"]+/g, (match) => {
    return blobToRealUrl.get(match) || match;
  });
  return result;
}

// Media upload adapter for TipTap (images + videos)
const imageUploadConfig = {
  accept: 'image/*,video/*',
  maxSize: 100 * 1024 * 1024, // 100MB (for videos)
  upload: async (file: File, onProgress?: (percent: number) => void) => {
    const result = await simpleUpload(file, (event) => {
      onProgress?.(event.percent);
    });
    // Build the real URL
    let realUrl = result.url;
    if (realUrl && !realUrl.startsWith('http') && !realUrl.startsWith(apiBase)) {
      realUrl = `${apiBase}${realUrl}`;
    }
    // Normalize direct-url to view
    realUrl = realUrl.replace(/\/direct-url/g, '/view');
    return realUrl;
  },
};

// Generate option IDs: A, B, C, D, ...
function getOptionId(index: number): string {
  return String.fromCharCode(65 + index);
}

// Add a new option
function addOption() {
  if (options.value.length >= 10) {
    message.warning('最多支持10个选项');
    return;
  }
  const index = options.value.length;
  options.value.push({ id: getOptionId(index), text: '' });
}

// Remove an option
function removeOption(index: number) {
  if (options.value.length <= 2) {
    message.warning('至少需要2个选项');
    return;
  }
  const removed = options.value[index];
  options.value.splice(index, 1);
  // Re-index option IDs
  options.value.forEach((opt, i) => {
    opt.id = getOptionId(i);
  });
  // Remove from correct answers if present
  correctAnswers.value = correctAnswers.value.filter((a) => a !== removed.id);
}

// Toggle correct answer for choice questions
function toggleCorrectAnswer(optionId: string) {
  const questionType = currentType.value;
  if (questionType === 'single_choice') {
    correctAnswers.value = [optionId];
  } else {
    const idx = correctAnswers.value.indexOf(optionId);
    if (idx >= 0) {
      correctAnswers.value.splice(idx, 1);
    } else {
      correctAnswers.value.push(optionId);
    }
  }
}

const [Form, formApi] = useVbenForm({
  schema: [
    {
      component: 'Input',
      fieldName: 'title',
      label: '题目标题',
      rules: 'required',
      componentProps: { placeholder: '请输入题目标题' },
    },
    {
      component: 'Select',
      fieldName: 'questionType',
      label: '题型',
      rules: 'required',
      componentProps: {
        options: QUESTION_TYPE_OPTIONS,
        placeholder: '请选择题型',
        class: 'w-full',
        onChange: (val: string) => onQuestionTypeChange(val),
      },
    },
    {
      component: 'Select',
      fieldName: 'difficulty',
      label: '难度',
      defaultValue: 1,
      componentProps: { options: DIFFICULTY_OPTIONS, class: 'w-full' },
    },
    {
      component: 'Select',
      fieldName: 'resourceType',
      label: '资源类型',
      defaultValue: 'private',
      componentProps: { options: RESOURCE_TYPE_OPTIONS, class: 'w-full' },
    },
    {
      component: 'Select',
      fieldName: 'examId',
      label: '所属考试',
      componentProps: {
        options: examOptions,
        placeholder: '请选择考试',
        class: 'w-full',
        onChange: (val: number) => {
          if (val) {
            loadSubjectOptions(val);
          } else {
            subjectOptions.value = [];
          }
          formApi.setValues({ subjectId: undefined, categoryId: undefined });
        },
      },
    },
    {
      component: 'Select',
      fieldName: 'subjectId',
      label: '所属科目',
      componentProps: {
        options: subjectOptions,
        placeholder: '请先选择考试',
        class: 'w-full',
      },
    },
    {
      component: 'Select',
      fieldName: 'categoryId',
      label: '章节分类',
      componentProps: {
        options: categoryOptions,
        placeholder: '请选择章节分类',
        class: 'w-full',
      },
    },
  ],
  showDefaultActions: false,
});

// Handle question type change from Select
function onQuestionTypeChange(val: string) {
  const oldType = currentType.value;
  currentType.value = val;
  if (val === oldType) return;
  // Reset type-specific data when type changes
  if (!CHOICE_TYPES.includes(val)) {
    options.value = [
      { id: 'A', text: '' },
      { id: 'B', text: '' },
    ];
    correctAnswers.value = [];
  }
  if (val !== 'true_false') {
    tfAnswer.value = 'true';
  }
}

const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    // Validate basic form fields
    const { valid } = await formApi.validate();
    if (!valid) return;

    // Validate stem (required)
    const stemText = stemHtml.value?.replace(/<[^>]*>/g, '').trim();
    if (!stemText) {
      message.warning('请输入题干内容');
      return;
    }

    const values = await formApi.getValues();
    const questionType = values.questionType as string;

    // Convert blob URLs back to real URLs before saving
    const realStemHtml = convertBlobUrlsToReal(stemHtml.value);
    const realAnalysisTextHtml = convertBlobUrlsToReal(analysisTextHtml.value);
    const realAnalysisMediaHtml = convertBlobUrlsToReal(analysisMediaHtml.value);
    const realEssayAnswerHtml = convertBlobUrlsToReal(essayAnswerHtml.value);

    // Build content and answer based on question type
    let contentJson = '';
    let answerJson = '';

    if (CHOICE_TYPES.includes(questionType)) {
      // Validate options
      const validOptions = options.value.filter((o) => o.text.trim());
      if (validOptions.length < 2) {
        message.warning('请至少填写2个选项');
        return;
      }
      if (correctAnswers.value.length === 0) {
        message.warning('请标记正确答案');
        return;
      }
      contentJson = JSON.stringify(validOptions);
      answerJson = JSON.stringify({ correct: correctAnswers.value });
    } else if (questionType === 'true_false') {
      contentJson = JSON.stringify({ type: 'true_false' });
      answerJson = JSON.stringify({ correct: tfAnswer.value });
    } else if (questionType === 'fill_blank') {
      // Count blanks in stem
      const blankCount = (realStemHtml.match(/_{3,}|（\s*）|\(\s*\)/g) || []).length;
      if (blankCount > 0) {
        let blanks: string[] = [];
        try {
          blanks = JSON.parse(realEssayAnswerHtml || '[]');
          if (!Array.isArray(blanks)) blanks = [];
        } catch {
          blanks = [];
        }
        if (blanks.length === 0 && realEssayAnswerHtml) {
          blanks = [realEssayAnswerHtml];
        }
        contentJson = JSON.stringify({ blankCount });
        answerJson = JSON.stringify({ blanks });
      } else {
        contentJson = '';
        answerJson = htmlToJson(realEssayAnswerHtml);
      }
    } else {
      // Essay, programming, etc.
      contentJson = '';
      answerJson = htmlToJson(realEssayAnswerHtml);
    }

    drawerApi.lock();

    const submitData = {
      ...values,
      stem: htmlToJson(realStemHtml),
      content: contentJson ? contentJson : htmlToJson(''),
      answer: answerJson,
      analysis: JSON.stringify({
        text: realAnalysisTextHtml || '',
        media: realAnalysisMediaHtml || '',
      }),
    };

    (id.value ? updateQuestion(id.value, submitData) : createQuestion(submitData))
      .then(async () => {
        await drawerApi.close();
        emits('success');
      })
      .catch(() => {
        drawerApi.unlock();
      });
  },
  async onOpenChange(isOpen) {
    if (isOpen) {
      const data = drawerApi.getData<QuestionApi.Question>();
      formApi.resetForm();
      // Clear blob URL mapping
      for (const blobUrl of blobToRealUrl.keys()) {
        URL.revokeObjectURL(blobUrl);
      }
      blobToRealUrl.clear();

      if (data?.id) {
        formData.value = data;
        id.value = data.id;
        await formApi.setValues({
          title: data.title,
          questionType: data.questionType,
          difficulty: data.difficulty,
          resourceType: data.resourceType,
          examId: data.examId || undefined,
          subjectId: data.subjectId || undefined,
          categoryId: data.categoryId || undefined,
        });

        // Load subject options if exam is set
        if (data.examId) {
          await loadSubjectOptions(data.examId);
          // Re-set subjectId after options are loaded
          if (data.subjectId) {
            await formApi.setValues({ subjectId: data.subjectId });
          }
        }
        currentType.value = data.questionType || '';

        // Decode and fix image URLs, then convert to blob URLs for display
        const rawStem = fixImageUrls(safeJsonParse(data.stem));
        const rawEssayAnswer = fixImageUrls(safeJsonParse(data.answer));

        // Parse analysis: handle new {text, media} format and legacy plain HTML
        let rawAnalysisText = '';
        let rawAnalysisMedia = '';
        try {
          const parsed = safeJsonParse(data.analysis);
          if (parsed && typeof parsed === 'object' && ('text' in parsed || 'media' in parsed)) {
            // New format: {text, media}
            rawAnalysisText = fixImageUrls(parsed.text || '');
            rawAnalysisMedia = fixImageUrls(parsed.media || '');
          } else if (typeof parsed === 'string') {
            // Legacy format: plain HTML string → put in text field
            rawAnalysisText = fixImageUrls(parsed);
          }
        } catch {
          // Fallback: treat as plain HTML
          rawAnalysisText = fixImageUrls(safeJsonParse(data.analysis) || '');
        }

        // Convert images to blob URLs (authenticated fetch)
        const [stemWithBlobs, analysisTextWithBlobs, analysisMediaWithBlobs, essayWithBlobs] = await Promise.all([
          convertMediaToBlobUrls(rawStem),
          convertMediaToBlobUrls(rawAnalysisText),
          convertMediaToBlobUrls(rawAnalysisMedia),
          convertMediaToBlobUrls(rawEssayAnswer),
        ]);

        stemHtml.value = stemWithBlobs;
        analysisTextHtml.value = analysisTextWithBlobs;
        analysisMediaHtml.value = analysisMediaWithBlobs;
        essayAnswerHtml.value = essayWithBlobs;

        // Decode type-specific content and answer
        const questionType = data.questionType || '';
        if (CHOICE_TYPES.includes(questionType)) {
          // Parse options from content
          try {
            const parsed = JSON.parse(data.content || '[]');
            if (Array.isArray(parsed) && parsed.length > 0) {
              options.value = parsed.map((o: any, i: number) => ({
                id: o.id || getOptionId(i),
                text: o.text || o.label || '',
              }));
            }
          } catch {
            options.value = [
              { id: 'A', text: '' },
              { id: 'B', text: '' },
            ];
          }
          // Parse correct answer
          try {
            const ans = JSON.parse(data.answer || '{}');
            correctAnswers.value = ans.correct || [];
          } catch {
            correctAnswers.value = [];
          }
        } else if (questionType === 'true_false') {
          try {
            const ans = JSON.parse(data.answer || '{}');
            tfAnswer.value = ans.correct || 'true';
          } catch {
            tfAnswer.value = 'true';
          }
        }
        // essayAnswerHtml already set above with blob URLs
      } else {
        formData.value = undefined;
        id.value = undefined;
        currentType.value = '';
        stemHtml.value = '';
        analysisTextHtml.value = '';
        analysisMediaHtml.value = '';
        essayAnswerHtml.value = '';
        options.value = [
          { id: 'A', text: '' },
          { id: 'B', text: '' },
        ];
        correctAnswers.value = [];
        tfAnswer.value = 'true';
      }
    }
  },
});

const drawerTitle = computed(() => {
  return id.value ? '编辑题目' : '新增题目';
});
</script>

<template>
  <Drawer :title="drawerTitle" class="w-[960px]">
    <div class="flex flex-col gap-6 p-4">
      <!-- Basic fields -->
      <Form />

      <!-- Stem (always shown, required) -->
      <div>
        <label class="mb-2 block text-sm font-medium">
          题干 <span class="text-red-500">*</span>
        </label>
        <VbenTiptap
          v-model="stemHtml"
          :image-upload="imageUploadConfig"
          :min-height="180"
          :max-height="350"
          placeholder="请输入题目内容..."
        />
      </div>

      <!-- Choice question options -->
      <div v-if="isChoiceType" class="rounded-lg border border-gray-200 p-4">
        <div class="mb-3 flex items-center justify-between">
          <label class="text-sm font-medium">
            选项
            <span class="ml-1 text-xs text-gray-400">
              （点击选项标记正确答案{{ currentType === 'single_choice' ? '，单选' : '，可多选' }}）
            </span>
          </label>
          <Button size="small" @click="addOption">
            <Plus class="mr-1 size-3" />
            添加选项
          </Button>
        </div>

        <div class="space-y-3">
          <div
            v-for="(opt, index) in options"
            :key="index"
            class="flex items-start gap-3"
          >
            <!-- Correct answer indicator -->
            <div
              class="mt-2 flex size-7 shrink-0 cursor-pointer items-center justify-center rounded-full border-2 text-sm font-bold transition-colors"
              :class="
                correctAnswers.includes(opt.id)
                  ? 'border-green-500 bg-green-500 text-white'
                  : 'border-gray-300 text-gray-400 hover:border-green-300'
              "
              @click="toggleCorrectAnswer(opt.id)"
            >
              {{ opt.id }}
            </div>

            <!-- Option text input -->
            <div class="flex-1">
              <Input
                v-model:value="opt.text"
                :placeholder="`选项 ${opt.id} 内容`"
                size="large"
              />
            </div>

            <!-- Remove button -->
            <Tooltip title="删除选项">
              <Button
                type="text"
                danger
                size="small"
                :disabled="options.length <= 2"
                @click="removeOption(index)"
              >
                ✕
              </Button>
            </Tooltip>
          </div>
        </div>

        <div
          v-if="correctAnswers.length > 0"
          class="mt-3 text-xs text-green-600"
        >
          ✓ 正确答案：{{ correctAnswers.join(', ') }}
        </div>
      </div>

      <!-- True/False selector -->
      <div v-if="isTrueFalse" class="rounded-lg border border-gray-200 p-4">
        <label class="mb-3 block text-sm font-medium">正确答案</label>
        <RadioGroup v-model:value="tfAnswer" button-style="solid">
          <Radio.Button value="true">正确 (✓)</Radio.Button>
          <Radio.Button value="false">错误 (✗)</Radio.Button>
        </RadioGroup>
      </div>

      <!-- Fill-blank or Essay answer -->
      <div v-if="!isChoiceType && !isTrueFalse">
        <label class="mb-2 block text-sm font-medium">
          {{ isFillBlank ? '各空答案' : '参考答案' }}
        </label>
        <div v-if="isFillBlank" class="mb-1 text-xs text-gray-400">
          请在题干中用 ___（三个下划线以上）标记空位，然后在此填写每个空的答案
        </div>
        <VbenTiptap
          v-model="essayAnswerHtml"
          :image-upload="imageUploadConfig"
          :min-height="120"
          :max-height="250"
          :placeholder="isFillBlank ? '按顺序填写每个空的答案...' : '参考答案...'"
        />
      </div>

      <!-- Analysis: text + media (always shown) -->
      <div>
        <label class="mb-2 block text-sm font-medium">文字解析</label>
        <VbenTiptap
          v-model="analysisTextHtml"
          :min-height="120"
          :max-height="250"
          placeholder="输入文字解析..."
        />
      </div>
      <div>
        <label class="mb-2 block text-sm font-medium">图片/视频解析</label>
        <VbenTiptap
          v-model="analysisMediaHtml"
          :image-upload="imageUploadConfig"
          :min-height="150"
          :max-height="300"
          placeholder="上传图片或视频解析..."
        />
      </div>
    </div>
  </Drawer>
</template>
