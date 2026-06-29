<script lang="ts" setup>
import { ref, onMounted } from 'vue';

import { Page } from '@vben/common-ui';

import { message } from 'ant-design-vue';
import {
  Button,
  Card,
  Form,
  FormItem,
  Input,
  Switch,
} from 'ant-design-vue';

import {
  getMobileSettings,
  updateMobileSettings,
} from '#/api/system/mobile-config';

const form = ref({
  noticeEnabled: false,
  noticeContent: '',
  appDownloadUrl: '',
  customerServiceUrl: '',
  aboutUs: '',
  agreementUrl: '',
  privacyUrl: '',
});

const loading = ref(false);
const saving = ref(false);

onMounted(async () => {
  loading.value = true;
  try {
    const data = await getMobileSettings();
    if (data) {
      form.value = { ...form.value, ...data };
    }
  } catch (error) {
    console.error('加载设置失败:', error);
  } finally {
    loading.value = false;
  }
});

async function handleSave() {
  saving.value = true;
  try {
    await updateMobileSettings(form.value);
    message.success('保存成功');
  } catch (error: any) {
    message.error(error.message || '保存失败');
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <Page
    auto-content-height
    description="配置移动端全局设置"
    title="移动端设置"
    :loading="loading"
  >
    <template #extra>
      <Button type="primary" :loading="saving" @click="handleSave">
        保存设置
      </Button>
    </template>

    <div class="max-w-3xl mx-auto space-y-6">
      <!-- 公告设置 -->
      <Card title="📢 公告设置">
        <Form layout="vertical" :model="form">
          <FormItem label="启用公告">
            <Switch
              v-model:checked="form.noticeEnabled"
              checked-children="启用"
              un-checked-children="禁用"
            />
          </FormItem>
          <FormItem v-if="form.noticeEnabled" label="公告内容">
            <Input.TextArea
              v-model:value="form.noticeContent"
              placeholder="请输入公告内容"
              :rows="4"
            />
          </FormItem>
        </Form>
      </Card>

      <!-- 链接设置 -->
      <Card title="🔗 链接设置">
        <Form layout="vertical" :model="form">
          <FormItem label="APP下载地址">
            <Input
              v-model:value="form.appDownloadUrl"
              placeholder="https://..."
            />
          </FormItem>
          <FormItem label="客服链接">
            <Input
              v-model:value="form.customerServiceUrl"
              placeholder="https://..."
            />
          </FormItem>
          <FormItem label="用户协议链接">
            <Input
              v-model:value="form.agreementUrl"
              placeholder="https://..."
            />
          </FormItem>
          <FormItem label="隐私政策链接">
            <Input
              v-model:value="form.privacyUrl"
              placeholder="https://..."
            />
          </FormItem>
        </Form>
      </Card>

      <!-- 关于我们 -->
      <Card title="ℹ️ 关于我们">
        <Form layout="vertical" :model="form">
          <FormItem label="关于我们内容">
            <Input.TextArea
              v-model:value="form.aboutUs"
              placeholder="请输入关于我们的介绍内容"
              :rows="6"
            />
          </FormItem>
        </Form>
      </Card>
    </div>
  </Page>
</template>
