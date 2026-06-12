<script lang="ts" setup>
import { onMounted, ref } from 'vue';

import {
  Button,
  Col,
  Form,
  FormItem,
  message,
  Row,
  Select,
  SelectOption,
} from 'ant-design-vue';

import { getPrivacySettings, updatePrivacySettings } from '#/api';
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

onMounted(() => {
  loadPrivacy();
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
  </div>
</template>
