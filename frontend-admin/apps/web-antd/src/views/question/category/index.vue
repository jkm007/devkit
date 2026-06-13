<script lang="ts" setup>
import { ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Button } from 'ant-design-vue';

import CategoryTab from './modules/category-tab.vue';
import ExamCategoryTab from './modules/exam-category-tab.vue';
import ExamTab from './modules/exam-tab.vue';
import SubjectTab from './modules/subject-tab.vue';

const activeKey = ref('examCategory');

const tabs = [
  { key: 'examCategory', label: '考试大类' },
  { key: 'exam', label: '具体考试' },
  { key: 'subject', label: '科目模块' },
  { key: 'category', label: '章节分类' },
];
</script>
<template>
  <Page auto-content-height>
    <div class="flex h-full flex-col">
      <div class="mb-3 flex gap-2">
        <Button
          v-for="tab in tabs"
          :key="tab.key"
          :type="activeKey === tab.key ? 'primary' : 'default'"
          @click="activeKey = tab.key"
        >
          {{ tab.label }}
        </Button>
      </div>
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
