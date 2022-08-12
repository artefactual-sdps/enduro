<script setup lang="ts">
import Header from "@/components/Header.vue";
import Sidebar from "@/components/Sidebar.vue";
import { DialogWrapper } from "vue3-promise-dialog";
</script>

<template>
  <div class="d-flex flex-column">
    <div class="visually-hidden-focusable p-3 border-bottom">
      <a class="btn btn-outline-primary" href="#main">Skip to main content</a>
    </div>
    <Header />
    <div class="flex-grow-1 d-flex" id="content">
      <Sidebar />
      <div class="flex-grow-1 px-2 pt-3 overflow-auto" id="main" role="main">
        <router-view></router-view>
      </div>
    </div>
    <DialogWrapper />
  </div>
</template>

<style scoped>
#content {
  height: calc(100vh - 63px);
}
</style>
