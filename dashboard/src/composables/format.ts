import moment from "moment";

export function formatDateTime(value: Date | undefined) {
  if (!value || Number.isNaN(value.getTime())) {
    return "";
  }
  return moment(value).format("YYYY-MM-DD HH:mm:ss");
}

export function formatDateTimeString(value: string) {
  const date = new Date(value);
  return moment(date).format("YYYY-MM-DD HH:mm:ss");
}

export function formatDuration(from: Date, to: Date) {
  const diff = moment(to).diff(from);
  return moment.duration(diff).humanize();
}

export function humanFileSize(bytes: number, precision = 0): string {
  const base = 1000;
  const units = ["bytes", "KB", "MB", "GB", "TB"];
  let i = 0;
  while (bytes >= base && i < units.length - 1) {
    bytes /= base;
    i++;
  }
  return `${bytes.toFixed(precision)} ${units[i]}`;
}
