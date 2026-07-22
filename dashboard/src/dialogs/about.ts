import AboutDialog from "@/components/AboutDialog.vue";
import { defineDialog } from "@/dialogs/dialog";

export const openAboutDialog = defineDialog<void>(AboutDialog, undefined);
