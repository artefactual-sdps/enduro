<script setup lang="ts">
import Header from "@/components/Header.vue";
import Dropdown from "bootstrap/js/dist/dropdown";
import Offcanvas from "bootstrap/js/dist/offcanvas";
import { onMounted } from "vue";
import { DialogWrapper } from "vue3-promise-dialog";
import TextAlignJustified from "~icons/akar-icons/text-align-justified";
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
      <div class="col-lg-10 d-flex">
        <form class="d-flex flex-grow-1" role="search">
          <input
            class="form-control me-2"
            type="search"
            placeholder="Search"
            aria-label="Search"
          />
          <button
            class="btn btn-link text-decoration-none ms-auto p-0"
            type="submit"
          >
            <IconSliderLine class="text-secondary" aria-hidden="true" />
            <span class="visually-hidden">Search</span>
          </button>
        </form>

        <div class="dropdown ms-3">
          <a
            href="#"
            class="link-dark text-decoration-none dropdown-toggle"
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
          <ul class="dropdown-menu dropdown-menu-end">
            <li><a class="dropdown-item" href="#">Profile</a></li>
            <li><hr class="dropdown-divider" /></li>
            <li><a class="dropdown-item" href="#">Sign out</a></li>
          </ul>
        </div>
      </div>
    </nav>
  </header>

  <div class="flex-grow-1 row">
    <div
      class="col-lg-2 offcanvas-lg offcanvas-start d-flex border-end bg-light pe-0"
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
      <div class="offcanvas-body d-flex flex-grow-1">
        <ul class="list-unstyled flex-grow-1">
          <li class="p-3">
            <router-link
              active-class="bg-white text-decoration-none text-dark"
              :to="{ name: 'packages' }"
            >
              <div class="row">
                <div class="col-3 d-flex justify-content-end p-0">
                  <IconBundleLine style="font-size: 1.5em" aria-hidden="true" />
                </div>
                <div class="col-9">Packages</div>
              </div></router-link
            >
          </li>
          <li class="p-3">
            <router-link active-class="bg-dark" :to="{ name: 'index' }">
              <div class="row">
                <div class="col-3 d-flex justify-content-end p-0">
                  <IconRackServerLine
                    style="font-size: 1.5em"
                    aria-hidden="true"
                  />
                </div>
                <div class="col-9">Locations</div>
              </div></router-link
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
/*
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


.search-block {
  padding: 0 2rem;
  flex-grow: 1;
  align-self: stretch;
  padding-left: 0;
  margin-bottom: 0;
  margin-left: 0;
  margin-right: 1rem;
}
input {
  min-height: 2.5rem;
  flex-grow: 1;
  width: 100%;
  border-color: #e1e1e1;
  border-style: solid;
  border-width: 1.5px;
}
input:hover {
  border-color: var(--enduro-purple);
}
.search-config {
  color: var(--enduro-purple);
}
*/
</style>
