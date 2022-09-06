<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import RawIconBundleLine from "~icons/clarity/bundle-line?raw&width=2em&height=2em";
import RawIconRackServerLine from "~icons/clarity/rack-server-line?raw&width=2em&height=2em";

const menuItems = [
  { routeName: "packages", icon: RawIconBundleLine, text: "Packages" },
  { routeName: "locations", icon: RawIconRackServerLine, text: "Locations" },
];

const layoutStore = useLayoutStore();
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
    <div class="offcanvas-header px-3">
      <h5 class="offcanvas-title" id="offcanvasLabel">Navigation</h5>
      <button
        type="button"
        class="btn-close"
        data-bs-dismiss="offcanvas"
        data-bs-target="#menu-offcanvas"
        aria-label="Close"
      ></button>
    </div>
    <div class="offcanvas-body d-flex flex-column flex-grow-1 pt-0">
      <nav aria-labelledby="offcanvasLabel">
        <ul class="list-unstyled flex-grow-1 mb-0">
          <li v-for="item in menuItems">
            <router-link
              class="d-block py-3 text-decoration-none sidebar-link"
              active-class="active"
              exact-active-class="exact-active"
              :to="{ name: item.routeName }"
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

      .sidebar-link .col-9 {
        font-size: 0.75 * $font-size-base;
      }
    }
  }
}
</style>
