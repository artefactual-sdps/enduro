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
        class="breadcrumb-item"
        :class="i == layoutStore.breadcrumb.length - 1 ? 'active' : ''"
        :aria-current="
          i == layoutStore.breadcrumb.length - 1 ? 'page' : undefined
        "
      >
        <router-link :to="item.route" v-if="item.route" class="text-primary">{{
          item.text
        }}</router-link>
        <template v-else>{{ item.text }}</template>
      </li>
    </ol>
  </nav>
</template>
