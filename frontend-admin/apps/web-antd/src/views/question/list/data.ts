import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridColumns } from '#/adapter/vxe-table';

export const QUESTION_TYPE_OPTIONS = [
  { label: '单选题', value: 'single_choice' },
  { label: '多选题', value: 'multiple_choice' },
  { label: '不定项选择题', value: 'indefinite_choice' },
  { label: '判断题', value: 'true_false' },
  { label: '填空题', value: 'fill_blank' },
  { label: '完形填空', value: 'cloze' },
  { label: '名词解释', value: 'term_explanation' },
  { label: '简答题', value: 'short_answer' },
  { label: '论述题', value: 'essay_question' },
  { label: '作文题', value: 'composition' },
  { label: '材料题', value: 'material' },
  { label: '案例分析题', value: 'case_analysis' },
  { label: '阅读理解题', value: 'reading' },
  { label: '匹配题', value: 'matching' },
  { label: '排序题', value: 'ordering' },
  { label: '分类题', value: 'classification' },
  { label: '听力题', value: 'listening' },
  { label: '口语题', value: 'speaking' },
  { label: '视频题', value: 'video' },
  { label: '文档题', value: 'document' },
  { label: '计算题', value: 'calculation' },
  { label: '证明题', value: 'proof' },
  { label: '操作题', value: 'operation' },
  { label: '编程题', value: 'programming' },
  { label: 'SQL题', value: 'sql' },
  { label: '代码阅读题', value: 'code_reading' },
  { label: '调试改错题', value: 'debugging' },
];

export const STATUS_OPTIONS = [
  { label: '草稿', value: 'draft', color: 'default' },
  { label: '待审核', value: 'pending', color: 'processing' },
  { label: '已驳回', value: 'rejected', color: 'error' },
  { label: '已发布', value: 'published', color: 'success' },
  { label: '已下架', value: 'archived', color: 'warning' },
];

export const DIFFICULTY_OPTIONS = [
  { label: '简单', value: 1, color: 'green' },
  { label: '中等', value: 2, color: 'orange' },
  { label: '困难', value: 3, color: 'red' },
];

export const RESOURCE_TYPE_OPTIONS = [
  { label: '公共', value: 'public' },
  { label: '私有', value: 'private' },
  { label: '分组', value: 'group' },
  { label: '指定用户', value: 'user' },
];

export function useQuestionFormSchema(): VbenFormSchema[] {
  return [
    { component: 'Input', fieldName: 'title', label: '标题' },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: QUESTION_TYPE_OPTIONS,
      },
      fieldName: 'questionType',
      label: '题型',
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: STATUS_OPTIONS,
      },
      fieldName: 'status',
      label: '状态',
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: DIFFICULTY_OPTIONS,
      },
      fieldName: 'difficulty',
      label: '难度',
    },
  ];
}

export function useQuestionColumns(): VxeTableGridColumns {
  return [
    { field: 'id', title: 'ID', width: 80 },
    { field: 'title', title: '题目标题', minWidth: 250 },
    {
      field: 'questionType',
      title: '题型',
      width: 120,
      formatter({ cellValue }: any) {
        const found = QUESTION_TYPE_OPTIONS.find((o) => o.value === cellValue);
        return found?.label || cellValue;
      },
    },
    {
      field: 'difficulty',
      title: '难度',
      width: 80,
      cellRender: { name: 'CellTag', options: DIFFICULTY_OPTIONS },
    },
    {
      field: 'status',
      title: '状态',
      width: 100,
      cellRender: { name: 'CellTag', options: STATUS_OPTIONS },
    },
    {
      field: 'resourceType',
      title: '资源类型',
      width: 100,
      formatter({ cellValue }: any) {
        const found = RESOURCE_TYPE_OPTIONS.find(
          (o) => o.value === cellValue,
        );
        return found?.label || cellValue;
      },
    },
    { field: 'createdBy', title: '创建人', width: 100 },
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
