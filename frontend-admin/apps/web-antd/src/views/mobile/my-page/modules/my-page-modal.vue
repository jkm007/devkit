<script lang="ts" setup>
import { ref, computed } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';
import { Form, FormItem, Input, InputNumber, Select, SelectOption, Switch } from 'ant-design-vue';

import {
  createMyPageMenu,
  updateMyPageMenu,
} from '#/api/system/mobile-config';

const emit = defineEmits(['success']);

const modalMode = ref<'create' | 'edit'>('create');
const currentItem = ref<any>(null);

const [Modal, modalApi] = useVbenModal({
  async onConfirm() {
    try {
      if (modalMode.value === 'create') {
        await createMyPageMenu(form.value);
        message.success('创建成功');
      } else if (currentItem.value) {
        await updateMyPageMenu(currentItem.value.id, form.value);
        message.success('更新成功');
      }
      emit('success');
      modalApi.close();
    } catch (error: any) {
      message.error(error.message || '操作失败');
    }
  },
});

const form = ref({
  title: '',
  icon: '',
  link: '',
  showBadge: false,
  badgeText: '',
  sortOrder: 0,
  status: 'enabled',
});

const iconOptions = [
  '📝', '📖', '🎯', '📊', '❓', '💡', '🔔', '📢',
  '📚', '🎓', '✏️', '📋', '🔍', '⭐', '🏆', '💪',
  '📱', '💻', '🌐', '📞', '💬', '❤️', '🎁', '🔑',
  '⚙️', '👤', '📦', '🏷️', '🎫', '📌', '🗓️', '💰',
];

const linkOptions = [
  { value: '/pages/wrong-book/index', label: '错题本' },
  { value: '/pages/favorites/index', label: '收藏夹' },
  { value: '/pages/history/index', label: '练习历史' },
  { value: '/pages/settings/index', label: '设置' },
  { value: '/pages/about/index', label: '关于我们' },
  { value: '/pages/feedback/index', label: '意见反馈' },
  { value: '/pages/announcement/index', label: '公告列表' },
  { value: '/pages/profile/index', label: '个人信息' },
];

const isEdit = computed(() => modalMode.value === 'edit');

function openModal(mode: 'create' | 'edit', data: any) {
  modalMode.value = mode;
  currentItem.value = data;

  if (mode === 'edit' && data) {
    form.value = { ...data };
  } else {
    form.value = {
      title: '',
      icon: '',
      link: '',
      showBadge: false,
      badgeText: '',
      sortOrder: 0,
      status: 'enabled',
    };
  }

  modalApi.open();
}

defineExpose({
  open: openModal,
});
</script>

<template>
  <Modal :title="isEdit ? '编辑菜单项' : '新增菜单项'">
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

      <FormItem label="链接页面" required>
        <Select
          v-model:value="form.link"
          placeholder="请选择或输入链接"
          show-search
          :options="linkOptions"
        />
      </FormItem>

      <FormItem label="显示角标">
        <Switch
          v-model:checked="form.showBadge"
          checked-children="是"
          un-checked-children="否"
        />
      </FormItem>

      <FormItem v-if="form.showBadge" label="角标文字">
        <Input
          v-model:value="form.badgeText"
          placeholder="如: NEW, HOT, 留空显示红点"
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
