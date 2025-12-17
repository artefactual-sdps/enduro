import { api } from "@/client";

function uploader(item: api.EnduroIngestSip | api.EnduroIngestBatch): string {
  return (
    item.uploaderName || item.uploaderEmail || item.uploaderUuid || "Unknown"
  );
}

export default uploader;
