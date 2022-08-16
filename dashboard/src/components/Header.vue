<script setup lang="ts">
import Breadcrumb from "@/components/Breadcrumb.vue";
import { useStateStore } from "@/stores/state";
import Offcanvas from "bootstrap/js/dist/offcanvas";
import { onMounted } from "vue";
import IconMenuLine from "~icons/clarity/menu-line";

const stateStore = useStateStore();

const offcanvas = $ref<HTMLElement | null>(null);

onMounted(() => {
  if (offcanvas) new Offcanvas(offcanvas);
});
</script>

<template>
  <header class="border-bottom sticky-top">
    <nav class="navbar navbar-expand-md p-0">
      <!-- Open offcanvas button, only visible in sm. -->
      <button
        ref="offcanvas"
        type="button"
        class="navbar-toggler btn btn-link text-decoration-none p-3"
        data-bs-toggle="offcanvas"
        data-bs-target="#menu-offcanvas"
        aria-controls="menu-offcanvas"
        aria-label="Open navigation"
      >
        <IconMenuLine
          class="text-dark mx-1"
          style="font-size: 1.5em"
          aria-hidden="true"
        />
      </button>

      <!-- Collapse/expand sidebar button, visible in md or higher. -->
      <button
        type="button"
        class="btn btn-link text-decoration-none p-3 d-none d-md-block"
        :class="stateStore.sidebarCollapsed ? 'sidebar-collapsed' : ''"
        :aria-label="
          (stateStore.sidebarCollapsed ? 'Expand' : 'Collapse') + ' navigation'
        "
        @click="stateStore.toggleSidebar()"
      >
        <IconMenuLine
          class="text-dark mx-1"
          style="font-size: 1.5em"
          aria-hidden="true"
        />
      </button>

      <router-link
        class="navbar-brand h1 mb-0 me-auto p-3 px-2 text-primary text-decoration-none d-flex align-items-center"
        :class="stateStore.sidebarCollapsed ? '' : 'ms-2'"
        :to="{ name: 'index' }"
      >
        <img src="/logo.png" alt="" height="30" class="me-2" />
        Enduro
      </router-link>

      <div class="flex-grow-1 d-none d-md-block">
        <span class="text-muted me-2">/</span>
        <Breadcrumb />
      </div>
    </nav>
  </header>
</template>

<style scoped>
.sidebar-collapsed {
  width: 90px;
  min-width: 90px;
}
</style>
