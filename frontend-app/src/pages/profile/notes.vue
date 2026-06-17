<template>
  <view class="notes-page">
    <!-- 头部 -->
    <view class="header">
      <text class="title">我的笔记</text>
      <text class="count">共 {{ total }} 条</text>
    </view>

    <!-- 列表 -->
    <view v-if="loading" class="loading">
      <text>加载中...</text>
    </view>
    <view v-else-if="list.length === 0" class="empty">
      <text class="empty-icon">📝</text>
      <text class="empty-text">暂无笔记</text>
      <text class="empty-hint">在题目详情页可添加笔记</text>
    </view>
    <view v-else class="list">
      <view
        v-for="item in list"
        :key="item.id"
        class="item"
        @click="goToDetail(item.questionId)"
      >
        <view class="item-header">
          <text class="question-title">{{ item.questionTitle || '题目' }}</text>
          <text class="time">{{ formatTime(item.createdAt || item.createTime) }}</text>
        </view>
        <text class="item-content">{{ item.content }}</text>
        <view class="item-actions">
          <text class="action-btn edit" @click.stop="editNote(item)">编辑</text>
          <text class="action-btn delete" @click.stop="removeNote(item)">删除</text>
        </view>
      </view>
    </view>

    <!-- 加载更多 -->
    <view v-if="list.length > 0 && list.length < total" class="load-more" @click="loadMore">
      <text>{{ loadingMore ? '加载中...' : '加载更多' }}</text>
    </view>

    <!-- 编辑弹窗 -->
    <view v-if="showEdit" class="edit-modal" @click.self="showEdit = false">
      <view class="edit-content">
        <text class="edit-title">编辑笔记</text>
        <textarea v-model="editText" class="edit-textarea" placeholder="请输入笔记内容" :maxlength="2000" />
        <view class="edit-actions">
          <button class="cancel-btn" @click="showEdit = false">取消</button>
          <button class="save-btn" @click="saveEdit">保存</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getNotes, updateNote, deleteNote } from '@/api/study';

const loading = ref(false);
const loadingMore = ref(false);
const list = ref<any[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = 20;

const showEdit = ref(false);
const editItem = ref<any>(null);
const editText = ref('');

onMounted(() => {
  loadList();
});

async function loadList() {
  loading.value = true;
  page.value = 1;
  try {
    const res = await getNotes({ page: page.value, pageSize });
    list.value = res.items || [];
    total.value = res.total || 0;
  } catch (e) {
    console.error('加载笔记失败:', e);
    list.value = [];
  } finally {
    loading.value = false;
  }
}

async function loadMore() {
  if (loadingMore.value) return;
  loadingMore.value = true;
  page.value++;
  try {
    const res = await getNotes({ page: page.value, pageSize });
    list.value = [...list.value, ...(res.items || [])];
  } catch (e) {
    console.error('加载更多失败:', e);
  } finally {
    loadingMore.value = false;
  }
}

function editNote(item: any) {
  editItem.value = item;
  editText.value = item.content || '';
  showEdit.value = true;
}

async function saveEdit() {
  if (!editText.value.trim()) {
    uni.showToast({ title: '请输入笔记内容', icon: 'none' });
    return;
  }
  try {
    await updateNote(editItem.value.id, { content: editText.value });
    editItem.value.content = editText.value;
    showEdit.value = false;
    uni.showToast({ title: '保存成功', icon: 'success' });
  } catch (e) {
    uni.showToast({ title: '保存失败', icon: 'none' });
  }
}

async function removeNote(item: any) {
  uni.showModal({
    title: '提示',
    content: '确定删除该笔记？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await deleteNote(item.id);
          list.value = list.value.filter(i => i.id !== item.id);
          total.value--;
          uni.showToast({ title: '已删除', icon: 'success' });
        } catch (e) {
          uni.showToast({ title: '删除失败', icon: 'none' });
        }
      }
    }
  });
}

function goToDetail(questionId: number) {
  if (questionId) {
    uni.navigateTo({ url: `/pages/question/detail?id=${questionId}` });
  }
}

function formatTime(time: string): string {
  if (!time) return '';
  const d = new Date(time);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  const days = Math.floor(diff / 86400000);
  if (days === 0) return '今天';
  if (days === 1) return '昨天';
  if (days < 7) return `${days}天前`;
  return `${d.getMonth() + 1}/${d.getDate()}`;
}
</script>

<style lang="scss" scoped>
.notes-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #fff;
  .title { font-size: 18px; font-weight: bold; color: #333; }
  .count { font-size: 13px; color: #999; }
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
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);

  .item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .question-title {
    font-size: 14px;
    font-weight: 500;
    color: #1890ff;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-right: 12px;
  }

  .time { font-size: 12px; color: #999; }

  .item-content {
    font-size: 14px;
    color: #333;
    line-height: 1.6;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .item-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid #f0f0f0;
  }

  .action-btn {
    font-size: 13px;
    padding: 4px 12px;
    border-radius: 16px;
    &.edit { color: #1890ff; background: #e6f7ff; }
    &.delete { color: #ff4d4f; background: #fff1f0; }
  }
}

.load-more {
  text-align: center;
  padding: 16px;
  color: #1890ff;
  font-size: 14px;
}

/* 编辑弹窗 */
.edit-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.edit-content {
  width: 90%;
  max-width: 400px;
  background: #fff;
  border-radius: 16px;
  padding: 24px;
}

.edit-title {
  font-size: 18px;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 16px;
}

.edit-textarea {
  width: 100%;
  min-height: 150px;
  padding: 12px;
  background: #f9f9f9;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
  font-size: 14px;
  line-height: 1.6;
}

.edit-actions {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.cancel-btn, .save-btn {
  flex: 1;
  height: 40px;
  border: none;
  border-radius: 20px;
  font-size: 15px;
}
.cancel-btn { background: #f5f5f5; color: #666; }
.save-btn { background: #1890ff; color: #fff; }
</style>
