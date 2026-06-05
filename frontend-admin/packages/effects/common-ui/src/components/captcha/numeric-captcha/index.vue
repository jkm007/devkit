<script setup lang="ts">
import type { NumericCaptchaProps } from '../types';

import { computed, onMounted, ref, useTemplateRef, watch } from 'vue';

import { $t } from '@vben/locales';

import { RotateCw } from '@vben/icons';

const props = withDefaults(defineProps<NumericCaptchaProps>(), {
  bgColor: '#f5f5f5',
  charLength: 4,
  fontSize: 32,
  height: 40,
  length: 4,
  serverCaptchaId: '',
  serverImage: '',
  textColor: '#333',
  width: 120,
});

const emit = defineEmits<{
  change: [code: string];
  success: [data: { captchaId: string; code: string }];
}>();

// 接收外部传入的 onSuccess 回调（通过 componentProps）
const propsAny = props as any;
const onSuccessCallback = propsAny.onSuccess as
  | ((data: { captchaId: string; code: string }) => void)
  | undefined;

const modelValue = defineModel<boolean>({ default: false });

const canvasRef = useTemplateRef<HTMLCanvasElement>('canvasRef');
const inputCode = ref('');
const currentCode = ref('');
const isVerified = ref(false);

const chars = '0123456789';

const charLength = computed(() => props.charLength ?? props.length);

/** 是否使用后端验证码 */
const isServerMode = computed(
  () => !!props.serverImage && !!props.serverCaptchaId,
);

/** 生成随机验证码（前端模式） */
function generateCode(): string {
  let code = '';
  for (let i = 0; i < charLength.value; i++) {
    code += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return code;
}

/** 随机颜色 */
function randomColor(min: number, max: number): string {
  const r = Math.floor(Math.random() * (max - min) + min);
  const g = Math.floor(Math.random() * (max - min) + min);
  const b = Math.floor(Math.random() * (max - min) + min);
  return `rgb(${r},${g},${b})`;
}

/** 绘制验证码（前端模式） */
function drawCaptcha() {
  const canvas = canvasRef.value;
  if (!canvas) return;

  const ctx = canvas.getContext('2d');
  if (!ctx) return;

  const { bgColor, fontSize, height, textColor, width } = props;

  canvas.width = width;
  canvas.height = height;

  // 背景
  ctx.fillStyle = bgColor;
  ctx.fillRect(0, 0, width, height);

  // 绘制干扰线
  for (let i = 0; i < 4; i++) {
    ctx.strokeStyle = randomColor(100, 200);
    ctx.beginPath();
    ctx.moveTo(Math.random() * width, Math.random() * height);
    ctx.lineTo(Math.random() * width, Math.random() * height);
    ctx.stroke();
  }

  // 绘制干扰点
  for (let i = 0; i < 20; i++) {
    ctx.fillStyle = randomColor(100, 200);
    ctx.beginPath();
    ctx.arc(
      Math.random() * width,
      Math.random() * height,
      1,
      0,
      2 * Math.PI,
    );
    ctx.fill();
  }

  // 绘制验证码文字
  const code = generateCode();
  currentCode.value = code;

  ctx.font = `${fontSize}px serif`;
  ctx.textBaseline = 'middle';

  const charWidth = width / (charLength.value + 1);
  for (let i = 0; i < charLength.value; i++) {
    ctx.fillStyle = textColor;
    const deg = (Math.random() * 30 * Math.PI) / 180;
    const x = charWidth * (i + 0.5);
    const y = height / 2;
    ctx.save();
    ctx.translate(x, y);
    ctx.rotate(deg);
    ctx.fillText(code[i], -fontSize / 4, 0);
    ctx.restore();
  }
}

/** 绘制后端图片到 canvas */
function drawServerImage() {
  const canvas = canvasRef.value;
  if (!canvas || !props.serverImage) return;

  const ctx = canvas.getContext('2d');
  if (!ctx) return;

  const img = new Image();
  img.onload = () => {
    canvas.width = img.width;
    canvas.height = img.height;
    ctx.drawImage(img, 0, 0);
  };
  img.src = props.serverImage;
}

/** 刷新验证码 */
function refresh() {
  inputCode.value = '';
  isVerified.value = false;
  modelValue.value = false;
  if (isServerMode.value) {
    // 后端模式：触发父组件重新获取验证码
    emit('change', '');
  } else {
    drawCaptcha();
  }
}

/** 验证 */
function verify() {
  if (isServerMode.value) {
    // 后端模式：返回用户输入，由后端验证
    if (inputCode.value.length === charLength.value) {
      isVerified.value = true;
      modelValue.value = true;
      const data = {
        captchaId: props.serverCaptchaId,
        code: inputCode.value,
      };
      emit('success', data);
      if (onSuccessCallback) {
        onSuccessCallback(data);
      }
      return true;
    }
    return false;
  }

  // 前端模式：本地比对
  if (inputCode.value.toLowerCase() === currentCode.value.toLowerCase()) {
    isVerified.value = true;
    modelValue.value = true;
    const data = { captchaId: 'local', code: currentCode.value };
    emit('success', data);
    if (onSuccessCallback) {
      onSuccessCallback(data);
    }
    return true;
  }
  refresh();
  return false;
}

watch(inputCode, (val) => {
  emit('change', val);
  if (val.length === charLength.value) {
    verify();
  }
});

watch(
  () => props.serverImage,
  () => {
    if (isServerMode.value) {
      inputCode.value = '';
      isVerified.value = false;
      modelValue.value = false;
      drawServerImage();
    }
  },
);

onMounted(() => {
  if (isServerMode.value) {
    drawServerImage();
  } else {
    drawCaptcha();
  }
});

defineExpose({ refresh, verify });
</script>

<template>
  <div class="flex items-center gap-2">
    <canvas
      ref="canvasRef"
      :height="height"
      :width="width"
      class="cursor-pointer rounded border border-border"
      @click="refresh"
    />
    <div class="relative">
      <input
        v-model="inputCode"
        :maxlength="charLength"
        :placeholder="
          $t('ui.captcha.numericPlaceholder') || '请输入验证码'
        "
        class="h-10 w-28 rounded-md border border-border bg-background px-3 text-sm outline-none focus:border-primary"
        @keyup.enter="verify"
      />
      <button
        class="absolute top-1/2 right-1 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground"
        type="button"
        @click="refresh"
      >
        <RotateCw class="size-3.5" />
      </button>
    </div>
  </div>
</template>
