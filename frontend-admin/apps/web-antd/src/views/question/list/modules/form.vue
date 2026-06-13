<script lang="ts" setup>
import type { QuestionApi } from '#/api/question/question';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { VbenTiptap } from '@vben/plugins/tiptap';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { simpleUpload } from '#/api/file';
import { createQuestion, updateQuestion } from '#/api/question/question';

import {
  DIFFICULTY_OPTIONS,
  QUESTION_TYPE_OPTIONS,
  RESOURCE_TYPE_OPTIONS,
} from '../data';

const emits = defineEmits(['success']);

const formData = ref<QuestionApi.Question>();
const id = ref<number>();

// Rich text field values
const stemHtml = ref('');
const contentHtml = ref('');
const answerHtml = ref('');
const analysisHtml = ref('');

// Decode JSON-encoded HTML string for TipTap
function safeJsonParse(jsonStr: string): string {
  if (!jsonStr || jsonStr === 'null') return '';
  try {
    const parsed = JSON.parse(jsonStr);
    return typeof parsed === 'string' ? parsed : jsonStr;
  } catch {
    return jsonStr;
  }
}

// Encode HTML string to JSON for backend
function htmlToJson(html: string): string {
  return JSON.stringify(html || '');
}

// Image upload adapter for TipTap
const imageUploadConfig = {
  accept: 'image/*',
  maxSize: 10 * 1024 * 1024, // 10MB
  upload: async (file: File, onProgress?: (percent: number) => void) => {
    const result = await simpleUpload(file, (event) => {
      onProgress?.(event.percent);
    });
    return result.url;
  },
};

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
  ],
  showDefaultActions: false,
});

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
    drawerApi.lock();

    const submitData = {
      ...values,
      stem: htmlToJson(stemHtml.value),
      content: htmlToJson(contentHtml.value),
      answer: htmlToJson(answerHtml.value),
      analysis: htmlToJson(analysisHtml.value),
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

      if (data?.id) {
        formData.value = data;
        id.value = data.id;
        await formApi.setValues({
          title: data.title,
          questionType: data.questionType,
          difficulty: data.difficulty,
          resourceType: data.resourceType,
        });
        // Decode JSON fields for TipTap editors
        stemHtml.value = safeJsonParse(data.stem);
        contentHtml.value = safeJsonParse(data.content);
        answerHtml.value = safeJsonParse(data.answer);
        analysisHtml.value = safeJsonParse(data.analysis);
      } else {
        formData.value = undefined;
        id.value = undefined;
        stemHtml.value = '';
        contentHtml.value = '';
        answerHtml.value = '';
        analysisHtml.value = '';
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

      <!-- Rich text fields -->
      <div class="space-y-5">
        <!-- Stem (required) -->
        <div>
          <label class="mb-2 block text-sm font-medium">
            题干 <span class="text-red-500">*</span>
          </label>
          <VbenTiptap
            v-model="stemHtml"
            :image-upload="imageUploadConfig"
            :min-height="200"
            :max-height="400"
            placeholder="请输入题干内容，支持富文本编辑和图片上传..."
          />
        </div>

        <!-- Content (options, etc.) -->
        <div>
          <label class="mb-2 block text-sm font-medium">题目内容</label>
          <div class="mb-1 text-xs text-gray-400">
            选择题填写选项、填空题填写空位等结构化内容
          </div>
          <VbenTiptap
            v-model="contentHtml"
            :image-upload="imageUploadConfig"
            :min-height="150"
            :max-height="300"
            placeholder="选项、填空等结构化内容..."
          />
        </div>

        <!-- Answer -->
        <div>
          <label class="mb-2 block text-sm font-medium">答案</label>
          <VbenTiptap
            v-model="answerHtml"
            :image-upload="imageUploadConfig"
            :min-height="120"
            :max-height="250"
            placeholder="标准答案..."
          />
        </div>

        <!-- Analysis -->
        <div>
          <label class="mb-2 block text-sm font-medium">解析</label>
          <VbenTiptap
            v-model="analysisHtml"
            :image-upload="imageUploadConfig"
            :min-height="150"
            :max-height="300"
            placeholder="答案解析..."
          />
        </div>
      </div>
    </div>
  </Drawer>
</template>
