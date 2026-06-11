import type { DescriptionsItemType } from '@vben/common-ui';

import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridColumns } from '#/adapter/vxe-table';
import type { SystemUserApi } from '#/api';

import { h } from 'vue';

import { Tag } from 'ant-design-vue';
import { z } from '#/adapter/form';

import { getGroupList, getRoleList } from '#/api';
import { $t } from '#/locales';

export function useFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('system.user.name'),
      rules: 'required',
    },
    {
      component: 'Input',
      fieldName: 'nickname',
      label: $t('system.user.nickname'),
    },
    {
      component: 'Input',
      fieldName: 'email',
      label: $t('system.user.email'),
      rules: z
        .string()
        .email($t('ui.formRules.invalidEmail'))
        .or(z.literal(''))
        .optional(),
    },
    {
      component: 'Input',
      fieldName: 'phone',
      label: $t('system.user.phone'),
      rules: z
        .string()
        .regex(/^1[3-9]\d{9}$/, $t('ui.formRules.invalidPhone'))
        .or(z.literal(''))
        .optional(),
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: $t('system.user.genderUnknown'), value: 0 },
          { label: $t('system.user.genderMale'), value: 1 },
          { label: $t('system.user.genderFemale'), value: 2 },
        ],
        optionType: 'button',
      },
      defaultValue: 0,
      fieldName: 'gender',
      label: $t('system.user.gender'),
    },
    {
      component: 'ApiTreeSelect',
      componentProps: {
        allowClear: true,
        api: getGroupList,
        class: 'w-full',
        labelField: 'name',
        valueField: 'id',
        childrenField: 'children',
      },
      fieldName: 'groupId',
      label: $t('system.user.group'),
      rules: 'required',
    },
    {
      component: 'ApiSelect',
      componentProps: {
        allowClear: true,
        api: async () => {
          const res = await getRoleList({ page: 1, pageSize: 100 });
          return res || [];
        },
        class: 'w-full',
        labelField: 'name',
        valueField: 'id',
        mode: 'multiple',
      },
      fieldName: 'roleIds',
      label: $t('system.user.roles'),
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
      label: $t('system.user.status'),
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: $t('system.user.remark'),
    },
  ];
}

export function useGridFormSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('system.user.name'),
    },
    { component: 'Input', fieldName: 'id', label: $t('system.user.id') },
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
      label: $t('system.user.status'),
    },
    {
      component: 'Input',
      fieldName: 'remark',
      label: $t('system.user.remark'),
    },
    {
      component: 'RangePicker',
      fieldName: 'createTime',
      label: $t('system.user.createTime'),
    },
  ];
}

/**
 * 用户详情描述列表项
 * @param row 用户数据
 */
export function useDescriptionItems(
  row?: SystemUserApi.SystemUser,
): DescriptionsItemType[] {
  const enabled = row?.status === 1;
  const genderMap: Record<number, string> = {
    0: $t('system.user.genderUnknown'),
    1: $t('system.user.genderMale'),
    2: $t('system.user.genderFemale'),
  };
  return [
    { label: $t('system.user.name'), content: row?.name },
    { label: $t('system.user.nickname'), content: row?.nickname },
    { label: $t('system.user.id'), content: row?.id },
    { label: $t('system.user.email'), content: row?.email },
    { label: $t('system.user.phone'), content: row?.phone },
    { label: $t('system.user.gender'), content: genderMap[row?.gender ?? 0] },
    { label: $t('system.user.group'), content: row?.groupId },
    {
      label: $t('system.user.status'),
      content: () =>
        h(
          Tag,
          {
            color: enabled ? 'success' : 'error',
          },
          {
            default: () =>
              enabled ? $t('common.enabled') : $t('common.disabled'),
          },
        ),
    },
    { label: $t('system.user.createTime'), content: row?.createTime },
    { label: $t('system.user.remark'), content: row?.remark },
  ];
}

export function useColumns<T = SystemUserApi.SystemUser>(
  onStatusChange?: (newStatus: any, row: T) => PromiseLike<boolean | undefined>,
): VxeTableGridColumns {
  return [
    {
      field: 'name',
      title: $t('system.user.name'),
      width: 120,
    },
    {
      field: 'nickname',
      title: $t('system.user.nickname'),
      width: 120,
    },
    {
      field: 'id',
      title: $t('system.user.id'),
      width: 80,
    },
    {
      field: 'email',
      title: $t('system.user.email'),
      minWidth: 160,
    },
    {
      field: 'phone',
      title: $t('system.user.phone'),
      width: 130,
    },
    {
      field: 'gender',
      title: $t('system.user.gender'),
      width: 80,
      slots: {
        default: ({ row }: { row: SystemUserApi.SystemUser }) => {
          const map: Record<number, string> = {
            0: $t('system.user.genderUnknown'),
            1: $t('system.user.genderMale'),
            2: $t('system.user.genderFemale'),
          };
          return (map[row.gender] || map[0]) as string;
        },
      },
    },
    {
      cellRender: {
        attrs: { beforeChange: onStatusChange },
        name: onStatusChange ? 'CellSwitch' : 'CellTag',
      },
      field: 'status',
      title: $t('system.user.status'),
      width: 100,
    },
    {
      field: 'remark',
      minWidth: 100,
      title: $t('system.user.remark'),
    },
    {
      field: 'createTime',
      title: $t('system.user.createTime'),
      width: 180,
    },
    {
      align: 'center',
      field: 'operation',
      fixed: 'right',
      slots: { default: 'action' },
      title: $t('system.user.operation'),
      width: 180,
    },
  ];
}
