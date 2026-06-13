<template>
  <view class="login-page">
    <view class="header">
      <text class="title">登录</text>
    </view>
    <view class="form">
      <input class="input" placeholder="用户名/邮箱/手机号" />
      <input class="input" type="password" placeholder="密码" />
      <button class="btn" type="primary">登录</button>
    </view>
    <view class="footer">
      <navigator url="/pages/login/register">注册账号</navigator>
      <navigator url="/pages/login/forget">忘记密码?</navigator>
    </view>
  </view>
</template>

<script setup lang="ts">
// 登录页 - Phase 3 开发
</script>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh;
  background-color: #fff;
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
    .input {
      border: 1px solid #ddd;
      border-radius: 8px;
      padding: 12px 16px;
      margin-bottom: 16px;
      font-size: 16px;
    }

    .btn {
      margin-top: 20px;
    }
  }

  .footer {
    margin-top: 20px;
    display: flex;
    justify-content: space-between;

    navigator {
      color: #1890ff;
      font-size: 14px;
    }
  }
}
</style>
