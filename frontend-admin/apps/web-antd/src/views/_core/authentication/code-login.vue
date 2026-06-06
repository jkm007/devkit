<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { Button, Form, FormItem, Input, message, Space, TabPane, Tabs } from 'ant-design-vue';

import { sendVerifyCodeApi } from '#/api/core/auth';
import { useAuthStore } from '#/store';
import { showCaptchaVerify } from '#/utils/captcha-verify';

defineOptions({ name: 'CodeLogin' });

const router = useRouter();
const authStore = useAuthStore();

// ==================== Tab ====================
const activeTab = ref('email');

// ==================== 邮箱登录 ====================
const emailForm = ref({
  email: '',
  code: '',
});

// ==================== 手机号登录（预留） ====================
const phoneForm = ref({
  phone: '',
  code: '',
});

// ==================== 验证码倒计时 ====================
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

// ==================== 发送验证码 ====================
async function handleSendEmailCode() {
  if (!emailForm.value.email) {
    message.warning('请先输入邮箱');
    return;
  }
  try {
    const { captchaId, captchaCode } = await showCaptchaVerify();
    await sendVerifyCodeApi({
      email: emailForm.value.email,
      purpose: 'login',
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

// ==================== 提交登录 ====================
async function handleEmailLogin() {
  if (!emailForm.value.email) {
    message.warning('请输入邮箱');
    return;
  }
  if (!emailForm.value.code || emailForm.value.code.length !== 6) {
    message.warning('请输入6位验证码');
    return;
  }

  const result = await authStore.authLoginByEmail({
    email: emailForm.value.email,
    code: emailForm.value.code,
  });
  if (!result?.userInfo) {
    message.error('登录失败，请检查邮箱和验证码');
  }
}

function handlePhoneLogin() {
  // TODO: 第二阶段实现短信验证码登录
  message.info('短信验证码登录即将上线');
}

function goToLogin() {
  router.push('/auth/login');
}
</script>

<template>
  <div class="mx-auto w-full max-w-[400px]">
    <h2 class="mb-2 text-center text-2xl font-bold">验证码登录 📲</h2>
    <p class="mb-6 text-center text-sm text-muted-foreground">
      使用邮箱或手机号验证码快速登录
    </p>

    <Tabs v-model:active-key="activeTab" centered>
      <!-- 邮箱登录 -->
      <TabPane key="email" tab="邮箱登录">
        <Form layout="vertical" :model="emailForm" @finish="handleEmailLogin">
          <FormItem label="邮箱" name="email" required>
            <Input
              v-model:value="emailForm.email"
              placeholder="请输入邮箱地址"
              size="large"
              type="email"
            />
          </FormItem>

          <FormItem label="验证码" name="code" required>
            <Space style="width: 100%;">
              <Input
                v-model:value="emailForm.code"
                placeholder="请输入6位验证码"
                :maxlength="6"
                size="large"
                style="flex: 1;"
              />
              <Button
                type="primary"
                size="large"
                :disabled="codeCountdown > 0"
                @click="handleSendEmailCode"
              >
                {{ codeCountdown > 0 ? `${codeCountdown}s` : '发送验证码' }}
              </Button>
            </Space>
          </FormItem>

          <FormItem>
            <Button
              type="primary"
              html-type="submit"
              block
              size="large"
              :loading="authStore.loginLoading"
            >
              登录
            </Button>
          </FormItem>
        </Form>
      </TabPane>

      <!-- 手机号登录（预留） -->
      <TabPane key="phone" tab="手机号登录">
        <Form layout="vertical" :model="phoneForm" @finish="handlePhoneLogin">
          <FormItem label="手机号" name="phone" required>
            <Input
              v-model:value="phoneForm.phone"
              placeholder="请输入手机号"
              size="large"
              :maxlength="11"
            />
          </FormItem>

          <FormItem label="验证码" name="code" required>
            <Space style="width: 100%;">
              <Input
                v-model:value="phoneForm.code"
                placeholder="请输入6位验证码"
                :maxlength="6"
                size="large"
                style="flex: 1;"
              />
              <Button
                type="primary"
                size="large"
                disabled
              >
                发送验证码
              </Button>
            </Space>
          </FormItem>

          <FormItem>
            <Button
              type="primary"
              html-type="submit"
              block
              size="large"
              disabled
            >
              登录（即将上线）
            </Button>
          </FormItem>
        </Form>
      </TabPane>
    </Tabs>

    <div class="mt-4 text-center">
      <a class="text-sm text-primary cursor-pointer" @click="goToLogin">
        ← 返回密码登录
      </a>
    </div>
  </div>
</template>
