import { createRouter, createWebHistory } from "vue-router/auto";
import { routes } from "vue-router/auto-routes";

import { useAuthStore } from "@/stores/auth";

const router = createRouter({
  history: createWebHistory("/"),
  strict: false,
  routes,
});

const signinRoutes: string[] = ["/user/signin", "/user/signin-callback"];
const protectedRoutes: Record<string, string[]> = {
  "/ingest/sips/": ["ingest:sips:list"],
  "/ingest/sips/[id]/": ["ingest:sips:read"],
  "/ingest/upload": ["ingest:sips:upload"],
  "/storage/aips/": ["storage:aips:list"],
  "/storage/aips/[id]/": ["storage:aips:read"],
  "/storage/locations/": ["storage:locations:list"],
  "/storage/locations/[id]/": ["storage:locations:read"],
  "/storage/locations/[id]/aips": ["storage:locations:aips:list"],
};

router.beforeEach(async (to, _, next) => {
  const authStore = useAuthStore();
  await authStore.loadUser();
  const routeName = to.name?.toString() || "";

  // TODO: Show alerts when redirecting.
  if (!authStore.isEnabled && signinRoutes.includes(routeName)) {
    next({ name: "/" });
  } else if (!authStore.isUserValid && !signinRoutes.includes(routeName)) {
    next({ name: "/user/signin" });
  } else if (
    protectedRoutes[routeName] !== undefined &&
    !authStore.checkAttributes(protectedRoutes[routeName])
  ) {
    next({ name: "/" });
  } else {
    next();
  }
});

export default router;
