<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';

import { onMounted, ref } from 'vue';

import { EchartsUI, useEcharts } from '@vben/plugins/echarts';

import { getStorageStats } from '#/api/system/user';

interface StorageStats {
  totalUsed: number;
  totalQuota: number;
  userCount: number;
  fileCount: number;
  byType: Array<{ type: string; count: number; size: number }>;
  topUsers: Array<{
    userId: number;
    userName: string;
    used: number;
    quota: number;
  }>;
}

const stats = ref<StorageStats | null>(null);

const formatSize = (bytes: number) => {
  if (bytes <= 0) return '0 B';
  if (bytes >= 1073741824) return `${(bytes / 1073741824).toFixed(1)} GB`;
  if (bytes >= 1048576) return `${(bytes / 1048576).toFixed(1)} MB`;
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(0)} KB`;
  return `${bytes} B`;
};

const typeLabel: Record<string, string> = {
  image: '图片',
  video: '视频',
  audio: '音频',
  document: '文档',
  other: '其他',
};

const typeColors: Record<string, string> = {
  image: '#5470c6',
  video: '#ee6666',
  audio: '#fac858',
  document: '#73c0de',
  other: '#91cc75',
};

const pieRef = ref<EchartsUIType>();
const barRef = ref<EchartsUIType>();
const { renderEcharts: renderPie } = useEcharts(pieRef);
const { renderEcharts: renderBar } = useEcharts(barRef);

onMounted(async () => {
  try {
    stats.value = await getStorageStats();
  } catch {
    return;
  }

  if (!stats.value) return;

  // 饼图: 按类型分布
  const pieData = stats.value.byType.map((item) => ({
    name: typeLabel[item.type] || item.type,
    value: item.size,
    itemStyle: { color: typeColors[item.type] || '#999' },
  }));
  renderPie({
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        return `${params.name}: ${formatSize(params.value)} (${params.percent}%)`;
      },
    },
    legend: {
      bottom: 0,
      textStyle: { fontSize: 12 },
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: true,
        itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
        label: { show: false },
        emphasis: {
          label: { show: true, fontSize: 14, fontWeight: 'bold' },
        },
        data: pieData,
      },
    ],
  });

  // 柱状图: Top 用户
  const topUsers = stats.value.topUsers.slice(0, 8);
  renderBar({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: any) => {
        const item = params[0];
        const user = topUsers[item.dataIndex];
        const quotaText =
          user.quota > 0 ? ` / ${formatSize(user.quota)}` : ' (不限)';
        return `${user.userName}<br/>已用: ${formatSize(user.used)}${quotaText}`;
      },
    },
    grid: { left: 60, right: 20, top: 10, bottom: 30 },
    xAxis: {
      type: 'value',
      axisLabel: {
        formatter: (val: number) => formatSize(val),
      },
    },
    yAxis: {
      type: 'category',
      data: topUsers.map((u) => u.userName),
      axisLabel: { fontSize: 12 },
    },
    series: [
      {
        type: 'bar',
        barWidth: 16,
        data: topUsers.map((u) => ({
          value: u.used,
          itemStyle: {
            color:
              u.quota > 0 && u.used / u.quota > 0.9
                ? '#ee6666'
                : u.quota > 0 && u.used / u.quota > 0.7
                  ? '#fac858'
                  : '#5470c6',
          },
        })),
      },
    ],
  });
});
</script>

<template>
  <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
    <!-- 左侧: 概览 + 饼图 -->
    <div>
      <div class="mb-3 text-sm font-medium text-gray-500">存储分布</div>
      <div v-if="stats" class="mb-3 flex gap-4 text-xs">
        <span
          >总用量: <strong>{{ formatSize(stats.totalUsed) }}</strong></span
        >
        <span
          >文件数: <strong>{{ stats.fileCount }}</strong></span
        >
        <span
          >用户数: <strong>{{ stats.userCount }}</strong></span
        >
      </div>
      <EchartsUI ref="pieRef" class="h-64" />
    </div>
    <!-- 右侧: Top 用户 -->
    <div>
      <div class="mb-3 text-sm font-medium text-gray-500">
        用户存储用量 Top 8
      </div>
      <EchartsUI ref="barRef" class="h-64" />
    </div>
  </div>
</template>
