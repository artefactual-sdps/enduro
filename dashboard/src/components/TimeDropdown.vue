<script setup lang="ts">
import { onMounted, ref, useTemplateRef } from "vue";

const emit = defineEmits<{ change: [field: string, value: string] }>();

type option = {
  value: string;
  label: string;
};

const options: option[] = [
  { value: "", label: "Any time" },
  { value: "3h", label: "The last 3 hours" },
  { value: "6h", label: "The last 6 hours" },
  { value: "12h", label: "The last 12 hours" },
  { value: "24h", label: "The last 24 hours" },
  { value: "3d", label: "The last 3 days" },
  { value: "7d", label: "The last 7 days" },
];

const props = defineProps<{
  fieldname: string;
}>();

const label = ref("Started");
const selected = ref(options[0]);
const dropdown = useTemplateRef("date-filter");

onMounted(() => {
  if (dropdown.value) {
    dropdown.value.style.display = "none";
  }
});

const toggle = () => {
  if (dropdown.value) {
    if (dropdown.value.style.display == "none") {
      dropdown.value.style.display = "block";
    } else if (dropdown.value.style.display == "block") {
      dropdown.value.style.display = "none";
    }
  }
};

const handleChange = (opt: option) => {
  selected.value = opt;

  if (opt.value === "") {
    label.value = "Started";
  } else {
    label.value = "Started: " + opt.label;
  }

  toggle();
  emit("change", props.fieldname, earliestTimeFromOption(opt));
};

const earliestTimeFromOption = (opt: option) => {
  const formatDate = (date: Date) => {
    let t = date.toISOString();
    t = t.split(".")[0] + "Z"; // remove milliseconds.
    return t;
  };

  // convert hours and days to milliseconds.
  const hour = 60 * 60 * 1000;
  const day = 24 * hour;

  let start = new Date();
  switch (opt.value) {
    case "3h":
      start = new Date(Date.now() - 3 * hour);
      break;
    case "6h":
      start = new Date(Date.now() - 6 * hour);
      break;
    case "12h":
      start = new Date(Date.now() - 12 * hour);
      break;
    case "24h":
      start = new Date(Date.now() - 24 * hour);
      break;
    case "3d":
      start = new Date(Date.now() - 3 * day);
      break;
    case "7d":
      start = new Date(Date.now() - 7 * day);
      break;
    default:
      return "";
  }

  return formatDate(start);
};
</script>

<template>
  <div class="dropdown">
    <button
      @click="toggle"
      class="btn btn-primary dropdown-toggle"
      type="button"
      data-bs-toggle="dropdown"
      aria-expanded="false"
    >
      {{ label }}
    </button>
    <ul ref="date-filter" class="dropdown-menu">
      <li v-for="item in options" :key="item.value">
        <a class="dropdown-item" href="#" @click.prevent="handleChange(item)">{{
          item.label
        }}</a>
      </li>
    </ul>
  </div>
</template>
