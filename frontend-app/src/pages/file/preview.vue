<template>
  <view class="file-preview-page">
    <view class="header">
      <view class="back-btn" @click="goBack">←</view>
      <text class="title">{{ fileName }}</text>
    </view>

    <!-- 加载中 -->
    <view v-if="loading" class="loading">
      <text>加载中...</text>
    </view>

    <!-- 图片预览 -->
    <view v-else-if="fileType === 'image'" class="image-preview">
      <image :src="fileUrl" mode="aspectFit" class="preview-image" />
    </view>

    <!-- 视频预览 -->
    <view v-else-if="fileType === 'video'" class="video-preview">
      <video :src="fileUrl" controls class="preview-video" />
    </view>

    <!-- 音频预览 -->
    <view v-else-if="fileType === 'audio'" class="audio-preview">
      <view class="audio-info">
        <text class="audio-icon">🎵</text>
        <text class="audio-name">{{ fileName }}</text>
      </view>
      <view class="audio-controls">
        <button class="play-btn" @click="toggleAudio">{{ isPlaying ? '暂停' : '播放' }}</button>
      </view>
    </view>

    <!-- 文档预览 -->
    <view v-else-if="fileType === 'document'" class="document-preview">
      <!-- #ifdef H5 -->
      <iframe :src="fileUrl" class="doc-iframe" />
      <!-- #endif -->
      <!-- #ifdef MP-WEIXIN -->
      <view class="doc-hint">
        <text class="hint-icon">📄</text>
        <text class="hint-text">请在外部应用中查看文档</text>
        <button class="open-btn" @click="openInApp">打开文档</button>
      </view>
      <!-- #endif -->
      <!-- #ifdef APP-PLUS -->
      <web-view :src="fileUrl" />
      <!-- #endif -->
    </view>

    <!-- 不支持的类型 -->
    <view v-else class="unsupported">
      <text class="unsupported-icon">📁</text>
      <text class="unsupported-text">暂不支持预览此文件类型</text>
      <text class="unsupported-hint">{{ fileExtension }}</text>
      <button class="download-btn" @click="downloadFile">下载文件</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { getFileUrl } from '@/api/file';

const fileId = ref(0);
const loading = ref(true);
const fileUrl = ref('');
const fileName = ref('');
const isPlaying = ref(false);

// #ifdef APP-PLUS
let audioContext: any = null;
// #endif

const fileType = computed(() => {
  const ext = fileExtension.value.toLowerCase();
  if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg'].includes(ext)) return 'image';
  if (['mp4', 'webm', 'ogg', 'mov'].includes(ext)) return 'video';
  if (['mp3', 'wav', 'ogg', 'aac', 'flac'].includes(ext)) return 'audio';
  if (['pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'txt'].includes(ext)) return 'document';
  return 'unknown';
});

const fileExtension = computed(() => {
  return fileName.value.split('.').pop() || '';
});

onMounted(() => {
  const pages = getCurrentPages();
  const currentPage = pages[pages.length - 1] as any;
  fileId.value = Number(currentPage.options?.id) || 0;
  fileName.value = decodeURIComponent(currentPage.options?.name || '文件');
  fetchFileUrl();
});

async function fetchFileUrl() {
  loading.value = true;
  try {
    const res = await getFileUrl(fileId.value, 'preview');
    fileUrl.value = res.url;
  } catch {
    // 降级处理：使用默认 URL
    fileUrl.value = `/api/v1/files/${fileId.value}/preview`;
  } finally {
    loading.value = false;
  }
}

function toggleAudio() {
  // #ifdef APP-PLUS
  if (!audioContext) {
    audioContext = uni.createInnerAudioContext();
    audioContext.src = fileUrl.value;
  }
  if (isPlaying.value) {
    audioContext.pause();
  } else {
    audioContext.play();
  }
  isPlaying.value = !isPlaying.value;
  // #endif
}

function openInApp() {
  // #ifdef MP-WEIXIN
  uni.downloadFile({
    url: fileUrl.value,
    success: (res) => {
      uni.openDocument({ filePath: res.tempFilePath, showMenu: true });
    },
  });
  // #endif
}

function downloadFile() {
  uni.downloadFile({
    url: fileUrl.value,
    success: (res) => {
      uni.saveFile({
        tempFilePath: res.tempFilePath,
        success: () => uni.showToast({ title: '保存成功', icon: 'success' }),
      });
    },
  });
}

function goBack() {
  uni.navigateBack();
}

// 组件卸载时清理音频上下文
onUnmounted(() => {
  // #ifdef APP-PLUS
  if (audioContext) {
    audioContext.destroy();
    audioContext = null;
  }
  // #endif
});
</script>

<style lang="scss" scoped>
.file-preview-page { min-height: 100vh; background: #000; }

.header { position: fixed; top: 0; left: 0; right: 0; z-index: 100; display: flex; align-items: center; padding: 12px 16px; background: rgba(0,0,0,0.7); }
.back-btn { width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 18px; }
.title { flex: 1; color: #fff; font-size: 16px; margin-left: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.loading { display: flex; align-items: center; justify-content: center; min-height: 100vh; color: #999; }

.image-preview { display: flex; align-items: center; justify-content: center; min-height: 100vh; padding-top: 56px; }
.preview-image { max-width: 100%; max-height: 100vh; }

.video-preview { min-height: 100vh; padding-top: 56px; }
.preview-video { width: 100%; aspect-ratio: 16 / 9; }

.audio-preview { min-height: 100vh; padding-top: 56px; display: flex; flex-direction: column; align-items: center; justify-content: center; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.audio-info { display: flex; flex-direction: column; align-items: center; margin-bottom: 32px; }
.audio-icon { font-size: 48px; margin-bottom: 12px; }
.audio-name { color: #fff; font-size: 18px; }
.audio-controls { .play-btn { background: rgba(255,255,255,0.2); color: #fff; border: none; border-radius: 24px; padding: 12px 32px; font-size: 16px; } }

.document-preview { min-height: 100vh; padding-top: 56px; }
.doc-iframe { width: 100%; height: calc(100vh - 56px); border: none; }
.doc-hint { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: calc(100vh - 56px); background: #fff; }
.hint-icon { font-size: 48px; margin-bottom: 12px; }
.hint-text { font-size: 14px; color: #666; margin-bottom: 16px; }
.open-btn { background: #1890ff; color: #fff; border: none; border-radius: 20px; padding: 8px 24px; }

.unsupported { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 100vh; padding-top: 56px; background: #fff; }
.unsupported-icon { font-size: 48px; margin-bottom: 12px; }
.unsupported-text { font-size: 16px; color: #333; margin-bottom: 8px; }
.unsupported-hint { font-size: 13px; color: #999; margin-bottom: 20px; }
.download-btn { background: #1890ff; color: #fff; border: none; border-radius: 20px; padding: 10px 24px; }
</style>
