<script setup lang="ts">
import type { FunctionalComponent, SVGAttributes } from "vue";
import { useRoute } from "vue-router/auto";
import type { RouteLocationResolved } from "vue-router/auto";

const route = useRoute();

type Tab = {
  icon?: FunctionalComponent<SVGAttributes, {}>;
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
      <li class="nav-item d-flex" v-for="tab in tabs">
        <router-link
          v-if="tab.show"
          :to="tab.route"
          class="nav-link text-primary text-nowrap d-flex align-items-center"
          :class="{ active: isActive(tab) }"
        >
          <span v-html="tab.icon" class="me-2 text-dark" aria-hidden="true" />{{
            tab.text
          }}
        </router-link>
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
