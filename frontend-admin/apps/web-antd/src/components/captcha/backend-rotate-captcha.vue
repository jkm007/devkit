<script setup lang="ts">
/**
 * 后端旋转验证码组件
 * go-captcha rotation 模型：
 * - Master: 已旋转的主图（220x220）
 * - Thumb: 缩略图（80-100），用户旋转它来对齐主图
 * 用户调整角度后点击确认按钮提交验证
 */
import { computed, ref, watch } from 'vue';

import { $t } from '@vben/locales';

import { SliderCaptcha } from '@vben/common-ui';

interface Props {
  /** 后端验证码主图片（已旋转） */
  serverImage: string;
  /** 后端验证码缩略图（用户可旋转） */
  serverThumb?: string;
  /** 验证码 ID */
  serverCaptchaId: string;
  /** 图片尺寸 */
  imageSize?: number;
  /** 缩略图尺寸 */
  thumbSize?: number;
  /** 刷新回调 */
  onRefresh?: () => void;
}

const props = withDefaults(defineProps<Props>(), {
  serverThumb: '',
  imageSize: 220,
  thumbSize: 100,
});

const emit = defineEmits<{
  success: [data: { captchaCode: string; captchaId: string }];
  refresh: [];
}>();

const modelValue = defineModel<boolean>({ default: false });

// 当前旋转角度（用户通过滑块调整缩略图）
const currentAngle = ref(0);
const isVerified = ref(false);
const startTime = ref(0);
const sliderMoveX = ref(0);

// 滑块轨道宽度（补偿 action button）
const SLIDER_TRACK_WIDTH = props.imageSize + 46;

// 将滑块移动距离映射到角度
// moveX 范围: 0 ~ imageSize，映射到 0° ~ 360°
const mappedAngle = computed(() => {
  const ratio = sliderMoveX.value / props.imageSize;
  // ratio 0 = 0°, ratio 1 = 360°
  return Math.round(ratio * 360);
});

// 缩略图旋转样式
const thumbStyle = computed(() => ({
  transform: `rotateZ(${mappedAngle.value}deg)`,
}));

// 处理滑块开始拖动
function handleSliderStart() {
  startTime.value = Date.now();
}

// 处理滑块移动
function handleSliderMove(data: { moveX: number }) {
  sliderMoveX.value = data.moveX;
  currentAngle.value = mappedAngle.value;
}

// 处理滑块释放 - 自动验证
function handleSliderEnd() {
  if (isVerified.value) return;
  if (currentAngle.value < 10) {
    return; // 角度太小不提交
  }
  // 自动提交验证
  isVerified.value = true;
  modelValue.value = true;

  emit('success', {
    captchaId: props.serverCaptchaId,
    captchaCode: JSON.stringify({ angle: currentAngle.value }),
  });
}

// 刷新验证码
function handleRefresh() {
  currentAngle.value = 0;
  sliderMoveX.value = 0;
  isVerified.value = false;
  modelValue.value = false;
  startTime.value = 0;
  emit('refresh');
  props.onRefresh?.();
}

// 监听图片变化时重置状态
watch(
  () => props.serverImage,
  () => {
    currentAngle.value = 0;
    sliderMoveX.value = 0;
    isVerified.value = false;
    modelValue.value = false;
    startTime.value = 0;
  },
);

defineExpose({ refresh: handleRefresh });
</script>

<template>
  <div class="backend-rotate-captcha">
    <!-- 主图片区域 -->
    <div
      class="relative cursor-pointer overflow-hidden rounded-full border border-border shadow-md"
      :style="{ width: `${imageSize}px`, height: `${imageSize}px` }"
      @click="handleRefresh"
    >
      <!-- 主图片（后端已旋转，不需要再旋转） -->
      <img
        :src="serverImage"
        alt="verify"
        class="w-full rounded-full"
      />

      <!-- 缩略图（用户可旋转） -->
      <div
        v-if="serverThumb"
        class="absolute flex items-center justify-center rounded-full"
        :style="{
          width: `${thumbSize}px`,
          height: `${thumbSize}px`,
          top: `${(imageSize - thumbSize) / 2}px`,
          left: `${(imageSize - thumbSize) / 2}px`,
          backgroundColor: 'rgba(255,255,255,0.3)',
        }"
      >
        <img
          :src="serverThumb"
          :style="thumbStyle"
          class="rounded-full shadow-lg"
          alt="thumb"
        />
      </div>

      <!-- 验证成功遮罩 -->
      <div
        v-if="isVerified"
        class="absolute inset-0 flex items-center justify-center rounded-full bg-green-500/30"
      >
        <span class="text-lg font-bold text-white">✓</span>
      </div>

      <!-- 提示文字 -->
      <div
        class="absolute bottom-3 left-0 z-10 block h-7 w-full text-center text-xs leading-[30px] text-white"
      >
        <div class="bg-black/30">
          {{ $t('ui.captcha.sliderRotateDefaultTip') || '拖动滑块旋转图片对齐' }}
        </div>
      </div>
    </div>

    <!-- 当前角度显示 -->
    <div class="mt-2 text-center text-sm text-foreground/80">
      当前旋转: {{ currentAngle }}°
    </div>

    <!-- 滑块条 - 使用 is-slot 模式，释放时自动验证 -->
    <div class="mt-3" :style="{ width: `${SLIDER_TRACK_WIDTH}px` }">
      <SliderCaptcha
        v-model="modelValue"
        is-slot
        :text="$t('ui.captcha.sliderDefaultText')"
        :success-text="$t('ui.captcha.sliderSuccessText')"
        @start="handleSliderStart"
        @move="handleSliderMove"
        @end="handleSliderEnd"
      />
    </div>
  </div>
</template>

<style scoped>
.backend-rotate-captcha {
  display: flex;
  flex-direction: column;
  align-items: center;
}
</style>