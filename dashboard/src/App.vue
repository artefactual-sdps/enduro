<script setup lang="ts">
import { watch } from "vue";
import { DialogWrapper } from "vue3-promise-dialog";

import { client } from "@/client";
import Header from "@/components/Header.vue";
import Sidebar from "@/components/Sidebar.vue";
import { useAuthStore } from "@/stores/auth";

const authStore = useAuthStore();
authStore.loadConfig();

// Connect to the monitor APIs when the user is loaded
// successfully or if authentication is disabled.
watch(
  () => authStore.isUserValid,
  (valid) => {
    if (valid) {
      client.ingest.ingestMonitorRequest().then(() => {
        client.connectIngestMonitor();
      });
      client.storage.storageMonitorRequest().then(() => {
        client.connectStorageMonitor();
      });
    }
  },
  { immediate: true },
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
