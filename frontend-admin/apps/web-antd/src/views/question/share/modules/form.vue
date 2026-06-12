<script lang="ts" setup>
import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { useVbenForm } from '#/adapter/form';
import { createQuestionShare } from '#/api/question/share';

import { SHARE_TYPE_OPTIONS } from '../data';

const emits = defineEmits(['success']);

const id = ref<number>();

const [Form, formApi] = useVbenForm({
  schema: [
    {
      component: 'InputNumber',
      fieldName: 'questionId',
      label: '题目ID',
      rules: 'required',
      componentProps: { min: 1, class: 'w-full' },
    },
    {
      component: 'Select',
      fieldName: 'shareType',
      label: '分享类型',
      rules: 'required',
      componentProps: {
        options: SHARE_TYPE_OPTIONS,
        class: 'w-full',
      },
    },
    {
      component: 'InputNumber',
      fieldName: 'targetId',
      label: '目标ID',
      componentProps: { min: 0, class: 'w-full' },
    },
    {
      component: 'InputNumber',
      fieldName: 'expireHours',
      label: '有效时间(小时)',
      componentProps: { min: 0, class: 'w-full' },
    },
    {
      component: 'InputNumber',
      fieldName: 'maxAccess',
      label: '最大访问次数',
      componentProps: { min: 0, class: 'w-full' },
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
    createQuestionShare(values)
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
  return '创建分享';
});
</script>
<template>
  <Drawer :title="drawerTitle">
    <Form />
  </Drawer>
</template>
