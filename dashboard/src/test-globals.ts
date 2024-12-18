export const setup = () => {
  // Set the local timezone to Newfoundland for consistent test results. We are
  // using Regina because it has a constant offset (no DST) from UTC (-06:00).
  process.env.TZ = "America/Regina";
};
