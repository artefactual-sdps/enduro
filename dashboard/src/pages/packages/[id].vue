<script setup lang="ts">
import { runtime } from "@/client";
import PackagePendingAlert from "@/components/PackagePendingAlert.vue";
import PageLoadingAlert from "@/components/PageLoadingAlert.vue";
import { usePackageStore } from "@/stores/package";
import { useAsyncState } from "@vueuse/core";
import { useRoute } from "vue-router";

const route = useRoute();
const packageStore = usePackageStore();

const { execute, error } = useAsyncState(
  packageStore.fetchCurrent(route.params.id.toString()),
  null
);
</script>

<template>
  <div>
    <div class="container-xxl pt-3" v-if="error">
      <PageLoadingAlert :execute="execute" :error="error" />
    </div>

    <div class="container-xxl pt-3" v-if="packageStore.current">
      <!-- Alert -->
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

      <div class="pt-3">
        <router-view></router-view>
      </div>
    </div>
  </div>
</template>
