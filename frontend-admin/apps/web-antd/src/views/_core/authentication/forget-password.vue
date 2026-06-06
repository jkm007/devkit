<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { Button, Form, FormItem, Input, InputPassword, message, Space } from 'ant-design-vue';

import { resetPasswordApi, sendVerifyCodeApi } from '#/api/core/auth';
import {
  validateCode,
  validateEmail,
  validatePassword,
  validatePasswordMatch,
  VALIDATION_MESSAGES,
} from '#/utils/form-validation';
import { showCaptchaVerify } from '#/utils/captcha-verify';

defineOptions({ name: 'ForgetPassword' });

const router = useRouter();
const loading = ref(false);

// 表单数据
const formData = ref({
  email: '',
  emailCode: '',
  newPassword: '',
  confirmPassword: '',
});

// 验证码倒计时
const codeCountdown = ref(0);
let countdownTimer: ReturnType<typeof setInterval> | null = null;

function startCountdown() {
  codeCountdown.value = 60;
  countdownTimer = setInterval(() => {
    codeCountdown.value--;
    if (codeCountdown.value <= 0 && countdownTimer) {
      clearInterval(countdownTimer);
      countdownTimer = null;
    }
  }, 1000);
}

async function handleSendCode() {
  if (!validateEmail(formData.value.email)) {
    message.warning(VALIDATION_MESSAGES.emailRequired);
    return;
  }
  try {
    // 先弹出图形验证码
    const { captchaId, captchaCode } = await showCaptchaVerify();
    await sendVerifyCodeApi({
      email: formData.value.email,
      purpose: 'reset_password',
      captchaId,
      captchaCode,
    });
    message.success('验证码已发送到您的邮箱');
    startCountdown();
  } catch (e: any) {
    if (e?.message !== '用户取消验证码验证') {
      message.error(e?.message || '发送失败');
    }
  }
}

async function handleSubmit() {
  // 表单校验
  if (!validateEmail(formData.value.email)) {
    message.warning(VALIDATION_MESSAGES.emailRequired);
    return;
  }
  if (!validateCode(formData.value.emailCode)) {
    message.warning(VALIDATION_MESSAGES.codeInvalid);
    return;
  }
  if (!validatePassword(formData.value.newPassword)) {
    message.warning(VALIDATION_MESSAGES.passwordMinLength);
    return;
  }
  if (!validatePasswordMatch(formData.value.newPassword, formData.value.confirmPassword)) {
    message.warning(VALIDATION_MESSAGES.passwordMismatch);
    return;
  }

  loading.value = true;
  try {
    await resetPasswordApi({
      email: formData.value.email,
      emailCode: formData.value.emailCode,
      newPassword: formData.value.newPassword,
      confirmPassword: formData.value.confirmPassword,
    });
    message.success('密码重置成功，请登录');
    router.push('/auth/login');
  } catch (e: any) {
    message.error(e?.message || '重置失败');
  } finally {
    loading.value = false;
  }
}

function goToLogin() {
  router.push('/auth/login');
}
</script>

<template>
  <div class="mx-auto w-full max-w-[400px]">
    <h2 class="mb-6 text-center text-2xl font-bold">重置密码 🔑</h2>

    <Form layout="vertical" :model="formData" @finish="handleSubmit">
      <FormItem label="邮箱" name="email" required>
        <Input
          v-model:value="formData.email"
          placeholder="请输入注册时的邮箱地址"
          size="large"
          type="email"
        />
      </FormItem>

      <FormItem label="验证码" name="emailCode" required>
        <Space style="width: 100%;">
          <Input
            v-model:value="formData.emailCode"
            placeholder="请输入6位验证码"
            :maxlength="6"
            size="large"
            style="flex: 1;"
          />
          <Button
            type="primary"
            size="large"
            :disabled="codeCountdown > 0"
            @click="handleSendCode"
          >
            {{ codeCountdown > 0 ? `${codeCountdown}s` : '发送验证码' }}
          </Button>
        </Space>
      </FormItem>

      <FormItem label="新密码" name="newPassword" required>
        <InputPassword
          v-model:value="formData.newPassword"
          placeholder="请输入新密码（至少6个字符）"
          size="large"
        />
      </FormItem>

      <FormItem label="确认新密码" name="confirmPassword" required>
        <InputPassword
          v-model:value="formData.confirmPassword"
          placeholder="请再次输入新密码"
          size="large"
        />
      </FormItem>

      <FormItem>
        <Button
          type="primary"
          html-type="submit"
          block
          size="large"
          :loading="loading"
        >
          重置密码
        </Button>
      </FormItem>

      <div class="text-center">
        <a class="text-primary" @click="goToLogin">返回登录</a>
      </div>
    </Form>
  </div>
</template>
