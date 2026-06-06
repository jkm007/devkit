<script lang="ts" setup>
import type { AccountApi } from '#/api';

import { onMounted, ref } from 'vue';

import { VbenAvatar } from '@vben/common-ui';

import {
  Button,
  Col,
  DatePicker,
  Form,
  FormItem,
  Input,
  message,
  Radio,
  RadioGroup,
  Row,
} from 'ant-design-vue';

import { getUserInfo, updateProfile } from '#/api';
import { $t } from '#/locales';

import AvatarUpload from '#/components/avatar-upload/index.vue';

const loading = ref(false);
const saving = ref(false);
const userInfo = ref<AccountApi.UserInfo>({} as AccountApi.UserInfo);
const avatarUploadRef = ref<InstanceType<typeof AvatarUpload> | null>(null);

const formState = ref({
  nickname: '',
  email: '',
  phone: '',
  gender: 0,
  birthday: '',
  bio: '',
});

async function loadUserInfo() {
  loading.value = true;
  try {
    const res = await getUserInfo();
    userInfo.value = res;
    formState.value = {
      nickname: res.nickname || '',
      email: res.email || '',
      phone: res.phone || '',
      gender: res.gender || 0,
      birthday: res.birthday || '',
      bio: res.bio || '',
    };
  } catch {
    message.error('加载用户信息失败');
  } finally {
    loading.value = false;
  }
}

async function handleSave() {
  saving.value = true;
  try {
    await updateProfile(formState.value);
    message.success($t('account.profile.saveSuccess'));
    await loadUserInfo();
  } catch {
    message.error('保存失败');
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  loadUserInfo();
});

// 打开头像上传
function handleAvatarChange() {
  avatarUploadRef.value?.open();
}

// 头像上传成功后刷新用户信息
async function handleAvatarSuccess(url: string) {
  userInfo.value.avatar = url;
}
</script>

<template>
  <div v-loading="loading">
    <div class="mb-6 flex items-center gap-4">
      <div class="avatar-wrapper cursor-pointer" @click="handleAvatarChange">
        <VbenAvatar
          :src="userInfo.avatar"
          :alt="userInfo.username || 'avatar'"
          :size="72"
        />
        <div class="avatar-overlay">
          <span class="i-ant-design:camera-outlined text-lg" />
          <span class="text-xs">{{ $t('file.avatar.change') }}</span>
        </div>
      </div>
      <div>
        <div class="text-xl font-medium">
          {{ userInfo.nickname || userInfo.username }}
        </div>
        <div class="text-foreground/50 mt-1">
          {{ userInfo.username }}
          <span v-if="userInfo.roles?.length" class="ml-2 text-xs">
            {{ userInfo.roles.join(', ') }}
          </span>
        </div>
      </div>
    </div>

    <!-- 头像上传组件 -->
    <AvatarUpload ref="avatarUploadRef" @success="handleAvatarSuccess" />

    <Form layout="vertical" :model="formState">
      <Row :gutter="16">
        <Col :span="12">
          <FormItem :label="$t('account.profile.nickname')">
            <Input
              v-model:value="formState.nickname"
              :placeholder="$t('account.profile.nickname')"
            />
          </FormItem>
        </Col>
        <Col :span="12">
          <FormItem :label="$t('account.profile.email')">
            <Input
              v-model:value="formState.email"
              :placeholder="$t('account.profile.email')"
            />
          </FormItem>
        </Col>
      </Row>

      <Row :gutter="16">
        <Col :span="12">
          <FormItem :label="$t('account.profile.phone')">
            <Input
              v-model:value="formState.phone"
              :placeholder="$t('account.profile.phone')"
            />
          </FormItem>
        </Col>
        <Col :span="12">
          <FormItem :label="$t('account.profile.gender')">
            <RadioGroup v-model:value="formState.gender">
              <Radio :value="0">{{ $t('account.profile.genderUnknown') }}</Radio>
              <Radio :value="1">{{ $t('account.profile.genderMale') }}</Radio>
              <Radio :value="2">{{ $t('account.profile.genderFemale') }}</Radio>
            </RadioGroup>
          </FormItem>
        </Col>
      </Row>

      <Row :gutter="16">
        <Col :span="12">
          <FormItem :label="$t('account.profile.birthday')">
            <DatePicker
              v-model:value="formState.birthday"
              :placeholder="$t('account.profile.birthday')"
              class="w-full"
              value-format="YYYY-MM-DD"
            />
          </FormItem>
        </Col>
      </Row>

      <FormItem :label="$t('account.profile.bio')">
        <Input.TextArea
          v-model:value="formState.bio"
          :placeholder="$t('account.profile.bioPlaceholder')"
          :rows="3"
          :maxlength="500"
          show-count
        />
      </FormItem>

      <FormItem>
        <Button type="primary" :loading="saving" @click="handleSave">
          {{ $t('account.profile.save') }}
        </Button>
      </FormItem>
    </Form>
  </div>
</template>

<style scoped>
.avatar-wrapper {
  position: relative;
  border-radius: 50%;
  overflow: hidden;
}

.avatar-wrapper:hover .avatar-overlay {
  opacity: 1;
}

.avatar-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 50%;
  opacity: 0;
  transition: opacity 0.2s ease;
  color: white;
}
</style>
