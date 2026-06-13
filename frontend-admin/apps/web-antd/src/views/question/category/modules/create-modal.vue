<script lang="ts" setup>
import { ref } from 'vue';

import { Input, InputNumber, Modal, Radio, RadioGroup } from 'ant-design-vue';

const visible = ref(false);
const formData = ref({
  name: '',
  code: '',
  sortOrder: 0,
  status: 1,
});

const emit = defineEmits(['success']);

function open() {
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
    title="新增"
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
          placeholder="请输入名称"
        />
      </div>

      <div>
        <label class="block text-sm font-medium mb-2">编码</label>
        <Input
          v-model:value="formData.code"
          placeholder="请输入编码"
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