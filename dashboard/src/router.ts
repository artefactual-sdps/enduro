import { useLayoutStore } from "./stores/layout";
import { createRouter, createWebHistory } from "vue-router";
import routes from "~pages";

const router = createRouter({
  history: createWebHistory("/"),
  routes,
  strict: false,
});

router.beforeEach((to, _, next) => {
  const layoutStore = useLayoutStore();
  const publicRoutes: Array<string | undefined> = [
    "user-signin",
    "user-signin-callback",
  ];
  if (!layoutStore.isUserValid && !publicRoutes.includes(to.name?.toString()))
    next({ name: "user-signin" });
  else next();
});

export default router;
