<script lang="ts" setup>
import type { ExamApi } from '#/api/question/category';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import { createExam, updateExam } from '#/api/question/category';

import { useExamDrawerSchema } from '../data';

const emits = defineEmits(['success']);

const formData = ref<ExamApi.Exam>();
const id = ref<number>();
const categoryOptions = ref<any[]>([]);

const [Form, formApi] = useVbenForm({
  schema: useExamDrawerSchema(categoryOptions.value),
  showDefaultActions: false,
});

const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const values = await formApi.getValues();
    drawerApi.lock();
    (id.value ? updateExam(id.value, values) : createExam(values))
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
      if (data?.categoryOptions) {
        categoryOptions.value = data.categoryOptions;
        formApi.updateSchema([
          {
            fieldName: 'examCategoryId',
            componentProps: { options: categoryOptions.value },
          },
        ]);
      }
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
  return id.value ? '编辑考试' : '新增考试';
});
</script>
<template>
  <Drawer :title="drawerTitle">
    <Form />
  </Drawer>
</template>
