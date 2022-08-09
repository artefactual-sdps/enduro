<script setup lang="ts">
import Collapse from "bootstrap/js/dist/collapse";
import Dropdown from "bootstrap/js/dist/dropdown";
import Offcanvas from "bootstrap/js/dist/offcanvas";
import { onMounted } from "vue";
import { DialogWrapper } from "vue3-promise-dialog";
import IconAnalyticsLine from "~icons/clarity/analytics-line";
import IconBlocksGroupLine from "~icons/clarity/blocks-group-line";
import IconBundleLine from "~icons/clarity/bundle-line";
import IconFileGroupLine from "~icons/clarity/file-group-line";
import IconHomeLine from "~icons/clarity/home-line";
import IconMenuLine from "~icons/clarity/menu-line";
import IconProcessOnVmLine from "~icons/clarity/process-on-vm-line";
import IconRackServerLine from "~icons/clarity/rack-server-line";
import IconSearchLine from "~icons/clarity/search-line";
import IconSettingsLine from "~icons/clarity/settings-line";
import IconShieldCheckLine from "~icons/clarity/shield-check-line";
import IconSliderLine from "~icons/clarity/slider-line";

const offcanvas = $ref<HTMLElement | null>(null);
const collapse = $ref<HTMLElement | null>(null);
const dropdown = $ref<HTMLElement | null>(null);

onMounted(() => {
  if (offcanvas) new Offcanvas(offcanvas);
  if (collapse) new Collapse(collapse);
  if (dropdown) new Dropdown(dropdown);
});
</script>

<template>
  <header class="border-bottom">
    <nav class="navbar navbar-expand-md">
      <div class="container-fluid">
        <button
          ref="offcanvas"
          type="button"
          class="navbar-toggler btn btn-link text-decoration-none"
          data-bs-toggle="offcanvas"
          data-bs-target="#menu-offcanvas"
          aria-controls="menu-offcanvas"
          aria-label="Toggle menu"
        >
          <IconMenuLine class="text-dark fs-2" aria-hidden="true" />
        </button>
        <div class="col-2">
          <router-link
            class="navbar-brand me-auto me-md-3 p-2 text-enduro-primary"
            :to="{ name: 'index' }"
          >
            <img
              src="/logo.png"
              alt=""
              height="35"
              class="d-inline-block align-text-middle"
            />
            Enduro</router-link
          >
        </div>

        <div class="d-none d-md-block px-1">
          <nav aria-label="breadcrumb">
            <ol class="breadcrumb mb-0">
              <li class="breadcrumb-item active" aria-current="page">
                / Packages
              </li>
            </ol>
          </nav>
        </div>

        <button
          ref="collapse"
          type="button"
          class="navbar-toggler btn btn-link text-decoration-none ms-auto"
          data-bs-toggle="collapse"
          data-bs-target="#search-collapse"
          aria-controls="search-collapse"
          aria-expanded="false"
          aria-label="Toggle search"
        >
          <IconSearchLine class="text-dark fs-3" aria-hidden="true" />
        </button>

        <div class="collapse navbar-collapse py-3 py-md-0" id="search-collapse">
          <form class="d-flex flex-grow-1" role="search">
            <input
              class="form-control"
              type="search"
              placeholder="Search"
              aria-label="Search"
            />
            <button type="submit" class="btn btn-link text-decoration-none">
              <IconSliderLine class="text-secondary" aria-hidden="true" />
              <span class="visually-hidden">Search</span>
            </button>
          </form>
        </div>

        <div class="dropdown mx-2">
          <button
            ref="dropdown"
            type="button"
            class="btn btn-link text-dark text-decoration-none dropdown-toggle p-2"
            data-bs-toggle="dropdown"
            aria-expanded="false"
          >
            <img
              src="https://github.com/mdo.png"
              alt="mdo"
              width="32"
              height="32"
              class="rounded-circle"
            />
          </button>
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
      class="col-md-2 offcanvas-md offcanvas-start d-flex border-end bg-light pe-0"
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
              class="d-block p-3 text-decoration-none text-dark"
              active-class="bg-enduro-primary text-white"
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
          <li>
            <router-link
              class="d-block p-3 text-decoration-none text-dark"
              active-class="bg-enduro-primary text-white"
              :to="{ name: 'index' }"
            >
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
    <div class="col-md-10 ps-md-0">
      <router-view></router-view>
    </div>
  </div>
</template>
