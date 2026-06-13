<script lang="ts" setup>
import type { QuestionCategoryApi } from '#/api/question/category';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import {
  createQuestionCategory,
  updateQuestionCategory,
} from '#/api/question/category';

import { useCategoryDrawerSchema } from '../data';

const emits = defineEmits(['success']);

const formData = ref<QuestionCategoryApi.QuestionCategory>();
const id = ref<number>();
const examOptions = ref<any[]>([]);
const subjectOptions = ref<any[]>([]);
const parentOptions = ref<any[]>([]);
const onExamChangeCallback = ref<((examId: number) => Promise<any[]>) | null>(null);

const [Form, formApi] = useVbenForm({
  schema: useCategoryDrawerSchema(
    examOptions.value,
    subjectOptions.value,
    parentOptions.value,
  ),
  showDefaultActions: false,
});

const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const values = await formApi.getValues();
    drawerApi.lock();
    (
      id.value
        ? updateQuestionCategory(id.value, values)
        : createQuestionCategory(values)
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
      const data = drawerApi.getData<any>();
      formApi.resetForm();
      if (data?.examOptions) examOptions.value = data.examOptions;
      if (data?.subjectOptions) subjectOptions.value = data.subjectOptions;
      if (data?.parentOptions) parentOptions.value = data.parentOptions;
      if (data?.onExamChange) onExamChangeCallback.value = data.onExamChange;
      formApi.updateSchema([
        {
          fieldName: 'examId',
          componentProps: {
            options: examOptions.value,
            onChange(value: number) {
              if (onExamChangeCallback.value && value) {
                onExamChangeCallback.value(value).then((options) => {
                  subjectOptions.value = options;
                  formApi.updateSchema([
                    {
                      fieldName: 'subjectId',
                      componentProps: { options },
                    },
                  ]);
                  formApi.setValues({ subjectId: undefined });
                });
              } else {
                subjectOptions.value = [];
                formApi.updateSchema([
                  {
                    fieldName: 'subjectId',
                    componentProps: { options: [] },
                  },
                ]);
                formApi.setValues({ subjectId: undefined });
              }
            },
          },
        },
        {
          fieldName: 'subjectId',
          componentProps: { options: subjectOptions.value },
        },
        {
          fieldName: 'parentId',
          componentProps: { treeData: parentOptions.value },
        },
      ]);
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
  return id.value ? '编辑分类' : '新增分类';
});
</script>
<template>
  <Drawer :title="drawerTitle">
    <Form />
  </Drawer>
</template>
