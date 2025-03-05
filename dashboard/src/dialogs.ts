import { openDialog } from "vue3-promise-dialog";

import LocationDialog from "./components/LocationDialog.vue";

export async function openLocationDialog(currentLocationId?: string) {
  return await openDialog(LocationDialog, {
    currentLocationId: currentLocationId,
  });
}
