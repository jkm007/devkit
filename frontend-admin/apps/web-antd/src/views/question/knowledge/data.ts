import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridColumns } from '#/adapter/vxe-table';

export function useKnowledgePointFormSchema(): VbenFormSchema[] {
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

export function useKnowledgePointDrawerSchema(
  examOptions: any[] = [],
  subjectOptions: any[] = [],
  categoryOptions: any[] = [],
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
      component: 'Select',
      fieldName: 'categoryId',
      label: '所属分类',
      componentProps: {
        options: categoryOptions,
        placeholder: '请选择分类（可选）',
        allowClear: true,
        class: 'w-full',
      },
    },
    {
      component: 'TreeSelect',
      fieldName: 'parentId',
      label: '父级知识点',
      componentProps: {
        treeData: parentOptions,
        placeholder: '无（顶级知识点）',
        allowClear: true,
        class: 'w-full',
        fieldNames: { children: 'children', label: 'name', value: 'id' },
      },
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: '知识点名称',
      rules: 'required',
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '编码',
    },
    {
      component: 'Slider',
      fieldName: 'importance',
      label: '重要程度',
      defaultValue: 3,
      componentProps: {
        min: 1,
        max: 5,
        marks: { 1: '低', 2: '较低', 3: '中', 4: '较高', 5: '高' },
      },
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

export function useKnowledgePointColumns<T = any>(
  onStatusChange?: (newStatus: any, row: T) => PromiseLike<boolean | undefined>,
): VxeTableGridColumns {
  return [
    { field: 'id', title: 'ID', width: 80 },
    { field: 'name', title: '知识点名称', minWidth: 200 },
    { field: 'code', title: '编码', width: 120 },
    {
      field: 'importance',
      title: '重要程度',
      width: 100,
      formatter({ cellValue }: any) {
        const labels: Record<number, string> = {
          1: '低',
          2: '较低',
          3: '中',
          4: '较高',
          5: '高',
        };
        return labels[cellValue] || '中';
      },
    },
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
