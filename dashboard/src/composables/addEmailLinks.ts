import DOMPurify from "dompurify";

export function addEmailLinks(input: string): string {
  // Sanitize and replace email addresses with mailto links.
  return DOMPurify.sanitize(input).replace(
    /([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})(?=\s|$|[.,!?])/g,
    '<a href="mailto:$1">$1</a>',
  );
}
