<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';

import { Page } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { Button, Card, Empty, Input, InputNumber, message, Modal, RadioGroup, Tree } from 'ant-design-vue';

import { useAccess } from '@vben/access';

import CreateModal from './modules/create-modal.vue';

import {
  createExam,
  createExamCategory,
  createQuestionCategory,
  createSubject,
  deleteExam,
  deleteExamCategory,
  deleteQuestionCategory,
  deleteSubject,
  getExamAll,
  getExamCategoryAll,
  getQuestionCategoryAll,
  getSubjectAll,
  updateExam,
  updateExamCategory,
  updateQuestionCategory,
  updateSubject,
} from '#/api/question/category';

const { hasAccessByCodes } = useAccess();

// Create modal ref
const createModalRef = ref();
const createType = ref<'examCategory' | 'exam' | 'subject' | 'category' | null>(null);
const createParentId = ref<number>(0);

// Tree data structure
interface TreeNode {
  key: string;
  title: string;
  type: 'examCategory' | 'exam' | 'subject' | 'category';
  icon?: string;
  children?: TreeNode[];
  data: any; // Original data
  isLeaf?: boolean;
}

const treeData = ref<TreeNode[]>([]);
const selectedNode = ref<TreeNode | null>(null);
const expandedKeys = ref<Array<string | number>>([]);
const selectedKeys = ref<Array<string | number>>([]);

// Loading state
const loading = ref(false);
const saving = ref(false);

// Form data for editing
const formData = ref<any>({});

// Build tree data from API results
async function loadTreeData() {
  loading.value = true;
  try {
    // Load all data in sequence (subjects need exams data)
    const examCategories = await getExamCategoryAll();
    const exams = await getExamAll();
    const categories = await getQuestionCategoryAll();

    // Load all subjects for all exams
    const allSubjects: any[] = [];
    for (const exam of exams || []) {
      const subjects = await getSubjectAll(exam.id);
      allSubjects.push(...(subjects || []).map(s => ({ ...s, examId: exam.id })));
    }

    // Build tree
    const tree: TreeNode[] = [];

    // Level 1: Exam Categories
    for (const category of examCategories || []) {
      const categoryNode: TreeNode = {
        key: `examCategory-${category.id}`,
        title: category.name,
        type: 'examCategory',
        icon: '📁',
        data: category,
        children: [],
      };

      // Level 2: Exams
      const categoryExams = (exams || []).filter(e => e.examCategoryId === category.id);
      for (const exam of categoryExams) {
        const examNode: TreeNode = {
          key: `exam-${exam.id}`,
          title: exam.name,
          type: 'exam',
          icon: '📂',
          data: exam,
          children: [],
        };

        // Level 3: Subjects
        const examSubjects = (allSubjects || []).filter(s => s.examId === exam.id);
        for (const subject of examSubjects) {
          const subjectNode: TreeNode = {
            key: `subject-${subject.id}`,
            title: subject.name,
            type: 'subject',
            icon: '📄',
            data: subject,
            children: [],
          };

          // Level 4+: Categories (can be multi-level)
          const subjectCategories = (categories || []).filter(c => c.subjectId === subject.id);
          const categoryTree = buildCategoryTree(subjectCategories);
          subjectNode.children = categoryTree;

          examNode.children!.push(subjectNode);
        }

        categoryNode.children!.push(examNode);
      }

      tree.push(categoryNode);
    }

    treeData.value = tree;

    // Expand first level by default
    if (tree.length > 0) {
      expandedKeys.value = [tree[0]?.key].filter(Boolean) as string[];
    }
  } catch (error: any) {
    message.error(error?.message || '加载分类数据失败');
  } finally {
    loading.value = false;
  }
}

// Build category tree (multi-level)
function buildCategoryTree(categories: any[], parentId: number = 0): TreeNode[] {
  const nodes: TreeNode[] = [];
  const children = categories.filter(c => c.parentId === parentId);

  for (const category of children) {
    const node: TreeNode = {
      key: `category-${category.id}`,
      title: category.name,
      type: 'category',
      icon: '📝',
      data: category,
      children: buildCategoryTree(categories, category.id),
    };

    if (node.children!.length === 0) {
      node.isLeaf = true;
    }

    nodes.push(node);
  }

  return nodes;
}

// Handle tree node selection
function handleSelect(keys: (string | number)[], { node }: any) {
  if (keys.length > 0) {
    selectedKeys.value = keys;
    selectedNode.value = node;
    loadFormData(node);
  }
}

// Load form data based on node type
function loadFormData(node: TreeNode) {
  formData.value = { ...node.data };
}

// Get form schema based on node type
const formFields = ref<any[]>([]);

watch(selectedNode, (node) => {
  if (!node) {
    formFields.value = [];
    return;
  }

  switch (node.type) {
    case 'examCategory':
      formFields.value = [
        { label: '名称', field: 'name', type: 'input', required: true },
        { label: '编码', field: 'code', type: 'input' },
        { label: '排序', field: 'sortOrder', type: 'number', default: 0 },
        { label: '状态', field: 'status', type: 'radio', options: [{ label: '启用', value: 1 }, { label: '禁用', value: 0 }], default: 1 },
      ];
      break;
    case 'exam':
      formFields.value = [
        { label: '考试名称', field: 'name', type: 'input', required: true },
        { label: '编码', field: 'code', type: 'input' },
        { label: '描述', field: 'description', type: 'textarea' },
        { label: '排序', field: 'sortOrder', type: 'number', default: 0 },
        { label: '状态', field: 'status', type: 'radio', options: [{ label: '启用', value: 1 }, { label: '禁用', value: 0 }], default: 1 },
      ];
      break;
    case 'subject':
      formFields.value = [
        { label: '科目名称', field: 'name', type: 'input', required: true },
        { label: '编码', field: 'code', type: 'input' },
        { label: '排序', field: 'sortOrder', type: 'number', default: 0 },
        { label: '状态', field: 'status', type: 'radio', options: [{ label: '启用', value: 1 }, { label: '禁用', value: 0 }], default: 1 },
      ];
      break;
    case 'category':
      formFields.value = [
        { label: '分类名称', field: 'name', type: 'input', required: true },
        { label: '排序', field: 'sortOrder', type: 'number', default: 0 },
        { label: '状态', field: 'status', type: 'radio', options: [{ label: '启用', value: 1 }, { label: '禁用', value: 0 }], default: 1 },
      ];
      break;
  }
});

// Save form data
async function handleSave() {
  if (!selectedNode.value) return;

  saving.value = true;
  try {
    const data = formData.value;
    const type = selectedNode.value.type;
    const id = selectedNode.value.data.id;

    switch (type) {
      case 'examCategory':
        await updateExamCategory(id, data);
        break;
      case 'exam':
        await updateExam(id, data);
        break;
      case 'subject':
        await updateSubject(id, data);
        break;
      case 'category':
        await updateQuestionCategory(id, data);
        break;
    }

    message.success('保存成功');
    await loadTreeData();
    // Re-select the node
    selectedKeys.value = [selectedNode.value.key];
  } catch (error: any) {
    message.error(error?.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

// Create new node
async function handleCreate() {
  if (!selectedNode.value) {
    message.warning('请先选择一个节点');
    return;
  }

  const parentType = selectedNode.value.type;
  const parentId = selectedNode.value.data.id;
  const parentCode = selectedNode.value.data.code || '';

  // Determine child type
  switch (parentType) {
    case 'examCategory':
      createType.value = 'exam';
      break;
    case 'exam':
      createType.value = 'subject';
      break;
    case 'subject':
      createType.value = 'category';
      break;
    case 'category':
      createType.value = 'category';
      break;
    default:
      message.warning('无法在此节点下创建子级');
      return;
  }

  createParentId.value = parentId;
  createModalRef.value?.open({
    parentCode,
    nodeType: createType.value,
  });
}

// Handle create success
async function handleCreateSuccess(data: any) {
  try {
    switch (createType.value) {
      case 'exam':
        await createExam({ ...data, examCategoryId: createParentId.value, createdBy: 1 });
        break;
      case 'subject':
        await createSubject({ ...data, examId: createParentId.value, createdBy: 1 });
        break;
      case 'category':
        await createQuestionCategory({
          ...data,
          subjectId: selectedNode.value!.data.subjectId || createParentId.value,
          parentId: createParentId.value,
          createdBy: 1,
        });
        break;
    }

    message.success('创建成功');
    await loadTreeData();
  } catch (error: any) {
    message.error(error?.message || '创建失败');
  }
}

// Create root node (exam category)
function handleCreateRoot() {
  createType.value = 'examCategory';
  createParentId.value = 0;
  createModalRef.value?.open({
    parentCode: '',
    nodeType: 'examCategory',
  });
}

// Handle create root success
async function handleCreateRootSuccess(data: any) {
  try {
    await createExamCategory({
      ...data,
      createdBy: 1,
    });

    message.success('创建成功');
    await loadTreeData();
  } catch (error: any) {
    message.error(error?.message || '创建失败');
  }
}

// Delete node
async function handleDelete() {
  if (!selectedNode.value) return;

  const nodeType = selectedNode.value.type;
  const nodeId = selectedNode.value.data.id;
  const nodeTitle = selectedNode.value.title;

  Modal.confirm({
    title: '确认删除',
    content: `确定删除【${nodeTitle}】吗？删除后将同时删除其所有子节点。`,
    onOk: async () => {
      try {
        switch (nodeType) {
          case 'examCategory':
            await deleteExamCategory(nodeId);
            break;
          case 'exam':
            await deleteExam(nodeId);
            break;
          case 'subject':
            await deleteSubject(nodeId);
            break;
          case 'category':
            await deleteQuestionCategory(nodeId);
            break;
        }

        message.success('删除成功');
        selectedNode.value = null;
        selectedKeys.value = [];
        await loadTreeData();
      } catch (error: any) {
        message.error(error?.message || '删除失败');
      }
    },
  });
}

// Helper: get node type label
function getNodeTypeLabel(type: 'examCategory' | 'exam' | 'subject' | 'category'): string {
  const labels = {
    examCategory: '考试大类',
    exam: '具体考试',
    subject: '科目模块',
    category: '章节分类',
  };
  return labels[type];
}

onMounted(() => {
  loadTreeData();
});
</script>

<template>
  <Page auto-content-height>
    <CreateModal
      ref="createModalRef"
      @success="createType === 'examCategory' ? handleCreateRootSuccess($event) : handleCreateSuccess($event)"
    />

    <div class="flex h-full gap-4">
      <!-- Left: Tree -->
      <Card class="w-1/3 overflow-auto" title="分类科目树" :loading="loading">
        <template #extra>
          <Button
            v-if="hasAccessByCodes(['question:category:add'])"
            type="primary"
            size="small"
            @click="handleCreateRoot"
          >
            <Plus class="size-4" />
            新增大类
          </Button>
        </template>

        <Tree
          v-if="treeData.length > 0"
          :tree-data="treeData"
          :expanded-keys="expandedKeys"
          :selected-keys="selectedKeys"
          block-node
          @select="handleSelect"
          @expand="(keys) => expandedKeys = keys"
        >
          <template #title="{ title, data }">
            <span>
              <span class="mr-1">{{ data?.icon || '📄' }}</span>
              <span>{{ title }}</span>
            </span>
          </template>
        </Tree>

        <Empty v-else description="暂无分类数据" />
      </Card>

      <!-- Right: Detail Panel -->
      <Card class="w-2/3" title="详细信息">
        <div v-if="selectedNode" class="p-4">
          <!-- Node info header -->
          <div class="mb-6 flex items-center justify-between">
            <div>
              <h3 class="text-lg font-medium">{{ selectedNode.title }}</h3>
              <div class="mt-1 text-sm text-gray-500">
                类型：{{ getNodeTypeLabel(selectedNode.type) }}
                <span v-if="selectedNode.data.code"> | 编码：{{ selectedNode.data.code }}</span>
              </div>
            </div>
            <div class="flex gap-2">
              <Button
                v-if="hasAccessByCodes(['question:category:add'])"
                type="primary"
                size="small"
                @click="handleCreate"
              >
                <Plus class="size-4" />
                新增子级
              </Button>
              <Button
                v-if="hasAccessByCodes(['question:category:edit'])"
                size="small"
                @click="handleSave"
                :loading="saving"
              >
                保存
              </Button>
              <Button
                v-if="hasAccessByCodes(['question:category:delete'])"
                danger
                size="small"
                @click="handleDelete"
              >
                删除
              </Button>
            </div>
          </div>

          <!-- Form fields -->
          <div class="space-y-4">
            <div v-for="field in formFields" :key="field.field">
              <label class="block text-sm font-medium mb-2">
                {{ field.label }}
                <span v-if="field.required" class="text-red-500">*</span>
              </label>

              <!-- Input -->
              <Input
                v-if="field.type === 'input'"
                v-model:value="formData[field.field]"
                :placeholder="`请输入${field.label}`"
              />

              <!-- Number -->
              <InputNumber
                v-else-if="field.type === 'number'"
                v-model:value="formData[field.field]"
                :min="0"
                class="w-full"
              />

              <!-- Textarea -->
              <Input.TextArea
                v-else-if="field.type === 'textarea'"
                v-model:value="formData[field.field]"
                :placeholder="`请输入${field.label}`"
                :rows="3"
              />

              <!-- Radio -->
              <RadioGroup
                v-else-if="field.type === 'radio'"
                v-model:value="formData[field.field]"
                :options="field.options"
                button-style="solid"
              />
            </div>
          </div>
        </div>

        <Empty v-else description="请在左侧树中选择一个节点" />
      </Card>
    </div>
  </Page>
</template>