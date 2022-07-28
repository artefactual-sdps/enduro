<script setup lang="ts">
import Offcanvas from "bootstrap/js/dist/offcanvas";
import { onMounted } from "vue";
import TextAlignJustified from "~icons/akar-icons/text-align-justified";

const offcanvasButton = $ref<HTMLElement | null>(null);

onMounted(() => {
  if (offcanvasButton) new Offcanvas(offcanvasButton);
});
</script>

<template>
  <header class="navbar navbar-expand-lg border-bottom">
    <nav class="container-fluid gap-2" aria-label="Main navigation">
      <button
        class="bg-white text-enduro-primary border border-opacity-75 border-enduro-primary rounded-3 d-lg-none"
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
        class="navbar-brand h1 m-0 p-2 text-enduro-primary border border-opacity-75 border-enduro-primary rounded-3"
        :to="{ name: 'index' }"
        >Enduro</router-link
      >
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
