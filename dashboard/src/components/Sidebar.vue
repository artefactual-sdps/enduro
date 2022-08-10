<script setup lang="ts">
import RawIconBundleLine from "~icons/clarity/bundle-line?raw&width=2em&height=2em";
import RawIconRackServerLine from "~icons/clarity/rack-server-line?raw&width=2em&height=2em";
import IconAnalyticsLine from "~icons/clarity/analytics-line";
import IconBlocksGroupLine from "~icons/clarity/blocks-group-line";
import IconFileGroupLine from "~icons/clarity/file-group-line";
import IconHomeLine from "~icons/clarity/home-line";
import IconProcessOnVmLine from "~icons/clarity/process-on-vm-line";
import IconSearchLine from "~icons/clarity/search-line";
import IconSettingsLine from "~icons/clarity/settings-line";
import IconShieldCheckLine from "~icons/clarity/shield-check-line";
import IconSliderLine from "~icons/clarity/slider-line";

const menuItems = [
  { routeName: "packages", icon: RawIconBundleLine, text: "Packages" },
  { routeName: "locations", icon: RawIconRackServerLine, text: "Locations" },
];
</script>

<template>
  <div
    class="col-md-2 p-0 offcanvas-md offcanvas-start d-flex flex-grow-1 border-end bg-light"
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
    <div class="offcanvas-body d-flex flex-grow-1">
      <ul class="list-unstyled flex-grow-1">
        <li v-for="item in menuItems">
          <router-link
            class="d-block py-3 text-decoration-none text-dark"
            active-class="bg-enduro-primary text-white"
            :to="{ name: item.routeName }"
          >
            <div class="container-fluid">
              <div class="row">
                <div class="col-3 d-flex justify-content-end p-0">
                  <span v-html="item.icon" aria-hidden="true" />
                </div>
                <div class="col-9 d-flex align-items-center">
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
