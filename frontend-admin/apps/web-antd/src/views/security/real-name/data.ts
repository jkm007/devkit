import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridColumns } from '#/adapter/vxe-table';
import type { SystemRealNameApi } from '#/api';

import { h } from 'vue';

import { Tag } from 'ant-design-vue';

import { $t } from '#/locales';

export function useGridFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'realName',
      label: $t('system.realName.realName'),
    },
    {
      component: 'Input',
      fieldName: 'userId',
      label: $t('system.realName.userId'),
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: $t('system.realName.statusPending'), value: 0 },
          { label: $t('system.realName.statusApproved'), value: 1 },
          { label: $t('system.realName.statusRejected'), value: 2 },
        ],
      },
      fieldName: 'status',
      label: $t('system.realName.status'),
    },
    {
      component: 'RangePicker',
      fieldName: 'createTime',
      label: $t('system.realName.submittedAt'),
    },
  ];
}

const statusMap: Record<number, { color: string; key: string }> = {
  0: { color: 'processing', key: 'statusPending' },
  1: { color: 'success', key: 'statusApproved' },
  2: { color: 'error', key: 'statusRejected' },
};

export function useColumns(): VxeTableGridColumns {
  return [
    {
      field: 'id',
      title: 'ID',
      width: 80,
    },
    {
      field: 'userId',
      title: $t('system.realName.userId'),
      width: 100,
    },
    {
      field: 'username',
      title: $t('system.realName.username'),
      width: 120,
    },
    {
      field: 'realName',
      title: $t('system.realName.realName'),
      width: 120,
    },
    {
      field: 'idCard',
      title: $t('system.realName.idCard'),
      minWidth: 180,
    },
    {
      field: 'status',
      title: $t('system.realName.status'),
      width: 100,
      slots: {
        default: ({ row }: { row: SystemRealNameApi.RealNameApplication }) => {
          const fallback = { color: 'default', key: 'statusPending' };
          const s = statusMap[row.status] ?? fallback;
          return h(Tag, { color: s.color }, () => $t(`system.realName.${s.key}`));
        },
      },
    },
    {
      field: 'submittedAt',
      title: $t('system.realName.submittedAt'),
      width: 180,
    },
    {
      align: 'center',
      field: 'operation',
      fixed: 'right',
      slots: { default: 'action' },
      title: $t('system.realName.operation'),
      width: 180,
    },
  ];
}
