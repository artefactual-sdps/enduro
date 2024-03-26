import { useLayoutStore } from "./stores/layout";
import { createRouter, createWebHistory } from "vue-router/auto";

const router = createRouter({
  history: createWebHistory("/"),
  strict: false,
});

const publicRoutes: Array<string> = ["/user/signin", "/user/signin-callback"];

// Send unauthenticated users to the sign-in page.
router.beforeEach(async (to, _, next) => {
  const layoutStore = useLayoutStore();
  await layoutStore.loadUser();
  const routeName = to.name?.toString() || "";
  if (!layoutStore.isUserValid && !publicRoutes.includes(routeName)) {
    next({ name: "/user/signin" });
  } else {
    next();
  }
});

export default router;
