import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'mdi:shield-lock',
      order: 9994,
      title: '安全管理',
    },
    name: 'Security',
    path: '/security',
    children: [
      {
        path: '/security/real-name',
        name: 'SecurityRealName',
        meta: {
          icon: 'lucide:badge-check',
          title: '实名审核',
        },
        component: () => import('#/views/system/real-name/list.vue'),
      },
      {
        path: '/security/security-log',
        name: 'SecurityLog',
        meta: {
          icon: 'lucide:shield-check',
          title: '安全日志',
        },
        component: () => import('#/views/system/security-log/list.vue'),
      },
      {
        path: '/security/risk',
        name: 'SecurityRisk',
        meta: {
          icon: 'lucide:shield-alert',
          title: '风险评分监控',
        },
        component: () => import('#/views/system/risk/index.vue'),
      },
    ],
  },
];

export default routes;
