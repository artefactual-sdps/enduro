<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import IconHomeLine from "~icons/clarity/home-line";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
layoutStore.updateBreadcrumb([]);
</script>

<template>
  <div class="container-xxl">
    <h1 class="d-flex mb-3">
      <IconHomeLine class="me-3 text-dark" />Welcome<span
        v-if="authStore.isEnabled"
        >, {{ authStore.getUserDisplayName }}</span
      >!
    </h1>
    <p>
      Enduro is a new application under development by Artefactual Systems.
      Originally created as a more stable replacement for
      <a href="https://www.archivematica.org/" target="_blank">Archivematica</a
      >'s automation tools library of scripts, it has since evolved into a
      flexible tool to be paired with preservation applications to provide
      initial ingest activities such as content and structure validation,
      packaging, and more.
    </p>
    <p>
      See the
      <a href="https://enduro.readthedocs.io/" target="_blank"
        >Enduro documentation</a
      >
      for more information!
    </p>
  </div>
</template>
