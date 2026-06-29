<template>
  <view class="invite-manage-page">
    <view class="header">
      <text class="title">{{ className }}</text>
      <text class="subtitle">邀请码管理</text>
    </view>

    <!-- 生成邀请码 -->
    <view class="create-section">
      <view class="form-row">
        <input
          v-model.number="maxUses"
          class="input"
          placeholder="最大使用次数（0=无限制）"
          type="number"
        />
        <input
          v-model="expireAt"
          class="input"
          placeholder="过期时间（可选）"
          type="datetime-local"
        />
      </view>
      <button class="gen-btn" @click="handleCreate">生成邀请码</button>
    </view>

    <!-- 邀请码列表 -->
    <view class="invite-list">
      <view
        v-for="inv in invitations"
        :key="inv.id"
        class="invite-item"
      >
        <view class="invite-info">
          <text class="code">{{ inv.code }}</text>
          <text class="usage">
            已用 {{ inv.usedCount }} / {{ inv.maxUses > 0 ? inv.maxUses : '∞' }}
          </text>
          <text v-if="inv.expireAt" class="expire">
            过期：{{ formatDate(inv.expireAt) }}
          </text>
        </view>
        <view class="invite-status">
          <text class="tag" :class="inv.status === 1 ? 'active' : 'disabled'">
            {{ inv.status === 1 ? '有效' : '已禁用' }}
          </text>
          <button
            v-if="inv.status === 1"
            class="disable-btn"
            size="mini"
            @click="handleDisable(inv)"
          >
            禁用
          </button>
        </view>
      </view>
      <view v-if="invitations.length === 0" class="empty">暂无邀请码</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { onLoad } from '@dcloudio/uni-app';
import { getClassInvitations, createInvitation, disableInvitation, getClassDetail } from '@/api/class';
import type { ClassInvitation } from '@/api/class';

const classId = ref(0);
const className = ref('');
const invitations = ref<ClassInvitation[]>([]);
const maxUses = ref(0);
const expireAt = ref('');

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
    const res = await getClassInvitations(classId.value);
    invitations.value = res || [];
  } catch (e: any) {
    uni.showToast({ title: e?.message || '加载失败', icon: 'none' });
  }
}

function handleCreate() {
  const data: any = {};
  if (maxUses.value > 0) data.maxUses = maxUses.value;
  if (expireAt.value) data.expireAt = new Date(expireAt.value).toISOString();

  createInvitation(classId.value, data)
    .then(() => {
      uni.showToast({ title: '生成成功', icon: 'success' });
      maxUses.value = 0;
      expireAt.value = '';
      loadData();
    })
    .catch((e: any) => {
      uni.showToast({ title: e?.message || '生成失败', icon: 'none' });
    });
}

function handleDisable(inv: ClassInvitation) {
  uni.showModal({
    title: '确认禁用',
    content: `确定要禁用邀请码 ${inv.code} 吗？`,
    success: async (res) => {
      if (!res.confirm) return;
      try {
        await disableInvitation(classId.value, inv.id);
        uni.showToast({ title: '已禁用', icon: 'success' });
        loadData();
      } catch (e: any) {
        uni.showToast({ title: e?.message || '操作失败', icon: 'none' });
      }
    },
  });
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
}
</script>

<style lang="scss" scoped>
.invite-manage-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.header {
  background: linear-gradient(135deg, #1890ff 0%, #36cfc9 100%);
  padding: 20px 16px;
  color: #fff;

  .title { font-size: 18px; font-weight: bold; display: block; }
  .subtitle { font-size: 13px; opacity: 0.8; }
}

.create-section {
  margin: 12px 16px;
  background: #fff;
  border-radius: 12px;
  padding: 16px;

  .form-row {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;

    .input {
      flex: 1;
      background: #f5f7fa;
      border: 1px solid #e8e8e8;
      border-radius: 8px;
      padding: 8px 12px;
      font-size: 13px;
    }
  }

  .gen-btn {
    width: 100%;
    background: #1890ff;
    color: #fff;
    border: none;
    border-radius: 8px;
    padding: 10px;
    font-size: 14px;
  }
}

.invite-list {
  margin: 0 16px;
  background: #fff;
  border-radius: 12px;
  padding: 8px 12px;

  .invite-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;
    border-bottom: 1px solid #f0f0f0;

    &:last-child { border-bottom: none; }

    .invite-info {
      flex: 1;

      .code {
        font-family: monospace;
        font-size: 16px;
        font-weight: 600;
        display: block;
        color: #333;
      }

      .usage {
        font-size: 12px;
        color: #999;
        margin-right: 8px;
      }

      .expire {
        font-size: 11px;
        color: #ccc;
      }
    }

    .invite-status {
      display: flex;
      gap: 8px;
      align-items: center;

      .tag {
        font-size: 11px;
        padding: 2px 8px;
        border-radius: 4px;
        color: #fff;
      }
      .tag.active { background: #52c41a; }
      .tag.disabled { background: #d9d9d9; }

      .disable-btn {
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
