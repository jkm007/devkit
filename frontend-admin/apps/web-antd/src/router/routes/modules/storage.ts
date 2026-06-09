import type { RouteRecordRaw } from 'vue-router';

import { BasicLayout } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      icon: 'mdi:cloud-sync',
      order: 9995,
      title: '存储管理',
    },
    name: 'Storage',
    path: '/storage',
    children: [
      {
        path: '/storage/storage-bucket',
        name: 'StorageBucket',
        meta: {
          icon: 'mdi:database',
          title: '存储桶管理',
        },
        component: () => import('#/views/storage/storage-bucket/index.vue'),
      },
      {
        path: '/storage/storage-config',
        name: 'StorageConfig',
        meta: {
          icon: 'mdi:server-network',
          title: '存储配置',
        },
        component: () => import('#/views/storage/storage-config/index.vue'),
      },
      {
        path: '/storage/file-type-rule',
        name: 'StorageFileTypeRule',
        meta: {
          icon: 'mdi:file-check',
          title: '文件类型规则',
        },
        component: () => import('#/views/storage/file-type-rule/index.vue'),
      },
    ],
  },
];

export default routes;
