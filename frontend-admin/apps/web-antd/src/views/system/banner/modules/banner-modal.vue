<template>
  <VbenModal
    v-model:visible="visible"
    :title="mode === 'create' ? '新增轮播图' : '编辑轮播图'"
    width="600px"
    @ok="handleSubmit"
    @cancel="handleCancel"
  >
    <VbenForm
      ref="formRef"
      :model="formData"
      :rules="rules"
      :label-width="100"
    >
      <VbenFormItem label="标题" name="title">
        <VbenInput v-model:value="formData.title" placeholder="请输入轮播图标题" />
      </VbenFormItem>

      <VbenFormItem label="图片" name="image">
        <VbenUpload
          v-model:value="formData.image"
          :max-count="1"
          accept="image/*"
          list-type="picture-card"
          @change="handleImageChange"
        />
      </VbenFormItem>

      <VbenFormItem label="链接类型" name="linkType">
        <VbenSelect v-model:value="formData.linkType" :options="linkTypeOptions" />
      </VbenFormItem>

      <VbenFormItem v-if="formData.linkType !== 'none'" label="链接" name="link">
        <VbenInput v-model:value="formData.link" placeholder="请输入链接地址" />
      </VbenFormItem>

      <VbenFormItem label="排序" name="sortOrder">
        <VbenInputNumber v-model:value="formData.sortOrder" min="0" max="999" />
      </VbenFormItem>

      <VbenFormItem v-if="mode === 'edit'" label="状态" name="status">
        <VbenSelect v-model:value="formData.status" :options="statusOptions" />
      </VbenFormItem>
    </VbenForm>
  </VbenModal>
</template>

<script lang="ts" setup>
import { ref, watch, computed } from 'vue';
import { VbenModal, VbenForm, VbenFormItem, VbenInput, VbenSelect, VbenInputNumber, VbenUpload } from '#/components';
import { createBanner, updateBanner } from '#/api/system/banner';
import type { Banner } from '#/api/system/banner';

const props = defineProps<{
  visible: boolean;
  banner: Banner | null;
  mode: 'create' | 'edit';
}>();

const emit = defineEmits<{
  'update:visible': [value: boolean];
  'success': [];
}>();

const formRef = ref();
const formData = ref({
  title: '',
  image: '',
  link: '',
  linkType: 'none',
  sortOrder: 0,
  status: 'enabled',
});

const rules = {
  title: [{ required: true, message: '请输入标题' }],
  image: [{ required: true, message: '请上传图片' }],
};

const linkTypeOptions = [
  { label: '内部链接', value: 'internal' },
  { label: '外部链接', value: 'external' },
  { label: '无链接', value: 'none' },
];

const statusOptions = [
  { label: '启用', value: 'enabled' },
  { label: '禁用', value: 'disabled' },
];

const visible = computed({
  get: () => props.visible,
  set: (value: boolean) => emit('update:visible', value),
});

watch(
  () => props.banner,
  (banner) => {
    if (banner && props.mode === 'edit') {
      formData.value = {
        title: banner.title,
        image: banner.image,
        link: banner.link || '',
        linkType: banner.linkType || 'none',
        sortOrder: banner.sortOrder || 0,
        status: banner.status || 'enabled',
      };
    } else {
      formData.value = {
        title: '',
        image: '',
        link: '',
        linkType: 'none',
        sortOrder: 0,
        status: 'enabled',
      };
    }
  },
);

function handleImageChange(fileList: any[]) {
  if (fileList.length > 0) {
    formData.value.image = fileList[0].url || fileList[0].response?.data?.url || '';
  }
}

async function handleSubmit() {
  try {
    await formRef.value?.validate();

    if (props.mode === 'create') {
      await createBanner(formData.value);
    } else if (props.banner) {
      await updateBanner(props.banner.id, formData.value);
    }

    emit('success');
  } catch (error: any) {
    console.error('操作失败:', error);
  }
}

function handleCancel() {
  visible.value = false;
}
</script>