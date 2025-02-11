import { openDialog } from "vue3-promise-dialog";

import PackageLocationDialogVue from "./components/PackageLocationDialog.vue";

export async function openPackageLocationDialog(currentLocationId?: string) {
  return await openDialog(PackageLocationDialogVue, {
    currentLocationId: currentLocationId,
  });
}
