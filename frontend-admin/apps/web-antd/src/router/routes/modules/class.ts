import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'mdi:account-group',
      order: 9997,
      title: '班级管理',
    },
    name: 'Class',
    path: '/class',
    children: [
      {
        path: '/class/list',
        name: 'ClassList',
        meta: {
          icon: 'mdi:account-group-outline',
          title: '班级列表',
        },
        component: () => import('#/views/class/list/index.vue'),
      },
    ],
  },
];

export default routes;
