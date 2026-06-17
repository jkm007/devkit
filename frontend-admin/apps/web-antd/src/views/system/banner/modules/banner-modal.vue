<script lang="ts" setup>
import type { Banner } from '#/api/system/banner';

import { ref } from 'vue';

import { useVbenDrawer, useVbenForm } from '@vben/common-ui';
import { message, Upload } from 'ant-design-vue';

import { createBanner, updateBanner } from '#/api/system/banner';
import { simpleUpload } from '#/api/file';

const emit = defineEmits<{
  success: [];
}>();

const modalMode = ref<'create' | 'edit'>('create');
const currentBanner = ref<Banner | null>(null);
const imageUrl = ref('');
const uploading = ref(false);

const linkTypeOptions = [
  { label: '内部链接', value: 'page' },
  { label: '外部链接', value: 'external' },
  { label: '无链接', value: 'none' },
];

const statusOptions = [
  { label: '启用', value: 'enabled' },
  { label: '禁用', value: 'disabled' },
];

// 自定义图片上传
async function handleImageUpload(file: File) {
  const isImage = file.type.startsWith('image/');
  if (!isImage) {
    message.error('只能上传图片文件');
    return false;
  }
  const isLt2M = file.size / 1024 / 1024 < 2;
  if (!isLt2M) {
    message.error('图片大小不能超过 2MB');
    return false;
  }

  uploading.value = true;
  try {
    const result = await simpleUpload(file);
    imageUrl.value = result.url;
    formApi.setValues({ image: result.url });
    message.success('图片上传成功');
  } catch (error: any) {
    message.error(error.message || '图片上传失败');
  } finally {
    uploading.value = false;
  }
  return false; // 阻止自动上传
}

function useFormSchema() {
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
      component: 'Input',
      componentProps: {
        placeholder: '请输入或上传图片URL',
      },
      fieldName: 'image',
      label: '图片URL',
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
        if: (values: any) => values.linkType !== 'none',
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
      imageUrl.value = '';

      if (data) {
        modalMode.value = 'edit';
        currentBanner.value = data;
        imageUrl.value = data.image || '';
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
    <div class="banner-form">
      <Form />

      <!-- 图片上传区域 -->
      <div class="image-upload-section">
        <div class="upload-label">快速上传图片</div>
        <div v-if="imageUrl" class="image-preview">
          <img :src="imageUrl" alt="轮播图" class="preview-image" />
          <div class="image-actions">
            <Upload
              :show-upload-list="false"
              :before-upload="handleImageUpload"
              accept="image/*"
            >
              <a-button size="small" :loading="uploading">重新上传</a-button>
            </Upload>
            <a-button
              size="small"
              danger
              @click="
                imageUrl = '';
                formApi.setValues({ image: '' });
              "
            >
              删除
            </a-button>
          </div>
        </div>
        <Upload
          v-else
          :show-upload-list="false"
          :before-upload="handleImageUpload"
          accept="image/*"
        >
          <div class="upload-trigger">
            <div class="upload-icon">+</div>
            <div class="upload-text">
              {{ uploading ? '上传中...' : '点击上传图片' }}
            </div>
          </div>
        </Upload>
      </div>
    </div>
  </Drawer>
</template>

<style scoped>
.banner-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.image-upload-section {
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.upload-label {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
  color: #333;
}

.image-preview {
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #d9d9d9;
}

.preview-image {
  width: 100%;
  height: 200px;
  object-fit: cover;
  display: block;
}

.image-actions {
  display: flex;
  gap: 8px;
  padding: 8px;
  background: #fafafa;
  border-top: 1px solid #f0f0f0;
}

.upload-trigger {
  width: 100%;
  height: 150px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 1px dashed #d9d9d9;
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.3s;
}

.upload-trigger:hover {
  border-color: #1890ff;
}

.upload-icon {
  font-size: 32px;
  color: #999;
  margin-bottom: 8px;
}

.upload-text {
  color: #666;
  font-size: 14px;
}
</style>
