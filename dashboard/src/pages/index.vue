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
      Enduro is a new application under development by
      <a href="https://www.artefactual.com/" target="_blank"
        >Artefactual Systems</a
      >. Originally created as a more stable replacement for Archivematica's
      <a href="https://github.com/artefactual/automation-tools" target="_blank"
        >automation-tools</a
      >
      library of scripts, it has since evolved into a flexible tool to be paired
      with preservation applications like
      <a href="https://www.archivematica.org/" target="_blank">Archivematica</a>
      and
      <a href="https://github.com/artefactual-labs/a3m" target="_blank">a3m</a>
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
</template>
