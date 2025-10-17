<script setup lang="ts">
import { openDialog } from "vue3-promise-dialog";

import AboutDialogVue from "@/components/AboutDialog.vue";
import Breadcrumb from "@/components/Breadcrumb.vue";
import InstitutionLogo from "@/components/InstitutionLogo.vue";
import { useLayoutStore } from "@/stores/layout";
import IconInfo from "~icons/clarity/info-standard-solid";
import IconMenu from "~icons/clarity/menu-line";

const layoutStore = useLayoutStore();

const showAbout = async () => await openDialog(AboutDialogVue);

const institution: { logo: string; name: string; url: string } = {
  logo: import.meta.env.VITE_INSTITUTION_LOGO,
  name: import.meta.env.VITE_INSTITUTION_NAME,
  url: import.meta.env.VITE_INSTITUTION_URL,
};
</script>

<template>
  <header class="bg-white border-bottom sticky-top">
    <nav class="navbar navbar-expand-md p-0">
      <!-- Open offcanvas button, only visible in sm. -->
      <button
        type="button"
        class="navbar-toggler btn btn-link text-decoration-none p-3"
        data-bs-toggle="offcanvas"
        data-bs-target="#menu-offcanvas"
        aria-controls="menu-offcanvas"
        aria-label="Open navigation"
      >
        <IconMenu
          class="text-dark mx-1"
          style="font-size: 1.5em"
          aria-hidden="true"
        />
      </button>

      <!-- Collapse/expand sidebar button, visible in md or higher. -->
      <button
        type="button"
        class="btn btn-link text-decoration-none p-3 d-none d-md-block"
        :class="layoutStore.sidebarCollapsed ? 'sidebar-collapsed' : ''"
        :aria-label="
          (layoutStore.sidebarCollapsed ? 'Expand' : 'Collapse') + ' navigation'
        "
        @click="layoutStore.toggleSidebar()"
      >
        <IconMenu
          class="text-dark mx-1"
          style="font-size: 1.5em"
          aria-hidden="true"
        />
      </button>

      <RouterLink
        class="navbar-brand h1 mb-0 me-auto p-3 px-2 text-primary text-decoration-none d-flex align-items-center fw-bold"
        :class="layoutStore.sidebarCollapsed ? '' : 'ms-2'"
        :to="{ name: '/' }"
      >
        <img src="/logo.png" alt="" height="30" class="me-2" />
        Enduro
      </RouterLink>

      <div class="flex-grow-1 d-none d-md-block">
        <Breadcrumb />
      </div>

      <InstitutionLogo
        :logo="institution.logo"
        :name="institution.name"
        :url="institution.url"
      />

      <button
        type="button"
        class="btn btn-link text-decoration-none p-3"
        aria-label="About Enduro"
      >
        <IconInfo
          class="text-primary mx-1"
          style="font-size: 1.5em"
          aria-hidden="true"
          @click="showAbout"
        />
      </button>
    </nav>
  </header>
</template>

<style lang="scss" scoped>
header {
  height: $header-height;
}

.sidebar-collapsed {
  width: $sidebar-collapsed-width;
  min-width: $sidebar-collapsed-width;
}
</style>
