<script setup lang="ts">
import { runtime } from "@/client";
import PackagePendingAlert from "@/components/PackagePendingAlert.vue";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import { usePackageStore } from "@/stores/package";
import { useAsyncState } from "@vueuse/core";
import { useRoute, useRouter } from "vue-router";

const route = useRoute();
const router = useRouter();
const packageStore = usePackageStore();

const { execute, error } = useAsyncState(
  () => {
    return packageStore.fetchCurrent(route.params.id.toString());
  },
  null,
  {
    onError: (err) => {
      const rerr = err as runtime.ResponseError;
      try {
        if (rerr.response.status == 404) {
          router.push({ name: "all" });
        }
      } catch (err) {}
    },
  }
);
</script>

<template>
  <div>
    <div class="container-fluid pt-3" v-if="error">
      <div class="container-xxl px-0">
        <PageLoadingAlert :execute="execute" :error="error"></PageLoadingAlert>
      </div>
    </div>

    <!-- Navigation bar -->
    <div
      class="container-fluid pt-3 packages-navbar"
      v-if="packageStore.current"
    >
      <div class="container-xxl px-0" v-if="packageStore.current">
        <div class="col">
          <PackagePendingAlert />
        </div>
        <!-- Breadcrumb -->
        <div class="col">
          <nav aria-label="breadcrumb">
            <ol class="breadcrumb">
              <li class="breadcrumb-item">
                <router-link :to="{ name: 'packages' }">Packages</router-link>
              </li>
              <li class="breadcrumb-item active" aria-current="page">
                {{ packageStore.current.name }}
              </li>
            </ol>
          </nav>
        </div>
        <!-- Navigation tabs -->
        <div class="col">
          <ul class="nav nav-tabs">
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
            <li class="nav-item">
              <router-link
                class="nav-link"
                exact-active-class="active"
                :to="{
                  name: 'packages-id-workflow',
                  params: { id: packageStore.current.id },
                }"
                >Workflow</router-link
              >
            </li>
          </ul>
        </div>
      </div>
    </div>

    <div class="container-xxl px-3">
      <div class="pt-3">
        <router-view></router-view>
      </div>
    </div>
  </div>
</template>
