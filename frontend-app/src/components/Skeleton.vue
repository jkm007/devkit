<template>
  <view class="skeleton">
    <view v-for="(item, i) in items" :key="i" class="skeleton-item" :class="item.class">
      <!-- 头像圆形骨架 -->
      <view v-if="item.type === 'avatar'" class="sk-avatar" />
      <!-- 文本行骨架 -->
      <view v-else-if="item.type === 'text'" class="sk-text" :style="{ width: item.width || '100%' }" />
      <!-- 标题骨架 -->
      <view v-else-if="item.type === 'title'" class="sk-title" :style="{ width: item.width || '60%' }" />
      <!-- 图片矩形骨架 -->
      <view v-else-if="item.type === 'image'" class="sk-image" :style="{ width: item.width || '100%', height: item.height || '160px' }" />
      <!-- 卡片容器 -->
      <view v-else-if="item.type === 'card'" class="sk-card">
        <view class="sk-title" :style="{ width: item.width || '50%' }" />
        <view class="sk-text" style="width: 90%" />
        <view class="sk-text" style="width: 70%" />
      </view>
      <!-- 列表项 -->
      <view v-else-if="item.type === 'list-item'" class="sk-list-item">
        <view class="sk-avatar" />
        <view class="sk-list-content">
          <view class="sk-text" style="width: 80%" />
          <view class="sk-text" style="width: 50%" />
        </view>
      </view>
      <!-- 按钮骨架 -->
      <view v-else-if="item.type === 'button'" class="sk-button" :style="{ width: item.width || '120px' }" />
    </view>
  </view>
</template>

<script setup lang="ts">
interface SkeletonItem {
  type: 'avatar' | 'text' | 'title' | 'image' | 'card' | 'list-item' | 'button';
  width?: string;
  height?: string;
  class?: string;
}

defineProps<{
  count?: number;      // 重复次数
  items?: SkeletonItem[]; // 自定义骨架项
}>();

const defaultItems: SkeletonItem[] = [
  { type: 'avatar' },
  { type: 'title', width: '60%' },
  { type: 'text', width: '100%' },
  { type: 'text', width: '80%' },
];
</script>

<style lang="scss" scoped>
.skeleton { padding: 16px; }
.skeleton-item { margin-bottom: 12px; }

/* 闪烁动画 */
@keyframes shimmer {
  0% { background-position: -200px 0; }
  100% { background-position: 200px 0; }
}

.sk-avatar {
  width: 40px; height: 40px; border-radius: 50%;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 400px 100%;
  animation: shimmer 1.5s infinite;
}

.sk-text {
  height: 14px; border-radius: 4px; margin-bottom: 8px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 400px 100%;
  animation: shimmer 1.5s infinite;
}

.sk-title {
  height: 20px; border-radius: 4px; margin-bottom: 12px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 400px 100%;
  animation: shimmer 1.5s infinite;
}

.sk-image {
  border-radius: 8px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 400px 100%;
  animation: shimmer 1.5s infinite;
}

.sk-card {
  padding: 16px; background: #fff; border-radius: 8px;
  .sk-title { margin-bottom: 8px; }
  .sk-text { margin-bottom: 6px; }
}

.sk-list-item {
  display: flex; align-items: center; gap: 12px; padding: 12px; background: #fff; border-radius: 8px;
}
.sk-list-content { flex: 1; }

.sk-button {
  height: 44px; border-radius: 22px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 400px 100%;
  animation: shimmer 1.5s infinite;
}
</style>
