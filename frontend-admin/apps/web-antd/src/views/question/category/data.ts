import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridColumns } from '#/adapter/vxe-table';

// ==================== 考试大类 ====================

export function useExamCategoryFormSchema(): VbenFormSchema[] {
  return [
    { component: 'Input', fieldName: 'name', label: '名称' },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 },
        ],
      },
      fieldName: 'status',
      label: '状态',
    },
  ];
}

export function useExamCategoryDrawerSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: '名称',
      rules: 'required',
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '编码',
    },
    {
      component: 'InputNumber',
      fieldName: 'sortOrder',
      label: '排序',
      defaultValue: 0,
      componentProps: { min: 0, class: 'w-full' },
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 },
        ],
        optionType: 'button',
      },
      defaultValue: 1,
      fieldName: 'status',
      label: '状态',
    },
  ];
}

export function useExamCategoryColumns<T = any>(
  onStatusChange?: (newStatus: any, row: T) => PromiseLike<boolean | undefined>,
): VxeTableGridColumns {
  return [
    { field: 'id', title: 'ID', width: 80 },
    { field: 'name', title: '名称', minWidth: 200 },
    { field: 'code', title: '编码', width: 120 },
    { field: 'level', title: '层级', width: 80 },
    { field: 'sortOrder', title: '排序', width: 80 },
    {
      cellRender: {
        attrs: { beforeChange: onStatusChange },
        name: onStatusChange ? 'CellSwitch' : 'CellTag',
      },
      field: 'status',
      title: '状态',
      width: 100,
    },
    { field: 'createTime', title: '创建时间', width: 180 },
    {
      align: 'center',
      field: 'operation',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 150,
    },
  ];
}

// ==================== 具体考试 ====================

export function useExamFormSchema(): VbenFormSchema[] {
  return [
    { component: 'Input', fieldName: 'name', label: '名称' },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 },
        ],
      },
      fieldName: 'status',
      label: '状态',
    },
  ];
}

export function useExamDrawerSchema(
  categoryOptions: any[] = [],
): VbenFormSchema[] {
  return [
    {
      component: 'Select',
      fieldName: 'examCategoryId',
      label: '所属考试大类',
      rules: 'required',
      componentProps: {
        options: categoryOptions,
        placeholder: '请选择考试大类',
        class: 'w-full',
      },
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: '考试名称',
      rules: 'required',
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '编码',
    },
    {
      component: 'Textarea',
      fieldName: 'description',
      label: '描述',
    },
    {
      component: 'InputNumber',
      fieldName: 'sortOrder',
      label: '排序',
      defaultValue: 0,
      componentProps: { min: 0, class: 'w-full' },
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 },
        ],
        optionType: 'button',
      },
      defaultValue: 1,
      fieldName: 'status',
      label: '状态',
    },
  ];
}

export function useExamColumns<T = any>(
  onStatusChange?: (newStatus: any, row: T) => PromiseLike<boolean | undefined>,
): VxeTableGridColumns {
  return [
    { field: 'id', title: 'ID', width: 80 },
    { field: 'name', title: '考试名称', minWidth: 200 },
    { field: 'code', title: '编码', width: 120 },
    { field: 'description', title: '描述', minWidth: 150 },
    { field: 'sortOrder', title: '排序', width: 80 },
    {
      cellRender: {
        attrs: { beforeChange: onStatusChange },
        name: onStatusChange ? 'CellSwitch' : 'CellTag',
      },
      field: 'status',
      title: '状态',
      width: 100,
    },
    { field: 'createTime', title: '创建时间', width: 180 },
    {
      align: 'center',
      field: 'operation',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 150,
    },
  ];
}

// ==================== 科目 ====================

export function useSubjectFormSchema(): VbenFormSchema[] {
  return [
    { component: 'Input', fieldName: 'name', label: '名称' },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 },
        ],
      },
      fieldName: 'status',
      label: '状态',
    },
  ];
}

export function useSubjectDrawerSchema(
  examOptions: any[] = [],
): VbenFormSchema[] {
  return [
    {
      component: 'Select',
      fieldName: 'examId',
      label: '所属考试',
      rules: 'required',
      componentProps: {
        options: examOptions,
        placeholder: '请选择考试',
        class: 'w-full',
      },
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: '科目名称',
      rules: 'required',
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '编码',
    },
    {
      component: 'InputNumber',
      fieldName: 'sortOrder',
      label: '排序',
      defaultValue: 0,
      componentProps: { min: 0, class: 'w-full' },
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 },
        ],
        optionType: 'button',
      },
      defaultValue: 1,
      fieldName: 'status',
      label: '状态',
    },
  ];
}

export function useSubjectColumns<T = any>(
  onStatusChange?: (newStatus: any, row: T) => PromiseLike<boolean | undefined>,
): VxeTableGridColumns {
  return [
    { field: 'id', title: 'ID', width: 80 },
    { field: 'name', title: '科目名称', minWidth: 200 },
    { field: 'code', title: '编码', width: 120 },
    { field: 'sortOrder', title: '排序', width: 80 },
    {
      cellRender: {
        attrs: { beforeChange: onStatusChange },
        name: onStatusChange ? 'CellSwitch' : 'CellTag',
      },
      field: 'status',
      title: '状态',
      width: 100,
    },
    { field: 'createTime', title: '创建时间', width: 180 },
    {
      align: 'center',
      field: 'operation',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 150,
    },
  ];
}

// ==================== 章节分类 ====================

export function useCategoryFormSchema(): VbenFormSchema[] {
  return [
    { component: 'Input', fieldName: 'name', label: '名称' },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 },
        ],
      },
      fieldName: 'status',
      label: '状态',
    },
  ];
}

export function useCategoryDrawerSchema(
  examOptions: any[] = [],
  subjectOptions: any[] = [],
  parentOptions: any[] = [],
): VbenFormSchema[] {
  return [
    {
      component: 'Select',
      fieldName: 'examId',
      label: '所属考试',
      componentProps: {
        options: examOptions,
        placeholder: '请选择考试（可选）',
        allowClear: true,
        class: 'w-full',
      },
    },
    {
      component: 'Select',
      fieldName: 'subjectId',
      label: '所属科目',
      componentProps: {
        options: subjectOptions,
        placeholder: '请选择科目（可选）',
        allowClear: true,
        class: 'w-full',
      },
    },
    {
      component: 'TreeSelect',
      fieldName: 'parentId',
      label: '父级分类',
      componentProps: {
        treeData: parentOptions,
        placeholder: '无（顶级分类）',
        allowClear: true,
        class: 'w-full',
        fieldNames: { children: 'children', label: 'name', value: 'id' },
      },
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: '分类名称',
      rules: 'required',
    },
    {
      component: 'InputNumber',
      fieldName: 'sortOrder',
      label: '排序',
      defaultValue: 0,
      componentProps: { min: 0, class: 'w-full' },
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 },
        ],
        optionType: 'button',
      },
      defaultValue: 1,
      fieldName: 'status',
      label: '状态',
    },
  ];
}

export function useCategoryColumns<T = any>(
  onStatusChange?: (newStatus: any, row: T) => PromiseLike<boolean | undefined>,
): VxeTableGridColumns {
  return [
    { field: 'id', title: 'ID', width: 80 },
    { field: 'name', title: '分类名称', minWidth: 200 },
    { field: 'level', title: '层级', width: 80 },
    { field: 'sortOrder', title: '排序', width: 80 },
    {
      cellRender: {
        attrs: { beforeChange: onStatusChange },
        name: onStatusChange ? 'CellSwitch' : 'CellTag',
      },
      field: 'status',
      title: '状态',
      width: 100,
    },
    { field: 'createTime', title: '创建时间', width: 180 },
    {
      align: 'center',
      field: 'operation',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 150,
    },
  ];
}
