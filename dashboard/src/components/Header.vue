<script setup lang="ts">
import Breadcrumb from "@/components/Breadcrumb.vue";
import { useStateStore } from "@/stores/state";
import Collapse from "bootstrap/js/dist/collapse";
import Dropdown from "bootstrap/js/dist/dropdown";
import Offcanvas from "bootstrap/js/dist/offcanvas";
import { onMounted } from "vue";
import IconMenuLine from "~icons/clarity/menu-line";

const stateStore = useStateStore();

const offcanvas = $ref<HTMLElement | null>(null);
//const collapse = $ref<HTMLElement | null>(null);
//const dropdown = $ref<HTMLElement | null>(null);

onMounted(() => {
  if (offcanvas) new Offcanvas(offcanvas);
  //if (collapse) new Collapse(collapse);
  //if (dropdown) new Dropdown(dropdown);
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
        @click="stateStore.sidebarCollapsed = !stateStore.sidebarCollapsed"
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

      <!-- SEARCH BOX STUFF
      <button
        ref="collapse"
        type="button"
        class="navbar-toggler btn btn-link text-decoration-none ms-auto"
        data-bs-toggle="collapse"
        data-bs-target="#search-collapse"
        aria-controls="search-collapse"
        aria-expanded="false"
        aria-label="Toggle search"
      >
        <IconSearchLine class="text-dark fs-3" aria-hidden="true" />
      </button>
      
      <div
        class="collapse navbar-collapse py-3 py-md-0"
        id="search-collapse"
      >
        <form class="d-flex flex-grow-1" role="search">
          <input
            class="form-control"
            type="search"
            placeholder="Search"
            aria-label="Search"
          />
          <button type="submit" class="btn btn-link text-decoration-none">
            <IconSliderLine class="text-secondary" aria-hidden="true" />
            <span class="visually-hidden">Search</span>
          </button>
        </form>
      </div>
      -->

      <!-- USER MENU STUFF
      <div class="dropdown me-3">
        <button
          ref="dropdown"
          type="button"
          class="btn btn-link text-dark text-decoration-none dropdown-toggle p-2"
          data-bs-toggle="dropdown"
          aria-expanded="false"
        >
          <img
            src="https://github.com/mdo.png"
            alt="mdo"
            width="32"
            height="32"
            class="rounded-circle"
          />
        </button>
        <ul class="dropdown-menu dropdown-menu-end">
          <li><a class="dropdown-item" href="#">Profile</a></li>
          <li><hr class="dropdown-divider" /></li>
          <li><a class="dropdown-item" href="#">Sign out</a></li>
        </ul>
      </div>
      -->
    </nav>
  </header>
</template>

<style scoped>
.sidebar-collapsed {
  width: 90px;
  min-width: 90px;
}
</style>
