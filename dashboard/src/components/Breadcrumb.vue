<script setup lang="ts">
import { useStateStore } from "@/stores/state";

const stateStore = useStateStore();
</script>

<template>
  <nav aria-label="Breadcrumb" class="d-inline-block">
    <ol class="breadcrumb mb-0">
      <li
        v-for="item in stateStore.breadcrumb"
        class="breadcrumb-item"
        :class="!item.routeName ? 'active' : ''"
      >
        <router-link :to="{ name: item.routeName }" v-if="item.routeName">{{
          item.text
        }}</router-link>
        <span v-else>{{ item.text }}</span>
      </li>
    </ol>
  </nav>
</template>
