<script setup lang="ts">
import { useLayoutStore } from "./stores/layout";
import Header from "@/components/Header.vue";
import Sidebar from "@/components/Sidebar.vue";
import { DialogWrapper } from "vue3-promise-dialog";

const layoutStore = useLayoutStore();
</script>

<template>
  <div class="d-flex flex-column min-vh-100">
    <div
      class="visually-hidden-focusable p-3 border-bottom"
      v-if="layoutStore.isUserValid"
    >
      <a class="btn btn-sm btn-outline-primary" href="#main"
        >Skip to main content</a
      >
    </div>
    <Header v-if="layoutStore.isUserValid" />
    <div class="flex-grow-1 d-flex">
      <Sidebar v-if="layoutStore.isUserValid" />
      <main class="flex-grow-1 d-flex px-2 pt-3" id="main">
        <router-view></router-view>
      </main>
    </div>
    <DialogWrapper v-if="layoutStore.isUserValid" />
  </div>
</template>
