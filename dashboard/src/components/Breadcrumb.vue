<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";

const layoutStore = useLayoutStore();
</script>

<template>
  <nav aria-label="Breadcrumb" class="d-inline-block">
    <ol class="breadcrumb mb-0">
      <li
        v-for="item in layoutStore.breadcrumb"
        class="breadcrumb-item"
        :class="!item.route ? 'active' : ''"
        :aria-current="!item.route ? 'page' : undefined"
      >
        <router-link
          :to="item.route"
          v-if="item.route"
          class="text-primary"
          >{{ item.text }}</router-link
        >
        <template v-else>{{ item.text }}</template>
      </li>
    </ol>
  </nav>
</template>
