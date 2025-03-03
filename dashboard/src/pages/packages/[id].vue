<script setup lang="ts">
import { useAsyncState } from "@vueuse/core";
import { useRoute, useRouter } from "vue-router/auto";

import PackagePendingAlert from "@/components/PackagePendingAlert.vue";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import Tabs from "@/components/Tabs.vue";
import { useAuthStore } from "@/stores/auth";
import { usePackageStore } from "@/stores/package";
import IconBundleLine from "~icons/clarity/bundle-line";
import RawIconDetailsLine from "~icons/clarity/details-line?raw&font-size=20px";

const route = useRoute("/packages/[id]");
const router = useRouter();
const authStore = useAuthStore();
const packageStore = usePackageStore();

const { execute, error } = useAsyncState(
  packageStore.fetchCurrent(route.params.id.toString()),
  null,
);

const tabs = [
  {
    icon: RawIconDetailsLine,
    text: "Summary",
    route: router.resolve({
      name: "/packages/[id]/",
      params: { id: route.params.id },
    }),
    show: authStore.checkAttributes(["ingest:sips:read"]),
  },
];
</script>

<template>
  <div class="container-xxl">
    <PageLoadingAlert v-if="error" :execute="execute" :error="error" />

    <PackagePendingAlert v-if="packageStore.current" />

    <h1 class="d-flex mb-3" v-if="packageStore.current">
      <IconBundleLine class="me-3 text-dark" />{{ packageStore.current.name }}
    </h1>

    <Tabs :tabs="tabs" param="id" />

    <router-view></router-view>
  </div>
</template>
