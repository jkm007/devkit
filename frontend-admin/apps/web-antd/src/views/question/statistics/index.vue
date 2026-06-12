<script lang="ts" setup>
import { onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Card, Col, Row, Statistic } from 'ant-design-vue';

import { getQuestionStats } from '#/api/question/question';

const stats = ref<any>({});
const loading = ref(true);

async function loadStats() {
  loading.value = true;
  try {
    const res = await getQuestionStats();
    stats.value = res || {};
  } catch {
    // ignore
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadStats();
});
</script>
<template>
  <Page auto-content-height>
    <Row :gutter="16">
      <Col :span="6">
        <Card>
          <Statistic title="总题目数" :value="stats.total || 0" />
        </Card>
      </Col>
      <Col :span="6">
        <Card>
          <Statistic
            title="已发布"
            :value="
              (stats.byStatus || []).find(
                (s: any) => s.status === 'published',
              )?.count || 0
            "
            value-style="color: #3f8600"
          />
        </Card>
      </Col>
      <Col :span="6">
        <Card>
          <Statistic
            title="草稿"
            :value="
              (stats.byStatus || []).find((s: any) => s.status === 'draft')
                ?.count || 0
            "
          />
        </Card>
      </Col>
      <Col :span="6">
        <Card>
          <Statistic
            title="待审核"
            :value="
              (stats.byStatus || []).find((s: any) => s.status === 'pending')
                ?.count || 0
            "
            value-style="color: #cf1322"
          />
        </Card>
      </Col>
    </Row>

    <Row :gutter="16" class="mt-4">
      <Col :span="12">
        <Card title="按题型统计">
          <div v-for="item in stats.byType || []" :key="item.questionType">
            {{ item.questionType }}: {{ item.count }} 题
          </div>
          <div v-if="!stats.byType?.length" class="text-gray-400">暂无数据</div>
        </Card>
      </Col>
      <Col :span="12">
        <Card title="按难度统计">
          <div v-for="item in stats.byDifficulty || []" :key="item.difficulty">
            {{ ['', '简单', '中等', '困难'][item.difficulty] || '未知' }}:
            {{ item.count }} 题
          </div>
          <div v-if="!stats.byDifficulty?.length" class="text-gray-400">
            暂无数据
          </div>
        </Card>
      </Col>
    </Row>
  </Page>
</template>
