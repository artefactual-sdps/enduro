import LocationDialog from "@/components/LocationDialog.vue";
import { defineDialog } from "@/dialogs/dialog";

export interface LocationDialogProps {
  currentLocationId?: string;
}

export const openLocationDialog = defineDialog<
  string | null,
  LocationDialogProps
>(LocationDialog, null);
