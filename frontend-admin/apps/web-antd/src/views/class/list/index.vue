<script lang="ts" setup>
import type { ClassApi } from '#/api/class';

import { computed, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  Button,
  Form,
  Input,
  message,
  Modal,
  Select,
  Space,
  Table,
  Tag,
} from 'ant-design-vue';

import {
  addClassMember,
  createClass,
  createClassInvitation,
  deleteClass,
  disableClassInvitation,
  getClassInvitations,
  getClassList,
  getClassMembers,
  removeClassMember,
  updateClass,
  updateClassMemberRole,
} from '#/api/class';

const searchKeyword = ref('');
const loading = ref(false);
const dataSource = ref<ClassApi.Class[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);

const pagination = computed(() => ({
  current: page.value,
  pageSize: pageSize.value,
  total: total.value,
  showSizeChanger: true,
  showTotal: (t: number) => `共 ${t} 条`,
}));

const columns = [
  { title: '班级名称', dataIndex: 'name', key: 'name' },
  { title: '邀请码', dataIndex: 'code', key: 'code' },
  { title: '成员数', dataIndex: 'memberCount', key: 'memberCount' },
  { title: '创建人', dataIndex: 'creatorName', key: 'creatorName' },
  { title: '状态', dataIndex: 'status', key: 'status' },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt' },
  { title: '操作', key: 'action', width: 220 },
];

async function loadList() {
  loading.value = true;
  try {
    const res = await getClassList({
      page: page.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value,
    });
    dataSource.value = res.items || [];
    total.value = res.total || 0;
  } catch (e: any) {
    message.error(e?.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  page.value = 1;
  loadList();
}

function onTableChange(p: any) {
  page.value = p.current;
  pageSize.value = p.pageSize;
  loadList();
}

const modalVisible = ref(false);
const modalTitle = ref('创建班级');
const modalForm = ref({ id: 0, name: '', description: '', status: 1 });

function openCreateModal() {
  modalTitle.value = '创建班级';
  modalForm.value = { id: 0, name: '', description: '', status: 1 };
  modalVisible.value = true;
}

function openEditModal(row: any) {
  modalTitle.value = '编辑班级';
  modalForm.value = {
    id: row.id,
    name: row.name,
    description: row.description,
    status: row.status,
  };
  modalVisible.value = true;
}

async function handleModalOk() {
  if (!modalForm.value.name) {
    message.warning('请输入班级名称');
    return;
  }
  try {
    if (modalForm.value.id) {
      await updateClass(modalForm.value.id, {
        name: modalForm.value.name,
        description: modalForm.value.description,
        status: modalForm.value.status,
      });
      message.success('更新成功');
    } else {
      await createClass({
        name: modalForm.value.name,
        description: modalForm.value.description,
      });
      message.success('创建成功');
    }
    modalVisible.value = false;
    loadList();
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

async function handleDelete(row: any) {
  Modal.confirm({
    title: '确认删除',
    content: `确定删除班级「${row.name}」吗？`,
    onOk: async () => {
      try {
        await deleteClass(row.id);
        message.success('删除成功');
        loadList();
      } catch (e: any) {
        message.error(e?.message || '删除失败');
      }
    },
  });
}

const memberModalVisible = ref(false);
const currentClassId = ref(0);
const memberList = ref<ClassApi.ClassMember[]>([]);
const memberLoading = ref(false);
const newMember = ref({ userId: '', role: 'student' });

async function openMemberModal(row: any) {
  currentClassId.value = row.id;
  memberModalVisible.value = true;
  await loadMembers();
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

async function handleAddMember() {
  if (!newMember.value.userId) {
    message.warning('请输入用户ID');
    return;
  }
  try {
    await addClassMember(currentClassId.value, {
      userId: Number(newMember.value.userId),
      role: newMember.value.role,
    });
    message.success('添加成功');
    newMember.value = { userId: '', role: 'student' };
    loadMembers();
  } catch (e: any) {
    message.error(e?.message || '添加失败');
  }
}

async function handleUpdateMemberRole(member: any) {
  try {
    await updateClassMemberRole(currentClassId.value, member.id, {
      role: member.role,
    });
    message.success('更新成功');
  } catch (e: any) {
    message.error(e?.message || '更新失败');
  }
}

async function handleRemoveMember(member: any) {
  try {
    await removeClassMember(currentClassId.value, member.id);
    message.success('移除成功');
    loadMembers();
  } catch (e: any) {
    message.error(e?.message || '移除失败');
  }
}

const invitationModalVisible = ref(false);
const invitationList = ref<ClassApi.ClassInvitation[]>([]);
const invitationLoading = ref(false);
const newInvitation = ref({ maxUses: 0, expireAt: '' });

async function openInvitationModal(row: any) {
  currentClassId.value = row.id;
  invitationModalVisible.value = true;
  await loadInvitations();
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

async function handleCreateInvitation() {
  try {
    const data: any = {};
    if (newInvitation.value.maxUses > 0) {
      data.maxUses = newInvitation.value.maxUses;
    }
    if (newInvitation.value.expireAt) {
      data.expireAt = new Date(newInvitation.value.expireAt).toISOString();
    }
    await createClassInvitation(currentClassId.value, data);
    message.success('创建成功');
    newInvitation.value = { maxUses: 0, expireAt: '' };
    loadInvitations();
  } catch (e: any) {
    message.error(e?.message || '创建失败');
  }
}

async function handleDisableInvitation(inv: any) {
  try {
    await disableClassInvitation(inv.id);
    message.success('已禁用');
    loadInvitations();
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  }
}

loadList();
</script>

<template>
  <Page title="班级管理">
    <div class="mb-4 flex gap-4">
      <Input
        v-model:value="searchKeyword"
        placeholder="搜索班级名称/邀请码"
        style="width: 300px"
        @pressEnter="onSearch"
      />
      <Button type="primary" @click="onSearch">搜索</Button>
      <Button type="primary" @click="openCreateModal">创建班级</Button>
    </div>

    <Table
      :columns="columns"
      :data-source="dataSource"
      :loading="loading"
      :pagination="pagination"
      row-key="id"
      @change="onTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'status'">
          <Tag v-if="record.status === 1" color="green">启用</Tag>
          <Tag v-else>禁用</Tag>
        </template>
        <template v-if="column.key === 'action'">
          <Space>
            <Button size="small" @click="openEditModal(record)">编辑</Button>
            <Button size="small" @click="openMemberModal(record)">成员</Button>
            <Button size="small" @click="openInvitationModal(record)">邀请码</Button>
            <Button size="small" danger @click="handleDelete(record)">删除</Button>
          </Space>
        </template>
      </template>
    </Table>

    <!-- 创建/编辑班级 -->
    <Modal
      v-model:open="modalVisible"
      :title="modalTitle"
      @ok="handleModalOk"
    >
      <Form layout="vertical">
        <Form.Item label="班级名称">
          <Input v-model:value="modalForm.name" />
        </Form.Item>
        <Form.Item label="班级描述">
          <Input.TextArea v-model:value="modalForm.description" />
        </Form.Item>
        <Form.Item v-if="modalForm.id" label="状态">
          <Select v-model:value="modalForm.status">
            <Select.Option :value="1">启用</Select.Option>
            <Select.Option :value="0">禁用</Select.Option>
          </Select>
        </Form.Item>
      </Form>
    </Modal>

    <!-- 成员管理 -->
    <Modal
      v-model:open="memberModalVisible"
      title="班级成员"
      width="700px"
      :footer="null"
    >
      <div class="mb-4 flex gap-2">
        <Input
          v-model:value="newMember.userId"
          placeholder="用户ID"
          style="width: 150px"
        />
        <Select v-model:value="newMember.role" style="width: 150px">
          <Select.Option value="student">同学</Select.Option>
          <Select.Option value="monitor">班级管理员</Select.Option>
          <Select.Option value="teacher">班主任/老师</Select.Option>
        </Select>
        <Button type="primary" @click="handleAddMember">添加</Button>
      </div>
      <Table
        :columns="[
          { title: '用户', dataIndex: 'nickname', key: 'nickname' },
          { title: '角色', dataIndex: 'role', key: 'role' },
          { title: '操作', key: 'action' },
        ]"
        :data-source="memberList"
        :loading="memberLoading"
        row-key="id"
        :pagination="false"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'role'">
            <Select
              v-model:value="record.role"
              style="width: 140px"
              @change="handleUpdateMemberRole(record)"
            >
              <Select.Option value="student">同学</Select.Option>
              <Select.Option value="monitor">班级管理员</Select.Option>
              <Select.Option value="teacher">班主任/老师</Select.Option>
            </Select>
          </template>
          <template v-if="column.key === 'action'">
            <Button size="small" danger @click="handleRemoveMember(record)">移除</Button>
          </template>
        </template>
      </Table>
    </Modal>

    <!-- 邀请码 -->
    <Modal
      v-model:open="invitationModalVisible"
      title="邀请码管理"
      width="600px"
      :footer="null"
    >
      <div class="mb-4 flex gap-2">
        <Input
          v-model:value="newInvitation.maxUses"
          placeholder="最大使用次数（0=无限制）"
          style="width: 200px"
          type="number"
        />
        <Input
          v-model:value="newInvitation.expireAt"
          placeholder="过期时间"
          style="width: 200px"
          type="datetime-local"
        />
        <Button type="primary" @click="handleCreateInvitation">生成</Button>
      </div>
      <Table
        :columns="[
          { title: '邀请码', dataIndex: 'code', key: 'code' },
          { title: '已用/上限', key: 'usage' },
          { title: '过期时间', dataIndex: 'expireAt', key: 'expireAt' },
          { title: '状态', dataIndex: 'status', key: 'status' },
          { title: '操作', key: 'action' },
        ]"
        :data-source="invitationList"
        :loading="invitationLoading"
        row-key="id"
        :pagination="false"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'usage'">
            {{ record.usedCount }} / {{ record.maxUses > 0 ? record.maxUses : '∞' }}
          </template>
          <template v-if="column.key === 'status'">
            {{ record.status === 1 ? '有效' : '已禁用' }}
          </template>
          <template v-if="column.key === 'action'">
            <Button
              v-if="record.status === 1"
              size="small"
              danger
              @click="handleDisableInvitation(record)"
            >
              禁用
            </Button>
          </template>
        </template>
      </Table>
    </Modal>
  </Page>
</template>
