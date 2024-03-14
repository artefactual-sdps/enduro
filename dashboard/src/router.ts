import { useLayoutStore } from "./stores/layout";
import { createRouter, createWebHistory } from "vue-router/auto";

const router = createRouter({
  history: createWebHistory("/"),
  strict: false,
});

router.beforeEach((to, _, next) => {
  const layoutStore = useLayoutStore();
  const publicRoutes: Array<string> = ["/user/signin", "/user/signin-callback"];
  const routeName = to.name?.toString() || "";
  if (!layoutStore.isUserValid && !publicRoutes.includes(routeName)) {
    next({ name: "/user/signin" });
  } else {
    next();
  }
});

export default router;
