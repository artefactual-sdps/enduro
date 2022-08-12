<script setup lang="ts">
import Header from "@/components/Header.vue";
import Sidebar from "@/components/Sidebar.vue";
import { DialogWrapper } from "vue3-promise-dialog";
</script>

<template>
  <div class="d-flex flex-column min-vh-100">
    <div class="visually-hidden-focusable p-3 border-bottom">
      <a class="btn btn-outline-primary" href="#main">Skip to main content</a>
    </div>
    <Header />
    <div class="flex-grow-1 d-flex">
      <Sidebar />
      <div class="flex-grow-1 p-2" id="main" role="main">
        <router-view></router-view>
      </div>
    </div>
    <DialogWrapper />
  </div>
</template>
