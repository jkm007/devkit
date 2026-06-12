<script lang="ts" setup>
import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import { createImportTask } from '#/api/question/import';

const emits = defineEmits(['success']);

const id = ref<number>();

const [Form, formApi] = useVbenForm({
  schema: [
    {
      component: 'InputNumber',
      fieldName: 'fileId',
      label: '文件ID',
      rules: 'required',
      componentProps: { min: 1, class: 'w-full' },
    },
    {
      component: 'Input',
      fieldName: 'fileName',
      label: '文件名',
      rules: 'required',
    },
    {
      component: 'Select',
      fieldName: 'fileType',
      label: '文件类型',
      rules: 'required',
      componentProps: {
        options: [
          { label: 'Excel', value: 'excel' },
          { label: 'Word', value: 'word' },
          { label: 'PDF', value: 'pdf' },
          { label: 'ZIP', value: 'zip' },
        ],
        class: 'w-full',
      },
    },
    {
      component: 'InputNumber',
      fieldName: 'targetCategoryId',
      label: '目标分类ID',
      componentProps: { min: 0, class: 'w-full' },
    },
    {
      component: 'Select',
      fieldName: 'targetResourceType',
      label: '目标资源类型',
      defaultValue: 'private',
      componentProps: {
        options: [
          { label: '私有', value: 'private' },
          { label: '公共', value: 'public' },
          { label: '分组', value: 'group' },
        ],
        class: 'w-full',
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
    createImportTask(values)
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
      formApi.resetForm();
      id.value = undefined;
    }
  },
});

const drawerTitle = computed(() => {
  return '新建导入任务';
});
</script>
<template>
  <Drawer :title="drawerTitle">
    <Form />
  </Drawer>
</template>
