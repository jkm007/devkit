import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    path: '/account',
    name: 'Account',
    meta: {
      icon: 'lucide:user-circle',
      order: 9998,
      title: $t('account.title'),
      hideInMenu: true,
    },
    children: [
      {
        path: '/account/index',
        name: 'AccountIndex',
        meta: {
          icon: 'lucide:user-circle',
          title: $t('account.title'),
          hideInMenu: true,
        },
        component: () => import('#/views/account/index.vue'),
      },
    ],
  },
];

export default routes;
