<script lang="ts" setup>
import { onUnmounted, ref } from 'vue';

import { useVbenModal, VCropper } from '@vben/common-ui';

import { message, Upload } from 'ant-design-vue';

import { simpleUpload } from '#/api/file';
import { updateAvatar } from '#/api/account';
import { useUserStore } from '@vben/stores';
import { $t } from '#/locales';

defineOptions({ name: 'AvatarUpload' });

const emit = defineEmits<{
  success: [url: string];
}>();

const userStore = useUserStore();

// 图片相关
const imageUrl = ref('');
const cropperRef = ref<InstanceType<typeof VCropper> | null>(null);

// 创建 Modal
const [Modal, modalApi] = useVbenModal({
  async onConfirm() {
    await handleSave();
  },
  onCancel() {
    handleClose();
  },
});

// 文件选择
function handleBeforeUpload(file: File) {
  // 验证文件类型
  const isImage = file.type.startsWith('image/');
  if (!isImage) {
    message.error($t('file.avatar.uploadTip'));
    return false;
  }

  // 验证文件大小（最大 5MB）
  const isLt5M = file.size / 1024 / 1024 < 5;
  if (!isLt5M) {
    message.error('图片大小不能超过 5MB');
    return false;
  }

  // 创建预览 URL
  imageUrl.value = URL.createObjectURL(file);
  return false; // 阻止自动上传，手动控制
}

// 保存头像
async function handleSave() {
  if (!cropperRef.value || !imageUrl.value) {
    message.error($t('file.avatar.uploadTip'));
    return;
  }

  modalApi.lock();
  let uploadedUrl: string | null = null;
  try {
    // 获取裁剪后的图片 Blob
    const blob = await cropperRef.value.getCropImage('image/jpeg', 0.92, 'blob', 200, 200);
    if (!blob) {
      message.error('裁剪失败');
      return;
    }

    // 创建 File 对象
    const croppedFile = new File([blob], 'avatar.jpg', { type: 'image/jpeg' });

    // 上传文件
    const uploadResult = await simpleUpload(croppedFile);
    uploadedUrl = uploadResult.url;

    // 保存旧头像 URL 用于回滚
    const oldAvatar = userStore.userInfo?.avatar;

    // 更新头像
    await updateAvatar({ avatar: uploadResult.url });

    // 更新用户信息（仅在 updateProfile 成功后）
    if (userStore.userInfo) {
      userStore.userInfo.avatar = uploadResult.url;
    }

    message.success($t('file.avatar.saveSuccess'));
    emit('success', uploadResult.url);
    handleClose();
  } catch (error: any) {
    // 上传成功但更新失败时，回滚 store
    if (uploadedUrl && userStore.userInfo) {
      // 通知用户更新失败，但文件已上传
      message.error('头像更新失败，请重试');
    } else {
      message.error(error?.message || $t('file.avatar.saveError'));
    }
  } finally {
    modalApi.unlock();
  }
}

// 关闭 Modal
function handleClose() {
  modalApi.close();
  if (imageUrl.value && imageUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(imageUrl.value);
  }
  imageUrl.value = '';
}

// 打开 Modal
function open() {
  imageUrl.value = '';
  modalApi.open();
}

// 组件卸载时清理 blob URL
onUnmounted(() => {
  if (imageUrl.value && imageUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(imageUrl.value);
  }
});

defineExpose({ open });
</script>

<template>
  <Modal
    :title="$t('file.avatar.title')"
    :confirm-text="$t('file.avatar.save')"
    :cancel-text="$t('common.cancel')"
    :closable="true"
    :width="600"
  >
    <div class="avatar-upload-container">
      <!-- 上传区域 -->
      <div v-if="!imageUrl" class="upload-area">
        <Upload.Dragger
          :show-upload-list="false"
          :before-upload="handleBeforeUpload"
          accept="image/*"
        >
          <div class="upload-content">
            <div class="i-ant-design:inbox-outlined text-4xl text-primary mb-4" />
            <p class="text-base">{{ $t('file.avatar.dragTip') }}</p>
            <p class="text-sm text-muted-foreground mt-2">{{ $t('file.avatar.formatTip') }}</p>
            <p class="text-sm text-muted-foreground">{{ $t('file.avatar.sizeTip') }}</p>
          </div>
        </Upload.Dragger>
      </div>

      <!-- 裁剪区域 -->
      <div v-else class="crop-area">
        <p class="text-sm text-muted-foreground text-center mb-4">
          {{ $t('file.avatar.cropTip') }}
        </p>
        <VCropper
          ref="cropperRef"
          :img="imageUrl"
          aspect-ratio="1:1"
          :width="400"
          :height="400"
        />
      </div>
    </div>
  </Modal>
</template>

<style scoped>
.avatar-upload-container {
  min-height: 300px;
  padding: 16px;
}

.upload-area {
  display: flex;
  justify-content: center;
  align-items: center;
}

.upload-content {
  padding: 40px;
  text-align: center;
}

.crop-area {
  display: flex;
  flex-direction: column;
  align-items: center;
}
</style>