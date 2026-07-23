import AipDeletionRequestDialog from "@/components/AipDeletionRequestDialog.vue";
import { defineDialog } from "@/dialogs/dialog";

export const openAipDeletionRequestDialog = defineDialog<string | null>(
  AipDeletionRequestDialog,
  null,
);
