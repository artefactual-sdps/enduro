import { api } from "@/client";

function uploader(sip: api.EnduroIngestSip): string {
  return sip.uploaderName || sip.uploaderEmail || sip.uploaderUuid || "Unknown";
}

export default uploader;
