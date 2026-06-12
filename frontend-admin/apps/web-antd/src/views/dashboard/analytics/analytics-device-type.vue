<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';

import { ref, watch } from 'vue';

import { EchartsUI, useEcharts } from '@vben/plugins/echarts';

const props = defineProps<{
  data?: Array<{ type: string; count: number }>;
}>();

const chartRef = ref<EchartsUIType>();
const { renderEcharts } = useEcharts(chartRef);

const typeLabel: Record<string, string> = {
  web: 'Web端',
  h5: 'H5端',
  app: 'App端',
  miniapp: '小程序',
};

const colors = ['#5470c6', '#91cc75', '#fac858', '#ee6666'];

function render(data: Array<{ type: string; count: number }>) {
  renderEcharts({
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, textStyle: { fontSize: 12 } },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: true,
        itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
        label: { show: false },
        emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
        data: data.map((d, i) => ({
          name: typeLabel[d.type] || d.type,
          value: d.count,
          itemStyle: { color: colors[i % colors.length] },
        })),
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
