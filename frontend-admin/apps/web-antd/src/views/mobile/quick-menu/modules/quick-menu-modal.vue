<script lang="ts" setup>
import { ref, computed } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { message } from 'ant-design-vue';
import { Form, FormItem, Input, InputNumber, Select, SelectOption, Switch } from 'ant-design-vue';

import {
  createQuickMenu,
  updateQuickMenu,
} from '#/api/system/mobile-config';

const emit = defineEmits(['success']);

const modalMode = ref<'create' | 'edit'>('create');
const currentItem = ref<any>(null);

const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    try {
      if (modalMode.value === 'create') {
        await createQuickMenu(form.value);
        message.success('创建成功');
      } else if (currentItem.value) {
        await updateQuickMenu(currentItem.value.id, form.value);
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
      const data = drawerApi.getData();

      if (data) {
        modalMode.value = 'edit';
        currentItem.value = data;
        form.value = { ...data };
      } else {
        modalMode.value = 'create';
        currentItem.value = null;
        form.value = {
          title: '',
          icon: '',
          link: '',
          linkType: 'page',
          sortOrder: 0,
          status: 'enabled',
        };
      }
    }
  },
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

const isEdit = computed(() => modalMode.value === 'edit');
</script>

<template>
  <Drawer :title="isEdit ? '编辑快捷菜单' : '新增快捷菜单'">
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
  </Drawer>
</template>
