<script setup lang="ts">
import router from "@/router";
import { useAuthStore } from "@/stores/auth";

useAuthStore()
  .signoutCallback()
  .then(() => router.push({ name: "/" }));
</script>

<template><div /></template>
