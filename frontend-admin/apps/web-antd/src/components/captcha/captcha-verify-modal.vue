<script setup lang="ts">
/**
 * 验证码验证弹窗（用于 API 拦截器）
 * 独立组件，自行管理弹窗状态和验证码加载
 * 通过全局事件触发，验证成功后通过事件返回 captchaId + captchaCode
 *
 * 事件协议：
 * - 监听 captcha:show-verify → 显示弹窗
 * - 发送 captcha:verify-result → 验证成功 { captchaId, captchaCode }
 * - 发送 captcha:verify-cancel  → 用户取消
 */
import { onMounted, onUnmounted, ref } from 'vue';

import { Modal } from 'ant-design-vue';

import BackendCaptcha from './backend-captcha.vue';
import { getCaptcha, getPublicSettings } from '#/api/system/settings';

// 弹窗状态
const visible = ref(false);
const loading = ref(false);

// 验证码数据
const captchaId = ref('');
const captchaImage = ref('');
const captchaThumb = ref('');
const captchaThumbY = ref(0);
const captchaType = ref<'point' | 'puzzle' | 'slider'>('slider');

async function loadCaptcha() {
  loading.value = true;
  try {
    const data = await getCaptcha(captchaType.value);
    if (data && data.captcha_id && data.image) {
      captchaId.value = data.captcha_id;
      captchaImage.value = data.image;
      captchaThumb.value = data.thumb || '';
      captchaThumbY.value = (data as any).thumb_y || 0;
      // 使用后端返回的类型（如果有效）
      if (data.type && ['slider', 'puzzle', 'point'].includes(data.type)) {
        captchaType.value = data.type as 'point' | 'puzzle' | 'slider';
      }
    }
  } catch {
    // ignore
  } finally {
    loading.value = false;
  }
}

async function handleShow() {
  visible.value = true;
  // 从公开配置中获取验证码类型
  try {
    const settings = await getPublicSettings();
    const type = settings?.captcha?.captcha_type;
    if (type && ['slider', 'puzzle', 'point'].includes(String(type).replace(/"/g, ''))) {
      captchaType.value = String(type).replace(/"/g, '') as 'point' | 'puzzle' | 'slider';
    }
  } catch {
    // 使用默认类型
  }
  loadCaptcha();
}

function handleSuccess(result: { captchaCode: string; captchaId: string; startTime: number }) {
  visible.value = false;
  window.dispatchEvent(
    new CustomEvent('captcha:verify-result', {
      detail: { captchaId: result.captchaId, captchaCode: result.captchaCode, startTime: result.startTime },
    }),
  );
}

function handleCancel() {
  visible.value = false;
  window.dispatchEvent(new CustomEvent('captcha:verify-cancel'));
}

function handleRefresh() {
  loadCaptcha();
}

onMounted(() => {
  window.addEventListener('captcha:show-verify', handleShow as EventListener);
});

onUnmounted(() => {
  window.removeEventListener('captcha:show-verify', handleShow as EventListener);
});
</script>

<template>
  <Modal
    v-model:open="visible"
    title="安全验证"
    :centered="true"
    :mask-closable="false"
    :keyboard="false"
    :width="380"
    :footer="null"
    @cancel="handleCancel"
  >
    <div style="min-height: 200px; display: flex; align-items: center; justify-content: center;">
      <BackendCaptcha
        v-if="!loading && captchaImage"
        :captcha-type="captchaType"
        :server-image="captchaImage"
        :server-thumb="captchaThumb"
        :server-thumb-y="captchaThumbY"
        :server-captcha-id="captchaId"
        @success="handleSuccess"
        @refresh="handleRefresh"
      />
      <div v-else-if="loading" style="color: #999;">加载中...</div>
      <div v-else style="color: #999;">验证码加载失败，点击
        <a @click="loadCaptcha">重试</a>
      </div>
    </div>
  </Modal>
</template>
