<script lang="ts" setup>
import type { Arrayable } from '@vueuse/core';
import type { FlattenedItem } from 'reka-ui';

import type { ClassType, Recordable } from '@vben-core/typings';

import type { TreeProps } from './types';

import { computed, onMounted, ref, watch, watchEffect } from 'vue';

import { ChevronRight, IconifyIcon } from '@vben-core/icons';
import { cn, get } from '@vben-core/shared/utils';

import { TreeItem, TreeRoot } from 'reka-ui';

import { Checkbox } from '../checkbox';
import { treePropsDefaults } from './types';

const props = withDefaults(defineProps<TreeProps>(), treePropsDefaults());

const emits = defineEmits<{
  expand: [value: FlattenedItem<Recordable<any>>];
  select: [value: FlattenedItem<Recordable<any>>];
}>();

interface InnerFlattenItem<T = Recordable<any>, P = number | string> {
  hasChildren: boolean;
  id: P;
  level: number;
  parentId: null | P;
  parents: P[];
  value: T;
}

function flatten<T = Recordable<any>, P = number | string>(
  items: T[],
  childrenField: string = 'children',
  level = 0,
  parentId: null | P = null,
  parents: P[] = [],
): InnerFlattenItem<T, P>[] {
  const result: InnerFlattenItem<T, P>[] = [];
  items.forEach((item) => {
    const children = get(item, childrenField) as Array<T>;
    const id = get(item, props.valueField) as P;
    const val: InnerFlattenItem<T, P> = {
      hasChildren: Array.isArray(children) && children.length > 0,
      id,
      level,
      parentId,
      parents: [...parents],
      value: item,
    };
    result.push(val);
    if (val.hasChildren)
      result.push(
        ...flatten(children, childrenField, level + 1, id, [...parents, id]),
      );
  });
  return result;
}

const flattenData = ref<Array<InnerFlattenItem>>([]);
const modelValue = defineModel<Arrayable<number | string>>();
const expanded = ref<Array<number | string>>(props.defaultExpandedKeys ?? []);

const treeValue = ref();
let lastTreeData: any = null;

onMounted(() => {
  watchEffect(() => {
    flattenData.value = flatten(props.treeData, props.childrenField);
    if (flattenData.value.length > 0) {
      updateTreeValue();
    }

    // 只在 treeData 变化时执行展开
    const currentTreeData = JSON.stringify(props.treeData);
    if (lastTreeData !== currentTreeData) {
      lastTreeData = currentTreeData;
      if (
        props.defaultExpandedLevel !== undefined &&
        props.defaultExpandedLevel > 0
      ) {
        expandToLevel(props.defaultExpandedLevel);
      }
    }
  });

  // 外部 modelValue 变化时同步内部状态
  watch(modelValue, () => {
    if (flattenData.value.length > 0) {
      updateTreeValue();
    }
  });
});

function getItemByValue(value: number | string) {
  return flattenData.value.find(
    (item) => get(item.value, props.valueField) === value,
  )?.value;
}

function updateTreeValue() {
  const val = modelValue.value;
  if (val === undefined) {
    treeValue.value = props.multiple ? [] : undefined;
  } else if (Array.isArray(val)) {
    if (val.length === 0) {
      treeValue.value = [];
    } else {
      const filteredValues = val.filter((v) => {
        const item = getItemByValue(v);
        return item && !get(item, props.disabledField);
      });
      treeValue.value = filteredValues.map((v) => getItemByValue(v));

      if (filteredValues.length !== val.length) {
        modelValue.value = filteredValues;
      }
    }
  } else {
    const item = getItemByValue(val);
    if (item && !get(item, props.disabledField)) {
      treeValue.value = item;
    } else {
      treeValue.value = props.multiple ? [] : undefined;
      modelValue.value = props.multiple ? [] : undefined;
    }
  }
}

function updateModelValue(val: Arrayable<Recordable<any>>) {
  if (Array.isArray(val)) {
    const filteredVal = val.filter((v) => !get(v, props.disabledField));
    modelValue.value = filteredVal.map((v) => get(v, props.valueField));
  } else {
    if (val && !get(val, props.disabledField)) {
      modelValue.value = get(val, props.valueField);
    }
  }
}

function expandToLevel(level: number) {
  const keys: string[] = [];
  flattenData.value.forEach((item) => {
    if (item.level <= level - 1) {
      keys.push(get(item.value, props.valueField));
    }
  });
  expanded.value = keys;
}

function collapseNodes(value: Arrayable<number | string>) {
  const keys = new Set(Array.isArray(value) ? value : [value]);
  expanded.value = expanded.value.filter((key) => !keys.has(key));
}

function expandNodes(value: Arrayable<number | string>) {
  const keys = [...(Array.isArray(value) ? value : [value])];
  keys.forEach((key) => {
    if (expanded.value.includes(key)) return;
    const item = getItemByValue(key);
    if (item) {
      expanded.value.push(key);
    }
  });
}

function expandAll() {
  expanded.value = flattenData.value
    .filter((item) => item.hasChildren)
    .map((item) => get(item.value, props.valueField));
}

function collapseAll() {
  expanded.value = [];
}

function checkAll() {
  if (!props.multiple) return;
  modelValue.value = [
    ...new Set(
      flattenData.value
        .filter((item) => !get(item.value, props.disabledField))
        .map((item) => get(item.value, props.valueField)),
    ),
  ];
  updateTreeValue();
}

function unCheckAll() {
  if (!props.multiple) return;
  modelValue.value = [];
  updateTreeValue();
}

function isNodeDisabled(item: FlattenedItem<Recordable<any>>) {
  return props.disabled || get(item.value, props.disabledField);
}

// 计算全选/半选状态
const selectAllStatus = computed<'indeterminate' | boolean>(() => {
  if (!props.multiple) return false;
  if (!modelValue.value || !Array.isArray(modelValue.value)) return false;

  const allValues = flattenData.value
    .filter((item) => !get(item.value, props.disabledField))
    .map((item) => get(item.value, props.valueField));

  const selectedCount = allValues.filter((v) =>
    (modelValue.value as (number | string)[]).includes(v),
  ).length;

  if (selectedCount === 0) return false;
  if (selectedCount === allValues.length) return true;
  return 'indeterminate';
});

function onSelectAllChange(checked: 'indeterminate' | boolean) {
  if (checked === true) {
    checkAll();
  } else {
    unCheckAll();
  }
}

function onToggle(item: FlattenedItem<Recordable<any>>) {
  emits('expand', item);
}
function onSelect(item: FlattenedItem<Recordable<any>>, isSelected: boolean) {
  if (isNodeDisabled(item)) {
    return;
  }

  // check-strictly 模式：点击父节点时向下传播到所有叶子节点
  if (props.checkStrictly && props.multiple && Array.isArray(modelValue.value)) {
    const nodeKey = get(item.value, props.valueField);
    const leaves = getNodeLeafKeys(nodeKey);
    if (leaves.length > 0) {
      const allSelected = leaves.every((leaf) =>
        (modelValue.value as unknown[]).includes(leaf),
      );
      if (allSelected) {
        // 全选 → 取消所有叶子，同时移除父节点 key
        modelValue.value = (modelValue.value as unknown[]).filter(
          (v) => !leaves.includes(v) && v !== nodeKey,
        );
      } else {
        // 未全选 → 添加所有叶子，移除父节点 key（状态由计算得出）
        const set = new Set(
          (modelValue.value as unknown[]).filter((v) => v !== nodeKey),
        );
        leaves.forEach((leaf) => set.add(leaf));
        modelValue.value = [...set];
      }
      updateTreeValue();
      emits('select', item);
      return;
    }
  }

  if (
    !props.checkStrictly &&
    props.multiple &&
    props.autoCheckParent &&
    isSelected
  ) {
    flattenData.value
      .find((i) => {
        return (
          get(i.value, props.valueField) === get(item.value, props.valueField)
        );
      })
      ?.parents?.filter((item) => !get(item, props.disabledField))
      ?.forEach((p) => {
        if (Array.isArray(modelValue.value) && !modelValue.value.includes(p)) {
          modelValue.value.push(p);
        }
      });
  }
  if (
    !props.checkStrictly &&
    props.multiple &&
    props.autoCheckParent &&
    !isSelected
  ) {
    flattenData.value
      .find((i) => {
        return (
          get(i.value, props.valueField) === get(item.value, props.valueField)
        );
      })
      ?.parents?.filter((item) => !get(item, props.disabledField))
      ?.toReversed()
      .forEach((p) => {
        const children = flattenData.value.filter((i) => {
          return (
            i.parents.length > 0 &&
            i.parents.includes(p) &&
            i.id !== item._id &&
            i.parentId === p
          );
        });
        if (Array.isArray(modelValue.value)) {
          const hasSelectedChild = children.some((child) =>
            (modelValue.value as unknown[]).includes(
              get(child.value, props.valueField),
            ),
          );
          if (!hasSelectedChild) {
            const index = modelValue.value.indexOf(p);
            if (index !== -1) {
              modelValue.value.splice(index, 1);
            }
          }
        }
      });
  }

  // 移除 modelValue 中的父节点 key（父节点状态由 isNodeAllSelected/isNodeIndeterminate 计算）
  if (props.checkStrictly && props.multiple && Array.isArray(modelValue.value)) {
    const nodeKey = get(item.value, props.valueField);
    const leaves = getNodeLeafKeys(nodeKey);
    if (leaves.length === 0) {
      // 当前是叶子节点，移除其所有祖先 key
      const parents = item.parents || [];
      for (const p of parents) {
        const parentKey = get(p, props.valueField);
        const idx = (modelValue.value as unknown[]).indexOf(parentKey);
        if (idx !== -1) {
          (modelValue.value as unknown[]).splice(idx, 1);
        }
      }
    }
  }

  updateTreeValue();
  emits('select', item);
}

// 预计算：每个节点的叶子节点 key 列表（只在 treeData 变化时重建）
const leafKeysMap = ref<Map<number | string, unknown[]>>(new Map());

function buildLeafKeysMap() {
  const map = new Map<number | string, unknown[]>();
  // 从叶子节点向上构建
  for (const item of flattenData.value) {
    const key = get(item.value, props.valueField);
    const children = flattenData.value.filter((i) => i.parentId === key);
    if (children.length === 0) {
      // 叶子节点，自身就是叶子
      map.set(key, [key]);
    }
  }
  // 自底向上：父节点的叶子 = 所有子节点的叶子之和
  // 重复几轮直到所有节点都有值
  for (let round = 0; round < 10; round++) {
    let changed = false;
    for (const item of flattenData.value) {
      const key = get(item.value, props.valueField);
      if (map.has(key)) continue;
      const children = flattenData.value.filter((i) => i.parentId === key);
      if (children.length === 0) continue;
      const allChildrenResolved = children.every((c) =>
        map.has(get(c.value, props.valueField)),
      );
      if (!allChildrenResolved) continue;
      const leaves: unknown[] = [];
      for (const c of children) {
        leaves.push(...(map.get(get(c.value, props.valueField)) || []));
      }
      map.set(key, leaves);
      changed = true;
    }
    if (!changed) break;
  }
  leafKeysMap.value = map;
}

// 监听 treeData 变化时重建 leafKeysMap
watch(
  () => flattenData.value.length,
  () => {
    if (flattenData.value.length > 0) {
      buildLeafKeysMap();
    }
  },
);

function getNodeLeafKeys(nodeKey: number | string): unknown[] {
  return leafKeysMap.value.get(nodeKey) || [];
}

// 判断节点是否半选（部分叶子节点选中）
function isNodeIndeterminate(nodeValue: Recordable<any>): boolean {
  if (!props.multiple) return false;
  if (!Array.isArray(modelValue.value)) return false;

  const nodeKey = get(nodeValue, props.valueField);
  const leaves = getNodeLeafKeys(nodeKey);
  if (leaves.length <= 1) return false;

  const selectedSet = new Set(modelValue.value as unknown[]);
  const selectedCount = leaves.filter((leaf) => selectedSet.has(leaf)).length;

  return selectedCount > 0 && selectedCount < leaves.length;
}

// 判断节点是否全选
function isNodeAllSelected(nodeValue: Recordable<any>): boolean {
  if (!props.multiple) return false;
  if (!Array.isArray(modelValue.value)) return false;

  const nodeKey = get(nodeValue, props.valueField);
  const leaves = getNodeLeafKeys(nodeKey);
  if (leaves.length === 0) return false;
  if (leaves.length === 1 && leaves[0] === nodeKey) {
    // 叶子节点
    return (modelValue.value as unknown[]).includes(nodeKey);
  }

  const selectedSet = new Set(modelValue.value as unknown[]);
  return leaves.every((leaf) => selectedSet.has(leaf));
}

defineExpose({
  collapseAll,
  collapseNodes,
  expandAll,
  expandNodes,
  checkAll,
  unCheckAll,
  expandToLevel,
  getItemByValue,
});
</script>
<template>
  <TreeRoot
    :get-key="(item) => get(item, valueField)"
    :get-children="(item) => get(item, childrenField)"
    :items="treeData"
    :model-value="treeValue"
    v-model:expanded="expanded as string[]"
    :default-expanded="defaultExpandedKeys as string[]"
    :propagate-select="!checkStrictly"
    :multiple="multiple"
    :disabled="disabled"
    :selection-behavior="allowClear || multiple ? 'toggle' : 'replace'"
    @update:model-value="updateModelValue"
    v-slot="{ flattenItems }"
    :class="
      cn(
        'text-blackA11 container list-none rounded-lg text-sm font-medium select-none',
        $attrs.class as unknown as ClassType,
        bordered ? 'border' : '',
      )
    "
  >
    <div
      :class="
        cn('my-0.5 flex w-full items-center p-1', bordered ? 'border-b' : '')
      "
      v-if="$slots.header"
    >
      <slot name="header"> </slot>
    </div>
    <div
      :class="
        cn('my-0.5 flex w-full items-center p-1', bordered ? 'border-b' : '')
      "
      v-if="treeData.length > 0 && showToggleAll"
    >
      <div
        class="flex size-5 flex-1 cursor-pointer items-center"
        @click="() => (expanded?.length > 0 ? collapseAll() : expandAll())"
      >
        <ChevronRight
          :class="{ 'rotate-90': expanded?.length > 0 }"
          class="text-foreground/80 hover:text-foreground size-4 cursor-pointer transition"
        />
        <div class="flex items-center gap-1 item-all-checkbox">
          <Checkbox
            v-if="multiple"
            :model-value="selectAllStatus"
            :indeterminate="selectAllStatus === 'indeterminate'"
            @click.stop
            @update:model-value="onSelectAllChange"
          />
          <span v-if="selectAllLabel">{{ selectAllLabel }}</span>
        </div>
      </div>
    </div>
    <TransitionGroup :name="transition ? 'fade' : ''">
      <TreeItem
        v-for="item in flattenItems"
        v-slot="{
          isExpanded,
          isSelected,
          isIndeterminate,
          handleSelect,
          handleToggle,
        }"
        :key="item._id"
        :style="{ 'margin-left': `${item.level - 1}rem` }"
        :class="
          cn('cursor-pointer', getNodeClass?.(item), {
            'data-[selected]:bg-accent': !multiple,
            'text-foreground/50 cursor-not-allowed': isNodeDisabled(item),
          })
        "
        v-bind="
          Object.assign(item.bind, {
            onfocus: isNodeDisabled(item) ? 'this.blur()' : undefined,
            disabled: isNodeDisabled(item),
          })
        "
        @select="
          (event: any) => {
            if (isNodeDisabled(item)) {
              event.preventDefault();
              event.stopPropagation();
              return;
            }
            if (event.detail.originalEvent.type === 'click') {
              event.preventDefault();
            }
            onSelect(item, event.detail.isSelected);
          }
        "
        @toggle="
          (event: any) => {
            if (event.detail.originalEvent.type === 'click') {
              event.preventDefault();
            }
            !isNodeDisabled(item) && onToggle(item);
          }
        "
        class="tree-node focus:ring-grass8 my-0.5 flex items-center rounded p-1 outline-hidden"
      >
        <!-- class="hover:ring-2" 鼠标移动上去时2px的圆环边框 -->
        <ChevronRight
          v-if="
            item.hasChildren &&
            Array.isArray(item.value[childrenField]) &&
            item.value[childrenField].length > 0
          "
          class="text-foreground/80 hover:text-foreground size-4 cursor-pointer transition"
          :class="{ 'rotate-90': isExpanded }"
          @click.stop="
            () => {
              handleToggle();
              onToggle(item);
            }
          "
        />
        <div v-else class="h-4 w-4"></div>
        <div class="flex items-center gap-1 item-checkbox">
          <Checkbox
            v-if="multiple"
            :model-value="isNodeIndeterminate(item.value) && !isNodeDisabled(item) ? 'indeterminate' : (isNodeAllSelected(item.value) && !isNodeDisabled(item))"
            :disabled="isNodeDisabled(item)"
            @click="
              (event: MouseEvent) => {
                if (isNodeDisabled(item)) {
                  event.preventDefault();
                  event.stopPropagation();
                  return;
                }
                handleSelect();
              }
            "
          />
          <div
            class="flex items-center gap-1 item-checkbox"
            :title="get(item.value, labelField)"
            @click="
              (event: MouseEvent) => {
                if (isNodeDisabled(item)) {
                  event.preventDefault();
                  event.stopPropagation();
                  return;
                }
                handleSelect();
              }
            "
          >
            <slot name="node" v-bind="item">
              <IconifyIcon
                class="size-4"
                v-if="showIcon && get(item.value, iconField)"
                :icon="get(item.value, iconField)"
              />
              {{ get(item.value, labelField) }}
            </slot>
          </div>
        </div>
        <div class="h-4 w-4"></div>
      </TreeItem>
    </TransitionGroup>
    <div
      :class="
        cn('my-0.5 flex w-full items-center p-1', bordered ? 'border-t' : '')
      "
      v-if="$slots.footer"
    >
      <slot name="footer"> </slot>
    </div>
  </TreeRoot>
</template>
<style lang="scss" scoped>
.container {
  position: relative;
  padding: 0;
  list-style-type: none;
}

.item {
  box-sizing: border-box;
  width: 100%;
  height: 30px;
  background-color: #f3f3f3;
  border: 1px solid #666;
}

.item-checkbox {
  width: 100%;
  overflow: hidden;
}

.item-all-checkbox {
  width: 100%;
  overflow: hidden;

  .text-label {
    margin-left: 8px;
  }
}

/* 1. 声明过渡效果 */
.fade-move,
.fade-enter-active,
.fade-leave-active {
  transition: all 0.5s cubic-bezier(0.55, 0, 0.1, 1);
}

/* 2. 声明进入和离开的状态 */
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: scaleY(0.01) translate(30px, 0);
}

/* 3. 确保离开的项目被移除出了布局流
      以便正确地计算移动时的动画效果。 */
.fade-leave-active {
  position: absolute;
}
</style>
