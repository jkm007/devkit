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
        path: '/question/category',
        name: 'QuestionCategory',
        meta: {
          icon: 'mdi:folder-tree',
          title: '分类科目',
        },
        component: () => import('#/views/question/category/index.vue'),
      },
    ],
  },
];

export default routes;
