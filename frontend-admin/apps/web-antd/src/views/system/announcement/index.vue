<script lang="ts" setup>
import { h, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Card,
  Empty,
  Form,
  Input,
  message,
  Modal,
  Table,
} from 'ant-design-vue';

import type { Notification } from '#/api/notification';
import { getAdminNotifications, publishAnnouncement } from '#/api/notification';

const formState = ref({
  title: '',
  content: '',
  link: '',
});
const publishing = ref(false);
const showPublishModal = ref(false);

const announcements = ref<Notification[]>([]);
const loading = ref(false);
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
});

const columns = [
  {
    title: '标题',
    dataIndex: 'title',
    key: 'title',
    width: 200,
  },
  {
    title: '内容',
    dataIndex: 'content',
    key: 'content',
    ellipsis: true,
  },
  {
    title: '跳转链接',
    dataIndex: 'link',
    key: 'link',
    width: 180,
    customRender: ({ text }: { text: string }) => {
      if (!text) return '-';
      return h('a', { href: text, target: '_blank', class: 'text-blue-500' }, text);
    },
  },
  {
    title: '发布时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 180,
    customRender: ({ text }: { text: string }) => {
      if (!text) return '-';
      return new Date(text).toLocaleString('zh-CN');
    },
  },
];

async function loadAnnouncements() {
  loading.value = true;
  try {
    const res = await getAdminNotifications({
      page: pagination.value.current,
      pageSize: pagination.value.pageSize,
      type: 'announcement',
    });
    announcements.value = res.items || [];
    pagination.value.total = res.total || 0;
  } catch {
    // 静默失败
  } finally {
    loading.value = false;
  }
}

function handleTableChange(pag: any) {
  pagination.value.current = pag.current;
  pagination.value.pageSize = pag.pageSize;
  loadAnnouncements();
}

async function handlePublish() {
  if (!formState.value.title.trim()) {
    message.warning('请输入公告标题');
    return;
  }
  if (!formState.value.content.trim()) {
    message.warning('请输入公告内容');
    return;
  }

  publishing.value = true;
  try {
    await publishAnnouncement({
      title: formState.value.title,
      content: formState.value.content,
      link: formState.value.link || undefined,
    });
    message.success('公告发布成功');
    showPublishModal.value = false;
    formState.value = { title: '', content: '', link: '' };
    loadAnnouncements();
  } catch (err: any) {
    message.error(err?.message || '发布失败');
  } finally {
    publishing.value = false;
  }
}

function openPublishModal() {
  formState.value = { title: '', content: '', link: '' };
  showPublishModal.value = true;
}

onMounted(loadAnnouncements);
</script>

<template>
  <Page auto-content-height>
    <Card title="公告管理" :bordered="true">
      <template #extra>
        <Button type="primary" @click="openPublishModal">
          发布公告
        </Button>
      </template>

      <Table
        :columns="columns"
        :data-source="announcements"
        :loading="loading"
        :pagination="pagination"
        row-key="id"
        @change="handleTableChange"
      >
        <template #emptyText>
          <Empty description="暂无公告" />
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="showPublishModal"
      title="发布公告"
      :confirm-loading="publishing"
      @ok="handlePublish"
      @cancel="showPublishModal = false"
    >
      <Form layout="vertical" :model="formState">
        <Form.Item label="公告标题" required>
          <Input
            v-model:value="formState.title"
            placeholder="请输入公告标题"
            :maxlength="200"
          />
        </Form.Item>
        <Form.Item label="公告内容" required>
          <Input.TextArea
            v-model:value="formState.content"
            placeholder="请输入公告内容"
            :rows="6"
            :maxlength="5000"
          />
        </Form.Item>
        <Form.Item label="跳转链接">
          <Input
            v-model:value="formState.link"
            placeholder="可选，点击公告后跳转的链接"
          />
        </Form.Item>
      </Form>
    </Modal>
  </Page>
</template>
