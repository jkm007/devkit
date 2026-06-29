<template>
  <view class="about-page">
    <view class="logo-section">
      <view class="logo">
        <text class="logo-text">📚</text>
      </view>
      <text class="app-name">题小助</text>
      <text class="version">版本：{{ version }}</text>
    </view>

    <view class="content-section">
      <view v-if="aboutContent" class="rich-content" v-html="aboutContent"></view>
      <view v-else class="default-content">
        <text class="desc">一款跨平台的智能学习辅助应用</text>
        <view class="feature-list">
          <text class="feature-title">主要功能</text>
          <text class="feature-item">📝 题库管理 — 支持 28 种题型</text>
          <text class="feature-item">🎯 智能练习 — 基于错题本智能推荐</text>
          <text class="feature-item">📕 错题本 — 自动收录错误题目</text>
          <text class="feature-item">⭐ 收藏功能 — 收藏重点题目</text>
          <text class="feature-item">🏫 班级系统 — 教师/学生班级管理</text>
        </view>
      </view>
    </view>

    <view class="footer">
      <text class="copyright">© 2026 题小助. All rights reserved.</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { request } from '@/api/request';

const aboutContent = ref('');
const version = ref('v1.0.0');

onMounted(async () => {
  try {
    const data = await request.get<any>('/mobile/settings');
    if (data?.aboutUs) {
      aboutContent.value = data.aboutUs;
    }
  } catch {
    // 使用默认内容
  }
  try {
    // #ifdef APP-PLUS
    const versionInfo = plus.runtime.version;
    if (versionInfo) version.value = versionInfo;
    // #endif
  } catch {
    // ignore
  }
});
</script>

<style lang="scss" scoped>
.about-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.logo-section {
  background: linear-gradient(135deg, #1890ff 0%, #36cfc9 100%);
  padding: 40px 16px 30px;
  display: flex;
  flex-direction: column;
  align-items: center;

  .logo {
    width: 72px;
    height: 72px;
    border-radius: 20px;
    background: rgba(255, 255, 255, 0.2);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 12px;

    .logo-text { font-size: 36px; }
  }

  .app-name {
    font-size: 22px;
    font-weight: bold;
    color: #fff;
    margin-bottom: 6px;
  }

  .version {
    font-size: 13px;
    color: rgba(255, 255, 255, 0.7);
  }
}

.content-section {
  margin: 16px;
  background: #fff;
  border-radius: 12px;
  padding: 20px 16px;

  .rich-content {
    font-size: 14px;
    line-height: 1.8;
    color: #333;
  }

  .default-content {
    .desc {
      font-size: 15px;
      color: #666;
      display: block;
      text-align: center;
      margin-bottom: 20px;
    }

    .feature-list {
      .feature-title {
        font-size: 15px;
        font-weight: 500;
        color: #333;
        display: block;
        margin-bottom: 12px;
      }

      .feature-item {
        font-size: 14px;
        color: #555;
        display: block;
        padding: 8px 0;
        border-bottom: 1px solid #f5f5f5;

        &:last-child { border-bottom: none; }
      }
    }
  }
}

.footer {
  text-align: center;
  padding: 30px 16px;

  .copyright {
    font-size: 12px;
    color: #999;
  }
}
</style>
