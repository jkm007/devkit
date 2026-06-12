<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';

import { nextTick, onMounted, ref } from 'vue';

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

onMounted(async () => {
  await nextTick();
  const data = props.data;
  if (!data || data.length === 0) return;

  renderEcharts({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
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
});
</script>

<template>
  <div v-if="!data || data.length === 0" class="flex h-64 items-center justify-center text-gray-400">
    暂无设备数据
  </div>
  <EchartsUI v-else ref="chartRef" />
</template>
