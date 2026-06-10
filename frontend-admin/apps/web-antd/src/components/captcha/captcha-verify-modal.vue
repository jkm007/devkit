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

import { Button, Input, Modal } from 'ant-design-vue';

import BackendCaptcha from './backend-captcha.vue';
import BackendRotateCaptcha from './backend-rotate-captcha.vue';
import { getCaptcha, getPublicSettings } from '#/api/system/settings';

// 弹窗状态
const visible = ref(false);
const loading = ref(false);

// 验证码数据
const captchaId = ref('');
const captchaImage = ref('');
const captchaThumb = ref('');
const captchaThumbY = ref(0);
const captchaType = ref<string>('slider');
const captchaHintText = ref('');
const captchaLength = ref(4); // 数字验证码长度

// 数字验证码输入
const numericCode = ref('');

async function loadCaptcha() {
  loading.value = true;
  numericCode.value = '';
  try {
    const data = await getCaptcha(captchaType.value);
    if (data && data.captcha_id && data.image) {
      captchaId.value = data.captcha_id;
      captchaImage.value = data.image;
      captchaThumb.value = data.thumb || '';
      captchaThumbY.value = data.thumb_y || 0;
      captchaHintText.value = data.hint_text || '';
      captchaLength.value = data.length || 4;
      // 使用后端返回的类型（如果有效）
      if (data.type && ['numeric', 'slider', 'puzzle', 'rotation', 'point'].includes(data.type)) {
        captchaType.value = data.type;
      }
    }
  } catch {
    // ignore
  } finally {
    loading.value = false;
  }
}

async function handleShow(event?: Event) {
  visible.value = true;
  // 优先使用后端指定的验证码类型（随机类型）
  const detail = (event as CustomEvent)?.detail;
  const serverType = detail?.captchaType;
  if (serverType && ['numeric', 'slider', 'puzzle', 'rotation', 'point'].includes(serverType)) {
    captchaType.value = serverType as any;
  } else {
    // 没有指定类型时，从公开配置中获取
    try {
      const settings = await getPublicSettings();
      const type = settings?.captcha?.captcha_type;
      if (type && ['numeric', 'slider', 'puzzle', 'rotation', 'point'].includes(String(type).replace(/"/g, ''))) {
        captchaType.value = String(type).replace(/"/g, '');
      }
    } catch {
      // 使用默认类型
    }
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

function handleNumericSubmit() {
  if (!numericCode.value || numericCode.value.length !== captchaLength.value) return;
  visible.value = false;
  window.dispatchEvent(
    new CustomEvent('captcha:verify-result', {
      detail: { captchaId: captchaId.value, captchaCode: numericCode.value, startTime: Date.now() },
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
    <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 12px 0;">
      <!-- 数字验证码 -->
      <div v-if="!loading && captchaImage && captchaType === 'numeric'" style="display: flex; flex-direction: column; align-items: center;">
        <img
          :src="captchaImage"
          alt="captcha"
          style="cursor: pointer; border: 1px solid #eee; border-radius: 4px;"
          @click="loadCaptcha"
        />
        <div style="margin-top: 12px; display: flex; gap: 8px; align-items: center;">
          <Input
            v-model:value="numericCode"
            :maxlength="captchaLength"
            :placeholder="`请输入 ${captchaLength} 位验证码`"
            style="width: 180px;"
            @press-enter="handleNumericSubmit"
          />
          <Button type="primary" :disabled="numericCode.length !== captchaLength" @click="handleNumericSubmit">
            确认
          </Button>
        </div>
      </div>

      <!-- 旋转验证码 -->
      <BackendRotateCaptcha
        v-else-if="!loading && captchaImage && captchaType === 'rotation'"
        :server-image="captchaImage"
        :server-thumb="captchaThumb"
        :server-captcha-id="captchaId"
        @success="handleSuccess"
        @refresh="handleRefresh"
      />

      <!-- 滑块/拼图/点选验证码 -->
      <BackendCaptcha
        v-else-if="!loading && captchaImage"
        :captcha-type="captchaType as 'point' | 'puzzle' | 'slider'"
        :server-image="captchaImage"
        :server-thumb="captchaThumb"
        :server-thumb-y="captchaThumbY"
        :server-captcha-id="captchaId"
        :hint-text="captchaHintText"
        @success="handleSuccess"
        @refresh="handleRefresh"
      />

      <div v-else-if="loading" style="color: #999; min-height: 200px; display: flex; align-items: center;">加载中...</div>
      <div v-else style="color: #999; min-height: 200px; display: flex; align-items: center;">验证码加载失败，点击
        <a @click="loadCaptcha">重试</a>
      </div>
    </div>
  </Modal>
</template>
