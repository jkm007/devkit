<template>
  <view class="class-detail-page">
    <view v-if="classInfo" class="class-card">
      <text class="name">{{ classInfo.name }}</text>
      <text class="desc">{{ classInfo.description || '暂无描述' }}</text>
      <view class="meta">
        <text class="code">邀请码：{{ classInfo.code }}</text>
        <text class="count">成员 {{ classInfo.memberCount }}</text>
      </view>
      <view class="actions">
        <view class="action-btn" @click="goToQuestions">
          <text class="icon">📝</text>
          <text>班级题目</text>
        </view>
        <view v-if="classInfo.myRole === 'student' || classInfo.myRole === 'monitor'" class="action-btn leave-btn" @click="handleLeave">
          <text class="icon">🚪</text>
          <text>退出班级</text>
        </view>
      </view>
    </view>

    <view class="section">
      <view class="section-header">
        <text class="section-title">班级成员</text>
      </view>
      <view v-if="memberLoading" class="loading">
        <text>加载中...</text>
      </view>
      <view v-else class="member-list">
        <view
          v-for="m in members"
          :key="m.id"
          class="member-item"
        >
          <view class="member-info">
            <text class="member-name">{{ m.nickname || m.username }}</text>
            <text class="member-role" :class="m.role">{{ getRoleText(m.role) }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { onLoad } from '@dcloudio/uni-app';
import { getClassDetail, getClassMembers, leaveClass } from '@/api/class';
import type { ClassInfo, ClassMember } from '@/api/class';

const classId = ref(0);
const classInfo = ref<ClassInfo | null>(null);
const members = ref<ClassMember[]>([]);
const memberLoading = ref(false);

onLoad((options: any) => {
  if (options?.id) {
    classId.value = Number(options.id);
    loadData();
  }
});

async function loadData() {
  try {
    classInfo.value = await getClassDetail(classId.value);
  } catch (e) {
    console.error('加载班级详情失败:', e);
  }
  await loadMembers();
}

async function loadMembers() {
  memberLoading.value = true;
  try {
    const res = await getClassMembers(classId.value, { page: 1, pageSize: 100 });
    members.value = res?.items || [];
  } catch (e) {
    console.error('加载成员失败:', e);
    members.value = [];
  } finally {
    memberLoading.value = false;
  }
}

function goToQuestions() {
  const name = classInfo.value?.name || '';
  uni.navigateTo({ url: `/pages/question/list?classId=${classId.value}&className=${encodeURIComponent(name)}` });
}

function getRoleText(role: string): string {
  const map: Record<string, string> = {
    student: '同学',
    monitor: '班级管理员',
    teacher: '班主任',
  };
  return map[role] || role;
}

function handleLeave() {
  uni.showModal({
    title: '确认退出',
    content: '退出后将无法再查看班级题目和内容，确定要退出吗？',
    success: async (res) => {
      if (!res.confirm) return;
      try {
        await leaveClass(classId.value);
        uni.showToast({ title: '已退出班级', icon: 'success' });
        setTimeout(() => { uni.navigateBack(); }, 1000);
      } catch (e: any) {
        uni.showToast({ title: e?.message || '退出失败', icon: 'none' });
      }
    },
  });
}
</script>

<style lang="scss" scoped>
.class-detail-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.class-card {
  background: linear-gradient(135deg, #1890ff 0%, #36cfc9 100%);
  padding: 24px 16px;
  color: #fff;

  .name {
    font-size: 20px;
    font-weight: bold;
    display: block;
    margin-bottom: 8px;
  }

  .desc {
    font-size: 13px;
    color: rgba(255, 255, 255, 0.8);
    display: block;
    margin-bottom: 16px;
  }

  .meta {
    display: flex;
    justify-content: space-between;
    font-size: 13px;
    color: rgba(255, 255, 255, 0.9);
    margin-bottom: 16px;
  }

  .actions {
    display: flex;
    gap: 12px;
  }

  .action-btn {
    background: rgba(255, 255, 255, 0.2);
    border-radius: 12px;
    padding: 12px 16px;
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;

    &:active { opacity: 0.7; }

    .icon { font-size: 18px; }
  }

  .leave-btn {
    background: rgba(255, 77, 77, 0.3);
  }
}

.section {
  margin: 12px 16px;
  background: #fff;
  border-radius: 12px;
  padding: 16px;

  .section-header {
    margin-bottom: 12px;
    .section-title {
      font-size: 16px;
      font-weight: 500;
      color: #333;
    }
  }
}

.loading { text-align: center; padding: 20px; color: #999; }

.member-list {
  .member-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;
    border-bottom: 1px solid #f5f5f5;

    &:last-child { border-bottom: none; }

    .member-name { font-size: 15px; color: #333; }
    .member-role {
      font-size: 11px;
      padding: 2px 8px;
      border-radius: 4px;
      color: #fff;

      &.student { background: #52c41a; }
      &.monitor { background: #faad14; }
      &.teacher { background: #1890ff; }
    }
  }
}
</style>
