<script setup lang="ts">
import Collapse from "bootstrap/js/dist/collapse";
import { onMounted } from "vue";

const el = $ref<HTMLElement | null>(null);

onMounted(() => {
  if (el) new Collapse(el);
});
</script>

<template>
  <nav class="navbar navbar-dark navbar-expand-sm">
    <div class="container-fluid">
      <router-link class="navbar-brand mb-0 h1" :to="{ name: 'index' }"
        >Enduro</router-link
      >

      <button
        class="navbar-toggler"
        type="button"
        ref="collapseButton"
        data-bs-toggle="collapse"
        data-bs-target="#navbarSupportedContent"
        aria-controls="navbarSupportedContent"
        aria-expanded="false"
        aria-label="Toggle navigation"
      >
        <span class="navbar-toggler-icon"></span>
      </button>

      <div class="collapse navbar-collapse" id="navbarSupportedContent">
        <ul class="navbar-nav me-auto">
          <li class="nav-item">
            <router-link
              class="nav-link"
              active-class="active"
              :to="{ name: 'packages' }"
              >Packages</router-link
            >
          </li>
        </ul>
      </div>
    </div>
  </nav>
</template>

<style lang="scss" scoped>
.navbar {
  user-select: none;
  background-color: var(--bs-enduro-primary);

  @media (min-width: 576px) {
    .nav-link.active {
      color: white;
      background-color: #fff2;
    }
  }

  .nav-link.active {
    border-radius: var(--bs-border-radius);
  }
}
</style>
