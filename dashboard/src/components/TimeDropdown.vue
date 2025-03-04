<script setup lang="ts">
import VueDatePicker from "@vuepic/vue-datepicker";
import type { ModelValue } from "@vuepic/vue-datepicker";
import "@vuepic/vue-datepicker/dist/main.css";
import Dropdown from "bootstrap/js/dist/dropdown";
import { onMounted, ref, watch } from "vue";

import IconClose from "~icons/clarity/close-line";

const emit = defineEmits<{
  change: [name: string, start: string, end: string];
}>();

type option = {
  value: string;
  label: string;
};

const options: option[] = [
  { value: "", label: "Select a time range" },
  { value: "3h", label: "The last 3 hours" },
  { value: "6h", label: "The last 6 hours" },
  { value: "12h", label: "The last 12 hours" },
  { value: "24h", label: "The last 24 hours" },
  { value: "3d", label: "The last 3 days" },
  { value: "7d", label: "The last 7 days" },
];

const props = defineProps<{
  name: string;
  label?: string;
  start?: Date;
  end?: Date;
}>();

const el = ref<HTMLElement | null>(null);
const dropdown = ref<Dropdown | null>(null);
const defaultLabel = props.label || "Time";
const btnLabel = ref(defaultLabel);
const selectedPreset = ref("");
const startTime = ref<Date | null>(null);
const endTime = ref<Date | null>(null);

onMounted(() => {
  if (el.value) dropdown.value = new Dropdown(el.value);
  if (props.start) {
    startTime.value = props.start;
    btnLabel.value = defaultLabel + ": Custom";
  }
  if (props.end) {
    endTime.value = props.end;
    btnLabel.value = defaultLabel + ": Custom";
  }
});

watch(selectedPreset, async (newValue) => {
  let sel = options.find((o) => o.value == newValue);
  if (!sel || sel.value === "") {
    return;
  }

  btnLabel.value = defaultLabel + ": " + sel.value;
  startTime.value = earliestTimeFromOption(sel.value);
  endTime.value = null;

  emitChange();
});

const emitChange = () => {
  emit(
    "change",
    props.name,
    formatDate(startTime.value),
    formatDate(endTime.value),
  );
};

const handleCustomTimeChange = () => {
  btnLabel.value = defaultLabel + ": Custom";
  selectedPreset.value = "";

  emitChange();
};

const handleStartChange = (modelData: ModelValue) => {
  startTime.value = modelData as Date;
  handleCustomTimeChange();
};

const handleEndChange = (modelData: ModelValue) => {
  endTime.value = modelData as Date;
  handleCustomTimeChange();
};

const reset = () => {
  btnLabel.value = defaultLabel;
  selectedPreset.value = "";
  startTime.value = null;
  endTime.value = null;

  emitChange();
};

const formatDate = (date: Date | null) => {
  if (!date) return "";
  let t = date.toISOString();
  t = t.split(".")[0] + "Z"; // remove milliseconds.
  return t;
};

const earliestTimeFromOption = (value: string) => {
  // convert hours and days to milliseconds.
  const hour = 60 * 60 * 1000;
  const day = 24 * hour;

  let start = new Date();
  switch (value) {
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
      return new Date(0);
  }

  return start;
};
</script>

<template>
  <div class="dropdown" ref="el">
    <button
      :id="'tdd-' + props.name + '-toggle'"
      class="btn btn-primary dropdown-toggle"
      type="button"
      data-bs-toggle="dropdown"
      aria-expanded="false"
    >
      {{ btnLabel }}
    </button>
    <button
      :id="'tdd-' + props.name + '-reset'"
      @click="reset()"
      class="btn btn-secondary"
      type="reset"
      aria-label="Reset time filter"
      v-show="startTime !== null || endTime !== null"
    >
      <IconClose />
    </button>
    <div :id="'tdd-' + props.name + '-menu'" class="dropdown-menu p-3">
      <h5>Preset range</h5>
      <select
        :id="'tdd-' + props.name + '-preset'"
        name="preset-times"
        class="form-select"
        aria-label="Select a time range"
        v-model="selectedPreset"
      >
        <option
          v-for="item in options"
          :key="item.value"
          :value="item.value"
          :selected="item.value == selectedPreset"
          :disabled="item.value == ''"
        >
          {{ item.label }}
        </option>
      </select>
      <hr />
      <h5>Custom range</h5>
      <div>
        <label :for="'tdd-' + props.name + '-start-input'">From</label>
        <VueDatePicker
          time-picker-inline
          :id="'tdd-' + props.name + '-start'"
          :name="'tdd-' + props.name + '-start-input'"
          v-model="startTime"
          placeholder="Start time"
          @update:model-value="handleStartChange"
        />
      </div>
      <div>
        <label :for="'tdd-' + props.name + '-end-input'">To</label>
        <VueDatePicker
          time-picker-inline
          :id="'tdd-' + props.name + '-end'"
          :name="'tdd-' + props.name + '-end-input'"
          v-model="endTime"
          placeholder="End time"
          @update:model-value="handleEndChange"
        />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.dropdown-menu {
  width: 300px;
}
</style>
