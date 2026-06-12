import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'mdi:book-open-variant',
      order: 9998,
      title: '题库管理',
    },
    name: 'Question',
    path: '/question',
    children: [
      {
        path: '/question/list',
        name: 'QuestionList',
        meta: {
          icon: 'mdi:file-question-outline',
          title: '题目管理',
        },
        component: () => import('#/views/question/list/index.vue'),
      },
      {
        path: '/question/import',
        name: 'QuestionImport',
        meta: {
          icon: 'mdi:import',
          title: '题目导入',
        },
        component: () => import('#/views/question/import/index.vue'),
      },
      {
        path: '/question/category',
        name: 'QuestionCategory',
        meta: {
          icon: 'mdi:folder-tree',
          title: '分类科目',
        },
        component: () => import('#/views/question/category/index.vue'),
      },
      {
        path: '/question/knowledge',
        name: 'QuestionKnowledge',
        meta: {
          icon: 'mdi:lightbulb-outline',
          title: '知识考点',
        },
        component: () => import('#/views/question/knowledge/index.vue'),
      },
      {
        path: '/question/source',
        name: 'QuestionSource',
        meta: {
          icon: 'mdi:source-branch',
          title: '来源标签',
        },
        component: () => import('#/views/question/source/index.vue'),
      },
      {
        path: '/question/audit',
        name: 'QuestionAudit',
        meta: {
          icon: 'mdi:check-decagram',
          title: '审核发布',
        },
        component: () => import('#/views/question/audit/index.vue'),
      },
      {
        path: '/question/statistics',
        name: 'QuestionStatistics',
        meta: {
          icon: 'mdi:chart-bar',
          title: '题库统计',
        },
        component: () => import('#/views/question/statistics/index.vue'),
      },
    ],
  },
];

export default routes;
