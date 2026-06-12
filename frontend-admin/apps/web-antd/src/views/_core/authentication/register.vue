<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';

import {
  Button,
  Form,
  FormItem,
  Input,
  InputPassword,
  message,
  Space,
} from 'ant-design-vue';

import { registerApi, sendVerifyCodeApi } from '#/api/core/auth';
import {
  validateCode,
  validateEmail,
  validatePassword,
  validatePasswordMatch,
  validateUsername,
  VALIDATION_MESSAGES,
} from '#/utils/form-validation';
import { showCaptchaVerify } from '#/utils/captcha-verify';

defineOptions({ name: 'Register' });

const router = useRouter();
const loading = ref(false);

// 表单数据
const formData = ref({
  username: '',
  email: '',
  emailCode: '',
  password: '',
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
    const { captchaId, captchaCode, startTime } = await showCaptchaVerify();
    await sendVerifyCodeApi({
      email: formData.value.email,
      purpose: 'register',
      captchaId,
      captchaCode,
      startTime,
    });
    message.success('验证码已发送到您的邮箱');
    startCountdown();
  } catch (e: any) {
    // 错误消息由 request.ts 的 errorMessageResponseInterceptor 统一显示
    // 这里只处理用户取消的情况，避免重复提示
    if (e?.message === '用户取消验证码验证') {
      // 用户主动取消，不提示
    }
  }
}

async function handleSubmit() {
  // 表单校验
  if (!validateUsername(formData.value.username)) {
    message.warning(VALIDATION_MESSAGES.usernameMinLength);
    return;
  }
  if (!validateEmail(formData.value.email)) {
    message.warning(VALIDATION_MESSAGES.emailRequired);
    return;
  }
  if (!validateCode(formData.value.emailCode)) {
    message.warning(VALIDATION_MESSAGES.codeInvalid);
    return;
  }
  if (!validatePassword(formData.value.password)) {
    message.warning(VALIDATION_MESSAGES.passwordMinLength);
    return;
  }
  if (
    !validatePasswordMatch(
      formData.value.password,
      formData.value.confirmPassword,
    )
  ) {
    message.warning(VALIDATION_MESSAGES.passwordMismatch);
    return;
  }

  loading.value = true;
  try {
    await registerApi({
      username: formData.value.username,
      email: formData.value.email,
      emailCode: formData.value.emailCode,
      password: formData.value.password,
      confirmPassword: formData.value.confirmPassword,
      registerSource: 'web',
    });
    message.success('注册成功，请登录');
    router.push('/auth/login');
  } catch (e: any) {
    message.error(e?.message || '注册失败');
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
    <h2 class="mb-6 text-center text-2xl font-bold">注册账号 🚀</h2>

    <Form layout="vertical" :model="formData" @finish="handleSubmit">
      <FormItem label="用户名" name="username" required>
        <Input
          v-model:value="formData.username"
          placeholder="请输入用户名（至少3个字符）"
          size="large"
        />
      </FormItem>

      <FormItem label="邮箱" name="email" required>
        <Input
          v-model:value="formData.email"
          placeholder="请输入邮箱地址"
          size="large"
          type="email"
        />
      </FormItem>

      <FormItem label="验证码" name="emailCode" required>
        <Space style="width: 100%">
          <Input
            v-model:value="formData.emailCode"
            placeholder="请输入6位验证码"
            :maxlength="6"
            size="large"
            style="flex: 1"
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

      <FormItem label="密码" name="password" required>
        <InputPassword
          v-model:value="formData.password"
          placeholder="请输入密码（至少6个字符）"
          size="large"
        />
      </FormItem>

      <FormItem label="确认密码" name="confirmPassword" required>
        <InputPassword
          v-model:value="formData.confirmPassword"
          placeholder="请再次输入密码"
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
          注册
        </Button>
      </FormItem>

      <div class="text-center">
        <span class="text-muted-foreground">已有账号？</span>
        <a class="ml-1 text-primary" @click="goToLogin">去登录</a>
      </div>
    </Form>
  </div>
</template>
