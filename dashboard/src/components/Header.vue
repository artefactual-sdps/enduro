<script setup lang="ts">
import Dropdown from "bootstrap/js/dist/dropdown";
import { onMounted, ref } from "vue";

import IconInfo from "~icons/clarity/info-standard-line";
import IconLogout from "~icons/clarity/logout-line";
import IconMenu from "~icons/clarity/menu-line";
import IconUser from "~icons/clarity/user-solid";

import AboutDialog from "@/components/AboutDialog.vue";
import Breadcrumb from "@/components/Breadcrumb.vue";
import InstitutionLogo from "@/components/InstitutionLogo.vue";
import ThemeToggle from "@/components/ThemeToggle.vue";
import { openDialog } from "@/dialogs/dialog";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const userMenuToggle = ref<HTMLElement | null>(null);

onMounted(() => {
  if (userMenuToggle.value) new Dropdown(userMenuToggle.value);
});

const showAbout = () => openDialog(AboutDialog);
</script>

<template>
  <header class="bg-body border-bottom sticky-top">
    <nav class="navbar navbar-expand-md p-0">
      <!-- Open offcanvas button, only visible in sm. -->
      <button
        type="button"
        class="navbar-toggler btn btn-link text-decoration-none border-0 p-3"
        data-bs-toggle="offcanvas"
        data-bs-target="#menu-offcanvas"
        aria-controls="menu-offcanvas"
        aria-label="Open navigation"
      >
        <IconMenu class="text-body fs-4 mx-1" aria-hidden="true" />
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
        <IconMenu class="text-body fs-4 mx-1" aria-hidden="true" />
      </button>

      <RouterLink
        class="navbar-brand h1 mb-0 me-auto p-3 px-2 text-primary text-decoration-none d-flex align-items-center fw-bold"
        :class="layoutStore.sidebarCollapsed ? '' : 'ms-2'"
        :to="{ name: '/' }"
      >
        <img src="/logo.png" alt="" height="30" class="me-2" />
        Enduro
      </RouterLink>

      <div class="flex-grow-1 d-none d-md-block overflow-hidden">
        <Breadcrumb />
      </div>

      <InstitutionLogo />

      <div class="dropdown">
        <button
          ref="userMenuToggle"
          type="button"
          class="btn btn-link p-3"
          data-bs-toggle="dropdown"
          aria-expanded="false"
          aria-label="Open user menu"
        >
          <IconUser class="text-primary fs-4 mx-1" aria-hidden="true" />
        </button>
        <ul class="dropdown-menu dropdown-menu-end mt-0">
          <li>
            <h6 class="dropdown-header">
              {{
                authStore.isEnabled
                  ? authStore.user?.profile.email
                  : "Unauthenticated"
              }}
            </h6>
          </li>
          <li>
            <ThemeToggle />
          </li>
          <li>
            <button
              type="button"
              class="dropdown-item d-flex align-items-center gap-3"
              @click="showAbout"
            >
              <IconInfo aria-hidden="true" />
              <span>About</span>
            </button>
          </li>
          <li v-if="authStore.isEnabled">
            <a
              class="dropdown-item d-flex align-items-center gap-3 text-decoration-none"
              href="#"
              @click.prevent="authStore.signoutRedirect()"
            >
              <IconLogout aria-hidden="true" />
              <span>Sign out</span>
            </a>
          </li>
        </ul>
      </div>
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
