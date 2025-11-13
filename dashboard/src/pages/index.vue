<script setup lang="ts">
import { onMounted } from "vue";

import { useAuthStore } from "@/stores/auth";
import { useCustomStore } from "@/stores/custom";
import { useLayoutStore } from "@/stores/layout";
import IconHome from "~icons/clarity/home-line";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const customStore = useCustomStore();

layoutStore.updateBreadcrumb([]);

onMounted(async () => {
  await customStore.loadHomeContent();
});
</script>

<template>
  <div class="container-xxl">
    <!-- Custom HTML content -->
    <template v-if="customStore.manifest?.homeUrl">
      <div v-if="customStore.homeLoading" class="text-center p-3">
        <div class="spinner-border text-muted" role="status">
          <span class="visually-hidden">Loading...</span>
        </div>
      </div>
      <div
        v-else-if="customStore.homeError"
        class="alert alert-warning"
        role="alert"
      >
        {{ customStore.homeError }}
      </div>
      <div
        v-else-if="customStore.homeContent"
        v-html="customStore.homeContent"
      ></div>
    </template>

    <!-- Default content -->
    <div v-if="!customStore.manifest?.homeUrl || customStore.homeError">
      <h1 class="d-flex mb-3">
        <IconHome class="me-3 text-dark" />Welcome<span
          v-if="authStore.isEnabled"
          >, {{ authStore.getUserDisplayName }}</span
        >!
      </h1>
      <p>
        Enduro is a new application under development by
        <a href="https://www.artefactual.com/" target="_blank"
          >Artefactual Systems</a
        >. Originally created as a more stable replacement for Archivematica's
        <a
          href="https://github.com/artefactual/automation-tools"
          target="_blank"
          >automation-tools</a
        >
        library of scripts, it has since evolved into a flexible tool to be
        paired with preservation applications like
        <a href="https://www.archivematica.org/" target="_blank"
          >Archivematica</a
        >
        and
        <a href="https://github.com/artefactual-labs/a3m" target="_blank"
          >a3m</a
        >
        to provide initial ingest activities such as content and structure
        validation, packaging, and more.
      </p>
      <p>
        See the
        <a href="https://enduro.readthedocs.io/" target="_blank"
          >Enduro documentation</a
        >
        for more information!
      </p>
    </div>
  </div>
</template>
