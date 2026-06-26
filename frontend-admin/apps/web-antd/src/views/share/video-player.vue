<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import { Page } from '@vben/common-ui';

import {
  Button,
  Select,
  SelectOption,
  Spin,
  Tooltip,
  message,
} from 'ant-design-vue';

import { getShareInfo, getShareFolderFiles, verifySharePassword } from '#/api/file';

defineOptions({ name: 'ShareVideoPlayer' });

const route = useRoute();
const loading = ref(true);
const shareInfo = ref<any>(null);
const error = ref('');

// 播放器状态
const videoRef = ref<HTMLVideoElement | null>(null);
const isPlaying = ref(false);
const currentTime = ref(0);
const duration = ref(0);
const volume = ref(1);
const playbackRate = ref(1);
const playMode = ref<'loop' | 'random' | 'sequential'>('sequential');
const currentPlayingFile = ref<any>(null);
const playlist = ref<any[]>([]);
const showPlaylist = ref(true);

// 播放速度选项
const speedOptions = [0.5, 0.75, 1, 1.25, 1.5, 2];

// 跳过片头片尾设置
const skipIntro = ref(0); // 跳过片头秒数
const skipOutro = ref(0); // 跳过片尾秒数
const showSkipSettings = ref(false);

// localStorage 键名
const STORAGE_KEY_PROGRESS = 'share_play_progress';
const STORAGE_KEY_VOLUME = 'share_play_volume';
const STORAGE_KEY_RATE = 'share_play_rate';
const STORAGE_KEY_MODE = 'share_play_mode';
const STORAGE_KEY_SKIP_INTRO = 'share_skip_intro';
const STORAGE_KEY_SKIP_OUTRO = 'share_skip_outro';

const shareCode = computed(() => route.params.code as string);
const fileId = computed(() => route.params.fileId as string);

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

function getShareFileUrl(file: any) {
  return `/api/v1/share/${shareCode.value}/file/${file.fileId}`;
}

// ==================== 播放进度保存/恢复 ====================

function savePlayProgress(fileId: number, time: number) {
  try {
    const progress = JSON.parse(localStorage.getItem(STORAGE_KEY_PROGRESS) || '{}');
    progress[fileId] = { time, updatedAt: Date.now() };
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

function saveSkipIntro(seconds: number) {
  skipIntro.value = seconds;
  localStorage.setItem(STORAGE_KEY_SKIP_INTRO, String(seconds));
}

function saveSkipOutro(seconds: number) {
  skipOutro.value = seconds;
  localStorage.setItem(STORAGE_KEY_SKIP_OUTRO, String(seconds));
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

    const savedSkipIntro = localStorage.getItem(STORAGE_KEY_SKIP_INTRO);
    if (savedSkipIntro) skipIntro.value = Number(savedSkipIntro);

    const savedSkipOutro = localStorage.getItem(STORAGE_KEY_SKIP_OUTRO);
    if (savedSkipOutro) skipOutro.value = Number(savedSkipOutro);
  } catch {
    // ignore
  }
}

// ==================== 播放器控制 ====================

function playFile(file: any) {
  currentPlayingFile.value = file;
  isPlaying.value = true;

  nextTick(() => {
    if (videoRef.value) {
      videoRef.value.src = getShareFileUrl(file);
      videoRef.value.playbackRate = playbackRate.value;
      videoRef.value.volume = volume.value;

      // 恢复播放进度，或跳过片头
      const savedTime = getPlayProgress(file.fileId);
      if (savedTime > 0) {
        videoRef.value.currentTime = savedTime;
      } else if (skipIntro.value > 0) {
        videoRef.value.currentTime = skipIntro.value;
      }

      videoRef.value.play().catch(() => {
        // 自动播放被阻止
      });
    }
  });
}

function togglePlay() {
  if (videoRef.value) {
    if (isPlaying.value) {
      videoRef.value.pause();
    } else {
      videoRef.value.play().catch(() => {});
    }
    isPlaying.value = !isPlaying.value;
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

  if (videoRef.value) {
    videoRef.value.currentTime = time;
  }
}

function changeSpeed(speed: any) {
  const rate = Number(speed) || 1;
  playbackRate.value = rate;
  savePlaybackRate(rate);

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

  if (videoRef.value) videoRef.value.volume = volume.value;
}

// ==================== 事件处理 ====================

function onTimeUpdate() {
  if (videoRef.value) {
    currentTime.value = videoRef.value.currentTime;
    duration.value = videoRef.value.duration;

    // 保存播放进度
    if (Math.floor(videoRef.value.currentTime) % 5 === 0) {
      savePlayProgress(currentPlayingFile.value?.fileId, videoRef.value.currentTime);
    }

    // 跳过片尾：当播放到 (总时长 - 片尾秒数) 时，自动播放下一个
    if (skipOutro.value > 0 && duration.value > 0) {
      const outroTime = duration.value - skipOutro.value;
      if (currentTime.value >= outroTime) {
        playNext();
      }
    }
  }
}

function onEnded() {
  if (playMode.value === 'loop') {
    if (videoRef.value) {
      videoRef.value.currentTime = 0;
      videoRef.value.play().catch(() => {});
    }
  } else {
    playNext();
  }
}

// ==================== 键盘快捷键 ====================

function handleKeydown(event: KeyboardEvent) {
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
      if (videoRef.value) {
        videoRef.value.currentTime = Math.max(0, videoRef.value.currentTime - 5);
      }
      break;
    }
    case 'ArrowRight': {
      event.preventDefault();
      if (videoRef.value) {
        videoRef.value.currentTime = Math.min(videoRef.value.duration, videoRef.value.currentTime + 5);
      }
      break;
    }
    case 'ArrowUp': {
      event.preventDefault();
      volume.value = Math.min(1, volume.value + 0.1);
      saveVolume(volume.value);
      if (videoRef.value) videoRef.value.volume = volume.value;
      break;
    }
    case 'ArrowDown': {
      event.preventDefault();
      volume.value = Math.max(0, volume.value - 0.1);
      saveVolume(volume.value);
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

// ==================== 数据加载 ====================

async function loadData() {
  try {
    loading.value = true;
    const info = await getShareInfo(shareCode.value);
    shareInfo.value = info;

    if (info.type === 'folder') {
      // 加载文件夹中的所有视频
      const result = await getShareFolderFiles(shareCode.value, {
        page: 1,
        pageSize: 1000,
      });
      const files = result?.items || [];
      playlist.value = files.filter((f: any) => f.contentType?.startsWith('video/'));

      // 找到当前文件
      const targetFile = playlist.value.find((f: any) => String(f.fileId) === fileId.value);
      if (targetFile) {
        playFile(targetFile);
      } else if (playlist.value.length > 0) {
        playFile(playlist.value[0]);
      }
    } else if (info.type === 'file') {
      // 单文件分享
      const file = {
        fileId: fileId.value || shareCode.value,
        fileName: info.fileName,
        fileSize: info.fileSize,
        contentType: info.contentType,
      };
      playlist.value = [file];
      playFile(file);
    }
  } catch (err: any) {
    error.value = err.message || '加载失败';
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  restoreSettings();
  loadData();
  document.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
  <Page title="视频播放">
    <div class="h-full flex flex-col bg-gray-900 text-white">
      <Spin :spinning="loading">
        <!-- 错误提示 -->
        <div v-if="error" class="flex items-center justify-center h-screen">
          <div class="text-center">
            <div class="text-4xl mb-3">⚠️</div>
            <p class="text-lg mb-4">{{ error }}</p>
            <Button type="primary" @click="loadData">重试</Button>
          </div>
        </div>

        <!-- 播放器 -->
        <div v-else-if="currentPlayingFile" class="flex flex-col h-screen">
          <!-- 视频区域 -->
          <div class="flex-1 flex items-center justify-center bg-black min-h-0">
            <video
              ref="videoRef"
              controls
              autoplay
              preload="auto"
              playsinline
              style="max-width: 100%; max-height: 100%; object-fit: contain;"
              @timeupdate="onTimeUpdate"
              @ended="onEnded"
              @play="isPlaying = true"
              @pause="isPlaying = false"
              @loadedmetadata="duration = videoRef?.duration || 0"
            />
          </div>

          <!-- 控制栏 -->
          <div class="bg-gray-800 p-4">
            <!-- 进度条 -->
            <div
              class="h-2 bg-gray-600 rounded-full cursor-pointer mb-3"
              @click="seekTo"
            >
              <div
                class="h-full bg-blue-500 rounded-full transition-all"
                :style="{ width: duration ? `${(currentTime / duration) * 100}%` : '0%' }"
              />
            </div>

            <!-- 控制按钮 -->
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-4">
                <!-- 上一个 -->
                <Tooltip title="上一个 (P)" v-if="playlist.length > 1">
                  <button class="text-2xl hover:text-blue-400 transition-colors" @click="playPrevious">
                    ⏮
                  </button>
                </Tooltip>

                <!-- 播放/暂停 -->
                <button class="text-3xl hover:text-blue-400 transition-colors" @click="togglePlay">
                  {{ isPlaying ? '⏸' : '▶️' }}
                </button>

                <!-- 下一个 -->
                <Tooltip title="下一个 (N)" v-if="playlist.length > 1">
                  <button class="text-2xl hover:text-blue-400 transition-colors" @click="playNext">
                    ⏭
                  </button>
                </Tooltip>

                <!-- 时间 -->
                <span class="text-gray-300">
                  {{ formatTime(currentTime) }} / {{ formatTime(duration) }}
                </span>

                <!-- 文件名 -->
                <span class="text-gray-400 truncate max-w-xs">
                  {{ currentPlayingFile.fileName }}
                </span>
              </div>

              <div class="flex items-center gap-4">
                <!-- 播放模式 -->
                <Tooltip :title="playModeTooltip" v-if="playlist.length > 1">
                  <button class="text-xl hover:text-blue-400 transition-colors" @click="togglePlayMode">
                    {{ playModeIcon }}
                  </button>
                </Tooltip>

                <!-- 播放速度 -->
                <div class="flex items-center gap-2">
                  <span class="text-gray-400 text-sm">速度</span>
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

                <!-- 音量 -->
                <div class="flex items-center gap-2">
                  <Tooltip :title="volume > 0 ? '静音 (M)' : '取消静音'">
                    <button class="text-xl hover:text-blue-400 transition-colors" @click="toggleMute">
                      {{ volume > 0 ? '🔊' : '🔇' }}
                    </button>
                  </Tooltip>
                </div>

                <!-- 跳过片头片尾 -->
                <Tooltip title="跳过片头片尾">
                  <button
                    class="text-xl hover:text-blue-400 transition-colors"
                    :class="{ 'text-blue-400': showSkipSettings }"
                    @click="showSkipSettings = !showSkipSettings"
                  >
                    ⏩
                  </button>
                </Tooltip>

                <!-- 播放列表 -->
                <Tooltip title="播放列表">
                  <button
                    class="text-xl hover:text-blue-400 transition-colors"
                    :class="{ 'text-blue-400': showPlaylist }"
                    @click="showPlaylist = !showPlaylist"
                  >
                    📋
                  </button>
                </Tooltip>
              </div>
            </div>

            <!-- 跳过片头片尾设置 -->
            <div v-if="showSkipSettings" class="bg-gray-700 px-4 py-3 flex items-center gap-6">
              <span class="text-gray-300 text-sm">跳过设置：</span>
              <div class="flex items-center gap-2">
                <span class="text-gray-400 text-sm">片头</span>
                <Select
                  :value="skipIntro"
                  size="small"
                  style="width: 80px"
                  @change="saveSkipIntro"
                >
                  <SelectOption :value="0">不跳过</SelectOption>
                  <SelectOption :value="5">5秒</SelectOption>
                  <SelectOption :value="10">10秒</SelectOption>
                  <SelectOption :value="15">15秒</SelectOption>
                  <SelectOption :value="30">30秒</SelectOption>
                  <SelectOption :value="60">60秒</SelectOption>
                  <SelectOption :value="90">90秒</SelectOption>
                </Select>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-gray-400 text-sm">片尾</span>
                <Select
                  :value="skipOutro"
                  size="small"
                  style="width: 80px"
                  @change="saveSkipOutro"
                >
                  <SelectOption :value="0">不跳过</SelectOption>
                  <SelectOption :value="5">5秒</SelectOption>
                  <SelectOption :value="10">10秒</SelectOption>
                  <SelectOption :value="15">15秒</SelectOption>
                  <SelectOption :value="30">30秒</SelectOption>
                  <SelectOption :value="60">60秒</SelectOption>
                  <SelectOption :value="90">90秒</SelectOption>
                </Select>
              </div>
              <span class="text-gray-500 text-xs">设置会保存，下次自动应用</span>
            </div>

            <!-- 快捷键提示 -->
            <div class="text-center text-xs text-gray-500 mt-3">
              空格: 暂停 | ←→: 快退/快进 5秒 | ↑↓: 音量 | M: 静音
              <span v-if="playlist.length > 1"> | N: 下一个 | P: 上一个</span>
            </div>
          </div>

          <!-- 播放列表 -->
          <div v-if="showPlaylist && playlist.length > 1" class="bg-gray-800 border-t border-gray-700 max-h-60 overflow-y-auto">
            <div class="px-4 py-2 text-sm text-gray-400 border-b border-gray-700 sticky top-0 bg-gray-800">
              播放列表 ({{ playlist.length }} 个视频)
            </div>
            <div
              v-for="(item, index) in playlist"
              :key="item.fileId"
              class="flex items-center gap-3 px-4 py-2 cursor-pointer hover:bg-gray-700 transition-colors"
              :class="{ 'bg-gray-700 text-blue-400': currentPlayingFile?.fileId === item.fileId }"
              @click="playFile(item)"
            >
              <span class="text-gray-500 w-6 text-right text-sm">{{ index + 1 }}</span>
              <span class="text-lg">🎬</span>
              <span class="flex-1 truncate">{{ item.fileName }}</span>
              <span class="text-sm text-gray-500">{{ formatFileSize(item.fileSize) }}</span>
              <span v-if="currentPlayingFile?.fileId === item.fileId" class="text-blue-400">
                {{ isPlaying ? '🔊' : '⏸' }}
              </span>
            </div>
          </div>
        </div>
      </Spin>
    </div>
  </Page>
</template>

<style scoped>
:deep(.ant-select-selector) {
  background-color: #374151 !important;
  border-color: #4b5563 !important;
  color: white !important;
}

:deep(.ant-select-arrow) {
  color: #9ca3af !important;
}
</style>
