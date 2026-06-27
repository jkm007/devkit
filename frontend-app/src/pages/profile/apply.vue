<template>
  <view class="apply-page">
    <!-- 可申请角色列表 -->
    <view class="section">
      <text class="section-title">可申请角色</text>
      <view v-if="roleLoading" class="loading">
        <text>加载中...</text>
      </view>
      <view v-else-if="roles.length === 0" class="empty">
        <text>暂无可申请的角色</text>
      </view>
      <view v-else class="role-list">
        <view
          v-for="role in roles"
          :key="role.id"
          class="role-item"
          :class="{ selected: selectedRole?.id === role.id }"
          @click="selectedRole = role"
        >
          <view class="role-info">
            <text class="role-name">{{ role.name }}</text>
            <text class="role-desc">{{ role.remark }}</text>
          </view>
          <text class="role-check" :class="{ active: selectedRole?.id === role.id }">
            {{ selectedRole?.id === role.id ? '✓' : '○' }}
          </text>
        </view>
      </view>
    </view>

    <!-- 申请理由 -->
    <view class="section" v-if="selectedRole">
      <text class="section-title">申请理由</text>
      <textarea
        v-model="reason"
        class="reason-input"
        placeholder="请简要说明申请理由..."
        maxlength="500"
        :auto-height="false"
        :show-confirm-bar="false"
      />
      <text class="char-count">{{ reason.length }}/500</text>
    </view>

    <!-- 提交按钮 -->
    <view class="action-section" v-if="selectedRole">
      <button class="submit-btn" :disabled="submitting" @click="handleSubmit">
        {{ submitting ? '提交中...' : '提交申请' }}
      </button>
    </view>

    <!-- 我的申请记录 -->
    <view class="section">
      <text class="section-title">我的申请记录</text>
      <view v-if="appLoading" class="loading">
        <text>加载中...</text>
      </view>
      <view v-else-if="applications.length === 0" class="empty">
        <text>暂无申请记录</text>
      </view>
      <view v-else class="app-list">
        <view
          v-for="app in applications"
          :key="app.id"
          class="app-item"
        >
          <view class="app-header">
            <text class="app-role">{{ app.roleName }}</text>
            <text class="app-status" :class="app.status">{{ getStatusText(app.status) }}</text>
          </view>
          <text class="app-reason">{{ app.reason || '无' }}</text>
          <text class="app-time">{{ formatTime(app.createdAt) }}</text>
          <text v-if="app.reviewNote" class="app-review">审核备注: {{ app.reviewNote }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { request } from '@/api/request';

interface Role {
  id: number;
  name: string;
  remark: string;
}

interface Application {
  id: number;
  roleName: string;
  reason: string;
  status: number;
  reviewNote: string;
  createdAt: string;
}

const roles = ref<Role[]>([]);
const applications = ref<Application[]>([]);
const selectedRole = ref<Role | null>(null);
const reason = ref('');
const submitting = ref(false);
const roleLoading = ref(false);
const appLoading = ref(false);

onMounted(() => {
  loadRoles();
  loadApplications();
});

async function loadRoles() {
  roleLoading.value = true;
  try {
    const res = await request.get<Role[]>('/auth/role-applications/available-roles');
    roles.value = Array.isArray(res) ? res : [];
  } catch {
    roles.value = [];
  } finally {
    roleLoading.value = false;
  }
}

async function loadApplications() {
  appLoading.value = true;
  try {
    const res = await request.get<any>('/auth/role-applications', {
      params: { page: 1, pageSize: 20 },
    });
    applications.value = res?.items || [];
  } catch {
    applications.value = [];
  } finally {
    appLoading.value = false;
  }
}

async function handleSubmit() {
  if (!selectedRole.value) {
    uni.showToast({ title: '请选择角色', icon: 'none' });
    return;
  }
  if (!reason.value.trim()) {
    uni.showToast({ title: '请填写申请理由', icon: 'none' });
    return;
  }

  submitting.value = true;
  try {
    await request.post('/auth/role-applications', {
      roleId: selectedRole.value.id,
      reason: reason.value,
    });
    uni.showToast({ title: '申请已提交', icon: 'success' });
    reason.value = '';
    selectedRole.value = null;
    loadApplications();
  } catch (e: any) {
    uni.showToast({ title: e?.message || '提交失败', icon: 'none' });
  } finally {
    submitting.value = false;
  }
}

function getStatusText(status: number): string {
  const map: Record<number, string> = {
    0: '待审核',
    1: '已通过',
    2: '已驳回',
  };
  return map[status] || '未知';
}

function formatTime(time: string): string {
  if (!time) return '';
  const date = new Date(time);
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
}
</script>

<style lang="scss" scoped>
.apply-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 30px;
}

.section {
  background: #fff;
  margin: 12px 16px;
  border-radius: 12px;
  padding: 16px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin-bottom: 12px;
  display: block;
}

.loading, .empty {
  text-align: center;
  padding: 30px 0;
  color: #999;
  font-size: 14px;
}

/* 角色列表 */
.role-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.role-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: #f8f8f8;
  border-radius: 8px;
  border: 2px solid transparent;
}

.role-item.selected {
  background: #e6f7ff;
  border-color: #1890ff;
}

.role-info {
  display: flex;
  flex-direction: column;
}

.role-name {
  font-size: 15px;
  font-weight: 500;
  color: #333;
}

.role-desc {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.role-check {
  font-size: 20px;
  color: #ccc;

  &.active {
    color: #1890ff;
  }
}

/* 申请理由 */
.reason-input {
  width: 100%;
  min-height: 100px;
  padding: 10px;
  background: #f8f8f8;
  border-radius: 8px;
  font-size: 14px;
  box-sizing: border-box;
}

.char-count {
  font-size: 12px;
  color: #999;
  text-align: right;
  display: block;
  margin-top: 4px;
}

/* 提交按钮 */
.action-section {
  padding: 0 16px;
}

.submit-btn {
  width: 100%;
  height: 44px;
  background: #1890ff;
  color: #fff;
  border: none;
  border-radius: 22px;
  font-size: 16px;
}

.submit-btn[disabled] {
  background: #ccc;
}

/* 申请记录 */
.app-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.app-item {
  padding: 12px;
  background: #f8f8f8;
  border-radius: 8px;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.app-role {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.app-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.app-status[data-status="0"] {
  background: #fff7e6;
  color: #fa8c16;
}

.app-status[data-status="1"] {
  background: #f6ffed;
  color: #52c41a;
}

.app-status[data-status="2"] {
  background: #fff2f0;
  color: #ff4d4f;
}

.app-reason {
  font-size: 13px;
  color: #666;
  display: block;
  margin-bottom: 4px;
}

.app-time {
  font-size: 12px;
  color: #999;
  display: block;
}

.app-review {
  font-size: 12px;
  color: #fa8c16;
  display: block;
  margin-top: 4px;
}
</style>
