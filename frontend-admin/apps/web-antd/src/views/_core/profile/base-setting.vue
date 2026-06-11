<script setup lang="ts">
import type { VbenFormSchema } from '#/adapter/form';

import { onMounted, ref } from 'vue';

import { ProfileBaseSetting } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { getRoleList, getUserInfoApi, updateProfile } from '#/api';

const profileBaseSettingRef = ref();

// 表单配置，无需使用 computed 包裹，因为 schema 内容是静态的
const formSchema: VbenFormSchema[] = [
  {
    fieldName: 'realName',
    component: 'Input',
    label: '姓名',
  },
  {
    fieldName: 'username',
    component: 'Input',
    label: '用户名',
    // 用户名不可编辑
    componentProps: {
      disabled: true,
    },
  },
  {
    fieldName: 'roles',
    component: 'ApiSelect',
    componentProps: {
      // 从 API 动态加载角色列表
      api: async () => {
        const res = await getRoleList({ page: 1, pageSize: 100 });
        return res || [];
      },
      labelField: 'name',
      valueField: 'name',
      mode: 'multiple',
      disabled: true,
    },
    label: '角色',
  },
  {
    fieldName: 'introduction',
    component: 'Textarea',
    label: '个人简介',
  },
];

/** 处理表单提交，将表单字段映射为后端 UpdateProfileRequest 格式并保存 */
async function handleSubmit(values: Record<string, any>) {
  try {
    // 表单字段 realName -> 后端 nickname，introduction -> 后端 bio
    await updateProfile({
      nickname: values.realName,
      bio: values.introduction,
    });
    message.success('资料更新成功');
  } catch {
    message.error('资料更新失败');
  }
}

onMounted(async () => {
  const data = await getUserInfoApi();
  profileBaseSettingRef.value.getFormApi().setValues(data);
});
</script>
<template>
  <ProfileBaseSetting
    ref="profileBaseSettingRef"
    :form-schema="formSchema"
    @submit="handleSubmit"
  />
</template>
