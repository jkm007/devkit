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
    ],
  },
];

export default routes;