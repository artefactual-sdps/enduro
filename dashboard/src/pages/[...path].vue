<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";

import { useCustomStore } from "@/stores/custom";
import { useLayoutStore } from "@/stores/layout";

const route = useRoute();
const customStore = useCustomStore();
const layoutStore = useLayoutStore();

const loading = ref(true);
const content = ref<string | null>(null);
const error = ref<string | null>(null);
const routePath = computed(() => route.path);

// Watch for route changes and load content.
watch(
  routePath,
  async (path) => {
    if (!customStore.manifest?.routes?.some((r) => r.path === path)) {
      loading.value = false;
      return;
    }

    loading.value = true;
    error.value = null;
    content.value = null;

    const loadedContent = await customStore.loadRouteContent(path);

    if (loadedContent) {
      content.value = loadedContent;
    } else {
      error.value =
        customStore.getRouteError(path) || "Failed to load custom content.";
    }

    loading.value = false;
  },
  { immediate: true },
);

// Get the route name for display from the manifest.
const routeName = computed(() => {
  const routeConfig = customStore.manifest?.routes?.find(
    (r) => r.path === routePath.value,
  );
  return routeConfig?.name || "";
});

// Update breadcrumb.
watch(
  routeName,
  (name) => {
    if (name) {
      layoutStore.updateBreadcrumb([{ text: name }]);
    } else {
      layoutStore.updateBreadcrumb([{ text: "Not found" }]);
    }
  },
  { immediate: true },
);
</script>

<template>
  <div class="container-xxl">
    <!-- Custom route content -->
    <template v-if="routeName">
      <div v-if="loading" class="text-center p-3">
        <div class="spinner-border text-muted" role="status">
          <span class="visually-hidden">Loading...</span>
        </div>
      </div>
      <div v-else-if="error" class="alert alert-warning" role="alert">
        {{ error }}
      </div>
      <div v-else-if="content" v-html="content"></div>
      <div v-else class="alert alert-info" role="alert">
        No content available for this route.
      </div>
    </template>

    <!-- 404 page -->
    <div v-else class="alert alert-warning" role="alert">
      <h4 class="alert-heading">Page not found!</h4>
      <p>We can't find the page you're looking for.</p>
      <hr />
      <router-link class="btn btn-warning" :to="{ name: '/' }"
        >Take me home</router-link
      >
    </div>
  </div>
</template>
