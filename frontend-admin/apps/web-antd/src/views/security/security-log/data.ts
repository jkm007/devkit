import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridColumns } from '#/adapter/vxe-table';
import type { SystemSecurityLogApi } from '#/api';

import { h } from 'vue';

import { Tag } from 'ant-design-vue';

import { $t } from '#/locales';

const eventTypeOptions = [
  { label: $t('account.eventType.login'), value: 'login' },
  { label: $t('account.eventType.logout'), value: 'logout' },
  { label: $t('account.eventType.login_fail'), value: 'login_fail' },
  { label: $t('account.eventType.password_change'), value: 'password_change' },
  { label: $t('account.eventType.password_reset'), value: 'password_reset' },
  { label: $t('account.eventType.bind_oauth'), value: 'bind_oauth' },
  { label: $t('account.eventType.unbind_oauth'), value: 'unbind_oauth' },
  { label: $t('account.eventType.account_lock'), value: 'account_lock' },
  { label: $t('account.eventType.account_deactivate'), value: 'account_deactivate' },
  { label: $t('account.eventType.user_create'), value: 'user_create' },
  { label: $t('account.eventType.user_update'), value: 'user_update' },
  { label: $t('account.eventType.user_delete'), value: 'user_delete' },
  { label: $t('account.eventType.role_create'), value: 'role_create' },
  { label: $t('account.eventType.role_update'), value: 'role_update' },
  { label: $t('account.eventType.role_delete'), value: 'role_delete' },
  { label: $t('account.eventType.menu_create'), value: 'menu_create' },
  { label: $t('account.eventType.menu_update'), value: 'menu_update' },
  { label: $t('account.eventType.menu_delete'), value: 'menu_delete' },
  { label: $t('account.eventType.group_create'), value: 'group_create' },
  { label: $t('account.eventType.group_update'), value: 'group_update' },
  { label: $t('account.eventType.group_delete'), value: 'group_delete' },
];

const eventTypeMap: Record<string, string> = Object.fromEntries(
  eventTypeOptions.map((o) => [o.value, o.label]),
);

export function useGridFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'userId',
      label: $t('system.securityLog.userId'),
    },
    {
      component: 'Input',
      fieldName: 'username',
      label: $t('system.securityLog.username'),
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: eventTypeOptions,
      },
      fieldName: 'eventType',
      label: $t('system.securityLog.eventType'),
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: $t('system.securityLog.statusSuccess'), value: 1 },
          { label: $t('system.securityLog.statusFail'), value: 0 },
        ],
      },
      fieldName: 'status',
      label: $t('system.securityLog.status'),
    },
    {
      component: 'Input',
      fieldName: 'ip',
      label: $t('system.securityLog.ip'),
    },
    {
      component: 'RangePicker',
      fieldName: 'createTime',
      label: $t('system.securityLog.createTime'),
    },
  ];
}

export function useColumns(): VxeTableGridColumns {
  return [
    {
      field: 'id',
      title: 'ID',
      width: 80,
    },
    {
      field: 'userId',
      title: $t('system.securityLog.userId'),
      width: 100,
    },
    {
      field: 'username',
      title: $t('system.securityLog.username'),
      width: 120,
    },
    {
      field: 'eventType',
      title: $t('system.securityLog.eventType'),
      width: 120,
      slots: {
        default: ({ row }: { row: SystemSecurityLogApi.SecurityLog }) => {
          return eventTypeMap[row.eventType] || row.eventType;
        },
      },
    },
    {
      field: 'eventDetail',
      title: $t('system.securityLog.eventDetail'),
      minWidth: 150,
    },
    {
      field: 'ip',
      title: $t('system.securityLog.ip'),
      width: 140,
    },
    {
      field: 'status',
      title: $t('system.securityLog.status'),
      width: 80,
      slots: {
        default: ({ row }: { row: SystemSecurityLogApi.SecurityLog }) => {
          const isSuccess = row.status === 1;
          return h(
            Tag,
            { color: isSuccess ? 'success' : 'error' },
            () =>
              isSuccess
                ? $t('system.securityLog.statusSuccess')
                : $t('system.securityLog.statusFail'),
          );
        },
      },
    },
    {
      field: 'createdAt',
      title: $t('system.securityLog.createTime'),
      width: 180,
    },
  ];
}
