<template>
  <view class="login-page">
    <!-- Logo -->
    <view class="logo-section">
      <image class="logo" src="/static/logo.png" mode="aspectFit" />
      <text class="app-name">题小助</text>
    </view>

    <!-- 登录表单 -->
    <view class="form-section">
      <!-- 登录方式切换 -->
      <view class="login-tabs">
        <view
          v-for="tab in loginTabs"
          :key="tab.key"
          class="tab-item"
          :class="{ active: activeTab === tab.key }"
          @click="activeTab = tab.key"
        >
          <text>{{ tab.label }}</text>
        </view>
      </view>

      <!-- 用户名密码登录 -->
      <view v-if="activeTab === 'password'" class="form-group">
        <view class="input-wrapper">
          <input
            v-model="form.username"
            class="input"
            placeholder="用户名/邮箱/手机号"
            :disabled="loading"
          />
        </view>
        <view class="input-wrapper">
          <input
            v-model="form.password"
            class="input"
            type="password"
            placeholder="密码"
            :disabled="loading"
          />
        </view>
        <!-- 验证码 -->
        <view v-if="showCaptcha" class="captcha-row">
          <view class="input-wrapper captcha-input">
            <input
              v-model="form.captchaCode"
              class="input"
              placeholder="验证码"
              :disabled="loading"
            />
          </view>
          <image
            v-if="captchaImage"
            class="captcha-img"
            :src="captchaImage"
            mode="aspectFit"
            @click="refreshCaptcha"
          />
        </view>
      </view>

      <!-- 邮箱验证码登录 -->
      <view v-if="activeTab === 'email'" class="form-group">
        <view class="input-wrapper">
          <input
            v-model="form.email"
            class="input"
            type="email"
            placeholder="邮箱地址"
            :disabled="loading"
          />
        </view>
        <view class="input-wrapper code-row">
          <input
            v-model="form.code"
            class="input"
            placeholder="验证码"
            :disabled="loading"
          />
          <button
            class="send-code-btn"
            :disabled="sendingCode || countdown > 0"
            @click="sendEmailCode"
          >
            {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
          </button>
        </view>
      </view>

      <!-- 手机验证码登录 -->
      <view v-if="activeTab === 'phone'" class="form-group">
        <view class="input-wrapper">
          <input
            v-model="form.phone"
            class="input"
            type="tel"
            placeholder="手机号"
            :disabled="loading"
          />
        </view>
        <view class="input-wrapper code-row">
          <input
            v-model="form.code"
            class="input"
            placeholder="验证码"
            :disabled="loading"
          />
          <button
            class="send-code-btn"
            :disabled="sendingCode || countdown > 0"
            @click="sendSmsCode"
          >
            {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
          </button>
        </view>
      </view>

      <!-- 登录按钮 -->
      <button
        class="login-btn"
        type="primary"
        :loading="loading"
        :disabled="loading || !canSubmit"
        @click="handleLogin"
      >
        登录
      </button>

      <!-- 底部链接 -->
      <view class="footer-links">
        <navigator url="/pages/login/register" class="link">注册账号</navigator>
        <navigator url="/pages/login/forget" class="link">忘记密码?</navigator>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue';
import { useUserStore } from '@/store/user';
import { getCaptcha } from '@/api/auth';

const userStore = useUserStore();

// 登录方式
const loginTabs = [
  { key: 'password', label: '密码登录' },
  { key: 'email', label: '邮箱登录' },
  { key: 'phone', label: '手机登录' },
];
const activeTab = ref('password');

// 表单数据
const form = ref({
  username: '',
  password: '',
  email: '',
  phone: '',
  code: '',
  captchaCode: '',
});

// 状态
const loading = ref(false);
const sendingCode = ref(false);
const countdown = ref(0);
const showCaptcha = ref(false);
const captchaImage = ref('');
const captchaId = ref('');
let countdownTimer: ReturnType<typeof setInterval> | null = null;

// 是否可以提交
const canSubmit = computed(() => {
  if (activeTab.value === 'password') {
    return form.value.username && form.value.password;
  }
  if (activeTab.value === 'email') {
    return form.value.email && form.value.code;
  }
  if (activeTab.value === 'phone') {
    return form.value.phone && form.value.code;
  }
  return false;
});

/**
 * 登录处理
 */
async function handleLogin() {
  if (!canSubmit.value) return;

  loading.value = true;
  try {
    if (activeTab.value === 'password') {
      await userStore.loginWithPassword(
        form.value.username,
        form.value.password,
        showCaptcha.value ? captchaId.value : undefined,
        showCaptcha.value ? form.value.captchaCode : undefined,
      );
    } else if (activeTab.value === 'email') {
      await userStore.loginWithEmail(form.value.email, form.value.code);
    } else if (activeTab.value === 'phone') {
      await userStore.loginWithPhone(form.value.phone, form.value.code);
    }

    // 登录成功，跳转首页
    uni.switchTab({ url: '/pages/index/index' });
  } catch (err: any) {
    uni.showToast({ title: err.message || '登录失败', icon: 'none' });

    // 如果需要验证码
    if (err.code === 403001) {
      showCaptcha.value = true;
      await refreshCaptcha();
    }
  } finally {
    loading.value = false;
  }
}

/**
 * 发送邮箱验证码
 */
async function sendEmailCode() {
  if (!form.value.email) {
    uni.showToast({ title: '请输入邮箱', icon: 'none' });
    return;
  }

  sendingCode.value = true;
  try {
    // TODO: 调用发送验证码接口
    startCountdown();
    uni.showToast({ title: '验证码已发送', icon: 'success' });
  } catch (err: any) {
    uni.showToast({ title: err.message || '发送失败', icon: 'none' });
  } finally {
    sendingCode.value = false;
  }
}

/**
 * 发送短信验证码
 */
async function sendSmsCode() {
  if (!form.value.phone) {
    uni.showToast({ title: '请输入手机号', icon: 'none' });
    return;
  }

  sendingCode.value = true;
  try {
    // TODO: 调用发送短信接口
    startCountdown();
    uni.showToast({ title: '验证码已发送', icon: 'success' });
  } catch (err: any) {
    uni.showToast({ title: err.message || '发送失败', icon: 'none' });
  } finally {
    sendingCode.value = false;
  }
}

/**
 * 获取验证码图片
 */
async function refreshCaptcha() {
  try {
    const res = await getCaptcha();
    captchaId.value = res.id;
    captchaImage.value = res.image;
  } catch {
    // 忽略
  }
}

/**
 * 开始倒计时
 */
function startCountdown() {
  countdown.value = 60;
  countdownTimer = setInterval(() => {
    countdown.value--;
    if (countdown.value <= 0) {
      clearInterval(countdownTimer!);
      countdownTimer = null;
    }
  }, 1000);
}

onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer);
    countdownTimer = null;
  }
});
</script>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #1890ff 0%, #f5f5f5 40%);
  padding: 60px 30px 30px;

  .logo-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-bottom: 50px;

    .logo {
      width: 80px;
      height: 80px;
      margin-bottom: 16px;
    }

    .app-name {
      font-size: 28px;
      font-weight: bold;
      color: #fff;
    }
  }

  .form-section {
    background: #fff;
    border-radius: 16px;
    padding: 30px 20px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);

    .login-tabs {
      display: flex;
      margin-bottom: 24px;
      border-bottom: 1px solid #eee;

      .tab-item {
        flex: 1;
        text-align: center;
        padding: 12px 0;
        font-size: 16px;
        color: #666;
        position: relative;

        &.active {
          color: #1890ff;
          font-weight: 500;

          &::after {
            content: '';
            position: absolute;
            bottom: -1px;
            left: 50%;
            transform: translateX(-50%);
            width: 40px;
            height: 3px;
            background: #1890ff;
            border-radius: 2px;
          }
        }
      }
    }

    .form-group {
      margin-bottom: 16px;

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

      .captcha-row {
        display: flex;
        gap: 12px;

        .captcha-input {
          flex: 1;
          margin-bottom: 0;
        }

        .captcha-img {
          width: 120px;
          height: 48px;
          border-radius: 8px;
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
    }

    .login-btn {
      height: 48px;
      line-height: 48px;
      font-size: 17px;
      border-radius: 24px;
      margin-top: 24px;
      background: linear-gradient(90deg, #1890ff, #36cfc9);
      border: none;

      &[disabled] {
        opacity: 0.6;
      }
    }

    .footer-links {
      display: flex;
      justify-content: space-between;
      margin-top: 20px;

      .link {
        font-size: 14px;
        color: #1890ff;
      }
    }
  }
}
</style>
