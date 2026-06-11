import type { VxeTableGridColumns } from '@vben/plugins/vxe-table';

import type { VbenFormSchema } from '#/adapter/form';
import type { SystemGroupApi } from '#/api/system/group';

import { z } from '#/adapter/form';
import { getGroupList } from '#/api/system/group';
import { getRoleList } from '#/api/system/role';
import { $t } from '#/locales';

/**
 * 获取编辑表单的字段配置。如果没有使用多语言，可以直接export一个数组常量
 */
export function useSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('system.group.groupName'),
      rules: z
        .string()
        .min(2, $t('ui.formRules.minLength', [$t('system.group.groupName'), 2]))
        .max(
          20,
          $t('ui.formRules.maxLength', [$t('system.group.groupName'), 20]),
        ),
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
      fieldName: 'pid',
      label: $t('system.group.parentGroup'),
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
      label: $t('system.group.roles'),
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
      label: $t('system.group.status'),
    },
    {
      component: 'Textarea',
      componentProps: {
        maxLength: 50,
        rows: 3,
        showCount: true,
      },
      fieldName: 'remark',
      label: $t('system.group.remark'),
      rules: z
        .string()
        .max(50, $t('ui.formRules.maxLength', [$t('system.group.remark'), 50]))
        .optional(),
    },
  ];
}

/**
 * 获取表格列配置
 */
export function useColumns(): VxeTableGridColumns<SystemGroupApi.SystemGroup> {
  return [
    {
      align: 'left',
      field: 'name',
      fixed: 'left',
      title: $t('system.group.groupName'),
      treeNode: true,
      width: 150,
    },
    {
      cellRender: { name: 'CellTag' },
      field: 'status',
      title: $t('system.group.status'),
      width: 100,
    },
    {
      field: 'createTime',
      title: $t('system.group.createTime'),
      width: 180,
    },
    {
      field: 'remark',
      title: $t('system.group.remark'),
    },
    {
      align: 'center',
      field: 'operation',
      fixed: 'right',
      slots: { default: 'action' },
      title: $t('system.group.operation'),
      width: 200,
    },
  ];
}
