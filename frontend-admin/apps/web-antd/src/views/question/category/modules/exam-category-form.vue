<script lang="ts" setup>
import type { ExamCategoryApi } from '#/api/question/category';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import { createExamCategory, updateExamCategory } from '#/api/question/category';

import { useExamCategoryDrawerSchema } from '../data';

const emits = defineEmits(['success']);

const formData = ref<ExamCategoryApi.ExamCategory>();
const id = ref<number>();

const [Form, formApi] = useVbenForm({
  schema: useExamCategoryDrawerSchema(),
  showDefaultActions: false,
});

const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const values = await formApi.getValues();
    drawerApi.lock();
    (id.value ? updateExamCategory(id.value, values) : createExamCategory(values))
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
      const data = drawerApi.getData<ExamCategoryApi.ExamCategory>();
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
  return id.value ? '编辑考试大类' : '新增考试大类';
});
</script>
<template>
  <Drawer :title="drawerTitle">
    <Form />
  </Drawer>
</template>
