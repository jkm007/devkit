<script setup lang="ts">
/**
 * 后端验证码包装组件
 * 支持 slider/puzzle/point 三种类型
 * 从后端获取图片，用户交互后返回验证数据
 *
 * 滑块/拼图模式：释放滑块时自动验证
 */
import { computed, ref, useTemplateRef, watch } from 'vue';

import { SliderCaptcha } from '@vben/common-ui';
import { $t } from '@vben/locales';

interface Props {
  /** 验证码类型 */
  captchaType: 'point' | 'puzzle' | 'slider';
  /** 后端验证码图片（带缺口的背景图） */
  serverImage: string;
  /** 后端验证码缩略图（滑块/拼图块） */
  serverThumb?: string;
  /** 缩略图的初始 Y 位置（从后端返回） */
  serverThumbY?: number;
  /** 验证码 ID */
  serverCaptchaId: string;
  /** 点选验证码的点数 */
  pointCount?: number;
  /** 刷新回调 */
  onRefresh?: () => void;
}

const props = withDefaults(defineProps<Props>(), {
  serverThumb: '',
  serverThumbY: 0,
  pointCount: 4,
});

const emit = defineEmits<{
  success: [data: { captchaCode: string; captchaId: string }];
  refresh: [];
}>();

const modelValue = defineModel<boolean>({ default: false });

// 滑块拖动距离（滑块轨道上的像素值）
const sliderMoveX = ref(0);
const isVerified = ref(false);

// 用户开始操作的时间戳
const startTime = ref(0);

// 图片容器 ref
const imageContainerRef = useTemplateRef<HTMLDivElement>('imageContainerRef');

// 点选验证码相关
const clickPoints = ref<Array<{ x: number; y: number }>>([]);
const pointChars = ['①', '②', '③', '④', '⑤', '⑥', '⑦', '⑧'];
const expectedPointCount = ref(4); // 期望点击的点数（从后端获取）

// 后端图片实际尺寸（go-captcha 配置）
// slider/puzzle: 320x200, point/click: 320x220
const IMAGE_WIDTH = 320;
const IMAGE_HEIGHT_SLIDER = 200;
const IMAGE_HEIGHT_POINT = 220;

// 滑块组件的 action button 宽度（约 40px，需要补偿）
const SLIDER_ACTION_WIDTH = 40;
const SLIDER_PADDING = 6;

// 滑块轨道宽度（需要比图片宽，以补偿 action button 的占用）
const SLIDER_TRACK_WIDTH = IMAGE_WIDTH + SLIDER_ACTION_WIDTH + SLIDER_PADDING;

// 滑块/拼图模式
const isSliderMode = computed(
  () => props.captchaType === 'slider' || props.captchaType === 'puzzle',
);

// 点选模式
const isPointMode = computed(() => props.captchaType === 'point');

// 直接映射：滑块拖动距离 = 图片 X 坐标
// 由于滑块轨道宽度已补偿 action button，moveX 可以直接作为 X 坐标
const mappedX = computed(() => {
  return Math.round(sliderMoveX.value);
});

// 滑块缩略图在图片上的位置（直接映射，1:1）
const thumbLeft = computed(() => {
  return Math.round(sliderMoveX.value);
});

// 处理滑块开始拖动
function handleSliderStart() {
  startTime.value = Date.now();
}

// 处理滑块移动
function handleSliderMove(data: { moveX: number }) {
  sliderMoveX.value = data.moveX;
}

// 处理滑块释放 - 自动验证
function handleSliderEnd() {
  if (isVerified.value) return;
  if (mappedX.value < 10) {
    return; // 太小不提交
  }
  // 自动提交验证
  isVerified.value = true;
  modelValue.value = true;

  // 提交坐标：X=用户拖动位置，Y=缩略图初始位置（答案Y）
  emit('success', {
    captchaId: props.serverCaptchaId,
    captchaCode: JSON.stringify({ x: mappedX.value, y: props.serverThumbY }),
  });
}

// 处理图片点击（点选模式）
function handleImageClick(e: MouseEvent) {
  if (isVerified.value) return;

  // 第一次点击时记录开始时间
  if (clickPoints.value.length === 0) {
    startTime.value = Date.now();
  }

  const container = imageContainerRef.value;
  if (!container) return;

  const rect = container.getBoundingClientRect();
  // 计算点击在图片上的实际坐标（基于后端图片尺寸）
  const clickX = e.clientX - rect.left;
  const clickY = e.clientY - rect.top;

  // 按比例映射到后端图片坐标
  const ratioX = IMAGE_WIDTH / rect.width;
  const ratioY = IMAGE_HEIGHT_POINT / rect.height;
  const x = Math.round(clickX * ratioX);
  const y = Math.round(clickY * ratioY);

  clickPoints.value.push({ x, y });

  // 点击数量够后自动验证
  if (clickPoints.value.length >= props.pointCount) {
    isVerified.value = true;
    modelValue.value = true;

    emit('success', {
      captchaId: props.serverCaptchaId,
      captchaCode: JSON.stringify(clickPoints.value),
    });
  }
}

// 撤销最后一个点
function undoLastPoint() {
  if (clickPoints.value.length > 0 && !isVerified.value) {
    clickPoints.value.pop();
  }
}

// 刷新验证码
function handleRefresh() {
  sliderMoveX.value = 0;
  isVerified.value = false;
  modelValue.value = false;
  clickPoints.value = [];
  startTime.value = 0;
  emit('refresh');
  props.onRefresh?.();
}

// 监听图片变化时重置状态
watch(
  () => props.serverImage,
  () => {
    sliderMoveX.value = 0;
    isVerified.value = false;
    modelValue.value = false;
    clickPoints.value = [];
    startTime.value = 0;
  },
);

defineExpose({ refresh: handleRefresh });
</script>

<template>
  <div class="backend-captcha">
    <!-- 滑块/拼图模式 -->
    <div v-if="isSliderMode" class="relative">
      <!-- 后端图片（带缺口） -->
      <div
        ref="imageContainerRef"
        class="relative overflow-hidden rounded border border-border"
        :style="{ width: `${IMAGE_WIDTH}px`, height: `${IMAGE_HEIGHT_SLIDER}px` }"
      >
        <img
          :src="serverImage"
          :alt="$t('ui.captcha.sliderDefaultText')"
          class="block h-full w-full"
          @click="handleRefresh"
        />

        <!-- 滑块缩略图（拼图块）- 跟随拖动，Y 位置由后端返回 -->
        <div
          v-if="serverThumb"
          class="absolute"
          :style="{
            top: `${serverThumbY}px`,
            left: `${thumbLeft}px`,
            transition: isVerified ? 'none' : 'left 0.05s linear',
          }"
        >
          <img
            :src="serverThumb"
            class="block"
            :style="{ maxHeight: `${IMAGE_HEIGHT_SLIDER}px` }"
          />
        </div>

        <!-- 验证成功遮罩 -->
        <div
          v-if="isVerified"
          class="absolute inset-0 flex items-center justify-center bg-green-500/30"
        >
          <span class="text-lg font-bold text-white">✓</span>
        </div>
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

    <!-- 点选模式 -->
    <div v-else-if="isPointMode" class="relative">
      <div
        ref="imageContainerRef"
        class="relative cursor-pointer overflow-hidden rounded border border-border"
        :style="{ width: `${IMAGE_WIDTH}px`, height: `${IMAGE_HEIGHT_POINT}px` }"
        @click="handleImageClick"
      >
        <img
          :src="serverImage"
          :alt="$t('ui.captcha.sliderDefaultText')"
          class="block h-full w-full"
        />

        <!-- 已点击的点标记 -->
        <div
          v-for="(point, index) in clickPoints"
          :key="index"
          class="absolute flex size-6 items-center justify-center rounded-full bg-red-500 text-xs text-white shadow"
          :style="{
            left: `${(point.x / IMAGE_WIDTH) * 100}%`,
            top: `${(point.y / IMAGE_HEIGHT_POINT) * 100}%`,
            transform: 'translate(-50%, -50%)',
          }"
        >
          {{ pointChars[index] || index + 1 }}
        </div>

        <!-- 验证成功遮罩 -->
        <div
          v-if="isVerified"
          class="absolute inset-0 flex items-center justify-center bg-green-500/30"
        >
          <span class="text-lg font-bold text-white">✓</span>
        </div>
      </div>

      <!-- 操作提示 -->
      <div class="mt-2 flex items-center justify-between text-sm">
        <span class="text-foreground/60">
          {{ $t('ui.captcha.clickInOrder') || '请依次点击文字' }}
          ({{ clickPoints.length }}/{{ props.pointCount }})
        </span>
        <button
          v-if="clickPoints.length > 0 && !isVerified"
          class="text-primary hover:underline"
          type="button"
          @click="undoLastPoint"
        >
          {{ $t('common.undo') || '撤销' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.backend-captcha {
  display: flex;
  flex-direction: column;
  align-items: center;
}
</style>