import moment from "moment";

export function FormatDateTime(value: Date | undefined) {
  if (!value || Number.isNaN(value.getTime())) {
    return "";
  }
  return moment(value).format("YYYY-MM-DD HH:mm:ss");
}

export function FormatDateTimeString(value: string) {
  const date = new Date(value);
  return moment(date).format("YYYY-MM-DD HH:mm:ss");
}

export function FormatDuration(from: Date, to: Date) {
  const diff = moment(to).diff(from);
  return moment.duration(diff).humanize();
}

export function HumanFileSize(bytes: number, precision: number = 0): string {
  const base = 1000;

  if (bytes >= base ** 4) {
    return `${(bytes / base ** 4).toFixed(precision)} TB`;
  } else if (bytes >= base ** 3) {
    return `${(bytes / base ** 3).toFixed(precision)} GB`;
  } else if (bytes >= base ** 2) {
    return `${(bytes / base ** 2).toFixed(precision)} MB`;
  } else if (bytes >= base) {
    return `${(bytes / base).toFixed(precision)} KB`;
  } else {
    return `${bytes} bytes`;
  }
}
