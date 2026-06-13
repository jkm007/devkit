<template>
  <view class="lazy-image-wrapper" :style="wrapperStyle" @click="handleClick">
    <!-- 加载中：显示占位或骨架 -->
    <view v-if="!loaded && !error" class="lazy-placeholder">
      <view v-if="skeleton" class="skeleton-shimmer" />
      <text v-else class="placeholder-text">{{ placeholderText }}</text>
    </view>

    <!-- 真实图片 -->
    <image
      v-show="loaded && !error"
      :src="src"
      :mode="mode"
      :class="['lazy-img', customClass]"
      :style="imageStyle"
      :lazy-load="true"
      @load="onLoad"
      @error="onError"
    />

    <!-- 加载失败：显示错误占位 -->
    <view v-if="error" class="lazy-error">
      <text class="error-icon">🖼️</text>
      <text class="error-text">{{ errorText }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const props = defineProps<{
  src: string;
  mode?: 'scaleToFill' | 'aspectFit' | 'aspectFill' | 'widthFix' | 'heightFix' | 'top' | 'bottom' | 'center';
  width?: string;
  height?: string;
  radius?: string;
  skeleton?: boolean;
  placeholderText?: string;
  errorText?: string;
  customClass?: string;
}>();

const emit = defineEmits<{
  (e: 'load', event: any): void;
  (e: 'error', event: any): void;
  (e: 'click'): void;
}>();

const loaded = ref(false);
const error = ref(false);

const wrapperStyle = computed(() => ({
  width: props.width || '100%',
  height: props.height || 'auto',
  borderRadius: props.radius || '0',
  overflow: 'hidden' as const,
  position: 'relative' as const,
}));

const imageStyle = computed(() => ({
  width: '100%',
  height: '100%',
  borderRadius: props.radius || '0',
}));

function onLoad(e: any) {
  loaded.value = true;
  emit('load', e);
}

function onError(e: any) {
  error.value = true;
  emit('error', e);
}

function handleClick() {
  emit('click');
}
</script>

<style lang="scss" scoped>
.lazy-image-wrapper { position: relative; display: inline-block; }

.lazy-placeholder {
  position: absolute; inset: 0; display: flex; align-items: center; justify-content: center;
  background: #f5f5f5;
}

.skeleton-shimmer {
  position: absolute; inset: 0;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 400px 100%;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { background-position: -200px 0; }
  100% { background-position: 200px 0; }
}

.placeholder-text { font-size: 12px; color: #ccc; }

.lazy-img { display: block; }

.lazy-error {
  position: absolute; inset: 0; display: flex; flex-direction: column;
  align-items: center; justify-content: center; background: #f9f9f9;
}
.error-icon { font-size: 24px; margin-bottom: 4px; }
.error-text { font-size: 12px; color: #999; }
</style>
