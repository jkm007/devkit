import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridColumns } from '#/adapter/vxe-table';

export function useAuditFormSchema(): VbenFormSchema[] {
  return [
    { component: 'Input', fieldName: 'title', label: '标题' },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: '待审核', value: 'pending' },
          { label: '已通过', value: 'published' },
          { label: '已驳回', value: 'rejected' },
        ],
      },
      defaultValue: 'pending',
      fieldName: 'status',
      label: '状态',
    },
  ];
}

export function useAuditColumns(): VxeTableGridColumns {
  return [
    { field: 'id', title: 'ID', width: 80 },
    { field: 'title', title: '题目标题', minWidth: 250 },
    {
      field: 'questionType',
      title: '题型',
      width: 120,
    },
    {
      field: 'status',
      title: '状态',
      width: 100,
      formatter({ cellValue }: any) {
        const map: Record<string, string> = {
          pending: '待审核',
          published: '已通过',
          rejected: '已驳回',
        };
        return map[cellValue] || cellValue;
      },
    },
    { field: 'rejectReason', title: '驳回原因', minWidth: 150 },
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
