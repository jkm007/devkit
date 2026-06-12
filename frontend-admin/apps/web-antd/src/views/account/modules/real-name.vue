<script lang="ts" setup>
import type { AccountApi } from '#/api';

import { onMounted, ref } from 'vue';

import {
  Button,
  Col,
  Descriptions,
  DescriptionsItem,
  Form,
  FormItem,
  Input,
  message,
  Row,
  Tag,
} from 'ant-design-vue';

import { getRealNameStatus, submitRealName } from '#/api';
import { $t } from '#/locales';

// ==================== 实名认证 ====================
const realNameLoading = ref(false);
const realNameStatus = ref<AccountApi.RealNameStatus>(
  {} as AccountApi.RealNameStatus,
);
const submitting = ref(false);
const realNameForm = ref({
  realName: '',
  idCard: '',
});

/**
 * Mask ID card number: show first 6 and last 4 digits, mask middle with ****
 * e.g. 110102**********1234
 */
function maskIdCard(idCard: string): string {
  if (!idCard || idCard.length < 11) {
    return idCard;
  }
  return `${idCard.slice(0, 6)}****${idCard.slice(-4)}`;
}

async function loadRealNameStatus() {
  realNameLoading.value = true;
  try {
    realNameStatus.value = await getRealNameStatus();
  } catch {
    realNameStatus.value = {} as AccountApi.RealNameStatus;
  } finally {
    realNameLoading.value = false;
  }
}

async function handleSubmitRealName() {
  if (!realNameForm.value.realName || !realNameForm.value.idCard) {
    message.warning($t('account.privacy.formRequired'));
    return;
  }
  submitting.value = true;
  try {
    await submitRealName(realNameForm.value);
    message.success($t('account.privacy.submitSuccess'));
    realNameForm.value = { realName: '', idCard: '' };
    await loadRealNameStatus();
  } catch {
    message.error('提交失败');
  } finally {
    submitting.value = false;
  }
}

function getStatusTag(status: number) {
  const map: Record<number, { color: string; text: string }> = {
    0: { color: 'default', text: $t('account.privacy.notVerified') },
    1: { color: 'success', text: $t('account.privacy.verified') },
    2: { color: 'processing', text: $t('account.privacy.pending') },
    3: { color: 'error', text: $t('account.privacy.rejected') },
  };
  return (
    map[status] || { color: 'default', text: $t('account.privacy.notVerified') }
  );
}

onMounted(() => {
  loadRealNameStatus();
});
</script>

<template>
  <div>
    <h3 class="mb-4 text-base font-medium">
      {{ $t('account.privacy.realName') }}
    </h3>
    <div v-loading="realNameLoading">
      <!-- 已有认证状态（非未认证状态） -->
      <div
        v-if="
          realNameStatus.status !== undefined && realNameStatus.status !== 0
        "
      >
        <Descriptions bordered :column="1">
          <DescriptionsItem :label="$t('account.privacy.realNameStatus')">
            <Tag :color="getStatusTag(realNameStatus.status).color">
              {{ getStatusTag(realNameStatus.status).text }}
            </Tag>
          </DescriptionsItem>
          <DescriptionsItem
            v-if="realNameStatus.status === 1 && realNameStatus.realName"
            :label="$t('account.privacy.realNameLabel')"
          >
            {{ realNameStatus.realName }}
          </DescriptionsItem>
          <DescriptionsItem
            v-if="realNameStatus.status === 1 && realNameStatus.idCard"
            :label="$t('account.privacy.idCardLabel')"
          >
            {{ maskIdCard(realNameStatus.idCard) }}
          </DescriptionsItem>
          <DescriptionsItem
            v-if="realNameStatus.rejectReason"
            :label="$t('account.privacy.rejectReason')"
          >
            <span class="text-red-500">{{
              realNameStatus.rejectReason
            }}</span>
          </DescriptionsItem>
        </Descriptions>

        <!-- 认证失败可以重新提交 -->
        <div v-if="realNameStatus.status === 3" class="mt-4">
          <Form layout="vertical" :model="realNameForm">
            <Row :gutter="16">
              <Col :span="12">
                <FormItem :label="$t('account.privacy.realNameLabel')">
                  <Input
                    v-model:value="realNameForm.realName"
                    :placeholder="$t('account.privacy.realNamePlaceholder')"
                  />
                </FormItem>
              </Col>
              <Col :span="12">
                <FormItem :label="$t('account.privacy.idCardLabel')">
                  <Input
                    v-model:value="realNameForm.idCard"
                    :placeholder="$t('account.privacy.idCardPlaceholder')"
                  />
                </FormItem>
              </Col>
            </Row>
            <FormItem>
              <Button
                type="primary"
                :loading="submitting"
                @click="handleSubmitRealName"
              >
                {{ $t('account.privacy.submitRealName') }}
              </Button>
            </FormItem>
          </Form>
        </div>
      </div>

      <!-- 未认证 - 显示提交表单 -->
      <div v-else>
        <Form layout="vertical" :model="realNameForm">
          <Row :gutter="16">
            <Col :span="12">
              <FormItem :label="$t('account.privacy.realNameLabel')">
                <Input
                  v-model:value="realNameForm.realName"
                  :placeholder="$t('account.privacy.realNamePlaceholder')"
                />
              </FormItem>
            </Col>
            <Col :span="12">
              <FormItem :label="$t('account.privacy.idCardLabel')">
                <Input
                  v-model:value="realNameForm.idCard"
                  :placeholder="$t('account.privacy.idCardPlaceholder')"
                />
              </FormItem>
            </Col>
          </Row>
          <FormItem>
            <Button
              type="primary"
              :loading="submitting"
              @click="handleSubmitRealName"
            >
              {{ $t('account.privacy.submitRealName') }}
            </Button>
          </FormItem>
        </Form>
      </div>
    </div>
  </div>
</template>
