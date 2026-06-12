<script lang="ts" setup>
import type { QuestionApi } from '#/api/question/question';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import { createQuestion, updateQuestion } from '#/api/question/question';

import {
  DIFFICULTY_OPTIONS,
  QUESTION_TYPE_OPTIONS,
  RESOURCE_TYPE_OPTIONS,
} from '../data';

const emits = defineEmits(['success']);

const formData = ref<QuestionApi.Question>();
const id = ref<number>();

const [Form, formApi] = useVbenForm({
  schema: [
    {
      component: 'Input',
      fieldName: 'title',
      label: '题目标题',
      rules: 'required',
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
      componentProps: {
        options: DIFFICULTY_OPTIONS,
        class: 'w-full',
      },
    },
    {
      component: 'Select',
      fieldName: 'resourceType',
      label: '资源类型',
      defaultValue: 'private',
      componentProps: {
        options: RESOURCE_TYPE_OPTIONS,
        class: 'w-full',
      },
    },
    {
      component: 'Textarea',
      fieldName: 'stem',
      label: '题干(JSON)',
      rules: 'required',
      componentProps: {
        rows: 6,
        placeholder: '请输入题干JSON内容',
      },
    },
    {
      component: 'Textarea',
      fieldName: 'content',
      label: '题目内容(JSON)',
      componentProps: {
        rows: 4,
        placeholder: '选项、填空等结构化内容',
      },
    },
    {
      component: 'Textarea',
      fieldName: 'answer',
      label: '答案(JSON)',
      componentProps: {
        rows: 3,
        placeholder: '标准答案',
      },
    },
    {
      component: 'Textarea',
      fieldName: 'analysis',
      label: '解析(JSON)',
      componentProps: {
        rows: 4,
        placeholder: '答案解析',
      },
    },
  ],
  showDefaultActions: false,
});

const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const values = await formApi.getValues();
    drawerApi.lock();
    (
      id.value ? updateQuestion(id.value, values) : createQuestion(values)
    )
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
        await formApi.setValues(data);
      } else {
        formData.value = undefined;
        id.value = undefined;
      }
    }
  },
});

const drawerTitle = computed(() => {
  return id.value ? '编辑题目' : '新增题目';
});
</script>
<template>
  <Drawer :title="drawerTitle" class="w-[800px]">
    <Form />
  </Drawer>
</template>
