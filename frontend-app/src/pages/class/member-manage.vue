<template>
  <view class="member-manage-page">
    <view class="header">
      <text class="title">{{ className }}</text>
      <text class="count">共 {{ members.length }} 人</text>
    </view>

    <!-- 添加成员 -->
    <view class="add-member">
      <input v-model="newMember.userId" class="input" placeholder="输入用户ID" type="number" />
      <select class="select" v-model="newMember.role">
        <option value="student">同学</option>
        <option value="monitor">班级管理员</option>
        <option value="teacher">班主任/老师</option>
      </select>
      <button class="add-btn" size="mini" @click="handleAdd">添加</button>
    </view>

    <!-- 成员列表 -->
    <view class="member-list">
      <view
        v-for="m in members"
        :key="m.id"
        class="member-item"
      >
        <view class="member-info">
          <text class="name">{{ m.nickname || '用户' + m.userId }}</text>
          <text class="role" :class="m.role">{{ getRoleText(m.role) }}</text>
        </view>
        <view class="actions" v-if="m.id !== myMemberId">
          <select
            class="role-select"
            :value="m.role"
            @change="handleChangeRole(m, $event)"
          >
            <option value="student">同学</option>
            <option value="monitor">班级管理员</option>
            <option value="teacher">班主任</option>
          </select>
          <button class="remove-btn" size="mini" @click="handleRemove(m)">移除</button>
        </view>
      </view>
      <view v-if="members.length === 0" class="empty">暂无成员</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { onLoad } from '@dcloudio/uni-app';
import { getClassMembers, addClassMember, removeClassMember, updateMemberRole, getClassDetail } from '@/api/class';
import type { ClassMember } from '@/api/class';

const classId = ref(0);
const className = ref('');
const members = ref<ClassMember[]>([]);
const myMemberId = ref(0);
const newMember = ref({ userId: '', role: 'student' });

onLoad((options: any) => {
  if (options?.id) {
    classId.value = Number(options.id);
  }
  if (options?.name) {
    className.value = decodeURIComponent(options.name);
  }
});

onMounted(async () => {
  await loadData();
});

async function loadData() {
  try {
    const detail = await getClassDetail(classId.value);
    if (detail) className.value = detail.name;
    const res = await getClassMembers(classId.value, { page: 1, pageSize: 100 });
    members.value = res?.items || [];
    // 找到当前用户的成员ID（通过接口返回的用户ID匹配，这里简化处理）
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' });
  }
}

async function handleAdd() {
  if (!newMember.value.userId) {
    uni.showToast({ title: '请输入用户ID', icon: 'none' });
    return;
  }
  try {
    await addClassMember(classId.value, {
      userId: Number(newMember.value.userId),
      role: newMember.value.role,
    });
    uni.showToast({ title: '添加成功', icon: 'success' });
    newMember.value = { userId: '', role: 'student' };
    loadData();
  } catch (e: any) {
    uni.showToast({ title: e?.message || '添加失败', icon: 'none' });
  }
}

async function handleRemove(m: ClassMember) {
  uni.showModal({
    title: '确认移除',
    content: `确定要移除成员「${m.nickname || '用户' + m.userId}」吗？`,
    success: async (res) => {
      if (!res.confirm) return;
      try {
        await removeClassMember(classId.value, m.id);
        uni.showToast({ title: '已移除', icon: 'success' });
        loadData();
      } catch (e: any) {
        uni.showToast({ title: e?.message || '移除失败', icon: 'none' });
      }
    },
  });
}

async function handleChangeRole(m: ClassMember, event: Event) {
  const role = (event.target as HTMLSelectElement).value;
  try {
    await updateMemberRole(classId.value, m.id, { role });
    uni.showToast({ title: '更新成功', icon: 'success' });
    loadData();
  } catch (e: any) {
    uni.showToast({ title: e?.message || '更新失败', icon: 'none' });
    loadData(); // 回滚
  }
}

function getRoleText(role: string): string {
  const map: Record<string, string> = { student: '同学', monitor: '班级管理员', teacher: '班主任' };
  return map[role] || role;
}
</script>

<style lang="scss" scoped>
.member-manage-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.header {
  background: linear-gradient(135deg, #1890ff 0%, #36cfc9 100%);
  padding: 20px 16px;
  color: #fff;
  display: flex;
  justify-content: space-between;
  align-items: center;

  .title { font-size: 18px; font-weight: bold; }
  .count { font-size: 13px; opacity: 0.8; }
}

.add-member {
  margin: 12px 16px;
  background: #fff;
  border-radius: 12px;
  padding: 12px;
  display: flex;
  gap: 8px;
  align-items: center;

  .input {
    flex: 1;
    background: #f5f7fa;
    border: 1px solid #e8e8e8;
    border-radius: 8px;
    padding: 8px 12px;
    font-size: 13px;
  }

  .select {
    width: 90px;
    background: #f5f7fa;
    border: 1px solid #e8e8e8;
    border-radius: 8px;
    padding: 8px;
    font-size: 12px;
  }

  .add-btn {
    background: #1890ff;
    color: #fff;
    border: none;
    font-size: 13px;
  }
}

.member-list {
  margin: 0 16px;
  background: #fff;
  border-radius: 12px;
  padding: 8px 12px;

  .member-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;
    border-bottom: 1px solid #f0f0f0;

    &:last-child { border-bottom: none; }

    .member-info {
      display: flex;
      align-items: center;
      gap: 8px;

      .name { font-size: 14px; color: #333; }
      .role {
        font-size: 11px;
        padding: 2px 6px;
        border-radius: 4px;
        color: #fff;
      }
      .role.student { background: #52c41a; }
      .role.monitor { background: #faad14; }
      .role.teacher { background: #1890ff; }
    }

    .actions {
      display: flex;
      gap: 8px;
      align-items: center;

      .role-select {
        font-size: 11px;
        padding: 4px;
        border: 1px solid #e8e8e8;
        border-radius: 4px;
      }

      .remove-btn {
        background: #fff;
        color: #ff4d4f;
        border: 1px solid #ff4d4f;
        font-size: 12px;
      }
    }
  }

  .empty {
    text-align: center;
    padding: 40px 0;
    color: #999;
    font-size: 14px;
  }
}
</style>
