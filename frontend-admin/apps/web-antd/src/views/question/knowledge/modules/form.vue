<script lang="ts" setup>
import type { KnowledgePointApi } from '#/api/question/knowledge';

import { computed, ref, watch } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import {
  createKnowledgePoint,
  updateKnowledgePoint,
} from '#/api/question/knowledge';
import { getSubjectAll } from '#/api/question/category';

import { useKnowledgePointDrawerSchema } from '../data';

const emits = defineEmits(['success']);

const formData = ref<KnowledgePointApi.KnowledgePoint>();
const id = ref<number>();
const examOptions = ref<any[]>([]);
const subjectOptions = ref<any[]>([]);
const categoryOptions = ref<any[]>([]);
const parentOptions = ref<any[]>([]);

const [Form, formApi] = useVbenForm({
  schema: useKnowledgePointDrawerSchema(
    examOptions.value,
    subjectOptions.value,
    categoryOptions.value,
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
        ? updateKnowledgePoint(id.value, values)
        : createKnowledgePoint(values)
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
      if (data?.categoryOptions) categoryOptions.value = data.categoryOptions;
      if (data?.parentOptions) parentOptions.value = data.parentOptions;
      // Sync options into form schema so selects display correctly
      await formApi.updateSchema([
        { fieldName: 'examId', componentProps: { options: examOptions.value } },
        { fieldName: 'subjectId', componentProps: { options: subjectOptions.value } },
        { fieldName: 'categoryId', componentProps: { options: categoryOptions.value } },
        { fieldName: 'parentId', componentProps: { treeData: parentOptions.value } },
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
  return id.value ? '编辑知识点' : '新增知识点';
});

// Watch examId changes to load subject options (cascading)
const examIdRef = formApi.useStore((state: any) => state.values?.examId);
watch(
  () => examIdRef.value,
  async (newExamId) => {
    if (newExamId) {
      try {
        const res = await getSubjectAll(newExamId);
        subjectOptions.value = (res || []).map((item: any) => ({
          label: item.name,
          value: item.id,
        }));
      } catch {
        subjectOptions.value = [];
      }
    } else {
      subjectOptions.value = [];
    }
    await formApi.updateSchema([
      { fieldName: 'subjectId', componentProps: { options: subjectOptions.value } },
    ]);
  },
);
</script>
<template>
  <Drawer :title="drawerTitle">
    <Form />
  </Drawer>
</template>
