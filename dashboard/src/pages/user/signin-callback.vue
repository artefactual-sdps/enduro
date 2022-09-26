<script setup lang="ts">
import auth from "@/auth";
import { useLayoutStore } from "@/stores/layout";
import { useRouter } from "vue-router";

const router = useRouter();
auth.signinCallback().then((user) => {
  useLayoutStore().setUser(user || null);
  router.push({ name: "index" });
});
</script>

<template></template>
