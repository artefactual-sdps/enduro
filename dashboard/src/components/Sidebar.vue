<script setup lang="ts">
import { useStateStore } from "@/stores/state";
import IconAnalyticsLine from "~icons/clarity/analytics-line";
import IconBlocksGroupLine from "~icons/clarity/blocks-group-line";
import RawIconBundleLine from "~icons/clarity/bundle-line?raw&width=2em&height=2em";
import RawIconCollapseLine from "~icons/clarity/collapse-line?raw&width=2em&height=2em";
import IconFileGroupLine from "~icons/clarity/file-group-line";
import IconProcessOnVmLine from "~icons/clarity/process-on-vm-line";
import RawIconRackServerLine from "~icons/clarity/rack-server-line?raw&width=2em&height=2em";
import IconSearchLine from "~icons/clarity/search-line";
import IconSettingsLine from "~icons/clarity/settings-line";
import IconShieldCheckLine from "~icons/clarity/shield-check-line";
import IconSliderLine from "~icons/clarity/slider-line";

const stateStore = useStateStore();

const menuItems = [
  { routeName: "packages", icon: RawIconBundleLine, text: "Packages" },
  { routeName: "locations", icon: RawIconRackServerLine, text: "Locations" },
];
</script>

<template>
  <div
    class="sidebar offcanvas-md offcanvas-start d-flex border-end bg-light"
    :class="stateStore.sidebarCollapsed ? 'collapsed' : ''"
    tabindex="-1"
    id="menu-offcanvas"
    aria-label="offcanvasLabel"
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
    <div class="offcanvas-body d-flex flex-column flex-grow-1">
      <ul class="list-unstyled flex-grow-1 mb-0">
        <li v-for="item in menuItems">
          <router-link
            class="d-block py-3 text-decoration-none text-dark sidebar-link"
            active-class="bg-enduro-primary text-white active"
            :to="{ name: item.routeName }"
          >
            <div class="container-fluid">
              <div class="row">
                <div
                  class="d-flex p-0"
                  :class="
                    stateStore.sidebarCollapsed
                      ? 'col-12 justify-content-center'
                      : 'col-3 justify-content-end'
                  "
                >
                  <span v-html="item.icon" aria-hidden="true" />
                </div>
                <div
                  class="col-9 d-flex align-items-center"
                  :class="stateStore.sidebarCollapsed ? 'd-none' : ''"
                >
                  {{ item.text }}
                </div>
              </div>
            </div></router-link
          >
        </li>
      </ul>
      <button
        type="button"
        class="btn btn-link text-decoration-none text-dark sidebar-link p-0 py-3 rounded-0 d-none d-md-block"
        @click="stateStore.toggleSidebar()"
      >
        <div class="container-fluid">
          <div class="row">
            <div
              class="d-flex p-0"
              :class="
                stateStore.sidebarCollapsed
                  ? 'col-12 justify-content-center'
                  : 'col-3 justify-content-end'
              "
            >
              <span
                v-html="RawIconCollapseLine"
                aria-hidden="true"
                :style="
                  stateStore.sidebarCollapsed
                    ? 'transform: rotate(90deg)'
                    : 'transform: rotate(270deg)'
                "
              />
            </div>
            <div
              class="col-9 d-flex align-items-center"
              :class="stateStore.sidebarCollapsed ? 'd-none' : ''"
            >
              <span v-if="stateStore.sidebarCollapsed">Expand</span>
              <span v-else>Collapse</span>
            </div>
          </div>
        </div>
      </button>
    </div>
  </div>
</template>

<style lang="scss">
.sidebar-link {
  &:hover,
  &:focus {
    background-color: shade-color($light, 25%) !important;

    &.active {
      background-color: shade-color($enduro-primary, 25%) !important;
    }
  }
}
</style>
