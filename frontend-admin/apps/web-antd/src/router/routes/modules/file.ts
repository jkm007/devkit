import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:folder',
      order: 9996,
      title: $t('file.title'),
    },
    name: 'File',
    path: '/file',
    children: [
      {
        path: '/file/list',
        name: 'FileList',
        meta: {
          icon: 'lucide:files',
          title: $t('file.list.title'),
        },
        component: () => import('#/views/file/list/index.vue'),
      },
      {
        path: '/file/share',
        name: 'FileShare',
        meta: {
          icon: 'lucide:share-2',
          title: '分享管理',
        },
        component: () => import('#/views/file/share/index.vue'),
      },
      {
        name: 'FileRecycle',
        path: '/file/recycle',
        component: () => import('#/views/file/recycle/index.vue'),
        meta: {
          title: '回收站',
          icon: 'lucide:trash-2',
        },
      },
    ],
  },
];

export default routes;
