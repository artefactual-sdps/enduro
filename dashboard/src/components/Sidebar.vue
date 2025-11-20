<script setup lang="ts">
import Collapse from "bootstrap/js/dist/collapse";
import Offcanvas from "bootstrap/js/dist/offcanvas";
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router/auto";

import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import IconAIPs from "~icons/clarity/bundle-line?width=2em&height=2em";
import IconCaret from "~icons/clarity/caret-line";
import IconHome from "~icons/clarity/home-line?width=2em&height=2em";
import IconLogout from "~icons/clarity/logout-line?width=2em&height=2em";
import IconUpload from "~icons/clarity/plus-circle-line?width=2em&height=2em";
import IconUser from "~icons/clarity/user-solid?width=2em&height=2em";
import IconSIPs from "~icons/octicon/package-dependencies-24?width=2em&height=2em";
import IconLocations from "~icons/octicon/server-24?width=2em&height=2em";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const router = useRouter();

const menuItems = [
  {
    route: router.resolve("/"),
    icon: IconHome,
    text: "Home",
    show: true,
  },
  {
    text: "INGEST",
    show: authStore.checkAttributes(["ingest:sips:list"]),
  },
  {
    route: router.resolve("/ingest/sips/"),
    icon: IconSIPs,
    text: "SIPs",
    show: authStore.checkAttributes(["ingest:sips:list"]),
  },
  {
    route: router.resolve("/ingest/upload"),
    icon: IconUpload,
    text: "Upload SIPs",
    show:
      authStore.checkAttributes(["ingest:sips:upload"]) ||
      authStore.checkAttributes([
        "ingest:sipsources:objects:list",
        "ingest:sips:create",
      ]),
  },
  {
    text: "STORAGE",
    show:
      authStore.checkAttributes(["storage:locations:list"]) ||
      authStore.checkAttributes(["storage:aips:list"]),
  },
  {
    route: router.resolve("/storage/locations/"),
    icon: IconLocations,
    text: "Locations",
    show: authStore.checkAttributes(["storage:locations:list"]),
  },
  {
    route: router.resolve("/storage/aips/"),
    icon: IconAIPs,
    text: "AIPs",
    show: authStore.checkAttributes(["storage:aips:list"]),
  },
];

let offcanvasInstance = <Offcanvas | null>null;
const offcanvas = ref<HTMLElement | null>(null);
const collapse = ref<HTMLElement | null>(null);

onMounted(() => {
  if (offcanvas.value) offcanvasInstance = new Offcanvas(offcanvas.value);
  if (collapse.value) new Collapse(collapse.value);
});

const closeOffcanvas = () => {
  if (offcanvasInstance) offcanvasInstance.hide();
};
</script>

<template>
  <div
    id="menu-offcanvas"
    ref="offcanvas"
    class="sidebar offcanvas-md offcanvas-start d-flex bg-light"
    :class="layoutStore.sidebarCollapsed ? 'collapsed' : ''"
    tabindex="-1"
    aria-labelledby="offcanvasLabel"
  >
    <div class="offcanvas-header">
      <h5 id="offcanvasLabel" class="offcanvas-title">Navigation</h5>
      <button
        type="button"
        class="btn-close"
        data-bs-dismiss="offcanvas"
        data-bs-target="#menu-offcanvas"
        aria-label="Close"
      />
    </div>
    <div class="offcanvas-body d-flex flex-grow-1 p-0">
      <nav
        aria-labelledby="offcanvasLabel"
        class="flex-grow-1 d-flex flex-column"
      >
        <ul class="list-unstyled flex-grow-1 mb-0">
          <li v-for="(item, i) in menuItems.filter((it) => it.show)" :key="i">
            <div
              v-if="!item.route"
              class="py-2 text-muted small"
              :class="layoutStore.sidebarCollapsed ? 'text-center' : 'ps-3'"
            >
              {{ item.text }}
            </div>
            <RouterLink
              v-else
              class="d-block py-3 text-decoration-none sidebar-link"
              active-class="active"
              exact-active-class="exact-active"
              :to="item.route"
              @click="closeOffcanvas"
            >
              <div class="container-fluid">
                <div class="row">
                  <div
                    class="d-flex p-0 col-3 justify-content-end"
                    :class="
                      layoutStore.sidebarCollapsed
                        ? 'col-md-12 justify-content-md-center'
                        : ''
                    "
                  >
                    <span aria-hidden="true">
                      <component :is="item.icon" />
                    </span>
                  </div>
                  <div
                    class="col-9 d-flex align-items-center"
                    :class="
                      layoutStore.sidebarCollapsed
                        ? 'col-md-12 justify-content-md-center pt-md-2'
                        : ''
                    "
                  >
                    {{ item.text }}
                  </div>
                </div>
              </div>
            </RouterLink>
          </li>
        </ul>
        <button
          v-if="authStore.isEnabled"
          id="user-menu-button"
          ref="collapse"
          type="button"
          class="btn btn-link d-block p-0 py-3 text-decoration-none text-dark sidebar-link rounded-0 collapsed border-top"
          data-bs-toggle="collapse"
          data-bs-target="#user-menu"
          aria-expanded="false"
          aria-controls="user-menu"
        >
          <div class="container-fluid">
            <div class="row">
              <div
                class="d-flex p-0 col-3 justify-content-end"
                :class="
                  layoutStore.sidebarCollapsed
                    ? 'col-md-12 justify-content-md-center'
                    : ''
                "
              >
                <span class="text-primary" aria-hidden="true">
                  <component :is="IconUser" />
                </span>
              </div>
              <div
                class="col-9 d-flex align-items-center"
                :class="
                  layoutStore.sidebarCollapsed
                    ? 'col-md-12 justify-content-md-center pt-md-2'
                    : ''
                "
              >
                <span class="text-truncate pe-1">{{
                  authStore.getUserDisplayName
                }}</span>
                <IconCaret class="ms-auto" />
              </div>
            </div>
          </div>
        </button>
        <div id="user-menu" class="collapse">
          <a
            class="d-block py-3 text-decoration-none text-dark sidebar-link"
            href="#"
            @click="authStore.signoutRedirect()"
          >
            <div class="container-fluid">
              <div class="row">
                <div
                  class="d-flex p-0 col-3 justify-content-end"
                  :class="
                    layoutStore.sidebarCollapsed
                      ? 'col-md-12 justify-content-md-center'
                      : ''
                  "
                >
                  <span aria-hidden="true">
                    <component :is="IconLogout" />
                  </span>
                </div>
                <div
                  class="col-9 d-flex align-items-center"
                  :class="
                    layoutStore.sidebarCollapsed
                      ? 'col-md-12 justify-content-md-center pt-md-2'
                      : ''
                  "
                >
                  Sign out
                </div>
              </div>
            </div>
          </a>
        </div>
      </nav>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.sidebar-link {
  color: $dark;

  &.active {
    color: $primary;
  }

  &.exact-active {
    color: $white;
    background-color: $primary;
  }

  &:hover,
  &:focus {
    background-color: shade-color($light, 25%) !important;

    &.exact-active {
      background-color: shade-color($primary, 25%) !important;
    }
  }
}

#user-menu-button {
  &:not(.collapsed) {
    background-color: $primary !important;
    color: $white !important;

    &:hover,
    &:focus {
      background-color: shade-color($primary, 25%) !important;
    }

    .col-3 span {
      color: $white !important;
    }

    .col-9 svg {
      transform: rotate(180deg);
    }
  }
}

@media (min-width: 768px) {
  .sidebar {
    position: sticky;
    top: $header-height;
    height: calc(100vh - $header-height);
    overflow-y: auto;
    overflow-x: hidden;
    border-right: $border-width $border-style $border-color;
    width: $sidebar-width;
    min-width: $sidebar-width;

    &.collapsed {
      width: $sidebar-collapsed-width;
      min-width: $sidebar-collapsed-width;

      [class^="col-"] {
        max-width: $sidebar-collapsed-width;
      }

      .sidebar-link .col-9 {
        font-size: 0.75 * $font-size-base;
      }
    }
  }
}
</style>
