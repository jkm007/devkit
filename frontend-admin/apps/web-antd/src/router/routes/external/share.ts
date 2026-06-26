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
  {
    path: '/share/:code/video/:fileId',
    name: 'ShareVideoPlayer',
    meta: {
      title: '视频播放',
      hideMenu: true,
      hideInTab: true,
      hideInBreadcrumb: true,
      ignoreAccess: true, // 无需登录即可访问
    },
    component: () => import('#/views/share/video-player.vue'),
  },
];

export default routes;
