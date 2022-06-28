import PackageLocationDialogVue from "./components/PackageLocationDialog.vue";
import { openDialog } from "vue3-promise-dialog";

export async function openPackageLocationDialog(currentLocation?: string) {
  return await openDialog(PackageLocationDialogVue, {
    currentLocation: currentLocation,
  });
}
