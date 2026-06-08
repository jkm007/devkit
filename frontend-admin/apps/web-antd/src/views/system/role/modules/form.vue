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

// 扁平化树，用于查找节点和子节点
const flatNodeMap = computed(() => {
  const map = new Map<string, any>();
  const flatten = (nodes: any[]) => {
    for (const node of nodes) {
      map.set(node.authCode, node);
      if (node.children) flatten(node.children);
    }
  };
  flatten(permissions.value);
  return map;
});

// 收集某个节点下所有叶子节点的 authCode
function collectLeafKeys(node: any): string[] {
  if (!node.children || node.children.length === 0) {
    return node.authCode.startsWith('__catalog_') ? [] : [node.authCode];
  }
  return node.children.flatMap((c: any) => collectLeafKeys(c));
}

// 判断父节点是否半选（部分子节点选中）
function isIndeterminate(node: any): boolean {
  if (!node.children || node.children.length === 0) return false;
  const leafKeys = collectLeafKeys(node);
  if (leafKeys.length === 0) return false;
  const selectedCount = leafKeys.filter((k) =>
    selectedKeys.value.includes(k),
  ).length;
  return selectedCount > 0 && selectedCount < leafKeys.length;
}

// 处理选中变化
function handleUpdateValue(keys: string[]) {
  // 找出本次操作的节点：在 keys 中但不在 selectedKeys 中的（新增），
  // 或在 selectedKeys 中但不在 keys 中的（移除）
  const oldSet = new Set(selectedKeys.value);
  const newSet = new Set(keys);
  let toggledKey = '';

  for (const k of newSet) {
    if (!oldSet.has(k)) {
      toggledKey = k;
      break;
    }
  }
  if (!toggledKey) {
    for (const k of oldSet) {
      if (!newSet.has(k)) {
        toggledKey = k;
        break;
      }
    }
  }

  // 检查被操作的节点是否是父节点
  const toggledNode = flatNodeMap.value.get(toggledKey);
  if (toggledNode?.children?.length > 0) {
    // 父节点：级联操作子节点
    const leafKeys = collectLeafKeys(toggledNode);
    const isNowSelected = newSet.has(toggledKey);

    if (isNowSelected) {
      // 选中父节点 → 添加所有叶子节点
      const merged = [...new Set([...keys, ...leafKeys])];
      selectedKeys.value = merged.filter((k) => !k.startsWith('__catalog_'));
    } else {
      // 取消父节点 → 移除所有叶子节点
      selectedKeys.value = keys.filter(
        (k) => !leafKeys.includes(k) && !k.startsWith('__catalog_'),
      );
    }
  } else {
    // 叶子节点：直接过滤
    selectedKeys.value = keys.filter((k) => !k.startsWith('__catalog_'));
  }

  // 同步到表单
  formApi.setValues({ permissions: selectedKeys.value });
}

// 监听表单值变化（编辑时加载已有权限）
watch(
  () => formData.value?.permissions,
  (val) => {
    if (Array.isArray(val)) {
      selectedKeys.value = [...val];
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
          selectedKeys.value = [...data.permissions];
        }
      }
    }
  },
});

/** 递归过滤：只保留有权限码的节点 */
function filterPermissionTree(nodes: any[]): any[] {
  return nodes
    .map((node) => {
      const children = node.children
        ? filterPermissionTree(node.children)
        : [];
      if (node.authCode || children.length > 0) {
        return {
          ...node,
          authCode: node.authCode || `__catalog_${node.id}`,
          children: children.length > 0 ? children : undefined,
        };
      }
      return null;
    })
    .filter(Boolean);
}

async function loadPermissions() {
  loadingPermissions.value = true;
  try {
    const res = await getMenuList();
    permissions.value = filterPermissionTree(res as unknown as DataNode[]);
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
  // 半选状态
  const authCode = node.value?.authCode;
  if (authCode) {
    const realNode = flatNodeMap.value.get(authCode);
    if (realNode && isIndeterminate(realNode)) {
      classes.push('tree-indeterminate');
    }
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

/* 半选状态样式：给 checkbox 添加背景色 */
:deep(.tree-indeterminate) {
  [data-state='unchecked'],
  button[role='checkbox'] {
    background-color: hsl(var(--primary) / 0.3) !important;
    border-color: hsl(var(--primary)) !important;
  }
}
</style>
