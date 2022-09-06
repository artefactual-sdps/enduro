<script setup lang="ts">
import Header from "@/components/Header.vue";
import Sidebar from "@/components/Sidebar.vue";
import { DialogWrapper } from "vue3-promise-dialog";
</script>

<template>
  <div class="d-flex flex-column">
    <div class="visually-hidden-focusable p-3 border-bottom">
      <a class="btn btn-sm btn-outline-primary" href="#main"
        >Skip to main content</a
      >
    </div>
    <Header />
    <div class="flex-grow-1 d-flex">
      <Sidebar />
      <main class="flex-grow-1 px-2 pt-3" id="main">
        <router-view></router-view>
      </main>
    </div>
    <DialogWrapper />
  </div>
</template>
