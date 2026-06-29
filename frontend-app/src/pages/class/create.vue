<template>
  <view class="create-class-page">
    <view class="form-card">
      <view class="form-item">
        <text class="label">班级名称 <text class="required">*</text></text>
        <input
          v-model="form.name"
          class="input"
          placeholder="请输入班级名称"
          maxlength="50"
        />
      </view>
      <view class="form-item">
        <text class="label">班级描述</text>
        <textarea
          v-model="form.description"
          class="textarea"
          placeholder="请输入班级描述（可选）"
          maxlength="200"
        />
      </view>
      <button class="submit-btn" :disabled="submitting" @click="handleSubmit">
        {{ submitting ? '创建中...' : '创建班级' }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { createClass } from '@/api/class';

const form = ref({ name: '', description: '' });
const submitting = ref(false);

function handleSubmit() {
  if (!form.value.name.trim()) {
    uni.showToast({ title: '请输入班级名称', icon: 'none' });
    return;
  }
  submitting.value = true;
  createClass({ name: form.value.name.trim(), description: form.value.description.trim() })
    .then((res) => {
      uni.showToast({ title: '创建成功', icon: 'success' });
      setTimeout(() => {
        uni.redirectTo({ url: `/pages/class/detail?id=${res.id}` });
      }, 800);
    })
    .catch((e: any) => {
      uni.showToast({ title: e?.message || '创建失败', icon: 'none' });
    })
    .finally(() => {
      submitting.value = false;
    });
}
</script>

<style lang="scss" scoped>
.create-class-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 16px;
}

.form-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px 16px;
}

.form-item {
  margin-bottom: 20px;

  .label {
    font-size: 14px;
    color: #333;
    font-weight: 500;
    display: block;
    margin-bottom: 8px;

    .required { color: #ff4d4f; }
  }

  .input {
    width: 100%;
    background: #f5f7fa;
    border: 1px solid #e8e8e8;
    border-radius: 8px;
    padding: 10px 12px;
    font-size: 14px;
  }

  .textarea {
    width: 100%;
    background: #f5f7fa;
    border: 1px solid #e8e8e8;
    border-radius: 8px;
    padding: 10px 12px;
    font-size: 14px;
    min-height: 80px;
  }
}

.submit-btn {
  background: #1890ff;
  color: #fff;
  border: none;
  border-radius: 8px;
  padding: 12px;
  font-size: 16px;
  margin-top: 24px;

  &:disabled { opacity: 0.6; }
  &:active:not(:disabled) { opacity: 0.8; }
}
</style>
