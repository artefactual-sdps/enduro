<script setup lang="ts">
import auth from "@/auth";
import { useLayoutStore } from "@/stores/layout";
import { useRouter } from "vue-router/auto";

const router = useRouter();
auth.signinCallback().then((user) => {
  useLayoutStore().setUser(user || null);
  router.push({ name: "/" });
});
</script>

<template></template>
