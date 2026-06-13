<script lang="ts" setup>
import type { QuestionSourceApi } from '#/api/question/source';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import {
  createQuestionSource,
  updateQuestionSource,
} from '#/api/question/source';

import { useSourceDrawerSchema } from '../data';

const emits = defineEmits(['success']);

const formData = ref<QuestionSourceApi.QuestionSource>();
const id = ref<number>();
const examOptions = ref<any[]>([]);

const [Form, formApi] = useVbenForm({
  schema: useSourceDrawerSchema(examOptions.value),
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
        ? updateQuestionSource(id.value, values)
        : createQuestionSource(values)
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
      // Sync options into form schema so selects display correctly
      await formApi.updateSchema([
        { fieldName: 'examId', componentProps: { options: examOptions.value } },
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
  return id.value ? '编辑来源' : '新增来源';
});
</script>
<template>
  <Drawer :title="drawerTitle">
    <Form />
  </Drawer>
</template>
