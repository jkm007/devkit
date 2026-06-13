<script lang="ts" setup>
import { ref, watch } from 'vue';

import { Input, InputNumber, message, Modal, Radio, RadioGroup } from 'ant-design-vue';

const visible = ref(false);
const modalTitle = ref('新增');
const parentCode = ref('');
const nodeType = ref<'examCategory' | 'exam' | 'subject' | 'category'>('examCategory');

const formData = ref({
  name: '',
  code: '',
  sortOrder: 0,
  status: 1,
});

const emit = defineEmits(['success']);

// Generate code based on name
watch(() => formData.value.name, (name) => {
  if (!name) return;

  // Auto-generate code from name (simple pinyin or abbreviation)
  const typeLabels = {
    examCategory: 'CAT',
    exam: 'EX',
    subject: 'SU',
    category: 'CA',
  };

  const prefix = parentCode.value || typeLabels[nodeType.value];
  const suffix = name.replace(/\s+/g, '').substring(0, 10);
  formData.value.code = `${prefix}-${suffix}`;
});

function open(params?: { parentCode?: string; nodeType?: 'examCategory' | 'exam' | 'subject' | 'category' }) {
  parentCode.value = params?.parentCode || '';
  nodeType.value = params?.nodeType || 'examCategory';

  const typeLabels = {
    examCategory: '考试大类',
    exam: '具体考试',
    subject: '科目模块',
    category: '章节分类',
  };

  modalTitle.value = `新增${typeLabels[nodeType.value]}`;

  formData.value = {
    name: '',
    code: '',
    sortOrder: 0,
    status: 1,
  };
  visible.value = true;
}

function handleOk() {
  if (!formData.value.name) {
    message.warning('请输入名称');
    return;
  }
  emit('success', formData.value);
  visible.value = false;
}

function handleCancel() {
  visible.value = false;
}

defineExpose({ open });
</script>

<template>
  <Modal
    v-model:open="visible"
    :title="modalTitle"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <div class="space-y-4">
      <div>
        <label class="block text-sm font-medium mb-2">
          名称 <span class="text-red-500">*</span>
        </label>
        <Input
          v-model:value="formData.name"
          placeholder="请输入名称（编码将自动生成）"
        />
      </div>

      <div>
        <label class="block text-sm font-medium mb-2">
          编码
          <span class="text-xs text-gray-500 ml-2">（自动生成，可修改）</span>
        </label>
        <Input
          v-model:value="formData.code"
          placeholder="编码"
        />
      </div>

      <div>
        <label class="block text-sm font-medium mb-2">排序</label>
        <InputNumber
          v-model:value="formData.sortOrder"
          :min="0"
          class="w-full"
        />
      </div>

      <div>
        <label class="block text-sm font-medium mb-2">状态</label>
        <RadioGroup
          v-model:value="formData.status"
          button-style="solid"
        >
          <Radio.Button :value="1">启用</Radio.Button>
          <Radio.Button :value="0">禁用</Radio.Button>
        </RadioGroup>
      </div>
    </div>
  </Modal>
</template>