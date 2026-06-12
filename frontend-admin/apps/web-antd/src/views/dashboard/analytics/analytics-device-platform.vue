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
  web: 'Web',
  h5: 'H5',
  ios: 'iOS',
  android: 'Android',
  miniapp: '小程序',
};

const colors = ['#5470c6', '#91cc75', '#ee6666', '#fac858', '#73c0de'];

onMounted(async () => {
  await nextTick();
  const data = props.data;
  if (!data || data.length === 0) return;

  renderEcharts({
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, textStyle: { fontSize: 12 } },
    series: [
      {
        type: 'pie',
        radius: '65%',
        center: ['50%', '45%'],
        avoidLabelOverlap: true,
        itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
        label: { show: true, formatter: '{b}: {c}' },
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
  <EchartsUI ref="chartRef" />
</template>
