<script lang="ts" setup>
import type { QuestionApi } from '#/api/question/question';

import { ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { VbenTiptapPreview } from '@vben/plugins/tiptap';

import { Tag } from 'ant-design-vue';

import {
  DIFFICULTY_OPTIONS,
  QUESTION_TYPE_OPTIONS,
  STATUS_OPTIONS,
} from '../data';

const questionData = ref<QuestionApi.Question | null>(null);

// Decode JSON-encoded HTML string
function safeJsonParse(jsonStr: string): string {
  if (!jsonStr || jsonStr === 'null') return '';
  try {
    const parsed = JSON.parse(jsonStr);
    return typeof parsed === 'string' ? parsed : jsonStr;
  } catch {
    return jsonStr;
  }
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
  const map: Record<number, string> = {
    1: 'green',
    2: 'orange',
    3: 'red',
  };
  return map[value] || 'default';
}

const [Drawer, drawerApi] = useVbenDrawer({
  onOpenChange(isOpen) {
    if (isOpen) {
      questionData.value = drawerApi.getData<QuestionApi.Question>();
    }
  },
});

// Computed content fields
function getStem(): string {
  return safeJsonParse(questionData.value?.stem || '');
}
function getContent(): string {
  return safeJsonParse(questionData.value?.content || '');
}
function getAnswer(): string {
  return safeJsonParse(questionData.value?.answer || '');
}
function getAnalysis(): string {
  return safeJsonParse(questionData.value?.analysis || '');
}
</script>

<template>
  <Drawer :footer="false" title="题目预览" class="w-[800px]">
    <div v-if="questionData" class="space-y-6 p-4">
      <!-- Header info -->
      <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
        <h2 class="mb-3 text-lg font-semibold">{{ questionData.title }}</h2>
        <div class="flex flex-wrap gap-3">
          <div class="flex items-center gap-1">
            <span class="text-sm text-gray-500">题型:</span>
            <Tag color="blue">
              {{ getLabel(QUESTION_TYPE_OPTIONS, questionData.questionType) }}
            </Tag>
          </div>
          <div class="flex items-center gap-1">
            <span class="text-sm text-gray-500">难度:</span>
            <Tag :color="getDifficultyColor(questionData.difficulty)">
              {{ getLabel(DIFFICULTY_OPTIONS, questionData.difficulty) }}
            </Tag>
          </div>
          <div class="flex items-center gap-1">
            <span class="text-sm text-gray-500">状态:</span>
            <Tag :color="getStatusColor(questionData.status)">
              {{ getLabel(STATUS_OPTIONS, questionData.status) }}
            </Tag>
          </div>
        </div>
      </div>

      <!-- Stem -->
      <div v-if="getStem()">
        <h3 class="mb-2 text-sm font-medium text-gray-700">题干</h3>
        <div class="rounded-lg border border-gray-200 p-4">
          <VbenTiptapPreview :content="getStem()" :min-height="100" />
        </div>
      </div>

      <!-- Content -->
      <div v-if="getContent()">
        <h3 class="mb-2 text-sm font-medium text-gray-700">题目内容</h3>
        <div class="rounded-lg border border-gray-200 p-4">
          <VbenTiptapPreview :content="getContent()" :min-height="80" />
        </div>
      </div>

      <!-- Answer -->
      <div v-if="getAnswer()">
        <h3 class="mb-2 text-sm font-medium text-gray-700">答案</h3>
        <div class="rounded-lg border border-gray-200 p-4">
          <VbenTiptapPreview :content="getAnswer()" :min-height="60" />
        </div>
      </div>

      <!-- Analysis -->
      <div v-if="getAnalysis()">
        <h3 class="mb-2 text-sm font-medium text-gray-700">解析</h3>
        <div class="rounded-lg border border-gray-200 p-4">
          <VbenTiptapPreview :content="getAnalysis()" :min-height="80" />
        </div>
      </div>

      <!-- Empty state -->
      <div
        v-if="!getStem() && !getContent() && !getAnswer() && !getAnalysis()"
        class="py-8 text-center text-gray-400"
      >
        暂无内容
      </div>
    </div>
  </Drawer>
</template>
