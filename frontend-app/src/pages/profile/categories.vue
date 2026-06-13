<template>
  <view class="categories-page">
    <view class="header">
      <text class="title">我的分类</text>
      <text class="subtitle">最多绑定 3 个分类（1 个主分类 + 2 个副分类）</text>
    </view>

    <!-- 已绑定列表 -->
    <view class="bound-section">
      <view v-if="bindings.length === 0" class="empty">
        <text class="empty-text">暂未绑定分类</text>
      </view>
      <view v-else>
        <view v-for="b in bindings" :key="b.id" class="binding-card">
          <view class="card-header">
            <text class="primary-tag" v-if="b.isPrimary">主</text>
            <text class="category-name">{{ b.categoryName || '分类 ' + b.categoryId }}</text>
          </view>
          <text class="bound-time">绑定于 {{ formatDate(b.boundAt) }}</text>
          <view class="card-actions">
            <view v-if="!b.isPrimary" class="action-btn" @click="setPrimary(b)">设为主分类</view>
            <view class="action-btn delete" @click="unbind(b)">✕ 解绑</view>
          </view>
        </view>
      </view>
    </view>

    <!-- 添加按钮 -->
    <button class="add-btn" :disabled="bindings.length >= 3" @click="showPicker = true">
      <text>{{ bindings.length >= 3 ? '已满 3 个分类' : '+ 添加分类' }}</text>
    </button>

    <!-- 提示 -->
    <view class="tip-section">
      <text class="tip-text">绑定分类后，首页和题库将默认显示该分类内容</text>
    </view>

    <!-- 分类选择弹窗 -->
    <up-popup v-model:show="showPicker" mode="bottom" round="12">
      <view class="picker-panel">
        <text class="picker-title">选择分类</text>
        <view class="category-list">
          <view v-for="cat in categories" :key="cat.id" class="category-item" @click="selectCategory(cat)">
            <text>{{ cat.name }}</text>
          </view>
        </view>
      </view>
    </up-popup>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getCategoryBindings, bindCategory, unbindCategory, setPrimaryCategory } from '@/api/user';

const bindings = ref<any[]>([]);
const showPicker = ref(false);
const categories = ref<any[]>([]);

onMounted(() => {
  loadBindings();
  loadCategories();
});

async function loadBindings() {
  try {
    bindings.value = await getCategoryBindings();
  } catch {
    // Mock 数据
    bindings.value = [
      { id: 1, categoryId: 101, categoryName: '网络协议', isPrimary: true, boundAt: '2024-01-10T10:00:00Z' },
      { id: 2, categoryId: 102, categoryName: '操作系统', isPrimary: false, boundAt: '2024-01-12T14:30:00Z' },
    ];
  }
}

async function loadCategories() {
  // Mock 全部分类列表
  const allCategories = [
    { id: 101, name: '网络协议' },
    { id: 102, name: '操作系统' },
    { id: 103, name: '数据结构' },
    { id: 104, name: '数据库' },
    { id: 105, name: '算法' },
  ];
  // 过滤已绑定的分类
  const boundIds = new Set(bindings.value.map(b => b.categoryId));
  categories.value = allCategories.filter(cat => !boundIds.has(cat.id));
}

async function selectCategory(cat: any) {
  showPicker.value = false;
  try {
    await bindCategory({ categoryId: cat.id, isPrimary: bindings.value.length === 0 });
    uni.showToast({ title: '绑定成功', icon: 'success' });
    await loadBindings();
    loadCategories(); // 刷新可选列表
  } catch (e: any) {
    uni.showToast({ title: e.message || '绑定失败', icon: 'none' });
  }
}

async function setPrimary(b: any) {
  try {
    await setPrimaryCategory(b.id);
    uni.showToast({ title: '已设为主分类', icon: 'success' });
    loadBindings();
  } catch {
    uni.showToast({ title: '操作失败', icon: 'none' });
  }
}

async function unbind(b: any) {
  uni.showModal({
    title: '确认解绑',
    content: `确定要解绑「${b.categoryName}」吗？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await unbindCategory(b.id);
          uni.showToast({ title: '已解绑', icon: 'success' });
          await loadBindings();
          loadCategories(); // 刷新可选列表
        } catch {
          uni.showToast({ title: '解绑失败', icon: 'none' });
        }
      }
    },
  });
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}
</script>

<style lang="scss" scoped>
.categories-page { min-height: 100vh; background: #f5f5f5; padding: 16px; }

.header { margin-bottom: 20px; }
.title { font-size: 20px; font-weight: bold; color: #333; display: block; }
.subtitle { font-size: 13px; color: #999; margin-top: 4px; display: block; }

.bound-section { background: #fff; border-radius: 12px; padding: 16px; margin-bottom: 16px; }
.empty { text-align: center; padding: 20px 0; }
.empty-text { font-size: 14px; color: #999; }

.binding-card { padding: 16px; background: #f9f9f9; border-radius: 8px; margin-bottom: 12px; }
.binding-card:last-child { margin-bottom: 0; }
.card-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.primary-tag { padding: 2px 8px; background: #1890ff; color: #fff; border-radius: 4px; font-size: 12px; }
.category-name { font-size: 16px; font-weight: 500; color: #333; }
.bound-time { font-size: 12px; color: #999; margin-bottom: 12px; display: block; }
.card-actions { display: flex; gap: 8px; }
.action-btn { padding: 6px 16px; background: #f5f7fa; border-radius: 16px; font-size: 13px; color: #666; }
.action-btn.delete { color: #ff4d4f; }

.add-btn { height: 48px; line-height: 48px; border: 2px dashed #d9d9d9; border-radius: 12px; background: transparent; color: #1890ff; font-size: 15px; }
.add-btn[disabled] { color: #ccc; border-color: #eee; }

.tip-section { margin-top: 16px; padding: 12px; background: #fffbe6; border-radius: 8px; }
.tip-text { font-size: 12px; color: #d48806; }

.picker-panel { padding: 20px; }
.picker-title { font-size: 16px; font-weight: 500; margin-bottom: 16px; display: block; }
.category-list { max-height: 300px; overflow-y: auto; }
.category-item { padding: 12px 16px; background: #f5f7fa; border-radius: 8px; margin-bottom: 8px; font-size: 14px; }
.category-item:active { background: #e6f7ff; }
</style>
