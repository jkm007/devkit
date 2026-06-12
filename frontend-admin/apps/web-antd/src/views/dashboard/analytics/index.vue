<script lang="ts" setup>
import type { AnalysisOverviewItem } from '@vben/common-ui';
import type { TabOption } from '@vben/types';

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

import AnalyticsStorage from './analytics-storage.vue';
import AnalyticsTrends from './analytics-trends.vue';
import AnalyticsVisitsData from './analytics-visits-data.vue';
import AnalyticsVisitsSales from './analytics-visits-sales.vue';
import AnalyticsVisitsSource from './analytics-visits-source.vue';
import AnalyticsVisits from './analytics-visits.vue';

const overviewItems: AnalysisOverviewItem[] = [
  {
    icon: SvgCardIcon,
    title: '用户量',
    totalTitle: '总用户量',
    totalValue: 120_000,
    value: 2000,
  },
  {
    icon: SvgCakeIcon,
    title: '访问量',
    totalTitle: '总访问量',
    totalValue: 500_000,
    value: 20_000,
  },
  {
    icon: SvgDownloadIcon,
    title: '下载量',
    totalTitle: '总下载量',
    totalValue: 120_000,
    value: 8000,
  },
  {
    icon: SvgBellIcon,
    title: '使用量',
    totalTitle: '总使用量',
    totalValue: 50_000,
    value: 5000,
  },
];

const chartTabs: TabOption[] = [
  {
    label: '流量趋势',
    value: 'trends',
  },
  {
    label: '月访问量',
    value: 'visits',
  },
];
</script>

<template>
  <Page title="分析页" content-class="p-5">
    <AnalysisOverview :items="overviewItems" />
    <AnalysisChartsTabs :tabs="chartTabs" class="mt-5">
      <template #trends>
        <AnalyticsTrends />
      </template>
      <template #visits>
        <AnalyticsVisits />
      </template>
    </AnalysisChartsTabs>

    <div class="mt-5 grid grid-cols-1 gap-5 md:grid-cols-3">
      <AnalysisChartCard title="访问数量">
        <AnalyticsVisitsData />
      </AnalysisChartCard>
      <AnalysisChartCard title="访问来源">
        <AnalyticsVisitsSource />
      </AnalysisChartCard>
      <AnalysisChartCard title="访问趋势">
        <AnalyticsVisitsSales />
      </AnalysisChartCard>
    </div>

    <div class="mt-5">
      <AnalysisChartCard title="存储分析">
        <AnalyticsStorage />
      </AnalysisChartCard>
    </div>
  </Page>
</template>
