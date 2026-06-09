<script lang="ts" setup>
import { onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';
import { $t } from '#/locales';

import {
  Alert,
  Badge,
  Button,
  Card,
  Col,
  Input,
  message,
  Popconfirm,
  Row,
  Spin,
  Statistic,
  Table,
  TableColumn,
  Tag,
} from 'ant-design-vue';

import { getRiskScores, getRiskStats, clearRiskScore } from '#/api/system/risk';

interface RiskScoreItem {
  ip: string;
  score: number;
  updatedAt: string;
  expireAt: string;
}

interface RiskStats {
  totalCount: number;
  triggerCount: number;
  blockCount: number;
  highRiskCount: number;
  triggerScore: number;
  blockScore: number;
  enabled: boolean;
}

const loading = ref(false);
const scores = ref<RiskScoreItem[]>([]);
const stats = ref<RiskStats | null>(null);
const searchIP = ref('');
const refreshing = ref(false);

async function loadData() {
  loading.value = true;
  try {
    const [scoresData, statsData] = await Promise.all([
      getRiskScores(100),
      getRiskStats(),
    ]);
    scores.value = scoresData;
    stats.value = statsData;
  } catch (e: any) {
    message.error('获取数据失败：' + (e?.message || '未知错误'));
  } finally {
    loading.value = false;
  }
}

async function handleRefresh() {
  refreshing.value = true;
  await loadData();
  refreshing.value = false;
  message.success('数据已刷新');
}

async function handleClear(ip: string) {
  try {
    await clearRiskScore(ip);
    message.success('已清除 ' + ip + ' 的风险评分');
    await loadData();
  } catch (e: any) {
    message.error('清除失败：' + (e?.message || '未知错误'));
  }
}

async function handleSearch() {
  if (!searchIP.value) {
    await loadData();
    return;
  }
  loading.value = true;
  try {
    const data = await getRiskScores(100);
    scores.value = data.filter((item: RiskScoreItem) =>
      item.ip.includes(searchIP.value),
    );
  } catch (e: any) {
    message.error('查询失败');
  } finally {
    loading.value = false;
  }
}

function getScoreColor(score: number): string {
  const triggerScore = stats.value?.triggerScore || 50;
  const blockScore = stats.value?.blockScore || 80;
  if (score >= blockScore) return 'red';
  if (score >= triggerScore) return 'orange';
  if (score >= 30) return 'yellow';
  return 'green';
}

function formatTime(timeStr: string): string {
  if (!timeStr) return '-';
  const date = new Date(timeStr);
  return date.toLocaleString('zh-CN');
}

onMounted(() => {
  loadData();
});
</script>

<template>
  <Page :title="$t('system.risk.title')" auto-content-height>
    <Spin :spinning="loading">
      <!-- 统计卡片 -->
      <Row :gutter="16" class="mb-4">
        <Col :span="4">
          <Card>
            <Statistic
              title="监控 IP 数"
              :value="stats?.totalCount || 0"
              :value-style="{ color: '#1890ff' }"
            />
          </Card>
        </Col>
        <Col :span="4">
          <Card>
            <Statistic
              title="触发验证码"
              :value="stats?.triggerCount || 0"
              :value-style="{ color: '#faad14' }"
            />
          </Card>
        </Col>
        <Col :span="4">
          <Card>
            <Statistic
              title="已拦截"
              :value="stats?.blockCount || 0"
              :value-style="{ color: '#ff4d4f' }"
            />
          </Card>
        </Col>
        <Col :span="4">
          <Card>
            <Statistic
              title="高风险 (>70)"
              :value="stats?.highRiskCount || 0"
              :value-style="{ color: '#ff4d4f' }"
            />
          </Card>
        </Col>
        <Col :span="4">
          <Card>
            <Statistic
              title="触发阈值"
              :value="stats?.triggerScore || 50"
              suffix="分"
            />
          </Card>
        </Col>
        <Col :span="4">
          <Card>
            <Statistic
              title="拦截阈值"
              :value="stats?.blockScore || 80"
              suffix="分"
            />
          </Card>
        </Col>
      </Row>

      <!-- 提示信息 -->
      <Alert
        v-if="stats?.enabled"
        message="风险评分系统已启用"
        description="对敏感接口进行多维度风险评估，高风险请求需要验证码验证或直接拦截。"
        type="info"
        show-icon
        class="mb-4"
      />
      <Alert
        v-else
        message="风险评分系统已关闭"
        description="所有请求均不做风险评估，建议启用以保护敏感接口。"
        type="warning"
        show-icon
        class="mb-4"
      />

      <!-- 操作栏 -->
      <Card class="mb-4">
        <Row :gutter="16" align="middle">
          <Col :span="8">
            <Input
              v-model:value="searchIP"
              placeholder="搜索 IP 地址..."
              allow-clear
              @press-enter="handleSearch"
            />
          </Col>
          <Col :span="4">
            <Button type="primary" @click="handleSearch">搜索</Button>
          </Col>
          <Col :span="4">
            <Button :loading="refreshing" @click="handleRefresh">
              刷新数据
            </Button>
          </Col>
        </Row>
      </Card>

      <!-- 风险评分列表 -->
      <Card title="风险评分列表">
        <Table :data-source="scores" :pagination="{ pageSize: 20 }" row-key="ip">
          <TableColumn title="IP 地址" data-index="ip" width="150">
            <template #default="{ record }">
              <span class="font-mono">{{ record.ip }}</span>
            </template>
          </TableColumn>
          <TableColumn title="风险分数" data-index="score" width="120" sorter>
            <template #default="{ record }">
              <Badge :color="getScoreColor(record.score)" :text="record.score" />
            </template>
          </TableColumn>
          <TableColumn title="风险等级" width="100">
            <template #default="{ record }">
              <Tag
                :color="getScoreColor(record.score)"
                v-if="record.score >= (stats?.blockScore || 80)"
              >
                拦截
              </Tag>
              <Tag
                :color="getScoreColor(record.score)"
                v-else-if="record.score >= (stats?.triggerScore || 50)"
              >
                验证码
              </Tag>
              <Tag color="green" v-else>正常</Tag>
            </template>
          </TableColumn>
          <TableColumn title="最后更新" data-index="updatedAt" width="180">
            <template #default="{ record }">
              {{ formatTime(record.updatedAt) }}
            </template>
          </TableColumn>
          <TableColumn title="过期时间" data-index="expireAt" width="180">
            <template #default="{ record }">
              {{ formatTime(record.expireAt) }}
            </template>
          </TableColumn>
          <TableColumn title="操作" width="100" fixed="right">
            <template #default="{ record }">
              <Popconfirm
                title="确定清除该 IP 的风险评分？"
                @confirm="handleClear(record.ip)"
              >
                <Button type="link" danger size="small">清除</Button>
              </Popconfirm>
            </template>
          </TableColumn>
        </Table>
      </Card>
    </Spin>
  </Page>
</template>

<style scoped>
.font-mono {
  font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
}
</style>