<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";

const layoutStore = useLayoutStore();
</script>

<template>
  <nav aria-label="Breadcrumb" class="d-flex">
    <span v-if="layoutStore.breadcrumb.length" class="text-muted me-2">/</span>
    <ol class="breadcrumb mb-0 flex-nowrap overflow-hidden">
      <li
        v-for="(item, i) in layoutStore.breadcrumb"
        :key="'breadcrumb-' + i"
        :class="[
          'breadcrumb-item',
          'text-nowrap',
          i === layoutStore.breadcrumb.length - 1
            ? 'active text-truncate'
            : 'flex-shrink-0',
        ]"
        :aria-current="
          i == layoutStore.breadcrumb.length - 1 ? 'page' : undefined
        "
      >
        <RouterLink v-if="item.route" :to="item.route" class="text-primary">
          {{ item.text }}
        </RouterLink>
        <template v-else>
          {{ item.text }}
        </template>
      </li>
    </ol>
  </nav>
</template>
