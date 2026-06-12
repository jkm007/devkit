<script setup lang="ts">
/**
 * 验证码弹框组件
 * 支持 slider/puzzle/rotation/point 四种类型
 *
 * 公开模式 (public=true)：用于登录等未认证场景，只获取图片，不验证
 * 私有模式 (public=false)：用于已认证场景（如系统设置测试），自动验证
 */
import { computed, ref, watch } from 'vue';

import { Modal } from 'ant-design-vue';

import BackendCaptcha from './backend-captcha.vue';
import BackendRotateCaptcha from './backend-rotate-captcha.vue';
import { getCaptcha, testCaptcha, verifyCaptcha } from '#/api/system/settings';

interface Props {
  /** 验证码类型: slider/puzzle/rotation/point */
  captchaType?: string;
  /** 弹框标题 */
  title?: string;
  /** 是否显示弹框 */
  visible?: boolean;
  /** 公开模式（无需认证），用于登录等场景 */
  public?: boolean;
  /** 验证结果：true=成功关闭, false=失败刷新, null=无操作 */
  verifyResult?: boolean | null;
  /** 验证失败消息 */
  verifyMessage?: string;
}

const props = withDefaults(defineProps<Props>(), {
  captchaType: 'slider',
  title: '',
  visible: false,
  public: false,
  verifyResult: null,
  verifyMessage: '',
});

const emit = defineEmits<{
  'update:visible': [value: boolean];
  success: [
    data: { captchaCode: string; captchaId: string; startTime?: number },
  ];
  fail: [message: string];
  close: [];
}>();

// ==================== 状态 ====================
const loading = ref(false);
const verifying = ref(false);
const pending = ref(false); // 等待后端验证
const errorMsg = ref(''); // 错误提示
const captchaId = ref('');
const captchaImage = ref('');
const captchaThumb = ref('');
const captchaThumbY = ref(0);
const captchaHintText = ref('');
const captchaChars = ref<string[]>([]);
const captchaPointCount = ref(4); // 点选验证码的点数
const captchaStartTime = ref(0);
const captchaResult = ref<{ valid: boolean; message: string } | null>(null);

// 组件引用
const backendCaptchaRef = ref();
const rotateCaptchaRef = ref();

// ==================== 计算属性 ====================
const modalVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});

const modalTitle = computed(() => {
  const typeLabels: Record<string, string> = {
    slider: '滑块验证',
    puzzle: '拼图验证',
    rotation: '旋转验证',
    point: '点选验证',
  };
  return props.title || typeLabels[props.captchaType] || '安全验证';
});

const isRotationType = computed(() => props.captchaType === 'rotation');
const isBackendType = computed(() =>
  ['slider', 'puzzle', 'point'].includes(props.captchaType),
);

// ==================== 方法 ====================
async function loadCaptcha() {
  if (!props.visible) return;

  loading.value = true;
  captchaResult.value = null;
  resetState();

  try {
    // 公开模式使用 getCaptcha，私有模式使用 testCaptcha
    const data = props.public
      ? await getCaptcha(props.captchaType)
      : await testCaptcha(props.captchaType);

    if (data && data.captcha_id && data.image) {
      captchaId.value = data.captcha_id;
      captchaImage.value = data.image;
      captchaThumb.value = data.thumb || '';
      captchaThumbY.value = data.thumb_y || 0;
      captchaHintText.value = data.hint_text || '';
      captchaChars.value = data.chars || [];
      captchaPointCount.value = data.chars?.length || 4; // 点选数量
      captchaStartTime.value = data.start_time || Date.now();
    } else {
      emit('fail', '获取验证码失败');
      modalVisible.value = false;
    }
  } catch (e: any) {
    emit('fail', `获取验证码失败：${e?.message || '未知错误'}`);
    modalVisible.value = false;
  } finally {
    loading.value = false;
  }
}

function resetState() {
  captchaId.value = '';
  captchaImage.value = '';
  captchaThumb.value = '';
  captchaThumbY.value = 0;
  captchaHintText.value = '';
  captchaChars.value = [];
  captchaStartTime.value = 0;
  captchaResult.value = null;
  if (backendCaptchaRef.value) {
    backendCaptchaRef.value.refresh?.();
  }
  if (rotateCaptchaRef.value) {
    rotateCaptchaRef.value.refresh?.();
  }
}

async function handleCaptchaSuccess(data: {
  captchaCode: string;
  captchaId: string;
  startTime?: number;
}) {
  // 公开模式：返回验证数据，等待父组件确认后再关闭
  // startTime 使用用户完成验证的时间（Date.now()），用于后端检测操作耗时
  if (props.public) {
    pending.value = true;
    errorMsg.value = '';
    emit('success', {
      captchaId: captchaId.value,
      captchaCode: data.captchaCode,
      startTime: Date.now(),
    });
    return;
  }

  // 私有模式：调用验证接口
  verifying.value = true;
  captchaResult.value = null;

  try {
    const result = await verifyCaptcha({
      captchaId: captchaId.value,
      captchaCode: data.captchaCode,
      startTime: Date.now(),
    });
    captchaResult.value = result;

    if (result.valid) {
      setTimeout(() => {
        emit('success', {
          captchaId: captchaId.value,
          captchaCode: data.captchaCode,
          startTime: Date.now(),
        });
        modalVisible.value = false;
      }, 500);
    } else {
      emit('fail', result.message);
      setTimeout(() => {
        loadCaptcha();
      }, 1000);
    }
  } catch (e: any) {
    captchaResult.value = { valid: false, message: e?.message || '验证失败' };
    emit('fail', e?.message || '验证失败');
    setTimeout(() => {
      loadCaptcha();
    }, 1000);
  } finally {
    verifying.value = false;
  }
}

function handleRefresh() {
  loadCaptcha();
}

function handleClose() {
  emit('close');
  modalVisible.value = false;
}

// ==================== 监听 ====================
watch(
  () => props.visible,
  (val) => {
    if (val) {
      loadCaptcha();
    } else {
      resetState();
    }
  },
);

watch(
  () => props.captchaType,
  () => {
    if (props.visible) {
      loadCaptcha();
    }
  },
);

// 监听验证结果
watch(
  () => props.verifyResult,
  (val) => {
    if (val === true) {
      // 验证成功：关闭弹窗
      pending.value = false;
      modalVisible.value = false;
    } else if (val === false) {
      // 验证失败：显示错误，刷新验证码，重置 pending 状态
      pending.value = false;
      errorMsg.value = props.verifyMessage || '验证码错误，请重试';
      loadCaptcha();
    }
  },
);
</script>

<template>
  <Modal
    v-model:open="modalVisible"
    :title="modalTitle"
    :width="400"
    :footer="null"
    :mask-closable="false"
    centered
    @cancel="handleClose"
  >
    <div class="captcha-modal-content">
      <!-- 加载状态 -->
      <div v-if="loading" class="flex items-center justify-center py-8">
        <span class="text-muted-foreground">加载中...</span>
      </div>

      <!-- 旋转验证码 -->
      <div
        v-else-if="isRotationType && captchaImage"
        class="flex flex-col items-center"
      >
        <BackendRotateCaptcha
          ref="rotateCaptchaRef"
          :server-image="captchaImage"
          :server-thumb="captchaThumb"
          :server-captcha-id="captchaId"
          :image-size="220"
          @refresh="handleRefresh"
          @success="handleCaptchaSuccess"
        />
      </div>

      <!-- 滑块/拼图/点选验证码 -->
      <div
        v-else-if="isBackendType && captchaImage"
        class="flex flex-col items-center"
      >
        <!-- 点选提示文字 -->
        <div
          v-if="captchaType === 'point' && captchaHintText"
          class="mb-2 text-center text-sm text-foreground"
        >
          {{ captchaHintText }}
        </div>

        <BackendCaptcha
          ref="backendCaptchaRef"
          :captcha-type="captchaType as 'slider' | 'puzzle' | 'point'"
          :server-image="captchaImage"
          :server-thumb="captchaThumb"
          :server-thumb-y="captchaThumbY"
          :server-captcha-id="captchaId"
          :point-count="captchaPointCount"
          @refresh="handleRefresh"
          @success="handleCaptchaSuccess"
        />

        <!-- 点选字符列表 -->
        <div
          v-if="captchaType === 'point' && captchaChars.length > 0"
          class="mt-2 text-center text-xs text-muted-foreground"
        >
          点击顺序：{{ captchaChars.join(' → ') }}
        </div>
      </div>

      <!-- 错误提示 -->
      <div v-if="errorMsg" class="mt-3 text-center">
        <span class="text-sm font-medium text-red-500">
          {{ errorMsg }}
        </span>
      </div>

      <!-- 验证中提示 -->
      <div v-if="pending" class="mt-3 text-center">
        <span class="text-sm text-blue-500">验证中...</span>
      </div>

      <!-- 验证结果提示（私有模式才显示） -->
      <div v-if="!public && captchaResult" class="mt-3 text-center">
        <span
          :class="captchaResult.valid ? 'text-green-500' : 'text-red-500'"
          class="text-sm font-medium"
        >
          {{ captchaResult.valid ? '验证通过 ✓' : '验证失败 ✗' }}
        </span>
        <span
          v-if="!captchaResult.valid"
          class="ml-2 text-xs text-muted-foreground"
        >
          {{ captchaResult.message }}
        </span>
      </div>
    </div>
  </Modal>
</template>

<style scoped>
.captcha-modal-content {
  min-height: 200px;
}
</style>
