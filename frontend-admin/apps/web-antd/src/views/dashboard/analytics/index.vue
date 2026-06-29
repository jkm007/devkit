<script lang="ts" setup>
import type { AnalysisOverviewItem } from '@vben/common-ui';
import type { TabOption } from '@vben/types';

import { onMounted, ref } from 'vue';

import {
  AnalysisChartCard,
  AnalysisChartsTabs,
  AnalysisOverview,
  Page,
} from '@vben/common-ui';
import {
  SvgBellIcon,
  SvgCakeIcon,
  SvgCardIcon,
  SvgDownloadIcon,
} from '@vben/icons';

import type { DashboardStats } from '#/api/system/dashboard';

import { getDashboardStats } from '#/api/system/dashboard';

import AnalyticsDevicePlatform from './analytics-device-platform.vue';
import AnalyticsDeviceType from './analytics-device-type.vue';
import AnalyticsEventTrend from './analytics-event-trend.vue';
import AnalyticsEventType from './analytics-event-type.vue';
import AnalyticsRecentLogins from './analytics-recent-logins.vue';
import AnalyticsStorage from './analytics-storage.vue';

const stats = ref<DashboardStats | null>(null);

const overviewItems = ref<AnalysisOverviewItem[]>([]);

const chartTabs: TabOption[] = [
  { label: '事件趋势', value: 'trend' },
  { label: '事件分布', value: 'events' },
];

onMounted(async () => {
  try {
    stats.value = await getDashboardStats();
    const s = stats.value;
    overviewItems.value = [
      {
        icon: SvgCardIcon,
        title: '用户总量',
        totalTitle: '活跃用户',
        totalValue: s.overview.activeUsers,
        value: s.overview.userCount,
      },
      {
        icon: SvgCakeIcon,
        title: '今日登录',
        totalTitle: '今日事件',
        totalValue: s.overview.todayEvents,
        value: s.overview.todayLogins,
      },
      {
        icon: SvgDownloadIcon,
        title: '文件数量',
        totalTitle: '存储用量',
        totalValue: s.overview.totalStorage,
        value: s.overview.fileCount,
      },
      {
        icon: SvgBellIcon,
        title: '在线设备',
        totalTitle: '本月事件',
        totalValue: s.eventsTrend.reduce(
          (sum, d) => sum + d.success + d.fail,
          0,
        ),
        value: s.overview.onlineDevices,
      },
    ];
  } catch {
    // ignore
  }
});
</script>

<template>
  <Page title="分析页" content-class="p-5">
    <AnalysisOverview :items="overviewItems" />

    <template v-if="stats">
      <AnalysisChartsTabs :tabs="chartTabs" class="mt-5">
        <template #trend>
          <AnalyticsEventTrend :data="stats.eventsTrend" />
        </template>
        <template #events>
          <AnalyticsEventType :data="stats.eventsByType" />
        </template>
      </AnalysisChartsTabs>

      <div class="mt-5 grid grid-cols-1 gap-5 md:grid-cols-3">
        <AnalysisChartCard title="设备类型分布">
          <AnalyticsDeviceType :data="stats.deviceByType" />
        </AnalysisChartCard>
        <AnalysisChartCard title="平台分布">
          <AnalyticsDevicePlatform :data="stats.deviceByPlatform" />
        </AnalysisChartCard>
        <AnalysisChartCard title="最近登录">
          <AnalyticsRecentLogins :data="stats.recentLogins" />
        </AnalysisChartCard>
      </div>

      <div class="mt-5">
        <AnalysisChartCard title="存储分析">
          <AnalyticsStorage />
        </AnalysisChartCard>
      </div>
    </template>
  </Page>
</template>
