<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";

const layoutStore = useLayoutStore();
</script>

<template>
  <span v-if="layoutStore.breadcrumb.length" class="text-muted me-2">/</span>
  <nav aria-label="Breadcrumb" class="d-inline-block">
    <ol class="breadcrumb mb-0">
      <li
        v-for="(item, i) in layoutStore.breadcrumb"
        :key="'breadcrumb-' + i"
        class="breadcrumb-item"
        :class="i == layoutStore.breadcrumb.length - 1 ? 'active' : ''"
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
