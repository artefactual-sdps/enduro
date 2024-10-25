<script setup lang="ts">
import { usePackageStore } from "@/stores/package";
import { PackageListStatusEnum } from "@/openapi-generator";
import IconCheckCircleLine from "~icons/clarity/check-circle-line";
import IconTimesCircleLine from "~icons/clarity/times-circle-line";
import IconPlayLine from "~icons/clarity/play-line";
import IconBarsLine from "~icons/clarity/bars-line";

import type { FunctionalComponent } from "vue";

const packageStore = usePackageStore();

class tab {
  value: string;
  label: string;
  icon?: FunctionalComponent;

  constructor(v: string, l: string, i?: FunctionalComponent) {
    this.value = v;
    this.label = l;
    this.icon = i;
  }
}

const tabs = [
  new tab("", "All"),
  new tab("done", "Done", IconCheckCircleLine),
  new tab("error", "Error", IconTimesCircleLine),
  new tab("in progress", "In progress", IconPlayLine),
  new tab("queued", "Queued", IconBarsLine),
];

const iconName = "IconCheckCircleLine";

function changeStatusFilter(s: string) {
  if (s == packageStore.filters.status) {
    return;
  }
  packageStore.filters.status = s as PackageListStatusEnum;
  packageStore.fetchPackages(1);
}
</script>

<template>
  <ul class="nav nav-tabs">
    <li v-for="tab in tabs">
      <a
        :class="[
          'nav-link',
          packageStore.filters.status == tab.value ? 'active' : '',
        ]"
        :aria-current="packageStore.filters.status == tab.value"
        href="#"
        @click="changeStatusFilter(tab.value)"
      >
        <component :is="tab.icon" />&nbsp;{{ tab.label }}
      </a>
    </li>
  </ul>
</template>
