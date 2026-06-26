<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Descriptions,
  DescriptionsItem,
  Image,
  InputPassword,
  message,
  Modal,
  Select,
  SelectOption,
  Spin,
  Table,
  Tag,
  Tooltip,
} from 'ant-design-vue';
import InputSearch from 'ant-design-vue/es/input/Search';

import { getShareInfo, getShareFolderFiles, verifySharePassword } from '#/api/file';

defineOptions({ name: 'SharePage' });

const route = useRoute();
const loading = ref(true);
const shareInfo = ref<any>(null);
const folderFiles = ref<any[]>([]);
const folderFilesTotal = ref(0);
const error = ref('');

// 密码验证相关
const needPassword = ref(false);
const passwordVerified = ref(false);
const passwordInput = ref('');
const passwordLoading = ref(false);
const passwordError = ref('');

const shareCode = route.params.code as string;

// 服务端筛选参数
const searchText = ref('');
const searchDebounceTimer = ref<ReturnType<typeof setTimeout> | null>(null);
const pagination = ref({ current: 1, pageSize: 100 });
const previewVisible = ref(false);
const previewUrl = ref('');
const previewName = ref('');
const previewContentType = ref('');
const previewType = ref<'audio' | 'image' | 'pdf' | 'video' | ''>('');

// ==================== 播放器状态 ====================
const audioRef = ref<HTMLAudioElement | null>(null);
const videoRef = ref<HTMLVideoElement | null>(null);
const isPlaying = ref(false);
const currentTime = ref(0);
const duration = ref(0);
const volume = ref(1);
const playbackRate = ref(1);
const playMode = ref<'loop' | 'random' | 'sequential'>('sequential');
const currentPlayingFile = ref<any>(null);
const playlist = ref<any[]>([]);
const showPlaylist = ref(false);

// 播放速度选项
const speedOptions = [0.5, 0.75, 1, 1.25, 1.5, 2];

// localStorage 键名
const STORAGE_KEY_PROGRESS = 'share_play_progress';
const STORAGE_KEY_VOLUME = 'share_play_volume';
const STORAGE_KEY_RATE = 'share_play_rate';
const STORAGE_KEY_MODE = 'share_play_mode';

// ==================== 工具函数 ====================

function formatTime(seconds: number): string {
  if (!seconds || isNaN(seconds)) return '00:00';
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
}

function formatFileSize(size: number): string {
  if (!size) return '-';
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`;
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`;
}

function formatDate(date: string): string {
  if (!date) return '永久有效';
  return new Date(date).toLocaleString();
}

function getFileTypeIcon(contentType: string): string {
  if (!contentType) return '📄';
  if (contentType.startsWith('audio/')) return '🎵';
  if (contentType.startsWith('video/')) return '🎬';
  if (contentType.startsWith('image/')) return '🖼️';
  if (contentType.includes('pdf')) return '📕';
  return '📄';
}

function isAudioFile(contentType: string): boolean {
  return contentType?.startsWith('audio/') || false;
}

function isVideoFile(contentType: string): boolean {
  return contentType?.startsWith('video/') || false;
}

function isMediaFile(contentType: string): boolean {
  return isAudioFile(contentType) || isVideoFile(contentType);
}

// 视频格式兼容性状态
const videoFormatSupported = ref(true);
const videoFormatError = ref('');

// 检查视频是否可以播放（通过实际加载检测）
function onVideoError(event: Event) {
  const video = event.target as HTMLVideoElement;
  if (video && video.error) {
    videoFormatSupported.value = false;
    const errorCodes: Record<number, string> = {
      1: '视频加载被中止',
      2: '网络错误，无法加载视频',
      3: '视频解码失败，格式可能不支持',
      4: '视频格式不支持或文件损坏',
    };
    videoFormatError.value = errorCodes[video.error.code] || '视频播放失败';
  }
}

// ==================== 播放进度保存/恢复 ====================

function savePlayProgress(fileId: number, time: number) {
  try {
    const progress = JSON.parse(localStorage.getItem(STORAGE_KEY_PROGRESS) || '{}');
    progress[fileId] = { time, updatedAt: Date.now() };
    // 只保留最近 100 条记录
    const keys = Object.keys(progress);
    if (keys.length > 100) {
      const sorted = keys.sort((a, b) => (progress[a]?.updatedAt || 0) - (progress[b]?.updatedAt || 0));
      if (sorted[0]) {
        delete progress[sorted[0]];
      }
    }
    localStorage.setItem(STORAGE_KEY_PROGRESS, JSON.stringify(progress));
  } catch {
    // ignore
  }
}

function getPlayProgress(fileId?: number): number {
  if (!fileId) return 0;
  try {
    const progress = JSON.parse(localStorage.getItem(STORAGE_KEY_PROGRESS) || '{}');
    return progress[fileId]?.time || 0;
  } catch {
    return 0;
  }
}

function saveVolume(vol: number) {
  localStorage.setItem(STORAGE_KEY_VOLUME, String(vol));
}

function savePlaybackRate(rate: number) {
  localStorage.setItem(STORAGE_KEY_RATE, String(rate));
}

function savePlayMode(mode: string) {
  localStorage.setItem(STORAGE_KEY_MODE, mode);
}

function restoreSettings() {
  try {
    const savedVolume = localStorage.getItem(STORAGE_KEY_VOLUME);
    if (savedVolume) volume.value = Number(savedVolume);

    const savedRate = localStorage.getItem(STORAGE_KEY_RATE);
    if (savedRate) playbackRate.value = Number(savedRate);

    const savedMode = localStorage.getItem(STORAGE_KEY_MODE);
    if (savedMode && ['sequential', 'loop', 'random'].includes(savedMode)) {
      playMode.value = savedMode as any;
    }
  } catch {
    // ignore
  }
}

// ==================== 播放器控制 ====================

function getShareFileUrl(file: any) {
  return `/api/v1/share/${shareCode}/file/${file.fileId}`;
}

function getPreviewType(file: any) {
  const contentType = file.contentType || '';
  const fileName = file.fileName || '';
  if (contentType.startsWith('image/')) return 'image';
  if (contentType.startsWith('video/')) return 'video';
  if (contentType.startsWith('audio/')) return 'audio';
  if (contentType.includes('pdf') || fileName.toLowerCase().endsWith('.pdf')) {
    return 'pdf';
  }
  return '';
}

function playFile(file: any) {
  const type = getPreviewType(file);
  const url = getShareFileUrl(file);

  if (type === 'audio') {
    currentPlayingFile.value = file;
    isPlaying.value = true;

    nextTick(() => {
      if (audioRef.value) {
        audioRef.value.src = url;
        audioRef.value.playbackRate = playbackRate.value;
        audioRef.value.volume = volume.value;

        // 恢复播放进度
        const savedTime = getPlayProgress(file.fileId);
        if (savedTime > 0) {
          audioRef.value.currentTime = savedTime;
        }

        audioRef.value.play().catch(() => {
          // 自动播放被阻止
        });
      }
    });
  } else if (type === 'video') {
    // 视频文件打开弹窗播放
    viewFileInModal(file);
  } else {
    // 非媒体文件，打开预览弹窗
    viewFileInModal(file);
  }
}

function togglePlay() {
  if (isAudioFile(currentPlayingFile.value?.contentType)) {
    if (audioRef.value) {
      if (isPlaying.value) {
        audioRef.value.pause();
      } else {
        audioRef.value.play().catch(() => {});
      }
      isPlaying.value = !isPlaying.value;
    }
  } else if (isVideoFile(currentPlayingFile.value?.contentType)) {
    if (videoRef.value) {
      if (isPlaying.value) {
        videoRef.value.pause();
      } else {
        videoRef.value.play().catch(() => {});
      }
      isPlaying.value = !isPlaying.value;
    }
  }
}

function playPrevious() {
  if (playlist.value.length === 0) return;

  const currentIndex = playlist.value.findIndex(
    (f) => f.fileId === currentPlayingFile.value?.fileId,
  );

  let prevIndex: number;
  if (playMode.value === 'random') {
    prevIndex = Math.floor(Math.random() * playlist.value.length);
  } else {
    prevIndex = currentIndex <= 0 ? playlist.value.length - 1 : currentIndex - 1;
  }

  playFile(playlist.value[prevIndex]);
}

function playNext() {
  if (playlist.value.length === 0) return;

  const currentIndex = playlist.value.findIndex(
    (f) => f.fileId === currentPlayingFile.value?.fileId,
  );

  let nextIndex: number;
  if (playMode.value === 'random') {
    nextIndex = Math.floor(Math.random() * playlist.value.length);
  } else if (playMode.value === 'loop') {
    nextIndex = currentIndex; // 单曲循环
  } else {
    nextIndex = currentIndex >= playlist.value.length - 1 ? 0 : currentIndex + 1;
  }

  playFile(playlist.value[nextIndex]);
}

function seekTo(event: MouseEvent) {
  const target = event.currentTarget as HTMLElement;
  const rect = target.getBoundingClientRect();
  const percent = (event.clientX - rect.left) / rect.width;
  const time = percent * duration.value;

  if (isAudioFile(currentPlayingFile.value?.contentType) && audioRef.value) {
    audioRef.value.currentTime = time;
  } else if (isVideoFile(currentPlayingFile.value?.contentType) && videoRef.value) {
    videoRef.value.currentTime = time;
  }
}

function changeVolume(event: MouseEvent) {
  const target = event.currentTarget as HTMLElement;
  const rect = target.getBoundingClientRect();
  const percent = Math.max(0, Math.min(1, (event.clientX - rect.left) / rect.width));
  volume.value = percent;
  saveVolume(percent);

  if (audioRef.value) audioRef.value.volume = percent;
  if (videoRef.value) videoRef.value.volume = percent;
}

function changeSpeed(speed: any) {
  const rate = Number(speed) || 1;
  playbackRate.value = rate;
  savePlaybackRate(rate);

  if (audioRef.value) audioRef.value.playbackRate = rate;
  if (videoRef.value) videoRef.value.playbackRate = rate;
}

function togglePlayMode() {
  const modes: Array<'loop' | 'random' | 'sequential'> = ['sequential', 'loop', 'random'];
  const currentIndex = modes.indexOf(playMode.value as 'loop' | 'random' | 'sequential');
  const nextIndex = (currentIndex + 1) % modes.length;
  const nextMode = modes[nextIndex];
  if (nextMode) {
    playMode.value = nextMode;
    savePlayMode(nextMode);
  }
}

function toggleMute() {
  if (volume.value > 0) {
    volume.value = 0;
  } else {
    volume.value = 1;
  }
  saveVolume(volume.value);

  if (audioRef.value) audioRef.value.volume = volume.value;
  if (videoRef.value) videoRef.value.volume = volume.value;
}

// ==================== 事件处理 ====================

function onTimeUpdate() {
  const player = isAudioFile(currentPlayingFile.value?.contentType)
    ? audioRef.value
    : videoRef.value;

  if (player) {
    currentTime.value = player.currentTime;
    duration.value = player.duration;

    // 每 5 秒保存一次进度
    if (Math.floor(player.currentTime) % 5 === 0) {
      savePlayProgress(currentPlayingFile.value?.fileId, player.currentTime);
    }
  }
}

function onEnded() {
  if (playMode.value === 'loop') {
    // 单曲循环
    const player = isAudioFile(currentPlayingFile.value?.contentType)
      ? audioRef.value
      : videoRef.value;
    if (player) {
      player.currentTime = 0;
      player.play().catch(() => {});
    }
  } else {
    // 播放下一首
    playNext();
  }
}

function onError() {
  isPlaying.value = false;
  message.error('播放失败，请重试');
}

// ==================== 预览弹窗 ====================

function viewFileInModal(file: any) {
  const type = getPreviewType(file);
  const url = getShareFileUrl(file);
  if (!type) {
    Modal.confirm({
      title: '无法预览',
      content: `文件 "${file.fileName}" 不支持在线预览，是否在新标签页打开？`,
      okText: '打开',
      cancelText: '取消',
      onOk: () => window.open(url, '_blank', 'noopener,noreferrer'),
    });
    return;
  }

  // 重置视频错误状态
  if (type === 'video') {
    videoFormatSupported.value = true;
    videoFormatError.value = '';
  }

  previewName.value = file.fileName;
  previewUrl.value = url;
  previewType.value = type;
  previewContentType.value = file.contentType || '';
  previewVisible.value = true;
}

function openPreviewInNewTab() {
  if (previewUrl.value) {
    window.open(previewUrl.value, '_blank', 'noopener,noreferrer');
  }
}

function downloadFile(file: any) {
  const url = getShareFileUrl(file);
  const link = document.createElement('a');
  link.href = url;
  link.download = file.fileName;
  link.rel = 'noopener noreferrer';
  link.click();
}

// 文件分享的下载按钮
function downloadSharedFile() {
  const url = `/api/v1/share/${shareCode}/file`;
  const link = document.createElement('a');
  link.href = url;
  link.download = shareInfo.value?.fileName || 'download';
  link.rel = 'noopener noreferrer';
  link.click();
}

// 打开视频弹窗
function openVideoModal() {
  // 重置视频错误状态
  videoFormatSupported.value = true;
  videoFormatError.value = '';

  previewName.value = shareInfo.value?.fileName || '视频';
  previewUrl.value = `/api/v1/share/${shareCode}/file`;
  previewType.value = 'video';
  previewContentType.value = shareInfo.value?.contentType || '';
  previewVisible.value = true;
}

// ==================== 键盘快捷键 ====================

function handleKeydown(event: KeyboardEvent) {
  // 如果正在输入，不处理快捷键
  if (
    event.target instanceof HTMLInputElement ||
    event.target instanceof HTMLTextAreaElement
  ) {
    return;
  }

  switch (event.code) {
    case 'Space': {
      event.preventDefault();
      togglePlay();
      break;
    }
    case 'ArrowLeft': {
      event.preventDefault();
      const player = isAudioFile(currentPlayingFile.value?.contentType)
        ? audioRef.value
        : videoRef.value;
      if (player) {
        player.currentTime = Math.max(0, player.currentTime - 5);
      }
      break;
    }
    case 'ArrowRight': {
      event.preventDefault();
      const player2 = isAudioFile(currentPlayingFile.value?.contentType)
        ? audioRef.value
        : videoRef.value;
      if (player2) {
        player2.currentTime = Math.min(player2.duration, player2.currentTime + 5);
      }
      break;
    }
    case 'ArrowUp': {
      event.preventDefault();
      volume.value = Math.min(1, volume.value + 0.1);
      saveVolume(volume.value);
      if (audioRef.value) audioRef.value.volume = volume.value;
      if (videoRef.value) videoRef.value.volume = volume.value;
      break;
    }
    case 'ArrowDown': {
      event.preventDefault();
      volume.value = Math.max(0, volume.value - 0.1);
      saveVolume(volume.value);
      if (audioRef.value) audioRef.value.volume = volume.value;
      if (videoRef.value) videoRef.value.volume = volume.value;
      break;
    }
    case 'KeyM': {
      event.preventDefault();
      toggleMute();
      break;
    }
    case 'KeyN': {
      event.preventDefault();
      playNext();
      break;
    }
    case 'KeyP': {
      event.preventDefault();
      playPrevious();
      break;
    }
  }
}

// ==================== 数据加载 ====================

async function loadShareInfo() {
  try {
    loading.value = true;
    const result = await getShareInfo(shareCode);
    shareInfo.value = result;

    // 检查是否需要密码
    if (result.hasPassword) {
      needPassword.value = true;
      loading.value = false;
      return;
    }

    // 无密码，直接加载内容
    passwordVerified.value = true;
    if (result.type === 'folder') {
      await loadFolderFiles();
    }
  } catch (err: any) {
    error.value = err.message || '分享不存在或已过期';
  } finally {
    loading.value = false;
  }
}

async function handleVerifyPassword() {
  if (!passwordInput.value) {
    passwordError.value = '请输入密码';
    return;
  }
  passwordLoading.value = true;
  passwordError.value = '';
  try {
    await verifySharePassword(shareCode, passwordInput.value);
    passwordVerified.value = true;
    needPassword.value = false;
    message.success('密码验证成功');

    // 验证成功后加载内容
    if (shareInfo.value?.type === 'folder') {
      await loadFolderFiles();
    }
  } catch (err: any) {
    passwordError.value = err.message || '密码错误';
  } finally {
    passwordLoading.value = false;
  }
}

async function loadFolderFiles() {
  try {
    const result = await getShareFolderFiles(shareCode, {
      page: pagination.value.current,
      pageSize: pagination.value.pageSize,
      keyword: searchText.value || undefined,
    });
    folderFiles.value = result?.items || [];
    folderFilesTotal.value = result?.total || 0;

    // 更新播放列表（只包含媒体文件）
    playlist.value = folderFiles.value.filter((f) => isMediaFile(f.contentType));
  } catch {
    folderFiles.value = [];
    folderFilesTotal.value = 0;
  }
}

// 搜索关键词变化时，防抖处理后重新加载
watch(searchText, () => {
  if (searchDebounceTimer.value) {
    clearTimeout(searchDebounceTimer.value);
  }
  searchDebounceTimer.value = setTimeout(() => {
    pagination.value.current = 1;
    loadFolderFiles();
  }, 300);
});

// 计算统计信息
const stats = computed(() => {
  const totalSize = folderFiles.value.reduce((sum, f) => sum + (f.fileSize || 0), 0);
  const audioCount = folderFiles.value.filter((f) => isAudioFile(f.contentType)).length;
  const videoCount = folderFiles.value.filter((f) => isVideoFile(f.contentType)).length;
  const mediaCount = audioCount + videoCount;
  return { totalSize, audioCount, videoCount, mediaCount };
});

// 播放模式图标
const playModeIcon = computed(() => {
  switch (playMode.value) {
    case 'sequential': return '🔁';
    case 'loop': return '🔂';
    case 'random': return '🔀';
    default: return '🔁';
  }
});

// 播放模式提示
const playModeTooltip = computed(() => {
  switch (playMode.value) {
    case 'sequential': return '顺序播放';
    case 'loop': return '单曲循环';
    case 'random': return '随机播放';
    default: return '顺序播放';
  }
});

// 文件列表表格列
const columns = [
  { title: '文件名', dataIndex: 'fileName', key: 'fileName', ellipsis: true },
  { title: '大小', dataIndex: 'fileSize', key: 'fileSize', width: 100 },
  { title: '类型', dataIndex: 'contentType', key: 'contentType', width: 120 },
  { title: '操作', key: 'action', width: 150 },
];

onMounted(() => {
  restoreSettings();
  loadShareInfo();
  document.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
  <Page title="文件分享">
    <div class="max-w-6xl mx-auto p-4">
      <Spin :spinning="loading">
        <!-- 密码验证界面 -->
        <Card v-if="needPassword && !passwordVerified" title="需要访问密码">
          <div class="max-w-sm mx-auto py-6">
            <div class="text-center mb-6">
              <span class="text-5xl mb-3 block">🔒</span>
              <p class="text-gray-500 mt-2">此分享内容需要输入密码才能访问</p>
            </div>
            <div class="space-y-4">
              <InputPassword
                v-model:value="passwordInput"
                placeholder="请输入访问密码"
                size="large"
                @press-enter="handleVerifyPassword"
              />
              <div v-if="passwordError" class="text-red-500 text-sm">{{ passwordError }}</div>
              <Button
                type="primary"
                block
                size="large"
                :loading="passwordLoading"
                @click="handleVerifyPassword"
              >
                验证密码
              </Button>
            </div>
          </div>
        </Card>

        <!-- 文件分享（单文件） -->
        <Card
          v-if="passwordVerified && shareInfo && shareInfo.type === 'file'"
          :title="shareInfo.fileName"
        >
          <!-- 图片预览 -->
          <div v-if="shareInfo.contentType?.startsWith('image/')" class="text-center">
            <Image
              :src="`/api/v1/share/${shareCode}/file`"
              class="max-w-full"
              style="max-height: 400px"
            />
          </div>

          <!-- PDF 预览 -->
          <div
            v-else-if="
              shareInfo.contentType?.includes('pdf') ||
              shareInfo.fileName?.toLowerCase().endsWith('.pdf')
            "
          >
            <iframe
              :src="`/api/v1/share/${shareCode}/file`"
              sandbox="allow-scripts allow-same-origin"
              referrerpolicy="no-referrer"
              style="width: 100%; height: 400px"
              frameborder="0"
            />
          </div>

          <!-- 视频预览 -->
          <div v-else-if="shareInfo.contentType?.startsWith('video/')">
            <div class="text-center py-6">
              <div class="text-6xl mb-4">🎬</div>
              <h3 class="text-xl font-medium mb-2">{{ shareInfo.fileName }}</h3>
              <p class="text-gray-500 mb-4">{{ formatFileSize(shareInfo.fileSize) }}</p>
              <div class="flex items-center justify-center gap-3">
                <Button type="primary" size="large" @click="openVideoModal">
                  ▶️ 播放视频
                </Button>
                <Button size="large" @click="downloadSharedFile">
                  📥 下载视频
                </Button>
              </div>
            </div>
          </div>

          <!-- 音频预览（增强播放器） -->
          <div v-else-if="shareInfo.contentType?.startsWith('audio/')">
            <div class="bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg p-6 text-white">
              <div class="text-center mb-4">
                <div class="text-6xl mb-3">🎵</div>
                <h3 class="text-lg font-medium truncate">{{ shareInfo.fileName }}</h3>
                <p class="text-sm text-white/70 mt-1">{{ shareInfo.sharerName }} 分享</p>
              </div>

              <!-- 进度条 -->
              <div class="mb-4">
                <div
                  class="h-2 bg-white/30 rounded-full cursor-pointer"
                  @click="seekTo"
                >
                  <div
                    class="h-full bg-white rounded-full transition-all"
                    :style="{ width: duration ? `${(currentTime / duration) * 100}%` : '0%' }"
                  />
                </div>
                <div class="flex justify-between text-xs mt-1 text-white/70">
                  <span>{{ formatTime(currentTime) }}</span>
                  <span>{{ formatTime(duration) }}</span>
                </div>
              </div>

              <!-- 控制按钮 -->
              <div class="flex items-center justify-center gap-4">
                <Tooltip :title="playModeTooltip">
                  <button
                    class="text-xl hover:scale-110 transition-transform"
                    @click="togglePlayMode"
                  >
                    {{ playModeIcon }}
                  </button>
                </Tooltip>

                <button
                  class="text-2xl hover:scale-110 transition-transform"
                  @click="playPrevious"
                >
                  ⏮
                </button>

                <button
                  class="text-4xl hover:scale-110 transition-transform"
                  @click="togglePlay"
                >
                  {{ isPlaying ? '⏸' : '▶️' }}
                </button>

                <button
                  class="text-2xl hover:scale-110 transition-transform"
                  @click="playNext"
                >
                  ⏭
                </button>

                <Tooltip title="静音">
                  <button
                    class="text-xl hover:scale-110 transition-transform"
                    @click="toggleMute"
                  >
                    {{ volume > 0 ? '🔊' : '🔇' }}
                  </button>
                </Tooltip>
              </div>

              <!-- 音量和速度控制 -->
              <div class="flex items-center justify-between mt-4 text-sm">
                <div class="flex items-center gap-2">
                  <span class="text-white/70">音量</span>
                  <div
                    class="w-20 h-1.5 bg-white/30 rounded-full cursor-pointer"
                    @click="changeVolume"
                  >
                    <div
                      class="h-full bg-white rounded-full"
                      :style="{ width: `${volume * 100}%` }"
                    />
                  </div>
                </div>

                <div class="flex items-center gap-2">
                  <span class="text-white/70">速度</span>
                  <Select
                    :value="playbackRate"
                    size="small"
                    style="width: 80px"
                    @change="changeSpeed"
                  >
                    <SelectOption v-for="s in speedOptions" :key="s" :value="s">
                      {{ s }}x
                    </SelectOption>
                  </Select>
                </div>
              </div>

              <!-- 快捷键提示 -->
              <div class="mt-4 text-center text-xs text-white/50">
                空格: 暂停 | ←→: 快退/快进 | ↑↓: 音量 | M: 静音 | N: 下一首 | P: 上一首
              </div>
            </div>

            <!-- 隐藏的音频元素 -->
            <audio
              ref="audioRef"
              preload="auto"
              @timeupdate="onTimeUpdate"
              @ended="onEnded"
              @error="onError"
              @play="isPlaying = true"
              @pause="isPlaying = false"
              @loadedmetadata="duration = audioRef?.duration || 0"
            />
          </div>

          <!-- 其他文件类型 -->
          <div v-else class="text-center py-8 text-gray-500">
            <div class="text-4xl mb-3">📄</div>
            <p>该文件类型不支持在线预览</p>
          </div>

          <!-- 文件信息 -->
          <Descriptions :column="2" class="mt-4">
            <DescriptionsItem label="文件名">{{ shareInfo.fileName }}</DescriptionsItem>
            <DescriptionsItem label="文件大小">{{ formatFileSize(shareInfo.fileSize) }}</DescriptionsItem>
            <DescriptionsItem label="分享者">
              <img
                v-if="shareInfo.sharerAvatar"
                :src="shareInfo.sharerAvatar"
                class="w-6 h-6 rounded-full inline-block mr-1"
              />
              {{ shareInfo.sharerName }}
            </DescriptionsItem>
            <DescriptionsItem label="过期时间">{{ formatDate(shareInfo.expireAt) }}</DescriptionsItem>
          </Descriptions>

          <!-- 下载按钮 -->
          <div class="mt-4 text-center">
            <Button type="primary" size="large" @click="downloadSharedFile">
              下载文件
            </Button>
          </div>
        </Card>

        <!-- 文件夹分享 -->
        <Card
          v-if="passwordVerified && shareInfo && shareInfo.type === 'folder'"
          :title="shareInfo.folderName"
        >
          <!-- 文件夹信息和统计 -->
          <div class="mb-4 flex flex-wrap items-center gap-4">
            <Descriptions :column="2" size="small">
              <DescriptionsItem label="文件夹">{{ shareInfo.folderName }}</DescriptionsItem>
              <DescriptionsItem label="文件总数">{{ folderFilesTotal }} 个</DescriptionsItem>
              <DescriptionsItem label="分享者">
                <img
                  v-if="shareInfo.sharerAvatar"
                  :src="shareInfo.sharerAvatar"
                  class="w-5 h-5 rounded-full inline-block mr-1"
                />
                {{ shareInfo.sharerName }}
              </DescriptionsItem>
              <DescriptionsItem label="过期时间">{{ formatDate(shareInfo.expireAt) }}</DescriptionsItem>
            </Descriptions>

            <!-- 媒体统计 -->
            <div class="flex items-center gap-3 ml-auto">
              <Tag v-if="stats.audioCount > 0" color="blue">
                🎵 音频: {{ stats.audioCount }}
              </Tag>
              <Tag v-if="stats.videoCount > 0" color="orange">
                🎬 视频: {{ stats.videoCount }}
              </Tag>
              <Tag color="default">
                📦 {{ formatFileSize(stats.totalSize) }}
              </Tag>
            </div>
          </div>

          <!-- 当前播放器（文件夹分享时显示） -->
          <div
            v-if="currentPlayingFile"
            class="mb-4 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg p-4 text-white"
          >
            <div class="flex items-center gap-4">
              <!-- 播放信息 -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-1">
                  <span class="text-lg">{{ getFileTypeIcon(currentPlayingFile.contentType) }}</span>
                  <span class="font-medium truncate">{{ currentPlayingFile.fileName }}</span>
                </div>

                <!-- 进度条 -->
                <div
                  class="h-1.5 bg-white/30 rounded-full cursor-pointer mb-2"
                  @click="seekTo"
                >
                  <div
                    class="h-full bg-white rounded-full transition-all"
                    :style="{ width: duration ? `${(currentTime / duration) * 100}%` : '0%' }"
                  />
                </div>

                <div class="flex items-center justify-between text-xs text-white/70">
                  <span>{{ formatTime(currentTime) }}</span>
                  <span>{{ formatTime(duration) }}</span>
                </div>
              </div>

              <!-- 控制按钮 -->
              <div class="flex items-center gap-3">
                <Tooltip :title="playModeTooltip">
                  <button
                    class="text-lg hover:scale-110 transition-transform"
                    @click="togglePlayMode"
                  >
                    {{ playModeIcon }}
                  </button>
                </Tooltip>

                <button
                  class="text-xl hover:scale-110 transition-transform"
                  @click="playPrevious"
                >
                  ⏮
                </button>

                <button
                  class="text-3xl hover:scale-110 transition-transform"
                  @click="togglePlay"
                >
                  {{ isPlaying ? '⏸' : '▶️' }}
                </button>

                <button
                  class="text-xl hover:scale-110 transition-transform"
                  @click="playNext"
                >
                  ⏭
                </button>

                <Tooltip title="静音">
                  <button
                    class="text-lg hover:scale-110 transition-transform"
                    @click="toggleMute"
                  >
                    {{ volume > 0 ? '🔊' : '🔇' }}
                  </button>
                </Tooltip>

                <Tooltip title="播放列表">
                  <button
                    class="text-lg hover:scale-110 transition-transform"
                    :class="{ 'text-yellow-300': showPlaylist }"
                    @click="showPlaylist = !showPlaylist"
                  >
                    📋
                  </button>
                </Tooltip>
              </div>
            </div>

            <!-- 播放速度和音量 -->
            <div class="flex items-center justify-between mt-3 text-sm">
              <div class="flex items-center gap-2">
                <span class="text-white/70">音量</span>
                <div
                  class="w-16 h-1 bg-white/30 rounded-full cursor-pointer"
                  @click="changeVolume"
                >
                  <div
                    class="h-full bg-white rounded-full"
                    :style="{ width: `${volume * 100}%` }"
                  />
                </div>
              </div>

              <div class="flex items-center gap-2">
                <span class="text-white/70">速度</span>
                <Select
                  :value="playbackRate"
                  size="small"
                  style="width: 70px"
                  @change="changeSpeed"
                >
                  <SelectOption v-for="s in speedOptions" :key="s" :value="s">
                    {{ s }}x
                  </SelectOption>
                </Select>
              </div>

              <div class="text-white/50 text-xs">
                空格: 暂停 | ←→: 快退/快进 | N/P: 切换
              </div>
            </div>

            <!-- 播放列表面板 -->
            <div
              v-if="showPlaylist"
              class="mt-3 bg-white/10 rounded-lg p-3 max-h-48 overflow-y-auto"
            >
              <div class="text-sm text-white/70 mb-2">
                播放列表 ({{ playlist.length }} 首)
              </div>
              <div
                v-for="(item, index) in playlist"
                :key="item.fileId"
                class="flex items-center gap-2 py-1.5 px-2 rounded cursor-pointer hover:bg-white/10 transition-colors"
                :class="{ 'bg-white/20': currentPlayingFile?.fileId === item.fileId }"
                @click="playFile(item)"
              >
                <span class="text-white/50 w-6 text-right text-xs">{{ index + 1 }}</span>
                <span class="text-sm">{{ getFileTypeIcon(item.contentType) }}</span>
                <span class="flex-1 truncate text-sm">{{ item.fileName }}</span>
                <span class="text-xs text-white/50">{{ formatFileSize(item.fileSize) }}</span>
                <span v-if="currentPlayingFile?.fileId === item.fileId" class="text-xs">
                  {{ isPlaying ? '🔊' : '⏸' }}
                </span>
              </div>
            </div>

            <!-- 隐藏的音频元素 -->
            <audio
              ref="audioRef"
              preload="auto"
              @timeupdate="onTimeUpdate"
              @ended="onEnded"
              @error="onError"
              @play="isPlaying = true"
              @pause="isPlaying = false"
              @loadedmetadata="duration = audioRef?.duration || 0"
            />

            <!-- 隐藏的视频元素 -->
            <video
              v-show="false"
              ref="videoRef"
              preload="auto"
              @timeupdate="onTimeUpdate"
              @ended="onEnded"
              @error="onError"
              @play="isPlaying = true"
              @pause="isPlaying = false"
              @loadedmetadata="duration = videoRef?.duration || 0"
            />
          </div>

          <!-- 搜索栏 -->
          <div class="mb-4 flex items-center gap-4">
            <InputSearch
              v-model:value="searchText"
              placeholder="搜索文件名"
              allow-clear
              style="width: 280px"
            />
            <div class="text-sm text-gray-500">
              共 {{ folderFilesTotal }} 个文件
              <span v-if="stats.mediaCount > 0">
               ，其中 {{ stats.mediaCount }} 个媒体文件
              </span>
            </div>
          </div>

          <!-- 文件列表 -->
          <Table
            :columns="columns"
            :data-source="folderFiles"
            :loading="loading"
            :pagination="{
              current: pagination.current,
              pageSize: pagination.pageSize,
              total: folderFilesTotal,
              showSizeChanger: true,
              showTotal: (total: number) => `共 ${total} 个文件`,
            }"
            row-key="fileId"
            size="small"
            @change="
              (pag: any) => {
                pagination.current = pag.current;
                pagination.pageSize = pag.pageSize;
                loadFolderFiles();
              }
            "
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'fileName'">
                <div class="flex items-center gap-2">
                  <span class="text-lg">{{ getFileTypeIcon(record.contentType) }}</span>
                  <span
                    class="truncate"
                    :class="{ 'text-blue-500 font-medium': currentPlayingFile?.fileId === record.fileId }"
                  >
                    {{ record.fileName }}
                  </span>
                  <span
                    v-if="currentPlayingFile?.fileId === record.fileId"
                    class="text-blue-500 text-xs"
                  >
                    {{ isPlaying ? '🔊 播放中' : '⏸ 已暂停' }}
                  </span>
                </div>
              </template>
              <template v-if="column.key === 'fileSize'">
                {{ formatFileSize(record.fileSize) }}
              </template>
              <template v-if="column.key === 'contentType'">
                <span class="text-sm text-gray-500">{{ record.contentType || '未知' }}</span>
              </template>
              <template v-if="column.key === 'action'">
                <div class="flex items-center gap-1">
                  <Button
                    v-if="isMediaFile(record.contentType)"
                    type="link"
                    size="small"
                    @click="playFile(record)"
                  >
                    {{ currentPlayingFile?.fileId === record.fileId && isPlaying ? '⏸ 暂停' : '▶ 播放' }}
                  </Button>
                  <Button
                    v-else
                    type="link"
                    size="small"
                    @click="viewFileInModal(record)"
                  >
                    预览
                  </Button>
                  <Button type="link" size="small" @click="downloadFile(record)">
                    下载
                  </Button>
                </div>
              </template>
            </template>
          </Table>

          <div
            v-if="!loading && folderFiles.length === 0"
            class="text-center py-8 text-gray-500"
          >
            <div class="text-4xl mb-3">📂</div>
            <p>{{ searchText ? '未找到匹配的文件' : '文件夹内暂无文件' }}</p>
          </div>
        </Card>

        <!-- 预览弹窗 -->
        <Modal
          v-model:open="previewVisible"
          :title="previewName"
          :footer="null"
          :width="
            previewType === 'video' ? 960 : previewType === 'audio' ? 500 : 800
          "
          :mask-closable="true"
          :keyboard="true"
          :destroy-on-close="true"
        >
          <div class="mb-3 text-right">
            <Button size="small" @click="openPreviewInNewTab">
              在新标签页打开
            </Button>
          </div>
          <div
            v-if="previewType === 'image' && previewUrl"
            class="text-center p-6"
          >
            <Image
              :src="previewUrl"
              class="max-w-full"
              style="max-height: 600px"
            />
          </div>
          <iframe
            v-else-if="previewType === 'pdf' && previewUrl"
            :src="previewUrl"
            sandbox="allow-scripts allow-same-origin"
            referrerpolicy="no-referrer"
            style="width: 100%; height: 600px; border: none"
          />
          <div v-else-if="previewType === 'video' && previewUrl" class="bg-black">
            <video
              :src="previewUrl"
              controls
              autoplay
              preload="auto"
              playsinline
              style="display: block; width: 100%; max-height: 70vh; background: #000"
              @error="onVideoError"
              @loadeddata="videoFormatSupported = true; videoFormatError = ''"
            />
            <!-- 播放失败提示 -->
            <div v-if="!videoFormatSupported" class="text-center py-6 bg-gray-900 text-white">
              <div class="text-4xl mb-3">⚠️</div>
              <p class="mb-4">{{ videoFormatError || '视频播放失败' }}</p>
              <Button type="primary" @click="openPreviewInNewTab">
                📥 在新标签页下载
              </Button>
            </div>
          </div>
          <div v-else-if="previewType === 'audio' && previewUrl" class="p-6">
            <div class="text-center mb-4">
              <div class="text-6xl mb-3">🎵</div>
              <p class="text-lg">{{ previewName }}</p>
            </div>
            <audio
              :src="previewUrl"
              controls
              autoplay
              preload="auto"
              style="width: 100%"
            />
          </div>
          <div v-else class="py-12 text-center text-gray-500">
            该文件类型不支持预览
          </div>
        </Modal>

        <!-- 错误提示 -->
        <Card v-if="error">
          <div class="text-center py-8">
            <div class="text-4xl mb-3 text-red-500">⚠️</div>
            <p class="text-lg">{{ error }}</p>
          </div>
        </Card>
      </Spin>
    </div>
  </Page>
</template>

<style scoped>
/* 自定义滚动条 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.3);
}

/* 选中行高亮 */
:deep(.ant-table-row:hover) {
  background-color: rgba(59, 130, 246, 0.05);
}
</style>
