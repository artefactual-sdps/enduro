import { createRouter, createWebHistory } from "vue-router/auto";
import { routes } from "vue-router/auto-routes";

import { useAuthStore } from "@/stores/auth";

const router = createRouter({
  history: createWebHistory("/"),
  strict: false,
  routes,
});

// Routes must end with a slash for comparison. The protected routes
// can specify multiple OR conditions for permissions. For example,
// the "/ingest/upload/" route requires either "ingest:sips:upload"
// or both "ingest:sipsources:objects:list" and "ingest:sips:create".
const signinRoutes: string[] = ["/user/signin/", "/user/signin-callback/"];
const protectedRoutes: Record<string, string[][]> = {
  "/ingest/sips/": [["ingest:sips:list"]],
  "/ingest/sips/[id]/": [["ingest:sips:read"]],
  "/ingest/upload/": [
    ["ingest:sips:upload"],
    ["ingest:sipsources:objects:list", "ingest:sips:create"],
  ],
  "/storage/aips/": [["storage:aips:list"]],
  "/storage/aips/[id]/": [["storage:aips:read"]],
  "/storage/locations/": [["storage:locations:list"]],
  "/storage/locations/[id]/": [["storage:locations:read"]],
  "/storage/locations/[id]/aips/": [["storage:locations:aips:list"]],
};

router.beforeEach(async (to, _, next) => {
  const authStore = useAuthStore();
  await authStore.loadUser();
  const routeName = to.name?.toString() || "";

  // Normalize route name to always end with slash for comparison.
  const name = routeName.endsWith("/") ? routeName : routeName + "/";

  // Helper function to check OR conditions.
  const checkRoutePermissions = (or: string[][]): boolean => {
    return or.some((and) => authStore.checkAttributes(and));
  };

  // TODO: Show alerts when redirecting.
  if (!authStore.isEnabled && signinRoutes.includes(name)) {
    next({ name: "/" });
  } else if (!authStore.isUserValid && !signinRoutes.includes(name)) {
    next({ name: "/user/signin" });
  } else if (
    protectedRoutes[name] !== undefined &&
    !checkRoutePermissions(protectedRoutes[name])
  ) {
    next({ name: "/" });
  } else {
    next();
  }
});

export default router;
