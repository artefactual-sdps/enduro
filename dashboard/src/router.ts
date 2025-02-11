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
  "/packages/": ["package:list"],
  "/packages/[id]/": ["package:read"],
  "/locations/": ["storage:location:list"],
  "/locations/[id]/": ["storage:location:read"],
  "/locations/[id]/packages": ["storage:location:listPackages"],
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
