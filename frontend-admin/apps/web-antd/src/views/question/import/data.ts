import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridColumns } from '#/adapter/vxe-table';

export const IMPORT_STATUS_OPTIONS = [
  { label: '已上传', value: 'uploaded', color: 'default' },
  { label: '解析中', value: 'parsing', color: 'processing' },
  { label: '已解析', value: 'parsed', color: 'success' },
  { label: '部分失败', value: 'partial_failed', color: 'warning' },
  { label: '失败', value: 'failed', color: 'error' },
  { label: '已确认', value: 'confirmed', color: 'success' },
  { label: '已发布', value: 'published', color: 'success' },
];

export function useImportFormSchema(): VbenFormSchema[] {
  return [
    { component: 'Input', fieldName: 'fileName', label: '文件名' },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: IMPORT_STATUS_OPTIONS,
      },
      fieldName: 'status',
      label: '状态',
    },
  ];
}

export function useImportColumns(): VxeTableGridColumns {
  return [
    { field: 'id', title: 'ID', width: 80 },
    { field: 'fileName', title: '文件名', minWidth: 200 },
    {
      field: 'fileType',
      title: '文件类型',
      width: 100,
    },
    {
      field: 'status',
      title: '状态',
      width: 100,
      formatter({ cellValue }: any) {
        const found = IMPORT_STATUS_OPTIONS.find((o) => o.value === cellValue);
        return found?.label || cellValue;
      },
    },
    {
      field: 'totalCount',
      title: '总题数',
      width: 80,
    },
    {
      field: 'successCount',
      title: '成功',
      width: 80,
    },
    {
      field: 'failedCount',
      title: '失败',
      width: 80,
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
