<script lang="ts" setup>
import type {
  WorkbenchProjectItem,
  WorkbenchQuickNavItem,
  WorkbenchTodoItem,
  WorkbenchTrendItem,
} from '@vben/common-ui';

import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

import {
  AnalysisChartCard,
  Page,
  WorkbenchHeader,
  WorkbenchProject,
  WorkbenchQuickNav,
  WorkbenchTodo,
  WorkbenchTrends,
} from '@vben/common-ui';
import { preferences } from '@vben/preferences';
import { useUserStore } from '@vben/stores';
import { openWindow } from '@vben/utils';

import { Button, Empty } from 'ant-design-vue';

import AnalyticsVisitsSource from '../analytics/analytics-visits-source.vue';

const userStore = useUserStore();

// 根据时间获取问候语
const greeting = computed(() => {
  const hour = new Date().getHours();
  if (hour >= 5 && hour < 12) return '早上好';
  if (hour >= 12 && hour < 14) return '中午好';
  if (hour >= 14 && hour < 18) return '下午好';
  return '晚上好';
});

// 项目数据 - 从后端接口获取，或根据实际业务填充
const projectItems: WorkbenchProjectItem[] = [];

// 快捷导航 - 使用项目中实际存在的路由
const quickNavItems: WorkbenchQuickNavItem[] = [
  {
    color: '#1fdaca',
    icon: 'ion:home-outline',
    title: '首页',
    url: '/',
  },
  {
    color: '#bf0c2c',
    icon: 'ion:grid-outline',
    title: '仪表盘',
    url: '/dashboard',
  },
  {
    color: '#3fb27f',
    icon: 'ion:settings-outline',
    title: '系统管理',
    url: '/system/user',
  },
  {
    color: '#4daf1bc9',
    icon: 'ion:key-outline',
    title: '用户管理',
    url: '/system/user',
  },
  {
    color: '#e18525',
    icon: 'ion:folder-outline',
    title: '文件管理',
    url: '/file/list',
  },
  {
    color: '#00d8ff',
    icon: 'ion:bar-chart-outline',
    title: '数据统计',
    url: '/analytics',
  },
];

// 待办事项 - 从后端接口获取，或根据实际业务填充
const todoItems = ref<WorkbenchTodoItem[]>([]);

// 最新动态 - 从后端接口获取，或根据实际业务填充
const trendItems: WorkbenchTrendItem[] = [];

const router = useRouter();

// 导航处理 - 外部链接新窗口打开，内部路由使用 router 跳转
function navTo(nav: WorkbenchProjectItem | WorkbenchQuickNavItem) {
  if (nav.url?.startsWith('http')) {
    openWindow(nav.url);
    return;
  }
  if (nav.url?.startsWith('/')) {
    router.push(nav.url).catch((error) => {
      console.error('Navigation failed:', error);
    });
  } else {
    console.warn(`Unknown URL for navigation item: ${nav.title} -> ${nav.url}`);
  }
}
</script>

<template>
  <Page :auto-content-height="false">
    <div class="p-5">
      <WorkbenchHeader
        :avatar="userStore.userInfo?.avatar || preferences.app.defaultAvatar"
      >
        <template #title>
          {{ greeting }}, {{ userStore.userInfo?.realName }}, 开始您一天的工作吧！
        </template>
        <template #description> 今日晴，20℃ - 32℃！ </template>
      </WorkbenchHeader>

      <div class="flex flex-col lg:flex-row">
        <div class="mr-4 w-full lg:w-3/5">
          <WorkbenchProject
            v-if="projectItems.length > 0"
            :items="projectItems"
            title="项目"
            @click="navTo"
          />
          <div v-else class="mb-5 rounded-lg bg-white p-6 dark:bg-[#1e1e2e]">
            <h3 class="mb-4 text-lg font-medium">项目</h3>
            <Empty
              :image="Empty.PRESENTED_IMAGE_SIMPLE"
              description="暂无项目"
            >
              <Button type="primary">立即创建</Button>
            </Empty>
          </div>
          <WorkbenchTrends :items="trendItems" class="mt-5" title="最新动态" />
        </div>
        <div class="w-full lg:w-2/5">
          <WorkbenchQuickNav
            :items="quickNavItems"
            class="lg:mt-0"
            title="快捷导航"
            @click="navTo"
          />
          <WorkbenchTodo
            v-if="todoItems.length > 0"
            :items="todoItems"
            class="mt-5"
            title="待办事项"
          />
          <div v-else class="mt-5 rounded-lg bg-white p-6 dark:bg-[#1e1e2e]">
            <h3 class="mb-4 text-lg font-medium">待办事项</h3>
            <Empty
              :image="Empty.PRESENTED_IMAGE_SIMPLE"
              description="暂无待办事项"
            >
              <Button type="primary">添加待办</Button>
            </Empty>
          </div>
          <AnalysisChartCard class="mt-5" title="访问来源">
            <AnalyticsVisitsSource />
          </AnalysisChartCard>
        </div>
      </div>
    </div>
  </Page>
</template>
