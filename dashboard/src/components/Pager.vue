<script setup lang="ts">
import { computed } from "vue";

import IconPrev from "~icons/bi/caret-left-fill";
import IconNext from "~icons/bi/caret-right-fill";

const props = defineProps({
  offset: {
    type: Number,
    default: 0,
  },
  limit: {
    type: Number,
    default: 20,
  },
  total: {
    type: Number,
    required: true,
  },
  maxVisiblePages: {
    type: Number,
    default: 7, // An odd number of pages makes the pager symmetrical.
  },
});

const emit = defineEmits<{
  (e: "page-change", page: number): void;
}>();

const currentPage = computed(() => {
  // Calculate current page based on offset and limit.
  return Math.floor(props.offset / props.limit) + 1;
});

const totalPages = computed(() => {
  // Calculate total pages based on total and limit.
  return Math.ceil(props.total / props.limit);
});

const visiblePages = computed(() => {
  if (totalPages.value <= props.maxVisiblePages) {
    // If total pages is less than max visible, show all pages.
    return Array.from({ length: totalPages.value }, (_, i) => i + 1);
  }

  // Calculate range of pages to show.
  const halfVisible = Math.floor(props.maxVisiblePages / 2);
  let startPage = Math.max(currentPage.value - halfVisible, 1);
  let endPage = startPage + props.maxVisiblePages - 1;

  // Adjust if end page exceeds total pages.
  if (endPage > totalPages.value) {
    endPage = totalPages.value;
    startPage = Math.max(endPage - props.maxVisiblePages + 1, 1);
  }

  return Array.from(
    { length: endPage - startPage + 1 },
    (_, i) => startPage + i,
  );
});

const goToPage = (page: number) => {
  if (page < 1 || page > totalPages.value || page === currentPage.value) {
    return;
  }
  emit("page-change", page);
};
</script>

<template>
  <nav role="navigation" aria-label="Pagination navigation">
    <ul class="pagination justify-content-center">
      <!-- Previous page -->
      <li class="page-item" :class="{ disabled: currentPage === 1 }">
        <a
          id="prev-page"
          class="page-link"
          href="#"
          aria-label="Go to previous page"
          title="Previous page"
          @click.prevent="goToPage(currentPage - 1)"
        >
          <IconPrev />
        </a>
      </li>

      <!-- First page and ellipses-->
      <template v-if="visiblePages[0] > 1">
        <li class="page-item">
          <a
            id="first-page"
            class="page-link"
            href="#"
            @click.prevent="goToPage(1)"
            >1</a
          >
        </li>
        <li v-if="visiblePages[0] > 2" class="page-item" aria-hidden="true">
          <a href="#" class="page-link disabled">…</a>
        </li>
      </template>

      <!-- Page numbers -->
      <li
        v-for="page in visiblePages"
        :key="page"
        class="page-item"
        :class="{ active: currentPage === page }"
      >
        <a
          :id="`page-${page}`"
          class="page-link"
          href="#"
          :aria-label="`Go to page ${page}`"
          @click.prevent="goToPage(page)"
          >{{ page }}</a
        >
      </li>

      <!-- Last page and ellipses-->
      <template v-if="visiblePages[visiblePages.length - 1] < totalPages">
        <li
          v-if="visiblePages[visiblePages.length - 1] < totalPages - 1"
          class="page-item"
          aria-hidden="true"
        >
          <a href="#" class="page-link disabled">…</a>
        </li>
        <li class="page-item">
          <a
            id="last-page"
            class="page-link"
            href="#"
            @click.prevent="goToPage(totalPages)"
            >{{ totalPages }}</a
          >
        </li>
      </template>

      <!-- Next page -->
      <li class="page-item" :class="{ disabled: currentPage === totalPages }">
        <a
          id="next-page"
          class="page-link"
          href="#"
          aria-label="Go to next page"
          title="Next page"
          @click.prevent="goToPage(currentPage + 1)"
        >
          <IconNext />
        </a>
      </li>
    </ul>
  </nav>
</template>
