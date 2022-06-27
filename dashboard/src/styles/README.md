Notes:

- `main.scss` is imported by `main.ts` - the global stylesheet for the app.
- `bootstrap-base.scss` is also made available to all components using
  `<style lang=scss>` via `vite.config.ts`. It does not impact on the final
  size of the stylesheets because `bootstra-base.scss` does not produce classes
  only variables and mixins.
