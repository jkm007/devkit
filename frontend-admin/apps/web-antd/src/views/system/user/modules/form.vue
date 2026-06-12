<script lang="ts" setup>
import type { DataNode } from 'ant-design-vue/es/tree';

import type { Recordable } from '@vben/types';

import type { SystemUserApi } from '#/api/system/user';

import { computed, nextTick, ref } from 'vue';

import { Tree, useVbenDrawer } from '@vben/common-ui';

import { Spin } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { getMenuList } from '#/api/system/menu';
import { createUser, updateUser } from '#/api/system/user';
import { $t } from '#/locales';

import { useFormSchema } from '../data';

const emits = defineEmits(['success']);

const formData = ref<SystemUserApi.SystemUser>();

const [Form, formApi] = useVbenForm({
  schema: useFormSchema(),
  showDefaultActions: false,
});

const permissions = ref<DataNode[]>([]);
const loadingPermissions = ref(false);

const id = ref();
const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const values = await formApi.getValues();
    // 存储配额：MB 转 bytes
    if (values.storageQuota !== undefined && values.storageQuota !== null) {
      values.storageQuota = Number(values.storageQuota) * 1024 * 1024;
    }
    drawerApi.lock();
    (id.value ? updateUser(id.value, values) : createUser(values))
      .then(() => {
        emits('success');
        drawerApi.close();
      })
      .catch(() => {
        drawerApi.unlock();
      });
  },

  async onOpenChange(isOpen) {
    if (isOpen) {
      const data = drawerApi.getData<SystemUserApi.SystemUser>();
      formApi.resetForm();

      if (data) {
        formData.value = data;
        id.value = data.id;
      } else {
        id.value = undefined;
      }

      if (permissions.value.length === 0) {
        await loadPermissions();
      }
      // Wait for Vue to flush DOM updates (form fields mounted)
      await nextTick();
      if (data) {
        // 存储配额：bytes 转 MB 用于显示（向下取整避免漂移）
        const formData = { ...data };
        if (formData.storageQuota && formData.storageQuota > 0) {
          formData.storageQuota = Math.floor(formData.storageQuota / 1024 / 1024);
        }
        formApi.setValues(formData);
      }
    }
  },
});

async function loadPermissions() {
  loadingPermissions.value = true;
  try {
    const res = await getMenuList();
    permissions.value = res as unknown as DataNode[];
  } finally {
    loadingPermissions.value = false;
  }
}

const getDrawerTitle = computed(() => {
  return formData.value?.id
    ? $t('common.edit', $t('system.user.name'))
    : $t('common.create', $t('system.user.name'));
});

function getNodeClass(node: Recordable<any>) {
  const classes: string[] = [];
  if (node.value?.type === 'button') {
    classes.push('inline-flex');
  }

  return classes.join(' ');
}
</script>
<template>
  <Drawer :title="getDrawerTitle">
    <Form>
      <template #permissions="slotProps">
        <Spin :spinning="loadingPermissions" :classes="{ root: 'w-full' }">
          <Tree
            :tree-data="permissions"
            multiple
            bordered
            :default-expanded-level="2"
            :get-node-class="getNodeClass"
            v-bind="slotProps"
            value-field="id"
            label-field="name"
          />
        </Spin>
      </template>
    </Form>
  </Drawer>
</template>
