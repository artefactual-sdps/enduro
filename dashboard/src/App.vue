<script setup lang="ts">
import Header from "./components/Header.vue";
</script>

<template>
  <div class="d-flex flex-column min-vh-100">
    <Header />
    <main class="flex-grow-1 pt-3 px-3">
      <router-view></router-view>
    </main>
  </div>
</template>
