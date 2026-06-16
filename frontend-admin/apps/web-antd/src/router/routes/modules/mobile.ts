import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'mdi:cellphone',
      order: 6,
      title: '移动端配置',
    },
    name: 'Mobile',
    path: '/mobile',
    redirect: '/mobile/banner',
    children: [
      {
        name: 'MobileBanner',
        path: 'banner',
        component: () => import('#/views/system/banner/list.vue'),
        meta: {
          icon: 'mdi:view-carousel',
          title: '轮播图管理',
        },
      },
      {
        name: 'MobileQuickMenu',
        path: 'quick-menu',
        component: () => import('#/views/mobile/quick-menu/index.vue'),
        meta: {
          icon: 'mdi:view-grid',
          title: '快捷菜单',
        },
      },
      {
        name: 'MobileMyPage',
        path: 'my-page',
        component: () => import('#/views/mobile/my-page/index.vue'),
        meta: {
          icon: 'mdi:account-cog',
          title: '我的页面',
        },
      },
      {
        name: 'MobileSettings',
        path: 'settings',
        component: () => import('#/views/mobile/settings/index.vue'),
        meta: {
          icon: 'mdi:cog',
          title: '移动端设置',
        },
      },
    ],
  },
];

export default routes;
