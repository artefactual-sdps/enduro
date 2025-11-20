<script setup lang="ts">
import type { FunctionalComponent, SVGAttributes } from "vue";
import { useRoute } from "vue-router/auto";
import type { RouteLocationResolved } from "vue-router/auto";

const route = useRoute();

type Tab = {
  icon?: FunctionalComponent<SVGAttributes>;
  text: string;
  route: RouteLocationResolved;
  show: boolean;
};

const { tabs, param } = defineProps<{
  tabs: Tab[];
  param: string;
}>();

function isActive(tab: Tab): boolean {
  return (
    tab.route.path == route.path && tab.route.query[param] == route.query[param]
  );
}
</script>

<template>
  <nav aria-label="Tabs" class="mb-3">
    <ul class="nav nav-tabs d-flex flex-nowrap">
      <li v-for="tab in tabs" :key="tab.text" class="nav-item d-flex">
        <RouterLink
          v-if="tab.show"
          :to="tab.route"
          class="nav-link text-primary text-nowrap d-flex align-items-center"
          :class="{ active: isActive(tab) }"
        >
          <span class="me-2 text-dark" aria-hidden="true">
            <component :is="tab.icon" v-if="tab.icon" />
          </span>
          {{ tab.text }}
        </RouterLink>
      </li>
    </ul>
  </nav>
</template>

<style scoped>
nav {
  overflow-x: auto;
  overflow-y: hidden;
}
</style>
