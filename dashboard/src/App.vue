<script setup lang="ts">
import Header from "@/components/Header.vue";
import { DialogWrapper } from "vue3-promise-dialog";
import IconAnalyticsLine from "~icons/clarity/analytics-line";
import IconBlocksGroupLine from "~icons/clarity/blocks-group-line";
import IconBundleLine from "~icons/clarity/bundle-line";
import IconFileGroupLine from "~icons/clarity/file-group-line";
import IconHomeLine from "~icons/clarity/home-line";
import IconProcessOnVmLine from "~icons/clarity/process-on-vm-line";
import IconRackServerLine from "~icons/clarity/rack-server-line";
import IconSettingsLine from "~icons/clarity/settings-line";
import IconShieldCheckLine from "~icons/clarity/shield-check-line";
</script>

<template>
  <!--
  <div class="d-flex flex-column min-vh-100" v-if="false">
    <Header />
    <main class="flex-grow-1">
      <router-view></router-view>
    </main>
    <DialogWrapper />
  </div>
  -->

  <div id="new-layout">
    <header
      class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow"
    >
      <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#"
        >Enduro</a
      >
      <button
        class="navbar-toggler position-absolute d-md-none collapsed"
        type="button"
        data-bs-toggle="collapse"
        data-bs-target="#sidebarMenu"
        aria-controls="sidebarMenu"
        aria-expanded="false"
        aria-label="Toggle navigation"
      >
        <span class="navbar-toggler-icon"></span>
      </button>
      <input
        class="form-control form-control-dark w-100 rounded-0 border-0"
        type="text"
        placeholder="Search"
        aria-label="Search"
      />
      <div class="navbar-nav">
        <div class="nav-item text-nowrap">
          <a class="nav-link px-3" href="#">Sign out</a>
        </div>
      </div>
    </header>

    <div class="container-fluid">
      <div class="row">
        <nav
          id="sidebarMenu"
          class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse"
        >
          <div class="position-sticky pt-3 sidebar-sticky">
            <ul class="nav flex-column">
              <li class="nav-item">
                <a class="nav-link" href="#">
                  <span
                    data-feather="file-text"
                    class="align-text-bottom"
                  ></span
                  ><IconHomeLine aria-hidden="true" />
                  Home
                </a>
              </li>
              <li class="nav-item">
                <router-link
                  class="nav-link"
                  active-class="active"
                  :to="{ name: 'packages' }"
                >
                  <span data-feather="home" class="align-text-bottom"></span>
                  <IconBundleLine aria-hidden="true" />
                  Packages
                </router-link>
              </li>
              <li class="nav-item">
                <a class="nav-link" href="#">
                  <span
                    data-feather="file-text"
                    class="align-text-bottom"
                  ></span>
                  <IconBlocksGroupLine aria-hidden="true" />
                  Batches
                </a>
              </li>
              <li class="nav-item">
                <router-link
                  class="nav-link"
                  active-class="active"
                  :to="{ name: 'locations' }"
                >
                  <span data-feather="file" class="align-text-bottom"></span>
                  <IconRackServerLine aria-hidden="true" />
                  Locations
                </router-link>
              </li>
              <li class="nav-item">
                <a class="nav-link" href="#">
                  <span
                    data-feather="file-text"
                    class="align-text-bottom"
                  ></span>
                  <IconProcessOnVmLine aria-hidden="true" />
                  Workflows
                </a>
              </li>
              <li class="nav-item">
                <a class="nav-link" href="#">
                  <span
                    data-feather="file-text"
                    class="align-text-bottom"
                  ></span>
                  <IconShieldCheckLine aria-hidden="true" />
                  Registry
                </a>
              </li>
              <li class="nav-item">
                <a class="nav-link" href="#">
                  <span
                    data-feather="file-text"
                    class="align-text-bottom"
                  ></span>
                  <IconAnalyticsLine aria-hidden="true" />
                  Reports
                </a>
              </li>
              <li class="nav-item">
                <a class="nav-link" href="#">
                  <span
                    data-feather="file-text"
                    class="align-text-bottom"
                  ></span>
                  <IconSettingsLine aria-hidden="true" />
                  Settings
                </a>
              </li>
            </ul>
          </div>
        </nav>

        <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
          <router-view></router-view>
        </main>
      </div>
    </div>
  </div>
</template>

<style>
body {
  font-size: 0.875rem;
}

.feather {
  width: 16px;
  height: 16px;
}

/*
 * Sidebar
 */

.sidebar {
  position: fixed;
  top: 0;
  /* rtl:raw:
  right: 0;
  */
  bottom: 0;
  /* rtl:remove */
  left: 0;
  z-index: 100; /* Behind the navbar */
  padding: 48px 0 0; /* Height of navbar */
  box-shadow: inset -1px 0 0 rgba(0, 0, 0, 0.1);
}

@media (max-width: 767.98px) {
  .sidebar {
    top: 5rem;
  }
}

.sidebar-sticky {
  height: calc(100vh - 48px);
  overflow-x: hidden;
  overflow-y: auto; /* Scrollable contents if viewport is shorter than content. */
}

.sidebar .nav-link {
  font-weight: 500;
  color: #333;
}

.sidebar .nav-link .feather {
  margin-right: 4px;
  color: #727272;
}

.sidebar .nav-link.active {
  color: #2470dc;
}

.sidebar .nav-link:hover .feather,
.sidebar .nav-link.active .feather {
  color: inherit;
}

.sidebar-heading {
  font-size: 0.75rem;
}

/*
 * Navbar
 */

.navbar-brand {
  padding-top: 0.75rem;
  padding-bottom: 0.75rem;
  background-color: rgba(0, 0, 0, 0.25);
  box-shadow: inset -1px 0 0 rgba(0, 0, 0, 0.25);
}

.navbar .navbar-toggler {
  top: 0.25rem;
  right: 1rem;
}

.navbar .form-control {
  padding: 0.75rem 1rem;
}

.form-control-dark {
  color: #fff;
  background-color: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.1);
}

.form-control-dark:focus {
  border-color: transparent;
  box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.25);
}
</style>
