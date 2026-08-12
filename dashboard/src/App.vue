<script setup lang="ts">
import Modal from "bootstrap/js/dist/modal";
import { watch } from "vue";

import Header from "@/components/Header.vue";
import Sidebar from "@/components/Sidebar.vue";
import DialogHost from "@/dialogs/DialogHost.vue";
import { useAuthStore } from "@/stores/auth";
import { useIngestMonitorStore } from "@/stores/ingestMonitor";
import { useStorageMonitorStore } from "@/stores/storageMonitor";

const authStore = useAuthStore();
authStore.loadConfig();

const ingestMonitor = useIngestMonitorStore();
const storageMonitor = useStorageMonitorStore();

// Connect to the monitor APIs when the user is loaded
// successfully or if authentication is disabled.
watch(
  () => authStore.isUserValid,
  (valid) => {
    if (valid) {
      ingestMonitor.connect();
      storageMonitor.connect();
    } else {
      // Hide any open modals when the user is logged out or invalidated.
      document
        .querySelectorAll<HTMLElement>(".modal.show")
        .forEach((el) => Modal.getInstance(el)?.hide());
    }
  },
  { immediate: true },
);
</script>

<template>
  <div class="d-flex flex-column min-vh-100">
    <div
      v-if="authStore.isUserValid"
      class="visually-hidden-focusable p-3 border-bottom"
    >
      <a class="btn btn-sm btn-outline-primary" href="#main"
        >Skip to main content</a
      >
    </div>
    <Header v-if="authStore.isUserValid" />
    <div class="flex-grow-1 d-flex">
      <Sidebar v-if="authStore.isUserValid" />
      <main
        id="main"
        class="flex-grow-1 d-flex min-w-0"
        :class="{ 'px-2 pt-3': authStore.isUserValid }"
        tabindex="-1"
      >
        <RouterView />
      </main>
    </div>
    <!-- Scope dialogs to valid sessions. Unmounting the host resolves an active
         dialog with undefined. -->
    <DialogHost v-if="authStore.isUserValid" />
  </div>
</template>
