import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridColumns } from '#/adapter/vxe-table';

export const SOURCE_TYPE_OPTIONS = [
  { label: '历年真题', value: 'real_exam' },
  { label: '模拟题', value: 'mock_exam' },
  { label: '章节练习', value: 'chapter_practice' },
  { label: '每日练习', value: 'daily_practice' },
  { label: '人工录入', value: 'manual' },
  { label: '文件导入', value: 'import' },
  { label: 'AI生成', value: 'ai_generated' },
  { label: '用户自建', value: 'user_created' },
];

export function useSourceFormSchema(): VbenFormSchema[] {
  return [
    { component: 'Input', fieldName: 'name', label: '名称' },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: SOURCE_TYPE_OPTIONS,
      },
      fieldName: 'sourceType',
      label: '来源类型',
    },
  ];
}

export function useSourceDrawerSchema(
  examOptions: any[] = [],
): VbenFormSchema[] {
  return [
    {
      component: 'Select',
      fieldName: 'sourceType',
      label: '来源类型',
      rules: 'required',
      componentProps: {
        options: SOURCE_TYPE_OPTIONS,
        placeholder: '请选择来源类型',
        class: 'w-full',
      },
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: '来源名称',
      rules: 'required',
    },
    {
      component: 'Select',
      fieldName: 'examId',
      label: '关联考试',
      componentProps: {
        options: examOptions,
        placeholder: '请选择考试（可选）',
        allowClear: true,
        class: 'w-full',
      },
    },
    {
      component: 'InputNumber',
      fieldName: 'year',
      label: '年份',
      componentProps: { min: 0, class: 'w-full' },
    },
    {
      component: 'Input',
      fieldName: 'region',
      label: '地区',
    },
    {
      component: 'Input',
      fieldName: 'paperName',
      label: '试卷名称',
    },
    {
      component: 'Input',
      fieldName: 'questionNo',
      label: '原题号',
    },
    {
      component: 'Textarea',
      fieldName: 'copyright',
      label: '版权说明',
    },
  ];
}

export function useSourceColumns(): VxeTableGridColumns {
  return [
    { field: 'id', title: 'ID', width: 80 },
    { field: 'name', title: '来源名称', minWidth: 200 },
    {
      field: 'sourceType',
      title: '来源类型',
      width: 120,
      formatter({ cellValue }: any) {
        const found = SOURCE_TYPE_OPTIONS.find((o) => o.value === cellValue);
        return found?.label || cellValue;
      },
    },
    { field: 'year', title: '年份', width: 80 },
    { field: 'region', title: '地区', width: 100 },
    { field: 'paperName', title: '试卷名称', minWidth: 150 },
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
