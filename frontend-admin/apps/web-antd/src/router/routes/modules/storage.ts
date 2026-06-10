import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
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
      {
        name: 'StorageSettings',
        path: '/storage/storage-settings',
        component: () => import('#/views/storage/storage-settings/index.vue'),
        meta: {
          title: '存储设置',
          icon: 'lucide:settings',
        },
      },
      {
        name: 'StorageManage',
        path: '/storage/storage-manage',
        component: () => import('#/views/storage/storage-manage/index.vue'),
        meta: {
          title: '存储管理',
          icon: 'lucide:database',
        },
      },
      {
        name: 'StorageTagRouting',
        path: '/storage/tag-routing',
        component: () => import('#/views/storage/tag-routing/index.vue'),
        meta: {
          title: '标签路由',
          icon: 'lucide:git-branch',
        },
      },
    ],
  },
];

export default routes;
