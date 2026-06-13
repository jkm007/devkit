<template>
  <view class="register-page">
    <view class="header">
      <text class="title">注册账号</text>
    </view>

    <view class="form">
      <view class="input-wrapper">
        <input v-model="form.username" class="input" placeholder="用户名" />
      </view>
      <view class="input-wrapper">
        <input v-model="form.password" class="input" type="password" placeholder="密码" />
      </view>
      <view class="input-wrapper">
        <input v-model="form.confirmPassword" class="input" type="password" placeholder="确认密码" />
      </view>
      <view class="input-wrapper">
        <input v-model="form.email" class="input" type="email" placeholder="邮箱（可选）" />
      </view>
      <view class="input-wrapper">
        <input v-model="form.phone" class="input" type="tel" placeholder="手机号（可选）" />
      </view>

      <button
        class="register-btn"
        type="primary"
        :loading="loading"
        :disabled="loading || !canSubmit"
        @click="handleRegister"
      >
        注册
      </button>
    </view>

    <view class="footer">
      <text class="hint">已有账号?</text>
      <navigator url="/pages/login/index" class="link">立即登录</navigator>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { register } from '@/api/auth';

const form = ref({
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
  phone: '',
});
const loading = ref(false);

const canSubmit = computed(() => {
  const pwd = form.value.password;
  const hasMinLength = pwd.length >= 8;
  return (
    form.value.username &&
    hasMinLength &&
    pwd === form.value.confirmPassword
  );
});

async function handleRegister() {
  if (!canSubmit.value) {
    if (form.value.password.length < 8) {
      uni.showToast({ title: '密码至少8位', icon: 'none' });
      return;
    }
    if (form.value.password !== form.value.confirmPassword) {
      uni.showToast({ title: '两次密码不一致', icon: 'none' });
      return;
    }
    return;
  }

  loading.value = true;
  try {
    await register({
      username: form.value.username,
      password: form.value.password,
      email: form.value.email || undefined,
      phone: form.value.phone || undefined,
    });
    uni.showToast({ title: '注册成功', icon: 'success' });
    setTimeout(() => {
      uni.navigateTo({ url: '/pages/login/index' });
    }, 1000);
  } catch (err: any) {
    uni.showToast({ title: err.message || '注册失败', icon: 'none' });
  } finally {
    loading.value = false;
  }
}
</script>

<style lang="scss" scoped>
.register-page {
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

    .register-btn {
      height: 48px;
      line-height: 48px;
      font-size: 17px;
      border-radius: 24px;
      margin-top: 24px;
      background: #1890ff;
      border: none;
    }
  }

  .footer {
    margin-top: 20px;
    text-align: center;

    .hint {
      color: #666;
      font-size: 14px;
    }

    .link {
      color: #1890ff;
      font-size: 14px;
      margin-left: 8px;
    }
  }
}
</style>
