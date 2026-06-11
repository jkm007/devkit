<script lang="ts" setup>
import type { DataNode } from 'ant-design-vue/es/tree';

import type { Recordable } from '@vben/types';

import type { SystemRoleApi } from '#/api/system/role';

import { computed, nextTick, ref, watch } from 'vue';

import { Tree, useVbenDrawer } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';

import { Spin } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { getMenuList } from '#/api/system/menu';
import { createRole, updateRole } from '#/api/system/role';
import { $t } from '#/locales';

import { useFormSchema } from '../data';

const emits = defineEmits(['success']);

const formData = ref<SystemRoleApi.SystemRole>();

const [Form, formApi] = useVbenForm({
  schema: useFormSchema(),
  showDefaultActions: false,
});

const permissions = ref<DataNode[]>([]);
const loadingPermissions = ref(false);
const selectedKeys = ref<string[]>([]);

// 收集某个节点下所有叶子节点的 authCode
function collectLeafKeys(node: any): string[] {
  if (!node.children || node.children.length === 0) {
    return node.authCode.startsWith('__catalog_') ? [] : [node.authCode];
  }
  return node.children.flatMap((c: any) => collectLeafKeys(c));
}

/** 为全选的父节点补上 authCode，让树组件显示完整勾选 */
function enrichWithParentKeys(leafKeys: string[]): string[] {
  const keySet = new Set(leafKeys);
  const result = [...leafKeys];

  function walk(nodes: any[]) {
    for (const node of nodes) {
      if (!node.children || node.children.length === 0) continue;
      walk(node.children);
      if (node.authCode && !keySet.has(node.authCode)) {
        const leaves = collectLeafKeys(node);
        if (leaves.length > 0 && leaves.every((k) => keySet.has(k))) {
          keySet.add(node.authCode);
          result.push(node.authCode);
        }
      }
    }
  }
  walk(permissions.value);
  return result;
}

// 处理选中变化 — 直接保留树组件返回的完整 keys（含父节点），保存时再过滤
function handleUpdateValue(keys: string[]) {
  selectedKeys.value = keys;
  // 同步到表单（只存叶子节点权限码，不含 __catalog_ 前缀）
  const leafOnly = keys.filter((k) => !k.startsWith('__catalog_'));
  formApi.setValues({ permissions: leafOnly });
}

// 监听表单值变化（编辑时加载已有权限）
watch(
  () => formData.value?.permissions,
  (val) => {
    if (Array.isArray(val)) {
      selectedKeys.value = enrichWithParentKeys(val);
    }
  },
  { immediate: true },
);

const id = ref();
const [Drawer, drawerApi] = useVbenDrawer({
  async onConfirm() {
    const { valid } = await formApi.validate();
    if (!valid) return;
    const values = await formApi.getValues();
    // 过滤掉占位用的 catalog key
    if (Array.isArray(values.permissions)) {
      values.permissions = values.permissions.filter(
        (p: string) => !p.startsWith('__catalog_'),
      );
    }
    drawerApi.lock();
    (id.value ? updateRole(id.value, values) : createRole(values))
      .then(async () => {
        await drawerApi.close();
        emits('success');
      })
      .catch(() => {
        drawerApi.unlock();
      });
  },

  async onOpenChange(isOpen) {
    if (isOpen) {
      const data = drawerApi.getData<SystemRoleApi.SystemRole>();
      formApi.resetForm();
      selectedKeys.value = [];

      if (data) {
        formData.value = data;
        id.value = data.id;
      } else {
        id.value = undefined;
      }

      if (permissions.value.length === 0) {
        await loadPermissions();
      }
      await nextTick();
      if (data) {
        formApi.setValues(data);
        if (Array.isArray(data.permissions)) {
          selectedKeys.value = enrichWithParentKeys(data.permissions);
        }
      }
    }
  },
});

/** 递归处理权限树：为没有 authCode 的节点生成占位符 */
function processPermissionTree(nodes: any[]): any[] {
  return nodes.map((node) => {
    const children = node.children ? processPermissionTree(node.children) : [];
    return {
      ...node,
      authCode: node.authCode || `__catalog_${node.id}`,
      children: children.length > 0 ? children : undefined,
    };
  });
}

async function loadPermissions() {
  loadingPermissions.value = true;
  try {
    const res = await getMenuList();
    permissions.value = processPermissionTree(res as unknown as DataNode[]);
  } finally {
    loadingPermissions.value = false;
  }
}

const getDrawerTitle = computed(() => {
  return formData.value?.id
    ? $t('common.edit', $t('system.role.name'))
    : $t('common.create', $t('system.role.name'));
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
      <template #permissions>
        <Spin :spinning="loadingPermissions" :classes="{ root: 'w-full' }">
          <Tree
            :tree-data="permissions"
            multiple
            bordered
            check-strictly
            :default-expanded-level="2"
            :get-node-class="getNodeClass"
            :model-value="selectedKeys"
            @update:model-value="(keys: string[]) => handleUpdateValue(keys)"
            value-field="authCode"
            label-field="meta.title"
            icon-field="meta.icon"
          >
            <template #node="{ value }">
              <IconifyIcon v-if="value.meta.icon" :icon="value.meta.icon" />
              {{ $t(value.meta.title) }}
            </template>
          </Tree>
        </Spin>
      </template>
    </Form>
  </Drawer>
</template>
<style lang="css" scoped>
:deep(.ant-tree-title) {
  .tree-actions {
    @apply ml-5 hidden;
  }
}

:deep(.ant-tree-title:hover) {
  .tree-actions {
    @apply ml-5 flex flex-auto justify-end;
  }
}

</style>
