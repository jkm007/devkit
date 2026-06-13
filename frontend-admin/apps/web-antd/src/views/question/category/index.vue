<script lang="ts" setup>
import { ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Button, Card } from 'ant-design-vue';

import CategoryTab from './modules/category-tab.vue';
import ExamCategoryTab from './modules/exam-category-tab.vue';
import ExamTab from './modules/exam-tab.vue';
import SubjectTab from './modules/subject-tab.vue';

const activeKey = ref('examCategory');

const tabs = [
  {
    key: 'examCategory',
    label: '考试大类',
    desc: '最顶层分类，区分不同考试领域（如：软考、小学教育、编程学习）'
  },
  {
    key: 'exam',
    label: '具体考试',
    desc: '考试大类下的具体考试项目（如：软考初级、一年级、应用编程）'
  },
  {
    key: 'subject',
    label: '科目模块',
    desc: '具体考试下的科目或技术方向（如：程序员、数学、前端开发）'
  },
  {
    key: 'category',
    label: '章节分类',
    desc: '科目下的详细知识点，可多级细分（如：数据结构 → 数组 → 排序）'
  },
];
</script>
<template>
  <Page auto-content-height>
    <div class="flex h-full flex-col">
      <!-- Tab buttons with descriptions -->
      <div class="mb-4">
        <div class="flex gap-2 mb-3">
          <Button
            v-for="tab in tabs"
            :key="tab.key"
            :type="activeKey === tab.key ? 'primary' : 'default'"
            @click="activeKey = tab.key"
          >
            {{ tab.label }}
          </Button>
        </div>
        <!-- Description card -->
        <Card class="bg-blue-50 border-blue-200">
          <div class="text-sm">
            <div class="font-medium text-blue-700 mb-1">当前层级：{{ tabs.find(t => t.key === activeKey)?.label }}</div>
            <div class="text-gray-600">{{ tabs.find(t => t.key === activeKey)?.desc }}</div>
            <div class="text-xs text-gray-500 mt-2">
              层级结构：考试大类 → 具体考试 → 科目模块 → 章节分类（可多级）
            </div>
          </div>
        </Card>
      </div>

      <!-- Tab content -->
      <div class="min-h-0 flex-1">
        <div v-show="activeKey === 'examCategory'" class="h-full">
          <ExamCategoryTab />
        </div>
        <div v-show="activeKey === 'exam'" class="h-full">
          <ExamTab />
        </div>
        <div v-show="activeKey === 'subject'" class="h-full">
          <SubjectTab />
        </div>
        <div v-show="activeKey === 'category'" class="h-full">
          <CategoryTab />
        </div>
      </div>
    </div>
  </Page>
</template>
