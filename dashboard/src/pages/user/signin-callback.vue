<script setup lang="ts">
import router from "@/router";
import { useAuthStore } from "@/stores/auth";

useAuthStore()
  .signinCallback()
  .then(() => router.push({ name: "/" }));
</script>

<template></template>
