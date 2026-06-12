import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridColumns } from '#/adapter/vxe-table';
import type { SystemRoleApi } from '#/api';

import { $t } from '#/locales';

export function useFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('system.role.roleName'),
      rules: 'required',
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: $t('common.enabled'), value: 1 },
          { label: $t('common.disabled'), value: 0 },
        ],
        optionType: 'button',
      },
      defaultValue: 1,
      fieldName: 'status',
      label: $t('system.role.status'),
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: '是', value: 1 },
          { label: '否', value: 0 },
        ],
        optionType: 'button',
      },
      defaultValue: 0,
      fieldName: 'allowApply',
      label: $t('system.role.allowApply'),
    },
    {
      component: 'InputNumber',
      fieldName: 'storageQuota',
      label: $t('system.role.storageQuota'),
      componentProps: {
        min: 0,
        step: 1024,
        placeholder: '0 = 不限制',
        class: 'w-full',
        addonAfter: 'MB',
      },
      help: '设置为 0 表示不限制存储空间',
      defaultValue: 0,
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: $t('system.role.remark'),
    },
    {
      component: 'Input',
      fieldName: 'permissions',
      formItemClass: 'items-start',
      label: $t('system.role.setPermissions'),
      modelPropName: 'modelValue',
    },
  ];
}

export function useGridFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('system.role.roleName'),
    },
    { component: 'Input', fieldName: 'id', label: $t('system.role.id') },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: $t('common.enabled'), value: 1 },
          { label: $t('common.disabled'), value: 0 },
        ],
      },
      fieldName: 'status',
      label: $t('system.role.status'),
    },
    {
      component: 'Input',
      fieldName: 'remark',
      label: $t('system.role.remark'),
    },
    {
      component: 'RangePicker',
      fieldName: 'createTime',
      label: $t('system.role.createTime'),
    },
  ];
}

export function useColumns<T = SystemRoleApi.SystemRole>(
  onStatusChange?: (newStatus: any, row: T) => PromiseLike<boolean | undefined>,
): VxeTableGridColumns {
  return [
    {
      field: 'name',
      title: $t('system.role.roleName'),
      width: 200,
    },
    {
      field: 'id',
      title: $t('system.role.id'),
      width: 200,
    },
    {
      cellRender: {
        attrs: { beforeChange: onStatusChange },
        name: onStatusChange ? 'CellSwitch' : 'CellTag',
      },
      field: 'status',
      title: $t('system.role.status'),
      width: 100,
    },
    {
      field: 'allowApply',
      title: $t('system.role.allowApply'),
      width: 120,
      formatter({ cellValue }: any) {
        return cellValue === 1 ? '是' : '否';
      },
    },
    {
      field: 'storageQuota',
      title: $t('system.role.storageQuota'),
      width: 130,
      formatter({ cellValue }: any) {
        if (!cellValue || cellValue <= 0) return '不限';
        if (cellValue >= 1073741824) return `${(cellValue / 1073741824).toFixed(1)} GB`;
        if (cellValue >= 1048576) return `${(cellValue / 1048576).toFixed(0)} MB`;
        return `${(cellValue / 1024).toFixed(0)} KB`;
      },
    },
    {
      field: 'remark',
      minWidth: 100,
      title: $t('system.role.remark'),
    },
    {
      field: 'createTime',
      title: $t('system.role.createTime'),
      width: 200,
    },
    {
      align: 'center',
      field: 'operation',
      fixed: 'right',
      slots: { default: 'action' },
      title: $t('system.role.operation'),
      width: 180,
    },
  ];
}
