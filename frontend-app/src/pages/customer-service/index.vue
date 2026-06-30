<template>
  <view class="customer-service-page">
    <view class="header-card">
      <text class="header-icon">💬</text>
      <text class="header-title">在线客服</text>
      <text class="header-subtitle">学习过程中遇到问题，随时联系我们</text>
    </view>

    <view class="section">
      <text class="section-title">联系方式</text>
      <view class="contact-list">
        <view class="contact-item" @click="copyText(servicePhone)">
          <text class="contact-icon">📞</text>
          <view class="contact-info">
            <text class="contact-label">客服电话</text>
            <text class="contact-value">{{ servicePhone || '暂未配置' }}</text>
          </view>
          <text class="contact-action">复制</text>
        </view>

        <view class="contact-item" @click="copyText(serviceEmail)">
          <text class="contact-icon">✉️</text>
          <view class="contact-info">
            <text class="contact-label">客服邮箱</text>
            <text class="contact-value">{{ serviceEmail || '暂未配置' }}</text>
          </view>
          <text class="contact-action">复制</text>
        </view>

        <view class="contact-item" @click="copyText(serviceWechat)">
          <text class="contact-icon">💬</text>
          <view class="contact-info">
            <text class="contact-label">客服微信</text>
            <text class="contact-value">{{ serviceWechat || '暂未配置' }}</text>
          </view>
          <text class="contact-action">复制</text>
        </view>
      </view>
    </view>

    <view class="section">
      <text class="section-title">服务时间</text>
      <view class="time-card">
        <text class="time-value">周一至周五 09:00 - 18:00</text>
        <text class="time-tip">非工作时间请留言，我们会尽快回复</text>
      </view>
    </view>

    <view class="section">
      <text class="section-title">常见问题</text>
      <view class="faq-list">
        <view
          v-for="(item, index) in faqList"
          :key="index"
          class="faq-item"
          @click="toggleFaq(index)"
        >
          <view class="faq-question">
            <text class="faq-q">Q</text>
            <text class="faq-text">{{ item.question }}</text>
            <text class="faq-arrow" :class="{ open: activeFaq === index }">›</text>
          </view>
          <view v-if="activeFaq === index" class="faq-answer">
            <text>{{ item.answer }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const servicePhone = ref('');
const serviceEmail = ref('');
const serviceWechat = ref('');

const activeFaq = ref<number | null>(null);

const faqList = ref([
  {
    question: '如何找回密码？',
    answer: '在登录页点击「忘记密码」，按照提示完成手机号/邮箱验证后即可重置密码。',
  },
  {
    question: '题目答案有误怎么办？',
    answer: '进入题目详情页，点击「纠错」按钮提交反馈，我们会尽快审核处理。',
  },
  {
    question: '如何加入班级？',
    answer: '在「我的」-「我的班级」中，点击「加入班级」并输入班级邀请码即可。',
  },
]);

function toggleFaq(index: number) {
  activeFaq.value = activeFaq.value === index ? null : index;
}

function copyText(text: string) {
  if (!text) {
    uni.showToast({ title: '暂未配置', icon: 'none' });
    return;
  }
  uni.setClipboardData({
    data: text,
    success: () => {
      uni.showToast({ title: '已复制', icon: 'success' });
    },
    fail: () => {
      uni.showToast({ title: '复制失败', icon: 'none' });
    },
  });
}
</script>

<style lang="scss" scoped>
.customer-service-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 30px;
}

.header-card {
  background: linear-gradient(135deg, #1890ff 0%, #36cfc9 100%);
  padding: 40px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #fff;

  .header-icon {
    font-size: 48px;
    margin-bottom: 12px;
  }

  .header-title {
    font-size: 22px;
    font-weight: bold;
    margin-bottom: 8px;
  }

  .header-subtitle {
    font-size: 13px;
    opacity: 0.9;
  }
}

.section {
  margin: 16px;

  .section-title {
    font-size: 15px;
    font-weight: 500;
    color: #333;
    margin-bottom: 12px;
    display: block;
  }
}

.contact-list,
.time-card,
.faq-list {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
}

.contact-item {
  display: flex;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }

  &:active {
    background: #f9f9f9;
  }

  .contact-icon {
    font-size: 24px;
    margin-right: 12px;
  }

  .contact-info {
    flex: 1;
    display: flex;
    flex-direction: column;

    .contact-label {
      font-size: 12px;
      color: #999;
      margin-bottom: 4px;
    }

    .contact-value {
      font-size: 15px;
      color: #333;
    }
  }

  .contact-action {
    font-size: 13px;
    color: #1890ff;
  }
}

.time-card {
  padding: 16px;
  display: flex;
  flex-direction: column;

  .time-value {
    font-size: 15px;
    color: #333;
    margin-bottom: 6px;
  }

  .time-tip {
    font-size: 12px;
    color: #999;
  }
}

.faq-item {
  border-bottom: 1px solid #f5f5f5;

  &:last-child {
    border-bottom: none;
  }

  .faq-question {
    display: flex;
    align-items: center;
    padding: 16px;

    .faq-q {
      width: 20px;
      height: 20px;
      line-height: 20px;
      text-align: center;
      background: #e6f7ff;
      color: #1890ff;
      border-radius: 4px;
      font-size: 12px;
      margin-right: 10px;
      flex-shrink: 0;
    }

    .faq-text {
      flex: 1;
      font-size: 14px;
      color: #333;
    }

    .faq-arrow {
      font-size: 18px;
      color: #ccc;
      transform: rotate(90deg);
      transition: transform 0.2s;

      &.open {
        transform: rotate(-90deg);
      }
    }
  }

  .faq-answer {
    padding: 0 16px 16px 46px;

    text {
      font-size: 13px;
      color: #666;
      line-height: 1.6;
    }
  }
}
</style>
