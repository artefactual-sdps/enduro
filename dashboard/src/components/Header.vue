<script setup lang="ts">
import Offcanvas from "bootstrap/js/dist/offcanvas";
import { onMounted } from "vue";
import TextAlignJustified from "~icons/akar-icons/text-align-justified";
import Header from "@/components/Header.vue";
import Dropdown from "bootstrap/js/dist/dropdown";
import { DialogWrapper } from "vue3-promise-dialog";
import IconAnalyticsLine from "~icons/clarity/analytics-line";
import IconBlocksGroupLine from "~icons/clarity/blocks-group-line";
import IconBundleLine from "~icons/clarity/bundle-line";
import IconFileGroupLine from "~icons/clarity/file-group-line";
import IconHomeLine from "~icons/clarity/home-line";
import IconMenuLine from "~icons/clarity/menu-line";
import IconProcessOnVmLine from "~icons/clarity/process-on-vm-line";
import IconRackServerLine from "~icons/clarity/rack-server-line";
import IconSettingsLine from "~icons/clarity/settings-line";
import IconShieldCheckLine from "~icons/clarity/shield-check-line";
import IconSliderLine from "~icons/clarity/slider-line";

const offcanvasButton = $ref<HTMLElement | null>(null);

const dropdown = $ref<HTMLElement | null>(null);
onMounted(() => {
  if (dropdown) new Dropdown(dropdown);
});

onMounted(() => {
  if (offcanvasButton) new Offcanvas(offcanvasButton);
});
</script>

<template>
  <header class="navbar navbar-expand-lg border-bottom">
    <nav class="container-fluid gap-2 row" aria-label="Main navigation">
      <div class="col-lg-2 d-flex justify-content-end">
        <button
          class="bg-white text-enduro-primary d-lg-none"
          type="button"
          ref="offcanvasButton"
          data-bs-toggle="offcanvas"
          data-bs-target="#offcanvas"
          aria-controls="offcanvas"
          aria-label="Toggle navigation"
        >
          <TextAlignJustified style="font-size: 2em" />
        </button>

        <router-link
          class="navbar-brand h1 m-0 p-2 text-enduro-primary"
          :to="{ name: 'index' }"
        >
          <img src="/logo.png" alt="" height="30" class="d-inline-block" />
          Enduro</router-link
        >
      </div>
      <div class="col-lg-10">
        <div class="avatar-profile">
          <div class="dropdown text-end">
            <a
              href="#"
              class="d-block link-dark text-decoration-none dropdown-toggle"
              data-bs-toggle="dropdown"
              ref="dropdown"
              aria-expanded="false"
            >
              <img
                src="https://github.com/mdo.png"
                alt="mdo"
                width="32"
                height="32"
                class="rounded-circle"
              />
            </a>
            <ul class="dropdown-menu text-small" style="">
              <li><a class="dropdown-item" href="#">Profile</a></li>
              <li><hr class="dropdown-divider" /></li>
              <li><a class="dropdown-item" href="#">Sign out</a></li>
            </ul>
          </div>
        </div>
      </div>
    </nav>
  </header>

  <div class="flex-grow-1 row">
    <div
      class="col-lg-2 offcanvas-lg offcanvas-start border-end bg-light"
      tabindex="-1"
      id="offcanvas"
      aria-label="offcanvasLabel"
    >
      <div class="offcanvas-header">
        <h5 class="offcanvas-title" id="offcanvasLabel">Navigation</h5>
        <button
          type="button"
          class="btn-close"
          data-bs-dismiss="offcanvas"
          data-bs-target="#offcanvas"
          aria-label="Close"
        ></button>
      </div>
      <div class="offcanvas-body">
        <ul class="list-unstyled p-lg-3">
          <li>
            <router-link active-class="active" :to="{ name: 'packages' }"
              >Packages</router-link
            >
          </li>
        </ul>
      </div>
    </div>
    <div class="col-lg-10 ps-lg-0">
      <router-view></router-view>
    </div>
  </div>
</template>

<style>
.avatar-profile {
  --bs-nav-link-padding-x: 0;
  --bes-nav-link-margin-y: 1rem;
  --bs-nav-link-padding-y: 1rem;
  --bs-nav-link-font-weight: ;
  --bs-nav-link-color: var(--bs-navbar-color);
  --bs-nav-link-hover-color: var(--bs-navbar-hover-color);
  --bs-nav-link-disabled-color: var(--bs-navbar-disabled-color);
  display: flex;
  flex-direction: column;
  padding-left: 0;
  margin-bottom: 0;
  margin-left: 1rem;
  margin-right: 1rem;
  list-style: none;
}
</style>
