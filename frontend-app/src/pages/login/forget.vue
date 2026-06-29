<template>
  <view class="forget-page">
    <view class="header">
      <text class="title">忘记密码</text>
    </view>

    <view class="form">
      <view class="input-wrapper">
        <input v-model="form.username" class="input" placeholder="用户名/邮箱/手机号" />
      </view>
      <view class="input-wrapper code-row">
        <input v-model="form.code" class="input" placeholder="验证码" />
        <button
          class="send-code-btn"
          :disabled="sendingCode || countdown > 0"
          @click="sendCode"
        >
          {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
        </button>
      </view>
      <view class="input-wrapper">
        <input v-model="form.newPassword" class="input" type="password" placeholder="新密码" />
      </view>
      <view class="input-wrapper">
        <input v-model="form.confirmPassword" class="input" type="password" placeholder="确认新密码" />
      </view>

      <button
        class="submit-btn"
        :loading="loading"
        :disabled="loading || !canSubmit"
        @click="handleSubmit"
      >
        重置密码
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue';
import { sendVerifyCode, resetPassword } from '@/api/auth';

const form = ref({
  username: '',
  code: '',
  newPassword: '',
  confirmPassword: '',
});
const loading = ref(false);
const sendingCode = ref(false);
const countdown = ref(0);
let countdownTimer: ReturnType<typeof setInterval> | null = null;
let lastSendTime = 0;
const COOLDOWN_MS = 60000; // 60秒冷却

const canSubmit = computed(() => {
  return (
    form.value.username &&
    form.value.code &&
    form.value.newPassword &&
    form.value.newPassword === form.value.confirmPassword
  );
});

async function sendCode() {
  // 先检查冷却时间
  const now = Date.now();
  if (now - lastSendTime < COOLDOWN_MS) {
    const remaining = Math.ceil((COOLDOWN_MS - (now - lastSendTime)) / 1000);
    uni.showToast({ title: `请 ${remaining} 秒后再试`, icon: 'none' });
    return;
  }

  if (!form.value.username) {
    uni.showToast({ title: '请输入用户名/邮箱/手机号', icon: 'none' });
    return;
  }
  sendingCode.value = true;
  try {
    await sendVerifyCode(form.value.username);
    lastSendTime = Date.now();
    countdown.value = 60;
    countdownTimer = setInterval(() => {
      countdown.value--;
      if (countdown.value <= 0) {
        clearInterval(countdownTimer!);
        countdownTimer = null;
      }
    }, 1000);
    uni.showToast({ title: '验证码已发送', icon: 'success' });
  } catch (e: any) {
    uni.showToast({ title: e.message || '发送失败', icon: 'none' });
  } finally {
    sendingCode.value = false;
  }
}

onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer);
    countdownTimer = null;
  }
});

async function handleSubmit() {
  if (!canSubmit.value) {
    if (form.value.newPassword !== form.value.confirmPassword) {
      uni.showToast({ title: '两次密码不一致', icon: 'none' });
      return;
    }
    return;
  }

  loading.value = true;
  try {
    await resetPassword({
      email: form.value.username,
      code: form.value.code,
      newPassword: form.value.newPassword,
    });
    uni.showToast({ title: '密码重置成功', icon: 'success' });
    setTimeout(() => {
      uni.navigateTo({ url: '/pages/login/index' });
    }, 1000);
  } catch (err: any) {
    uni.showToast({ title: err.message || '重置失败', icon: 'none' });
  } finally {
    loading.value = false;
  }
}
</script>

<style lang="scss" scoped>
.forget-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 40px 20px;

  .header {
    margin-bottom: 40px;

    .title {
      font-size: 28px;
      font-weight: bold;
      color: #333;
    }
  }

  .form {
    background: #fff;
    border-radius: 16px;
    padding: 24px 20px;

    .input-wrapper {
      background: #f5f7fa;
      border-radius: 8px;
      padding: 0 16px;
      margin-bottom: 12px;

      .input {
        height: 48px;
        font-size: 15px;
        color: #333;
      }
    }

    .code-row {
      display: flex;
      align-items: center;
      gap: 12px;

      .input {
        flex: 1;
      }

      .send-code-btn {
        width: 120px;
        height: 48px;
        line-height: 48px;
        font-size: 14px;
        background: #1890ff;
        color: #fff;
        border: none;
        border-radius: 8px;

        &[disabled] {
          background: #ccc;
          color: #999;
        }
      }
    }

    .submit-btn {
      height: 48px;
      line-height: 48px;
      font-size: 17px;
      border-radius: 24px;
      margin-top: 24px;
      background: #1890ff;
      border: none;
    }
  }
}
</style>
