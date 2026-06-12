<script lang="ts" setup>
import { Tag } from 'ant-design-vue';

defineProps<{
  data?: Array<{
    username: string;
    ip: string;
    device: string;
    location: string;
    status: number;
    createdAt: string;
  }>;
}>();

function formatTime(t: string) {
  if (!t) return '';
  const d = new Date(t);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
</script>

<template>
  <div class="max-h-64 overflow-y-auto">
    <table v-if="data && data.length > 0" class="w-full text-xs">
      <thead>
        <tr class="border-b text-left text-gray-500">
          <th class="py-1.5 font-medium">用户</th>
          <th class="py-1.5 font-medium">IP</th>
          <th class="py-1.5 font-medium">状态</th>
          <th class="py-1.5 font-medium">时间</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(item, idx) in data"
          :key="idx"
          class="border-b border-gray-50"
        >
          <td class="py-1.5">{{ item.username || '-' }}</td>
          <td class="py-1.5 text-gray-500">{{ item.ip }}</td>
          <td class="py-1.5">
            <Tag :color="item.status === 1 ? 'success' : 'error'" class="!m-0">
              {{ item.status === 1 ? '成功' : '失败' }}
            </Tag>
          </td>
          <td class="py-1.5 text-gray-400">{{ formatTime(item.createdAt) }}</td>
        </tr>
      </tbody>
    </table>
    <div v-else class="flex h-40 items-center justify-center text-gray-400">
      暂无登录记录
    </div>
  </div>
</template>
