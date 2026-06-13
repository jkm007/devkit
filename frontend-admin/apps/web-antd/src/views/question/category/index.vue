<script lang="ts" setup>
import { ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Tabs } from 'ant-design-vue';

import CategoryTab from './modules/category-tab.vue';
import ExamCategoryTab from './modules/exam-category-tab.vue';
import ExamTab from './modules/exam-tab.vue';
import SubjectTab from './modules/subject-tab.vue';

const activeKey = ref('examCategory');
</script>
<template>
  <Page auto-content-height>
    <Tabs v-model:active-key="activeKey" type="card" destroy-inactive-tab-pane>
      <Tabs.TabPane key="examCategory" tab="考试大类" force-render>
        <ExamCategoryTab />
      </Tabs.TabPane>
      <Tabs.TabPane key="exam" tab="具体考试" force-render>
        <ExamTab />
      </Tabs.TabPane>
      <Tabs.TabPane key="subject" tab="科目模块" force-render>
        <SubjectTab />
      </Tabs.TabPane>
      <Tabs.TabPane key="category" tab="章节分类" force-render>
        <CategoryTab />
      </Tabs.TabPane>
    </Tabs>
  </Page>
</template>
