<script setup lang="ts">
import type { FunctionalComponent, SVGAttributes } from "vue";
import type { RouteLocationResolved } from "vue-router/auto";

const { tabs } = defineProps<{
  tabs: {
    icon: FunctionalComponent<SVGAttributes, {}>;
    text: string;
    route: RouteLocationResolved;
  }[];
}>();


</script>

<template>
  <nav aria-label="Tabs" class="mb-3">
    <ul class="nav nav-tabs d-flex flex-nowrap">
      <li class="nav-item d-flex" v-for="tab in tabs">
        <router-link
          :to="tab.route"
          class="nav-link text-primary text-nowrap d-flex align-items-center"
          exact-active-class="active"
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
