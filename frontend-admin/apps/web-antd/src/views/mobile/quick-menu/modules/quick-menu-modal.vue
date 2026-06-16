<script lang="ts" setup>
import { ref, computed } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';
import { Form, FormItem, Input, InputNumber, Select, SelectOption, Switch } from 'ant-design-vue';

import {
  createQuickMenu,
  updateQuickMenu,
} from '#/api/system/mobile-config';

const props = defineProps<{
  data?: any;
  mode: 'create' | 'edit';
}>();

const emit = defineEmits(['success']);

const [Modal, modalApi] = useVbenModal({
  onConfirm: handleSubmit,
});

const form = ref({
  title: '',
  icon: '',
  link: '',
  linkType: 'page',
  sortOrder: 0,
  status: 'enabled',
});

const linkTypeOptions = [
  { value: 'page', label: '页面跳转' },
  { value: 'url', label: '外部链接' },
  { value: 'function', label: '功能' },
  { value: 'none', label: '无链接' },
];

const iconOptions = [
  '📝', '📖', '🎯', '📊', '❓', '💡', '🔔', '📢',
  '📚', '🎓', '✏️', '📋', '🔍', '⭐', '🏆', '💪',
  '📱', '💻', '🌐', '📞', '💬', '❤️', '🎁', '🔑',
];

const isEdit = computed(() => props.mode === 'edit');

function handleSubmit() {
  modalApi.lock();
  const api = isEdit.value
    ? updateQuickMenu(props.data.id, form.value)
    : createQuickMenu(form.value);

  api
    .then(() => {
      message.success(isEdit.value ? '更新成功' : '创建成功');
      emit('success');
      modalApi.close();
    })
    .catch((error: any) => {
      message.error(error.message || '操作失败');
    })
    .finally(() => {
      modalApi.lock(false);
    });
}

// 初始化表单
if (props.data) {
  form.value = { ...props.data };
}
</script>

<template>
  <Modal :title="isEdit ? '编辑快捷菜单' : '新增快捷菜单'">
    <Form layout="vertical" :model="form">
      <FormItem label="标题" required>
        <Input v-model:value="form.title" placeholder="请输入标题" />
      </FormItem>

      <FormItem label="图标" required>
        <Select v-model:value="form.icon" placeholder="请选择图标">
          <SelectOption v-for="icon in iconOptions" :key="icon" :value="icon">
            <span class="text-xl mr-2">{{ icon }}</span>
            <span>{{ icon }}</span>
          </SelectOption>
        </Select>
      </FormItem>

      <FormItem label="链接类型">
        <Select v-model:value="form.linkType">
          <SelectOption
            v-for="option in linkTypeOptions"
            :key="option.value"
            :value="option.value"
          >
            {{ option.label }}
          </SelectOption>
        </Select>
      </FormItem>

      <FormItem v-if="form.linkType !== 'none'" label="链接地址">
        <Input
          v-model:value="form.link"
          :placeholder="form.linkType === 'page' ? '如: /pages/practice/index' : '请输入URL'"
        />
      </FormItem>

      <FormItem label="排序">
        <InputNumber v-model:value="form.sortOrder" :min="0" style="width: 100%" />
      </FormItem>

      <FormItem label="状态">
        <Switch
          v-model:checked="form.status"
          checked-children="启用"
          un-checked-children="禁用"
          :checked-value="'enabled'"
          :un-checked-value="'disabled'"
        />
      </FormItem>
    </Form>
  </Modal>
</template>
