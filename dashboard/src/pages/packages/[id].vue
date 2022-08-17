<script setup lang="ts">
import PackagePendingAlert from "@/components/PackagePendingAlert.vue";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import { usePackageStore } from "@/stores/package";
import { useAsyncState } from "@vueuse/core";
import { useRoute } from "vue-router";
import IconBundleLine from "~icons/clarity/bundle-line";

const route = useRoute();
const packageStore = usePackageStore();

const { execute, error } = useAsyncState(
  packageStore.fetchCurrent(route.params.id.toString()),
  null
);
</script>

<template>
  <div class="container-xxl">
    <PageLoadingAlert v-if="error" :execute="execute" :error="error" />

    <!-- Alert -->
    <PackagePendingAlert v-if="packageStore.current" />

    <h1 class="d-flex mb-3">
      <IconBundleLine class="me-3 text-dark" />{{ packageStore.current?.name }}
    </h1>

    <!-- Navigation tabs -->
    <ul class="nav nav-tabs mb-3" v-if="packageStore.current">
      <li class="nav-item">
        <router-link
          class="nav-link"
          exact-active-class="active"
          :to="{
            name: 'packages-id',
            params: { id: packageStore.current.id },
          }"
          >Overview</router-link
        >
      </li>
    </ul>
    <router-view></router-view>
  </div>
</template>
