<script setup lang="ts">
import type { WorkbenchTrendItem } from '../typing';

import { computed } from 'vue';

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  VbenIcon,
} from '@vben-core/shadcn-ui';

import { sanitizeHtml } from './sanitize-html';

interface Props {
  items?: WorkbenchTrendItem[];
  title: string;
}

defineOptions({
  name: 'WorkbenchTrends',
});

const props = withDefaults(defineProps<Props>(), {
  items: () => [],
});

/** Sanitize HTML content for each trend item to prevent XSS. */
const sanitizedItems = computed(() =>
  props.items.map((item) => ({
    ...item,
    safeContent: sanitizeHtml(item.content),
  })),
);
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle class="text-lg">{{ title }}</CardTitle>
    </CardHeader>
    <CardContent class="flex flex-wrap p-5 pt-0">
      <ul class="w-full divide-y divide-border" role="list">
        <li
          v-for="item in sanitizedItems"
          :key="item.title"
          class="flex justify-between gap-x-6 py-5"
        >
          <div class="flex min-w-0 items-center gap-x-4">
            <VbenIcon
              :icon="item.avatar"
              alt=""
              class="size-10 flex-none rounded-full"
            />
            <div class="min-w-0 flex-auto">
              <p class="text-sm/6 font-semibold text-foreground">
                {{ item.title }}
              </p>
              <!-- v-html safe: content is sanitized via sanitizeHtml() -->
              <p
                class="mt-1 truncate text-xs/5 text-foreground/80 *:text-primary"
                v-html="item.safeContent"
              ></p>
            </div>
          </div>
          <div class="hidden h-full shrink-0 sm:flex sm:flex-col sm:items-end">
            <span class="mt-6 text-xs/6 text-foreground/80">
              {{ item.date }}
            </span>
          </div>
        </li>
      </ul>
    </CardContent>
  </Card>
</template>
