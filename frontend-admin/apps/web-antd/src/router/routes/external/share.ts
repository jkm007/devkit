import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/share/:code',
    name: 'SharePage',
    meta: {
      title: '文件分享',
      hideMenu: true,
      hideInTab: true,
      hideInBreadcrumb: true,
      ignoreAccess: true, // 无需登录即可访问
    },
    component: () => import('#/views/share/index.vue'),
  },
];

export default routes;