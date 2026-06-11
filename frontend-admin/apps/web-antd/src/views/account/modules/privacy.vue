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
  Select,
  SelectOption,
  Tag,
} from 'ant-design-vue';

import {
  getPrivacySettings,
  getRealNameStatus,
  submitRealName,
  updatePrivacySettings,
} from '#/api';
import { $t } from '#/locales';

// ==================== 隐私设置 ====================
const privacyLoading = ref(false);
const privacySaving = ref(false);
const privacyForm = ref({
  profileVisible: 1,
  realnameVisible: 1,
  emailVisible: 1,
  statsVisible: 1,
  classVisible: 1,
});

const visibilityOptions = [
  { label: $t('account.privacy.visibleAll'), value: 1 },
  { label: $t('account.privacy.visibleClass'), value: 2 },
  { label: $t('account.privacy.visibleSelf'), value: 3 },
];

async function loadPrivacy() {
  privacyLoading.value = true;
  try {
    const res = await getPrivacySettings();
    privacyForm.value = {
      profileVisible: res.profileVisible ?? 1,
      realnameVisible: res.realnameVisible ?? 1,
      emailVisible: res.emailVisible ?? 1,
      statsVisible: res.statsVisible ?? 1,
      classVisible: res.classVisible ?? 1,
    };
  } catch {
    // 静默处理
  } finally {
    privacyLoading.value = false;
  }
}

async function handleSavePrivacy() {
  privacySaving.value = true;
  try {
    await updatePrivacySettings(privacyForm.value);
    message.success($t('account.privacy.saveSuccess'));
  } catch {
    message.error('保存失败');
  } finally {
    privacySaving.value = false;
  }
}

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
  loadPrivacy();
  loadRealNameStatus();
});
</script>

<template>
  <div class="space-y-6">
    <!-- 隐私设置 -->
    <div>
      <h3 class="mb-4 text-base font-medium">
        {{ $t('account.privacy.privacySettings') }}
      </h3>
      <div v-loading="privacyLoading">
        <Form layout="vertical" :model="privacyForm">
          <Row :gutter="16">
            <Col :span="12">
              <FormItem :label="$t('account.privacy.profileVisible')">
                <Select v-model:value="privacyForm.profileVisible">
                  <SelectOption
                    v-for="opt in visibilityOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >
                    {{ opt.label }}
                  </SelectOption>
                </Select>
              </FormItem>
            </Col>
            <Col :span="12">
              <FormItem :label="$t('account.privacy.realnameVisible')">
                <Select v-model:value="privacyForm.realnameVisible">
                  <SelectOption
                    v-for="opt in visibilityOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >
                    {{ opt.label }}
                  </SelectOption>
                </Select>
              </FormItem>
            </Col>
          </Row>
          <Row :gutter="16">
            <Col :span="12">
              <FormItem :label="$t('account.privacy.emailVisible')">
                <Select v-model:value="privacyForm.emailVisible">
                  <SelectOption
                    v-for="opt in visibilityOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >
                    {{ opt.label }}
                  </SelectOption>
                </Select>
              </FormItem>
            </Col>
            <Col :span="12">
              <FormItem :label="$t('account.privacy.statsVisible')">
                <Select v-model:value="privacyForm.statsVisible">
                  <SelectOption
                    v-for="opt in visibilityOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >
                    {{ opt.label }}
                  </SelectOption>
                </Select>
              </FormItem>
            </Col>
          </Row>
          <Row :gutter="16">
            <Col :span="12">
              <FormItem :label="$t('account.privacy.classVisible')">
                <Select v-model:value="privacyForm.classVisible">
                  <SelectOption
                    v-for="opt in visibilityOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >
                    {{ opt.label }}
                  </SelectOption>
                </Select>
              </FormItem>
            </Col>
          </Row>
          <FormItem>
            <Button
              type="primary"
              :loading="privacySaving"
              @click="handleSavePrivacy"
            >
              {{ $t('account.profile.save') }}
            </Button>
          </FormItem>
        </Form>
      </div>
    </div>

    <!-- 实名认证 -->
    <div>
      <h3 class="mb-4 text-base font-medium">
        {{ $t('account.privacy.realName') }}
      </h3>
      <div v-loading="realNameLoading">
        <!-- 已有认证状态 -->
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
              v-if="realNameStatus.realName"
              :label="$t('account.privacy.realNameLabel')"
            >
              {{ realNameStatus.realName }}
            </DescriptionsItem>
            <DescriptionsItem
              v-if="realNameStatus.idCard"
              :label="$t('account.privacy.idCardLabel')"
            >
              {{ realNameStatus.idCard }}
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
  </div>
</template>
