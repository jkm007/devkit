<script lang="ts" setup>
import type { SubjectApi } from '#/api/question/category';

import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import { createSubject, updateSubject } from '#/api/question/category';

import { useSubjectDrawerSchema } from '../data';

const emits = defineEmits(['success']);

const formData = ref<SubjectApi.Subject>();
const id = ref<number>();
const examOptions = ref<any[]>([]);

const [Form, formApi] = useVbenForm({
  schema: useSubjectDrawerSchema(examOptions.value),
  showDefaultActions: false,
});

const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const values = await formApi.getValues();
    drawerApi.lock();
    (id.value ? updateSubject(id.value, values) : createSubject(values))
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
      if (data?.examOptions) {
        examOptions.value = data.examOptions;
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
  return id.value ? '编辑科目' : '新增科目';
});
</script>
<template>
  <Drawer :title="drawerTitle">
    <Form />
  </Drawer>
</template>
