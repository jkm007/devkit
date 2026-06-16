<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { Banner } from '#/api/system/banner';

import { ref } from 'vue';

import { useVbenDrawer, useVbenForm } from '@vben/common-ui';
import { message } from 'ant-design-vue';

import { createBanner, updateBanner } from '#/api/system/banner';

const emit = defineEmits<{
  success: [];
}>();

const modalMode = ref<'create' | 'edit'>('create');
const currentBanner = ref<Banner | null>(null);

const linkTypeOptions = [
  { label: '内部链接', value: 'internal' },
  { label: '外部链接', value: 'external' },
  { label: '无链接', value: 'none' },
];

const statusOptions = [
  { label: '启用', value: 'enabled' },
  { label: '禁用', value: 'disabled' },
];

function useFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      componentProps: {
        placeholder: '请输入轮播图标题',
      },
      fieldName: 'title',
      label: '标题',
      rules: 'required',
    },
    {
      component: 'Upload',
      componentProps: {
        accept: 'image/*',
        listType: 'picture-card',
        maxCount: 1,
      },
      fieldName: 'image',
      label: '图片',
      rules: 'required',
    },
    {
      component: 'Select',
      componentProps: {
        options: linkTypeOptions,
      },
      fieldName: 'linkType',
      label: '链接类型',
      defaultValue: 'none',
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: '请输入链接地址',
      },
      dependencies: {
        triggerFields: ['linkType'],
        if: (values) => values.linkType !== 'none',
        show: true,
      },
      fieldName: 'link',
      label: '链接',
    },
    {
      component: 'InputNumber',
      componentProps: {
        min: 0,
        max: 999,
      },
      fieldName: 'sortOrder',
      label: '排序',
      defaultValue: 0,
    },
    {
      component: 'Select',
      componentProps: {
        options: statusOptions,
      },
      fieldName: 'status',
      label: '状态',
      defaultValue: 'enabled',
      dependencies: {
        triggerFields: [],
        if: () => modalMode.value === 'edit',
        show: true,
      },
    },
  ];
}

const [Form, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
  },
  schema: useFormSchema(),
  showDefaultActions: false,
});

const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const values = await formApi.getValues();
    try {
      if (modalMode.value === 'create') {
        await createBanner(values);
        message.success('创建成功');
      } else if (currentBanner.value) {
        await updateBanner(currentBanner.value.id, values);
        message.success('更新成功');
      }
      emit('success');
      drawerApi.close();
    } catch (error: any) {
      message.error(error.message || '操作失败');
    }
  },
  async onOpenChange(isOpen) {
    if (isOpen) {
      const data = drawerApi.getData<Banner>();
      formApi.resetForm();

      if (data) {
        modalMode.value = 'edit';
        currentBanner.value = data;
        formApi.setValues({
          title: data.title,
          image: data.image,
          link: data.link || '',
          linkType: data.linkType || 'none',
          sortOrder: data.sortOrder || 0,
          status: data.status || 'enabled',
        });
      } else {
        modalMode.value = 'create';
        currentBanner.value = null;
      }
    }
  },
});
</script>

<template>
  <Drawer
    :title="modalMode === 'create' ? '新增轮播图' : '编辑轮播图'"
    class="w-[600px]"
  >
    <Form />
  </Drawer>
</template>
