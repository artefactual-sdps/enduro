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
