import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridColumns } from '#/adapter/vxe-table';

export const SHARE_TYPE_OPTIONS = [
  { label: '链接分享', value: 'link' },
  { label: '分组分享', value: 'group' },
  { label: '用户分享', value: 'user' },
];

export const SHARE_STATUS_OPTIONS = [
  { label: '有效', value: 1, color: 'success' },
  { label: '过期', value: 2, color: 'warning' },
  { label: '禁用', value: 3, color: 'error' },
];

export function useShareFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: SHARE_TYPE_OPTIONS,
      },
      fieldName: 'shareType',
      label: '分享类型',
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: SHARE_STATUS_OPTIONS,
      },
      fieldName: 'status',
      label: '状态',
    },
  ];
}

export function useShareColumns(): VxeTableGridColumns {
  return [
    { field: 'id', title: 'ID', width: 80 },
    { field: 'questionId', title: '题目ID', width: 80 },
    { field: 'shareCode', title: '分享码', width: 200 },
    {
      field: 'shareType',
      title: '分享类型',
      width: 100,
      formatter({ cellValue }: any) {
        const found = SHARE_TYPE_OPTIONS.find((o) => o.value === cellValue);
        return found?.label || cellValue;
      },
    },
    {
      field: 'status',
      title: '状态',
      width: 80,
      formatter({ cellValue }: any) {
        const found = SHARE_STATUS_OPTIONS.find((o) => o.value === cellValue);
        return found?.label || '有效';
      },
    },
    { field: 'accessCount', title: '访问次数', width: 80 },
    { field: 'maxAccess', title: '最大访问', width: 80 },
    { field: 'expireAt', title: '过期时间', width: 180 },
    { field: 'createTime', title: '创建时间', width: 180 },
    {
      align: 'center',
      field: 'operation',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 200,
    },
  ];
}
