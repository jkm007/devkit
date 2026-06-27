<template>
  <view class="class-list-page">
    <view class="header">
      <text class="title">我的班级</text>
      <text class="join-btn" @click="showJoinInput = true">加入班级</text>
    </view>

    <view v-if="loading" class="loading">
      <text>加载中...</text>
    </view>
    <view v-else-if="list.length === 0" class="empty">
      <text class="empty-icon">🏫</text>
      <text class="empty-text">暂无班级</text>
      <text class="empty-hint">点击右上角加入班级</text>
    </view>
    <view v-else class="list">
      <view
        v-for="item in list"
        :key="item.id"
        class="item"
        @click="goToDetail(item)"
      >
        <view class="item-header">
          <text class="name">{{ item.name }}</text>
          <text class="role" :class="item.myRole">{{ getRoleText(item.myRole) }}</text>
        </view>
        <text class="desc">{{ item.description || '暂无描述' }}</text>
        <view class="item-footer">
          <text class="info">成员 {{ item.memberCount }} · 邀请码 {{ item.code }}</text>
        </view>
      </view>
    </view>

    <!-- 加入班级弹窗 -->
    <uni-popup v-if="showJoinInput" type="dialog" ref="joinPopupRef">
      <uni-popup-dialog
        mode="input"
        title="加入班级"
        placeholder="请输入邀请码"
        @close="showJoinInput = false"
        @confirm="handleJoin"
      />
    </uni-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getMyClasses, joinClassByCode } from '@/api/class';
import type { ClassInfo } from '@/api/class';

const loading = ref(false);
const list = ref<ClassInfo[]>([]);
const showJoinInput = ref(false);
const joinPopupRef = ref<any>(null);

onMounted(() => {
  loadList();
});

async function loadList() {
  loading.value = true;
  try {
    const res = await getMyClasses();
    list.value = Array.isArray(res) ? res : [];
  } catch (e) {
    console.error('加载班级列表失败:', e);
    list.value = [];
  } finally {
    loading.value = false;
  }
}

async function handleJoin(code: string) {
  if (!code) {
    uni.showToast({ title: '请输入邀请码', icon: 'none' });
    return;
  }
  try {
    await joinClassByCode(code);
    uni.showToast({ title: '加入成功', icon: 'success' });
    showJoinInput.value = false;
    loadList();
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加入失败', icon: 'none' });
  }
}

function goToDetail(item: ClassInfo) {
  uni.navigateTo({ url: `/pages/class/detail?id=${item.id}` });
}

function getRoleText(role: string): string {
  const map: Record<string, string> = {
    student: '同学',
    monitor: '班级管理员',
    teacher: '班主任',
  };
  return map[role] || role;
}
</script>

<style lang="scss" scoped>
.class-list-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #fff;
  .title {
    font-size: 18px;
    font-weight: bold;
    color: #333;
  }
  .join-btn {
    font-size: 14px;
    color: #1890ff;
    padding: 4px 12px;
    border: 1px solid #1890ff;
    border-radius: 16px;
  }
}

.loading, .empty {
  text-align: center;
  padding: 60px 20px;
}
.empty-icon { font-size: 48px; display: block; margin-bottom: 16px; }
.empty-text { font-size: 16px; color: #666; display: block; }
.empty-hint { font-size: 13px; color: #999; display: block; margin-top: 8px; }

.list { padding: 12px 16px; }

.item {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);

  &:active { opacity: 0.7; }

  .item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .name {
    font-size: 16px;
    font-weight: 500;
    color: #333;
  }

  .role {
    font-size: 11px;
    padding: 2px 8px;
    border-radius: 4px;
    color: #fff;

    &.student { background: #52c41a; }
    &.monitor { background: #faad14; }
    &.teacher { background: #1890ff; }
  }

  .desc {
    font-size: 13px;
    color: #999;
    margin-bottom: 12px;
    display: block;
  }

  .item-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .info {
    font-size: 12px;
    color: #999;
  }
}
</style>
