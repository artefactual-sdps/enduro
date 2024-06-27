<script setup lang="ts">
import { watch } from "vue";
import { client } from "@/client";
import { useAuthStore } from "@/stores/auth";
import Header from "@/components/Header.vue";
import Sidebar from "@/components/Sidebar.vue";
import { DialogWrapper } from "vue3-promise-dialog";

const authStore = useAuthStore();

// Connect to the package monitor API when the user is loaded successfully.
watch(
  () => authStore.isUserValid,
  (valid) => {
    if (valid) {
      client.package.packageMonitorRequest().then(() => {
        client.connectPackageMonitor();
      });
    }
  },
);
</script>

<template>
  <div class="d-flex flex-column min-vh-100">
    <div
      class="visually-hidden-focusable p-3 border-bottom"
      v-if="authStore.isUserValid"
    >
      <a class="btn btn-sm btn-outline-primary" href="#main"
        >Skip to main content</a
      >
    </div>
    <Header v-if="authStore.isUserValid" />
    <div class="flex-grow-1 d-flex">
      <Sidebar v-if="authStore.isUserValid" />
      <main class="flex-grow-1 d-flex px-2 pt-3" id="main">
        <router-view></router-view>
      </main>
    </div>
    <DialogWrapper v-if="authStore.isUserValid" />
  </div>
</template>
