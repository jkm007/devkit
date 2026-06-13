<template>
  <view class="feedback-page">
    <view class="form-section">
      <view class="form-item">
        <text class="label">题目ID</text>
        <input class="input" type="number" v-model="form.questionId" placeholder="请输入题目ID" :disabled="isQuestionIdLocked" />
        <text v-if="isQuestionIdLocked" class="id-hint">题目ID已锁定，无法修改</text>
      </view>

      <view class="form-item">
        <text class="label">反馈类型</text>
        <view class="type-list">
          <view v-for="t in typeOptions" :key="t.value" class="type-btn" :class="{ active: form.feedbackType === t.value }" @click="form.feedbackType = t.value">
            {{ t.label }}
          </view>
        </view>
      </view>

      <view class="form-item">
        <text class="label">问题描述</text>
        <textarea class="textarea" v-model="form.description" placeholder="请详细描述您发现的问题" maxlength="500" />
        <text class="char-count">{{ form.description.length }}/500</text>
      </view>

      <view class="form-item">
        <text class="label">修改建议（选填）</text>
        <textarea class="textarea" v-model="form.suggestion" placeholder="您认为正确的答案或建议" maxlength="500" />
      </view>

      <button class="submit-btn" type="primary" :disabled="!canSubmit" @click="submitFeedback">
        提交纠错
      </button>
    </view>

    <!-- 我的纠错记录 -->
    <view class="history-section" v-if="historyList.length">
      <text class="section-title">我的纠错记录</text>
      <view class="history-list">
        <view v-for="h in historyList" :key="h.id" class="history-item" @click="goToDetail(h.id)">
          <view class="item-header">
            <text class="item-type">{{ getTypeLabel(h.feedbackType) }}</text>
            <text class="item-status" :class="h.status">{{ getStatusLabel(h.status) }}</text>
          </view>
          <text class="item-desc">{{ h.description }}</text>
          <text class="item-time">{{ formatDate(h.createdAt) }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { submitFeedback, getFeedbacks, type FeedbackItem } from '@/api/feedback';

const form = ref({
  questionId: 0,
  feedbackType: 'answer_error',
  description: '',
  suggestion: '',
});

// questionId 是否从 URL 参数填充（填充后禁止修改，防止 IDOR）
const isQuestionIdLocked = ref(false);

const historyList = ref<FeedbackItem[]>([]);

const typeOptions = [
  { label: '答案错误', value: 'answer_error' },
  { label: '内容错误', value: 'content_error' },
  { label: '选项错误', value: 'option_error' },
  { label: '其他', value: 'other' },
];

const canSubmit = computed(() => {
  return form.value.questionId > 0 && form.value.description.trim().length > 5;
});

onMounted(() => {
  const pages = getCurrentPages();
  const currentPage = pages[pages.length - 1] as any;
  const qid = currentPage.options?.questionId;
  if (qid) {
    form.value.questionId = parseInt(qid);
    isQuestionIdLocked.value = true; // 从 URL 填充后锁定，防止篡改
  }
  loadHistory();
});

async function loadHistory() {
  try {
    const res = await getFeedbacks({ page: 1, pageSize: 10 });
    historyList.value = res.items || [];
  } catch { /* ignore */ }
}

async function submitFeedback() {
  if (!canSubmit.value) return;
  uni.showLoading({ title: '提交中...' });
  try {
    await submitFeedback({
      questionId: form.value.questionId,
      feedbackType: form.value.feedbackType,
      description: form.value.description,
      suggestion: form.value.suggestion || undefined,
    });
    uni.showToast({ title: '提交成功', icon: 'success' });
    form.value.description = '';
    form.value.suggestion = '';
    loadHistory();
  } catch (e: any) {
    uni.showToast({ title: e.message || '提交失败', icon: 'none' });
  } finally {
    uni.hideLoading();
  }
}

function goToDetail(id: number) {
  uni.navigateTo({ url: `/pages/question/feedback-detail?id=${id}` });
}

function getTypeLabel(type: string): string {
  const map: Record<string, string> = { answer_error: '答案错误', content_error: '内容错误', option_error: '选项错误', other: '其他' };
  return map[type] || type;
}

function getStatusLabel(status: string): string {
  const map: Record<string, string> = { pending: '待处理', processing: '处理中', resolved: '已解决', rejected: '已驳回' };
  return map[status] || status;
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}
</script>

<style lang="scss" scoped>
.feedback-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 30px;

  .form-section {
    margin: 16px;
    background: #fff;
    border-radius: 12px;
    padding: 20px 16px;

    .form-item {
      margin-bottom: 20px;

      .label {
        font-size: 14px;
        font-weight: 500;
        color: #333;
        margin-bottom: 8px;
        display: block;
      }

      .input {
        width: 100%;
        height: 44px;
        background: #f5f7fa;
        border-radius: 8px;
        padding: 0 12px;
        font-size: 15px;
      }

      .id-hint {
        font-size: 12px;
        color: #1890ff;
        margin-top: 4px;
        display: block;
      }

      .type-list {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;

        .type-btn {
          padding: 8px 16px;
          background: #f5f7fa;
          border-radius: 16px;
          font-size: 13px;
          color: #666;

          &.active {
            background: #1890ff;
            color: #fff;
          }
        }
      }

      .textarea {
        width: 100%;
        min-height: 100px;
        background: #f5f7fa;
        border-radius: 8px;
        padding: 12px;
        font-size: 14px;
        box-sizing: border-box;
      }

      .char-count {
        font-size: 12px;
        color: #999;
        text-align: right;
        margin-top: 4px;
        display: block;
      }
    }

    .submit-btn {
      height: 44px;
      line-height: 44px;
      background: linear-gradient(90deg, #1890ff, #36cfc9);
      color: #fff;
      border: none;
      border-radius: 22px;
      font-size: 15px;

      &[disabled] { opacity: 0.5; }
    }
  }

  .history-section {
    margin: 0 16px;

    .section-title {
      font-size: 16px;
      font-weight: 500;
      color: #333;
      margin-bottom: 12px;
      display: block;
    }

    .history-list {
      .history-item {
        background: #fff;
        border-radius: 12px;
        padding: 14px;
        margin-bottom: 10px;

        .item-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 6px;

          .item-type {
            font-size: 12px;
            padding: 2px 8px;
            background: #e6f7ff;
            color: #1890ff;
            border-radius: 4px;
          }

          .item-status {
            font-size: 12px;
            &.pending { color: #faad14; }
            &.processing { color: #1890ff; }
            &.resolved { color: #52c41a; }
            &.rejected { color: #ff4d4f; }
          }
        }

        .item-desc {
          font-size: 14px;
          color: #333;
          line-height: 1.4;
          margin-bottom: 4px;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
          overflow: hidden;
        }

        .item-time {
          font-size: 11px;
          color: #999;
        }
      }
    }
  }
}
</style>
