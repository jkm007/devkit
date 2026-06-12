<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';

import { watch } from 'vue';

import { ref } from 'vue';

import { EchartsUI, useEcharts } from '@vben/plugins/echarts';

const props = defineProps<{
  data?: Array<{ date: string; success: number; fail: number }>;
}>();

const chartRef = ref<EchartsUIType>();
const { renderEcharts } = useEcharts(chartRef);

function render(data: Array<{ date: string; success: number; fail: number }>) {
  renderEcharts({
    tooltip: { trigger: 'axis' },
    legend: { data: ['成功', '失败'], bottom: 0 },
    grid: { left: '3%', right: '4%', bottom: '10%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: data.map((d) => d.date.slice(5)),
    },
    yAxis: { type: 'value' },
    series: [
      {
        name: '成功',
        type: 'line',
        smooth: true,
        areaStyle: { color: 'rgba(82,196,26,0.15)' },
        lineStyle: { color: '#52c41a' },
        itemStyle: { color: '#52c41a' },
        data: data.map((d) => d.success),
      },
      {
        name: '失败',
        type: 'line',
        smooth: true,
        areaStyle: { color: 'rgba(255,77,79,0.15)' },
        lineStyle: { color: '#ff4d4f' },
        itemStyle: { color: '#ff4d4f' },
        data: data.map((d) => d.fail),
      },
    ],
  });
}

watch(
  () => props.data,
  (val) => {
    if (val && val.length > 0) render(val);
  },
  { immediate: true },
);
</script>

<template>
  <EchartsUI ref="chartRef" />
</template>
