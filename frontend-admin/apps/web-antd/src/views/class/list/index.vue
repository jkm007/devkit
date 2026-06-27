<script lang="ts" setup>
import type { ClassApi } from '#/api/class';

import { ref } from 'vue';

import { Page, useVbenDrawer, useVbenModal } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { message, Tag } from 'ant-design-vue';

import { useVbenVxeGrid, VbenTableAction } from '#/adapter/vxe-table';
import {
  createClass,
  deleteClass,
  updateClass,
  getClassList,
  addClassMember,
  getClassMembers,
  removeClassMember,
  updateClassMemberRole,
  createClassInvitation,
  getClassInvitations,
  disableClassInvitation,
} from '#/api/class';

// ============ 主表格 ============
const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: 'Input',
        fieldName: 'keyword',
        label: '关键词',
        componentProps: { placeholder: '搜索班级名称/邀请码', allowClear: true },
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: '状态',
        componentProps: {
          options: [
            { label: '启用', value: 1 },
            { label: '禁用', value: 0 },
          ],
          placeholder: '全部',
          allowClear: true,
        },
      },
    ],
    submitOnChange: true,
  },
  gridOptions: {
    columns: [
      { field: 'id', title: 'ID', width: 80 },
      { field: 'name', title: '班级名称', minWidth: 180 },
      { field: 'code', title: '邀请码', width: 120 },
      { field: 'memberCount', title: '成员数', width: 100, align: 'center' },
      { field: 'creatorName', title: '创建人', width: 120 },
      {
        field: 'status',
        title: '状态',
        width: 100,
        align: 'center',
        slots: { default: 'status' },
      },
      { field: 'createdAt', title: '创建时间', width: 180 },
      { field: 'action', title: '操作', width: 280, fixed: 'right', slots: { default: 'action' } },
    ],
    height: 'auto',
    keepSource: true,
    pagerConfig: { pageSize: 20, pageSizeOpts: [10, 20, 50, 100] },
    proxyConfig: {
      ajax: {
        query: async ({ page }, formValues) => {
          return await getClassList({
            page: page.currentPage,
            pageSize: page.pageSize,
            ...(formValues as any),
          });
        },
      },
    },
    rowConfig: { keyField: 'id' },
    toolbarConfig: {
      custom: true,
      refresh: true,
      search: true,
      zoom: true,
    },
  } as any,
});

// ============ 创建/编辑班级 ============
interface ClassForm {
  id: number;
  name: string;
  description: string;
  status: number;
}

const classForm = ref<ClassForm>({ id: 0, name: '', description: '', status: 1 });
const classFormTitle = ref('创建班级');

const [ClassFormModal, classFormModalApi] = useVbenModal({
  onConfirm: async () => {
    if (!classForm.value.name) {
      message.warning('请输入班级名称');
      return;
    }
    try {
      if (classForm.value.id) {
        await updateClass(classForm.value.id, {
          name: classForm.value.name,
          description: classForm.value.description,
          status: classForm.value.status,
        });
        message.success('更新成功');
      } else {
        await createClass({
          name: classForm.value.name,
          description: classForm.value.description,
        });
        message.success('创建成功');
      }
      classFormModalApi.close();
      gridApi.reload();
    } catch (e: any) {
      message.error(e?.message || '操作失败');
    }
  },
});

function openCreateClass() {
  classFormTitle.value = '创建班级';
  classForm.value = { id: 0, name: '', description: '', status: 1 };
  classFormModalApi.open();
}

function openEditClass(row: any) {
  classFormTitle.value = '编辑班级';
  classForm.value = {
    id: row.id,
    name: row.name,
    description: row.description || '',
    status: row.status,
  };
  classFormModalApi.open();
}

function handleDelete(row: any) {
  deleteClass(row.id)
    .then(() => {
      message.success('删除成功');
      gridApi.reload();
    })
    .catch((e: any) => message.error(e?.message || '删除失败'));
}

// ============ 成员管理 ============
const memberList = ref<ClassApi.ClassMember[]>([]);
const currentClassId = ref(0);
const currentClassName = ref('');
const memberLoading = ref(false);
const newMember = ref({ userId: '', role: 'student' });

const [MemberDrawer, memberDrawerApi] = useVbenDrawer({
  onOpenChange: async (isOpen: boolean) => {
    if (isOpen) {
      await loadMembers();
    }
  },
});

async function openMemberModal(row: any) {
  currentClassId.value = row.id;
  currentClassName.value = row.name;
  memberDrawerApi.open();
}

async function loadMembers() {
  memberLoading.value = true;
  try {
    const res = await getClassMembers(currentClassId.value, { page: 1, pageSize: 100 });
    memberList.value = res.items || [];
  } catch (e: any) {
    message.error(e?.message || '加载成员失败');
  } finally {
    memberLoading.value = false;
  }
}

function handleAddMember() {
  if (!newMember.value.userId) {
    message.warning('请输入用户ID');
    return;
  }
  addClassMember(currentClassId.value, {
    userId: Number(newMember.value.userId),
    role: newMember.value.role,
  })
    .then(() => {
      message.success('添加成功');
      newMember.value = { userId: '', role: 'student' };
      loadMembers();
    })
    .catch((e: any) => message.error(e?.message || '添加失败'));
}

function handleUpdateMemberRole(member: any) {
  updateClassMemberRole(currentClassId.value, member.id, { role: member.role })
    .then(() => message.success('更新成功'))
    .catch((e: any) => message.error(e?.message || '更新失败'));
}

function handleRemoveMember(member: any) {
  removeClassMember(currentClassId.value, member.id)
    .then(() => {
      message.success('移除成功');
      loadMembers();
    })
    .catch((e: any) => message.error(e?.message || '移除失败'));
}

// ============ 邀请码管理 ============
const invitationList = ref<ClassApi.ClassInvitation[]>([]);
const invitationLoading = ref(false);
const newInvitation = ref({ maxUses: 0, expireAt: '' });

const [InvitationDrawer, invitationDrawerApi] = useVbenDrawer({
  onOpenChange: async (isOpen: boolean) => {
    if (isOpen) {
      await loadInvitations();
    }
  },
});

async function openInvitationModal(row: any) {
  currentClassId.value = row.id;
  currentClassName.value = row.name;
  invitationDrawerApi.open();
}

async function loadInvitations() {
  invitationLoading.value = true;
  try {
    const res = await getClassInvitations(currentClassId.value);
    invitationList.value = res || [];
  } catch (e: any) {
    message.error(e?.message || '加载邀请码失败');
  } finally {
    invitationLoading.value = false;
  }
}

function handleCreateInvitation() {
  const data: any = {};
  if (newInvitation.value.maxUses > 0) data.maxUses = newInvitation.value.maxUses;
  if (newInvitation.value.expireAt) data.expireAt = new Date(newInvitation.value.expireAt).toISOString();
  createClassInvitation(currentClassId.value, data)
    .then(() => {
      message.success('创建成功');
      newInvitation.value = { maxUses: 0, expireAt: '' };
      loadInvitations();
    })
    .catch((e: any) => message.error(e?.message || '创建失败'));
}

function handleDisableInvitation(inv: any) {
  disableClassInvitation(inv.id)
    .then(() => {
      message.success('已禁用');
      loadInvitations();
    })
    .catch((e: any) => message.error(e?.message || '操作失败'));
}
</script>

<template>
  <Page title="班级管理">
    <Grid table-title="班级列表" table-title-help="管理系统中的所有班级">
      <template #toolbar-left>
        <Button type="primary" @click="openCreateClass">
          <Plus class="mr-1 size-4" />
          创建班级
        </Button>
      </template>

      <template #status="{ row }">
        <Tag v-if="row.status === 1" color="green">启用</Tag>
        <Tag v-else>禁用</Tag>
      </template>

      <template #action="{ row }">
        <VbenTableAction
          :actions="[
            { name: '编辑', onClick: () => openEditClass(row) },
            { name: '成员', onClick: () => openMemberModal(row) },
            { name: '邀请码', onClick: () => openInvitationModal(row) },
          ]"
          :drop-down-actions="[
            { name: '删除', danger: true, onClick: () => handleDelete(row) },
          ]"
        />
      </template>
    </Grid>

    <!-- 创建/编辑班级 -->
    <ClassFormModal :title="classFormTitle" class="w-[500px]">
      <div class="space-y-4 py-4">
        <div>
          <label class="mb-1 block text-sm font-medium">班级名称 <span class="text-red-500">*</span></label>
          <input
            v-model="classForm.name"
            class="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
            placeholder="请输入班级名称"
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium">班级描述</label>
          <textarea
            v-model="classForm.description"
            class="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
            placeholder="请输入班级描述（可选）"
            rows="3"
          />
        </div>
        <div v-if="classForm.id">
          <label class="mb-1 block text-sm font-medium">状态</label>
          <select
            v-model="classForm.status"
            class="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
          >
            <option :value="1">启用</option>
            <option :value="0">禁用</option>
          </select>
        </div>
      </div>
    </ClassFormModal>

    <!-- 成员管理 -->
    <MemberDrawer title="班级成员" :description="currentClassName" class="w-[600px]">
      <div class="mb-4 flex gap-2">
        <input
          v-model="newMember.userId"
          class="w-32 rounded border border-gray-300 px-3 py-2 text-sm"
          placeholder="用户ID"
        />
        <select
          v-model="newMember.role"
          class="w-36 rounded border border-gray-300 px-3 py-2 text-sm"
        >
          <option value="student">同学</option>
          <option value="monitor">班级管理员</option>
          <option value="teacher">班主任/老师</option>
        </select>
        <Button type="primary" size="small" @click="handleAddMember">添加</Button>
      </div>

      <div v-if="memberLoading" class="py-8 text-center text-gray-400">加载中...</div>
      <div v-else-if="memberList.length === 0" class="py-8 text-center text-gray-400">暂无成员</div>
      <div v-else class="space-y-2">
        <div
          v-for="m in memberList"
          :key="m.id"
          class="flex items-center justify-between rounded border border-gray-200 px-3 py-2"
        >
          <div class="flex items-center gap-3">
            <span class="text-sm font-medium">{{ m.nickname || '用户' + m.userId }}</span>
          </div>
          <div class="flex items-center gap-2">
            <select
              :value="m.role"
              class="w-32 rounded border border-gray-300 px-2 py-1 text-xs"
              @change="m.role = ($event.target as HTMLSelectElement).value; handleUpdateMemberRole(m)"
            >
              <option value="student">同学</option>
              <option value="monitor">班级管理员</option>
              <option value="teacher">班主任/老师</option>
            </select>
            <Button size="small" danger @click="handleRemoveMember(m)">移除</Button>
          </div>
        </div>
      </div>
    </MemberDrawer>

    <!-- 邀请码管理 -->
    <InvitationDrawer title="邀请码管理" :description="currentClassName" class="w-[600px]">
      <div class="mb-4 flex gap-2">
        <input
          v-model.number="newInvitation.maxUses"
          class="w-40 rounded border border-gray-300 px-3 py-2 text-sm"
          placeholder="最大使用次数（0=无限制）"
          type="number"
        />
        <input
          v-model="newInvitation.expireAt"
          class="w-44 rounded border border-gray-300 px-3 py-2 text-sm"
          placeholder="过期时间"
          type="datetime-local"
        />
        <Button type="primary" size="small" @click="handleCreateInvitation">生成</Button>
      </div>

      <div v-if="invitationLoading" class="py-8 text-center text-gray-400">加载中...</div>
      <div v-else-if="invitationList.length === 0" class="py-8 text-center text-gray-400">暂无邀请码</div>
      <div v-else class="space-y-2">
        <div
          v-for="inv in invitationList"
          :key="inv.id"
          class="flex items-center justify-between rounded border border-gray-200 px-3 py-2"
        >
          <div>
            <span class="font-mono text-sm font-medium">{{ inv.code }}</span>
            <span class="ml-3 text-xs text-gray-400">
              已用 {{ inv.usedCount }} / {{ inv.maxUses > 0 ? inv.maxUses : '∞' }}
            </span>
          </div>
          <div class="flex items-center gap-2">
            <Tag v-if="inv.status === 1" color="green">有效</Tag>
            <Tag v-else>已禁用</Tag>
            <Button
              v-if="inv.status === 1"
              size="small"
              danger
              @click="handleDisableInvitation(inv)"
            >
              禁用
            </Button>
          </div>
        </div>
      </div>
    </InvitationDrawer>
  </Page>
</template>
