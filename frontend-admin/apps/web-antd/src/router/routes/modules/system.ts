import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'ion:settings-outline',
      order: 9997,
      title: $t('system.title'),
    },
    name: 'System',
    path: '/system',
    children: [
      {
        path: '/system/user',
        name: 'SystemUser',
        meta: {
          icon: 'mdi:user',
          title: $t('system.user.title'),
        },
        component: () => import('#/views/system/user/list.vue'),
      },
      {
        path: '/system/role',
        name: 'SystemRole',
        meta: {
          icon: 'mdi:account-group',
          title: $t('system.role.title'),
        },
        component: () => import('#/views/system/role/list.vue'),
      },
      {
        path: '/system/menu',
        name: 'SystemMenu',
        meta: {
          icon: 'mdi:menu',
          title: $t('system.menu.title'),
        },
        component: () => import('#/views/system/menu/list.vue'),
      },
      {
        path: '/system/group',
        name: 'SystemGroup',
        meta: {
          icon: 'charm:organisation',
          title: $t('system.group.title'),
        },
        component: () => import('#/views/system/group/list.vue'),
      },
      {
        path: '/system/settings',
        name: 'SystemSettings',
        meta: {
          icon: 'lucide:sliders-horizontal',
          title: $t('system.settings.title'),
        },
        component: () => import('#/views/system/settings/index.vue'),
      },
      {
        path: '/system/tag',
        name: 'SystemTag',
        meta: {
          icon: 'mdi:tag-multiple',
          title: '标签路由',
        },
        component: () => import('#/views/system/tag/index.vue'),
      },
      {
        path: '/system/role-application',
        name: 'SystemRoleApplication',
        meta: {
          icon: 'lucide:clipboard-check',
          title: '角色申请',
        },
        component: () => import('#/views/system/role-application/list.vue'),
      },
      {
        name: 'SystemScheduledTask',
        path: '/system/scheduled-task',
        component: () => import('#/views/system/scheduled-task/index.vue'),
        meta: {
          title: '定时任务',
          icon: 'lucide:clock',
        },
      },
      {
        name: 'SystemAnnouncement',
        path: '/system/announcement',
        component: () => import('#/views/system/announcement/index.vue'),
        meta: {
          title: '公告管理',
          icon: 'lucide:megaphone',
        },
      },
    ],
  },
];

export default routes;
