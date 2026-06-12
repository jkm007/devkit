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
  login: '登录',
  logout: '登出',
  'login-fail': '登录失败',
  'change-password': '修改密码',
  'reset-password': '重置密码',
  register: '注册',
  'update-profile': '更新资料',
  'bind-oauth': '绑定OAuth',
  'unbind-oauth': '解绑OAuth',
  'bind-wechat': '绑定微信',
  upload: '上传文件',
  delete: '删除文件',
};

function render(data: Array<{ type: string; count: number }>) {
  renderEcharts({
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      data: data.map((d) => typeLabel[d.type] || d.type),
      axisLabel: { rotate: data.length > 6 ? 30 : 0, fontSize: 11 },
    },
    yAxis: { type: 'value' },
    series: [
      {
        type: 'bar',
        barMaxWidth: 40,
        data: data.map((d) => d.count),
        itemStyle: {
          color: '#5470c6',
          borderRadius: [4, 4, 0, 0],
        },
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
