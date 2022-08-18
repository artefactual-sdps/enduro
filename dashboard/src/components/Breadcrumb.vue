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
        :class="!item.routeName ? 'active' : ''"
        :aria-current="!item.routeName ? 'page' : undefined"
      >
        <router-link
          :to="{ name: item.routeName }"
          v-if="item.routeName"
          class="text-primary"
          >{{ item.text }}</router-link
        >
        <template v-else>{{ item.text }}</template>
      </li>
    </ol>
  </nav>
</template>
