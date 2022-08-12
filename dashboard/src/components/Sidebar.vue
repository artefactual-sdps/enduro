<script setup lang="ts">
import { useStateStore } from "@/stores/state";
import IconAnalyticsLine from "~icons/clarity/analytics-line";
import IconBlocksGroupLine from "~icons/clarity/blocks-group-line";
import RawIconBundleLine from "~icons/clarity/bundle-line?raw&width=2em&height=2em";
import RawIconPinLine from "~icons/clarity/pin-line?raw&width=2em&height=2em";
import RawIconPinSolid from "~icons/clarity/pin-solid?raw&width=2em&height=2em";
import IconFileGroupLine from "~icons/clarity/file-group-line";
import IconProcessOnVmLine from "~icons/clarity/process-on-vm-line";
import RawIconRackServerLine from "~icons/clarity/rack-server-line?raw&width=2em&height=2em";
import IconSearchLine from "~icons/clarity/search-line";
import IconSettingsLine from "~icons/clarity/settings-line";
import IconShieldCheckLine from "~icons/clarity/shield-check-line";
import IconSliderLine from "~icons/clarity/slider-line";
import useEventListener from "@/composables/useEventListener";

const menuItems = [
  { routeName: "packages", icon: RawIconBundleLine, text: "Packages" },
  { routeName: "locations", icon: RawIconRackServerLine, text: "Locations" },
];

const stateStore = useStateStore();
const offcanvas = $ref<HTMLElement | null>(null);
var pinned = $ref<boolean>(false);

useEventListener($$(offcanvas), "mouseenter", (e) => {
  if (!pinned && !offcanvas?.classList.contains("show"))
    stateStore.expandSidebar();
});

useEventListener($$(offcanvas), "mouseleave", (e) => {
  if (!pinned && !offcanvas?.classList.contains("show"))
    stateStore.collapseSidebar();
});
</script>

<template>
  <div
    class="sidebar offcanvas-md offcanvas-start d-flex bg-light"
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
                  :class="stateStore.sidebarCollapsed ? 'd-md-none' : ''"
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
        @click="pinned = !pinned"
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
              <span
                v-html="RawIconPinSolid"
                class="text-primary"
                aria-hidden="true"
                v-if="pinned"
              />
              <span v-html="RawIconPinLine" aria-hidden="true" v-else />
            </div>
            <div
              class="col-9 d-flex align-items-center"
              :class="stateStore.sidebarCollapsed ? 'd-md-none' : ''"
            >
              <span v-if="!pinned">Pin</span>
              <span v-else>Unpin</span>
            </div>
          </div>
        </div>
      </button>
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
  }
}
</style>
