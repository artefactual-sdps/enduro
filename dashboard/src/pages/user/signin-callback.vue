<script setup lang="ts">
import auth from "@/auth";
import { client } from "@/client";
import { useLayoutStore } from "@/stores/layout";
import { useRouter } from "vue-router/auto";

const router = useRouter();
auth.signinCallback().then((user) => {
  useLayoutStore().setUser(user || null);
  if (user) {
    client.package.packageMonitorRequest().then(() => {
      client.connectPackageMonitor();
    });
  }
  router.push({ name: "/" });
});
</script>

<template></template>
