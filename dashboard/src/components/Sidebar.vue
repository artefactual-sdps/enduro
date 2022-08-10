<script setup lang="ts">
import IconAnalyticsLine from "~icons/clarity/analytics-line";
import IconBlocksGroupLine from "~icons/clarity/blocks-group-line";
import IconBundleLine from "~icons/clarity/bundle-line";
import IconFileGroupLine from "~icons/clarity/file-group-line";
import IconHomeLine from "~icons/clarity/home-line";
import IconProcessOnVmLine from "~icons/clarity/process-on-vm-line";
import IconRackServerLine from "~icons/clarity/rack-server-line";
import IconSearchLine from "~icons/clarity/search-line";
import IconSettingsLine from "~icons/clarity/settings-line";
import IconShieldCheckLine from "~icons/clarity/shield-check-line";
import IconSliderLine from "~icons/clarity/slider-line";
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
        <li>
          <router-link
            class="d-block py-3 text-decoration-none text-dark"
            active-class="bg-enduro-primary text-white"
            :to="{ name: 'packages' }"
          >
            <div class="container-fluid">
              <div class="row">
                <div class="col-3 d-flex justify-content-end p-0">
                  <IconBundleLine style="font-size: 1.5em" aria-hidden="true" />
                </div>
                <div class="col-9">Packages</div>
              </div>
            </div></router-link
          >
        </li>
        <li>
          <router-link
            class="d-block py-3 text-decoration-none text-dark"
            active-class="bg-enduro-primary text-white"
            :to="{ name: 'locations' }"
          >
            <div class="container-fluid">
              <div class="row">
                <div class="col-3 d-flex justify-content-end p-0">
                  <IconRackServerLine
                    style="font-size: 1.5em"
                    aria-hidden="true"
                  />
                </div>
                <div class="col-9">Locations</div>
              </div>
            </div></router-link
          >
        </li>
      </ul>
    </div>
  </div>
</template>
