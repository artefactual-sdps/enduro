<script setup lang="ts">
import { useStateStore } from "@/stores/state";
import IconAnalyticsLine from "~icons/clarity/analytics-line";
import IconBlocksGroupLine from "~icons/clarity/blocks-group-line";
import RawIconBundleLine from "~icons/clarity/bundle-line?raw&width=2em&height=2em";
import IconFileGroupLine from "~icons/clarity/file-group-line";
import IconProcessOnVmLine from "~icons/clarity/process-on-vm-line";
import RawIconRackServerLine from "~icons/clarity/rack-server-line?raw&width=2em&height=2em";
import IconSearchLine from "~icons/clarity/search-line";
import IconSettingsLine from "~icons/clarity/settings-line";
import IconShieldCheckLine from "~icons/clarity/shield-check-line";
import IconSliderLine from "~icons/clarity/slider-line";

const menuItems = [
  { routeName: "packages", icon: RawIconBundleLine, text: "Packages" },
  { routeName: "locations", icon: RawIconRackServerLine, text: "Locations" },
];

const stateStore = useStateStore();
</script>

<template>
  <div
    class="sidebar offcanvas-md offcanvas-start d-flex bg-light overflow-auto sticky-md-top"
    :class="stateStore.sidebarCollapsed ? 'collapsed' : ''"
    tabindex="-1"
    id="menu-offcanvas"
    aria-label="offcanvasLabel"
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
      <ul class="list-unstyled flex-grow-1 mb-0">
        <li v-for="item in menuItems">
          <router-link
            class="d-block py-3 text-decoration-none text-dark sidebar-link"
            active-class="bg-primary text-white active"
            :to="{ name: item.routeName }"
          >
            <div class="container-fluid">
              <div class="row">
                <div
                  class="d-flex p-0 col-3 justify-content-end"
                  :class="
                    stateStore.sidebarCollapsed
                      ? 'col-md-12 justify-content-md-center'
                      : ''
                  "
                >
                  <span v-html="item.icon" aria-hidden="true" />
                </div>
                <div
                  class="col-9 d-flex align-items-center"
                  :class="
                    stateStore.sidebarCollapsed
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
    </div>
  </div>
</template>

<style lang="scss" scoped>
.sidebar-link {
  &:hover,
  &:focus {
    background-color: shade-color($light, 25%) !important;

    &.active {
      background-color: shade-color($primary, 25%) !important;
    }
  }
}

@media (min-width: 768px) {
  .sidebar {
    border-right: $border-width $border-style $border-color;
    width: 200px;
    min-width: 200px;

    &.collapsed {
      width: 90px;
      min-width: 90px;

      .sidebar-link .col-9 {
        font-size: 0.75 * $font-size-base;
      }
    }
  }
}
</style>
