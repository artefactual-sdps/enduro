import BatchReviewConfirmDialog from "@/components/BatchReviewConfirmDialog.vue";
import { defineDialog } from "@/dialogs/dialog";

export interface BatchReviewConfirmDialogProps {
  heading: string;
  bodyHtml: string;
  confirmClass: "btn-primary" | "btn-danger";
}

export const openBatchReviewConfirmDialog = defineDialog<
  boolean,
  BatchReviewConfirmDialogProps
>(BatchReviewConfirmDialog, false);
