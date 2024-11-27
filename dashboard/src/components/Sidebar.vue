<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { useRouter } from "vue-router/auto";
import Collapse from "bootstrap/js/dist/collapse";
import Offcanvas from "bootstrap/js/dist/offcanvas";
import { onMounted, ref } from "vue";
import RawIconBundleLine from "~icons/clarity/bundle-line?raw&width=2em&height=2em";
import IconCaretLine from "~icons/clarity/caret-line";
import RawIconLogoutLine from "~icons/clarity/logout-line?raw&width=2em&height=2em";
import RawIconRackServerLine from "~icons/clarity/rack-server-line?raw&width=2em&height=2em";
import RawIconUserSolid from "~icons/clarity/user-solid?raw&width=2em&height=2em";
import RawIconHomeLine from "~icons/clarity/home-line?raw&width=2em&height=2em";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const router = useRouter();

const menuItems = [
  {
    route: router.resolve("/"),
    icon: RawIconHomeLine,
    text: "Home",
    show: true,
  },
  {
    route: router.resolve("/packages/"),
    icon: RawIconBundleLine,
    text: "Packages",
    show: authStore.checkAttributes(["package:list"]),
  },
  {
    route: router.resolve("/locations/"),
    icon: RawIconRackServerLine,
    text: "Locations",
    show: authStore.checkAttributes(["storage:location:list"]),
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
    class="sidebar offcanvas-md offcanvas-start d-flex bg-light"
    :class="layoutStore.sidebarCollapsed ? 'collapsed' : ''"
    tabindex="-1"
    id="menu-offcanvas"
    aria-labelledby="offcanvasLabel"
    ref="offcanvas"
  >
    <div class="offcanvas-header">
      <h5 class="offcanvas-title" id="offcanvasLabel">Navigation</h5>
      <button
        type="button"
        class="btn-close"
        data-bs-dismiss="offcanvas"
        data-bs-target="#menu-offcanvas"
        aria-label="Close"
      ></button>
    </div>
    <div class="offcanvas-body d-flex flex-grow-1 p-0">
      <nav
        aria-labelledby="offcanvasLabel"
        class="flex-grow-1 d-flex flex-column"
      >
        <ul class="list-unstyled flex-grow-1 mb-0">
          <li v-for="item in menuItems">
            <router-link
              v-if="item.show"
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
                    <span v-html="item.icon" aria-hidden="true" />
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
              </div></router-link
            >
          </li>
        </ul>
        <button
          v-if="authStore.isEnabled"
          ref="collapse"
          type="button"
          id="user-menu-button"
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
                <span
                  v-html="RawIconUserSolid"
                  class="text-primary"
                  aria-hidden="true"
                />
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
                <IconCaretLine class="ms-auto" />
              </div>
            </div>
          </div>
        </button>
        <div class="collapse" id="user-menu">
          <a
            class="d-block py-3 text-decoration-none text-dark sidebar-link"
            @click="authStore.signoutRedirect()"
            href="#"
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
                  <span v-html="RawIconLogoutLine" aria-hidden="true" />
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
