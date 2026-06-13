<template>
  <view class="content-block-renderer">
    <!-- 纯文本 -->
    <view v-if="block.type === 'text'" class="text-block">
      <text class="text-content">{{ block.content }}</text>
    </view>

    <!-- 图片 -->
    <view v-else-if="block.type === 'image'" class="image-block" @click="previewImage">
      <image
        :src="getImageUrl()"
        mode="widthFix"
        class="block-image"
        :lazy-load="true"
        @error="onImageError"
      />
      <view v-if="imageError" class="image-error">图片加载失败</view>
    </view>

    <!-- 视频 -->
    <view v-else-if="block.type === 'video'" class="video-block">
      <video
        :src="getMediaUrl()"
        :poster="videoPoster"
        controls
        class="block-video"
        @error="onVideoError"
      />
      <view v-if="videoError" class="media-error">视频加载失败</view>
    </view>

    <!-- 音频 -->
    <view v-else-if="block.type === 'audio'" class="audio-block">
      <view class="audio-player" @click="toggleAudio">
        <text class="audio-icon">{{ isPlaying ? '⏸️' : '▶️' }}</text>
        <text class="audio-name">{{ audioName }}</text>
        <text class="audio-duration">{{ audioDuration }}</text>
      </view>
      <view v-if="audioError" class="media-error">音频加载失败</view>
    </view>

    <!-- 文档（PDF/Word 等） -->
    <view v-else-if="block.type === 'document'" class="document-block">
      <view class="doc-preview" @click="openDocument">
        <text class="doc-icon">📄</text>
        <text class="doc-name">{{ documentName }}</text>
        <text class="doc-hint">点击查看</text>
      </view>
    </view>

    <!-- 公式（KaTeX） -->
    <view v-else-if="block.type === 'formula'" class="formula-block">
      <text class="formula-content">{{ block.content }}</text>
      <!-- H5 环境可引入 KaTeX 渲染，小程序显示原始文本 -->
    </view>

    <!-- 表格 -->
    <view v-else-if="block.type === 'table'" class="table-block">
      <scroll-view scroll-x class="table-scroll">
        <view class="table-wrapper" v-html="block.content"></view>
      </scroll-view>
    </view>

    <!-- 代码 -->
    <view v-else-if="block.type === 'code'" class="code-block">
      <view class="code-header">
        <text class="code-lang">{{ codeLang }}</text>
        <text class="code-copy" @click="copyCode">复制</text>
      </view>
      <scroll-view scroll-x class="code-scroll">
        <text class="code-content" selectable>{{ block.content }}</text>
      </scroll-view>
    </view>

    <!-- 引用 -->
    <view v-else-if="block.type === 'quote'" class="quote-block">
      <view class="quote-border">
        <text class="quote-content">{{ block.content }}</text>
      </view>
    </view>

    <!-- 分割线 -->
    <view v-else-if="block.type === 'divider'" class="divider-block">
      <view class="divider-line"></view>
    </view>

    <!-- 未知类型 -->
    <view v-else class="unknown-block">
      <text class="unknown-text">不支持的内容类型: {{ block.type }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import type { ContentBlock } from '@/api/types';

const props = defineProps<{
  block: ContentBlock;
}>();

// 图片
const imageError = ref(false);
function getImageUrl(): string {
  return props.block.url || `/api/v1/files/${props.block.fileId}/preview`;
}
function previewImage() {
  if (!imageError.value) {
    uni.previewImage({ urls: [getImageUrl()] });
  }
}
function onImageError() { imageError.value = true; }

// 视频
const videoError = ref(false);
const videoPoster = computed(() => {
  // 如果有视频封面图 URL
  return '';
});
function getMediaUrl(): string {
  return props.block.url || `/api/v1/files/${props.block.fileId}/preview`;
}
function onVideoError() { videoError.value = true; }

// 音频
const isPlaying = ref(false);
const audioError = ref(false);
const audioName = computed(() => {
  return props.block.content?.substring(0, 30) || '音频文件';
});
const audioDuration = ref('00:00');
// #ifdef APP-PLUS
let audioContext: any = null;
// #endif
function toggleAudio() {
  // #ifdef APP-PLUS
  if (!audioContext) {
    audioContext = uni.createInnerAudioContext();
    audioContext.src = getMediaUrl();
    audioContext.onTimeUpdate(() => {
      const current = audioContext.currentTime;
      const duration = audioContext.duration;
      audioDuration.value = formatTime(current) + ' / ' + formatTime(duration);
    });
    audioContext.onError(() => { audioError.value = true; });
  }
  if (isPlaying.value) { audioContext.pause(); }
  else { audioContext.play(); }
  isPlaying.value = !isPlaying.value;
  // #endif
  // #ifdef H5
  // H5 可使用 <audio> 标签，此处简化处理
  uni.showToast({ title: 'H5 环境请使用原生音频', icon: 'none' });
  // #endif
}
function formatTime(s: number): string {
  const m = Math.floor(s / 60);
  const sec = Math.floor(s % 60);
  return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
}

// 文档
const documentName = computed(() => {
  const url = props.block.url || '';
  return url.split('/').pop() || '文档';
});
function openDocument() {
  const url = props.block.url || `/api/v1/files/${props.block.fileId}/preview`;
  // #ifdef H5
  window.open(url, '_blank');
  // #endif
  // #ifdef MP-WEIXIN
  uni.downloadFile({
    url,
    success: (res) => {
      uni.openDocument({ filePath: res.tempFilePath, showMenu: true });
    },
  });
  // #endif
  // #ifdef APP-PLUS
  uni.navigateTo({ url: `/pages/webview/index?url=${encodeURIComponent(url)}&title=${encodeURIComponent(documentName.value)}` });
  // #endif
}

// 代码
const codeLang = computed(() => {
  // 从 content 中提取语言标识（如果有的话）
  const match = props.block.content?.match(/^```(\w+)/);
  return match ? match[1] : '';
});
function copyCode() {
  const content = props.block.content?.replace(/^```\w*\n?/, '').replace(/```$/, '') || '';
  uni.setClipboardData({ data: content, success: () => uni.showToast({ title: '已复制', icon: 'success' }) });
}
</script>

<style lang="scss" scoped>
.content-block-renderer { margin-bottom: 12px; }

.text-block { .text-content { font-size: 15px; line-height: 1.6; color: #333; white-space: pre-wrap; } }

.image-block { position: relative; border-radius: 8px; overflow: hidden; margin: 8px 0; }
.block-image { width: 100%; display: block; }
.image-error { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; background: #f5f5f5; color: #999; font-size: 13px; }

.video-block { margin: 8px 0; border-radius: 8px; overflow: hidden; }
.block-video { width: 100%; aspect-ratio: 16 / 9; background: #000; }

.audio-block { margin: 8px 0; }
.audio-player { display: flex; align-items: center; padding: 12px 16px; background: #f9f9f9; border-radius: 8px; gap: 12px; }
.audio-icon { font-size: 20px; }
.audio-name { flex: 1; font-size: 14px; color: #333; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.audio-duration { font-size: 12px; color: #999; font-family: monospace; }

.document-block { margin: 8px 0; }
.doc-preview { display: flex; flex-direction: column; align-items: center; padding: 20px; background: #f9f9f9; border-radius: 8px; border: 1px dashed #ddd; }
.doc-icon { font-size: 32px; margin-bottom: 8px; }
.doc-name { font-size: 14px; color: #333; margin-bottom: 4px; }
.doc-hint { font-size: 12px; color: #1890ff; }

.formula-block { padding: 12px; background: #f9f9f9; border-radius: 8px; margin: 8px 0; }
.formula-content { font-family: 'KaTeX_Math', serif; font-size: 15px; color: #333; }

.table-block { margin: 8px 0; border-radius: 8px; overflow: hidden; }
.table-scroll { width: 100%; }
.table-wrapper { min-width: max-content; }
.table-wrapper :deep(table) { border-collapse: collapse; width: 100%; }
.table-wrapper :deep(th), .table-wrapper :deep(td) { border: 1px solid #e8e8e8; padding: 8px 12px; text-align: left; }
.table-wrapper :deep(th) { background: #fafafa; font-weight: 500; }

.code-block { margin: 8px 0; border-radius: 8px; overflow: hidden; border: 1px solid #e8e8e8; }
.code-header { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: #f5f5f5; }
.code-lang { font-size: 12px; color: #666; }
.code-copy { font-size: 12px; color: #1890ff; }
.code-scroll { background: #1e1e1e; padding: 12px; max-height: 300px; }
.code-content { font-family: 'Fira Code', 'Consolas', monospace; font-size: 13px; color: #d4d4d4; white-space: pre; line-height: 1.5; }

.quote-block { margin: 8px 0; }
.quote-border { border-left: 4px solid #1890ff; padding: 8px 16px; background: #f9f9f9; border-radius: 0 8px 8px 0; }
.quote-content { font-size: 14px; color: #666; line-height: 1.6; font-style: italic; }

.divider-block { padding: 16px 0; }
.divider-line { height: 1px; background: #e8e8e8; }

.unknown-block { padding: 12px; background: #fff3e0; border-radius: 8px; }
.unknown-text { font-size: 12px; color: #f57c00; }
.media-error { padding: 12px; background: #fff3e0; border-radius: 8px; text-align: center; color: #f57c00; font-size: 13px; margin-top: 4px; }
</style>
