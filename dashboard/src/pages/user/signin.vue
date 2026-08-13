<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref } from "vue";

import IconDocumentation from "~icons/clarity/book-line";
import IconLogin from "~icons/clarity/login-line";
import IconShield from "~icons/clarity/shield-check-line";

import ThemeToggle from "@/components/ThemeToggle.vue";
import { useAuthStore } from "@/stores/auth";

const authStore = useAuthStore();
const signinButton = ref<HTMLButtonElement | null>(null);
const visualElement = ref<HTMLElement | null>(null);
const isRedirecting = ref(false);
const redirectError = ref(false);

const motionCurrent = { x: 0, y: 0 };
const motionTarget = { x: 0, y: 0 };
let motionFrame: number | null = null;
let motionPreference: MediaQueryList | null = null;

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
}

function setVisualMotion(element: HTMLElement, x: number, y: number) {
  element.style.setProperty("--signin-watermark-x", `${x * -24}px`);
  element.style.setProperty("--signin-watermark-y", `${y * -16}px`);
  element.style.setProperty(
    "--signin-watermark-rotate",
    `${-12 + x * -1.5}deg`,
  );
  element.style.setProperty("--signin-grid-x", `${x * 7}px`);
  element.style.setProperty("--signin-grid-y", `${y * 5}px`);
  element.style.setProperty("--signin-ambient-x", `${x * 30}px`);
  element.style.setProperty("--signin-ambient-y", `${y * 22}px`);
  element.style.setProperty("--signin-gradient-x", `${62 + x * 30}%`);
  element.style.setProperty("--signin-gradient-y", `${38 + y * 28}%`);
  element.style.setProperty("--signin-gradient-secondary-x", `${26 - x * 12}%`);
  element.style.setProperty("--signin-gradient-secondary-y", `${76 - y * 18}%`);
}

function animateVisualMotion() {
  if (!visualElement.value) {
    motionFrame = null;
    return;
  }

  const easing = 0.1;
  motionCurrent.x += (motionTarget.x - motionCurrent.x) * easing;
  motionCurrent.y += (motionTarget.y - motionCurrent.y) * easing;

  const remainingX = Math.abs(motionTarget.x - motionCurrent.x);
  const remainingY = Math.abs(motionTarget.y - motionCurrent.y);
  if (remainingX < 0.001 && remainingY < 0.001) {
    motionCurrent.x = motionTarget.x;
    motionCurrent.y = motionTarget.y;
    motionFrame = null;
  } else {
    motionFrame = window.requestAnimationFrame(animateVisualMotion);
  }

  setVisualMotion(visualElement.value, motionCurrent.x, motionCurrent.y);
}

function scheduleVisualMotion() {
  if (motionFrame === null) {
    motionFrame = window.requestAnimationFrame(animateVisualMotion);
  }
}

function handleDocumentPointerMove(event: PointerEvent) {
  if (event.pointerType !== "mouse") return;

  motionTarget.x = clamp((event.clientX / window.innerWidth - 0.5) * 2, -1, 1);
  motionTarget.y = clamp((event.clientY / window.innerHeight - 0.5) * 2, -1, 1);
  scheduleVisualMotion();
}

function resetVisualMotion() {
  motionTarget.x = 0;
  motionTarget.y = 0;
  scheduleVisualMotion();
}

function addVisualMotionListeners() {
  window.addEventListener("pointermove", handleDocumentPointerMove);
  window.addEventListener("blur", resetVisualMotion);
  document.addEventListener("mouseleave", resetVisualMotion);
}

function removeVisualMotionListeners() {
  window.removeEventListener("pointermove", handleDocumentPointerMove);
  window.removeEventListener("blur", resetVisualMotion);
  document.removeEventListener("mouseleave", resetVisualMotion);
}

function stopVisualMotion() {
  if (motionFrame !== null) {
    window.cancelAnimationFrame(motionFrame);
    motionFrame = null;
  }

  motionCurrent.x = 0;
  motionCurrent.y = 0;
  motionTarget.x = 0;
  motionTarget.y = 0;

  if (visualElement.value) {
    setVisualMotion(visualElement.value, 0, 0);
  }
}

function handleMotionPreferenceChange(event: MediaQueryListEvent) {
  if (event.matches) {
    removeVisualMotionListeners();
    stopVisualMotion();
  } else {
    addVisualMotionListeners();
  }
}

onMounted(() => {
  motionPreference = window.matchMedia("(prefers-reduced-motion: reduce)");
  motionPreference.addEventListener("change", handleMotionPreferenceChange);

  if (!motionPreference.matches) {
    addVisualMotionListeners();
  }
});

onUnmounted(() => {
  removeVisualMotionListeners();
  motionPreference?.removeEventListener("change", handleMotionPreferenceChange);

  if (motionFrame !== null) {
    window.cancelAnimationFrame(motionFrame);
  }
});

async function signin() {
  if (isRedirecting.value) return;

  redirectError.value = false;
  isRedirecting.value = true;

  try {
    await authStore.signinRedirect();
  } catch {
    isRedirecting.value = false;
    redirectError.value = true;
    await nextTick();
    signinButton.value?.focus();
  }
}
</script>

<template>
  <section class="signin-page flex-grow-1">
    <div
      ref="visualElement"
      class="signin-visual d-none d-lg-flex flex-column justify-content-between"
      aria-hidden="true"
    >
      <img class="signin-watermark" src="/logo.png" alt="" />

      <div class="signin-visual-brand d-flex align-items-center gap-3">
        <span class="signin-visual-logo">
          <img src="/logo.png" alt="" />
        </span>
        <span class="fs-4 fw-semibold">Enduro</span>
      </div>

      <div class="signin-visual-copy">
        <p class="signin-visual-title">
          Preservation workflows,<br />
          under control.
        </p>
        <p class="signin-visual-description mb-0">
          Reliably move digital content from ingest to storage.
        </p>
      </div>
    </div>

    <div class="signin-panel d-flex flex-column align-items-center px-4 py-5">
      <ThemeToggle compact class="signin-theme-toggle" />

      <div class="signin-card w-100">
        <div class="signin-brand d-flex align-items-center gap-3 mb-5">
          <img src="/logo.png" alt="" width="52" height="52" />
          <span class="fs-4 fw-semibold text-primary">Enduro</span>
        </div>

        <p class="signin-eyebrow text-primary mb-2">Secure workspace</p>
        <h1 id="signin-title" class="signin-title mb-3">Welcome to Enduro</h1>
        <p id="signin-description" class="signin-description mb-4">
          Sign in through your organization to manage digital preservation
          workflows.
        </p>

        <div v-if="redirectError" class="alert alert-danger mb-4" role="alert">
          We couldn't connect to the sign-in service. Please try again.
        </div>

        <button
          ref="signinButton"
          class="signin-button btn btn-primary btn-lg d-flex align-items-center justify-content-center gap-2 w-100"
          type="button"
          :disabled="isRedirecting"
          aria-describedby="signin-description signin-security-note"
          @click="signin"
        >
          <span
            v-if="isRedirecting"
            class="spinner-border spinner-border-sm"
            aria-hidden="true"
          ></span>
          <IconLogin v-else aria-hidden="true" class="fs-4" />
          {{
            isRedirecting
              ? "Redirecting to sign in…"
              : "Sign in with your organization"
          }}
        </button>

        <p
          v-if="isRedirecting"
          class="visually-hidden"
          role="status"
          aria-live="polite"
        >
          Redirecting to your identity provider.
        </p>

        <p
          id="signin-security-note"
          class="signin-security-note d-flex align-items-start gap-2 mb-0 mt-4"
        >
          <IconShield
            aria-hidden="true"
            class="flex-shrink-0 fs-5 text-primary"
          />
          <span>
            You’ll continue to your organization’s secure login page.
          </span>
        </p>
      </div>

      <footer class="signin-footer w-100 pt-4">
        <nav
          class="signin-footer-links"
          aria-label="Enduro and Artefactual resources"
        >
          <a
            class="signin-footer-link text-decoration-none"
            href="https://enduro.readthedocs.io/"
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Read the Enduro documentation (opens in a new tab)"
          >
            <span class="signin-footer-icon" aria-hidden="true">
              <IconDocumentation />
            </span>
            <span>Documentation</span>
          </a>
          <span class="signin-footer-separator" aria-hidden="true">·</span>
          <a
            class="signin-footer-link text-decoration-none"
            href="https://www.artefactual.com/"
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Visit the Artefactual website (opens in a new tab)"
          >
            <span class="signin-footer-icon" aria-hidden="true">
              <img src="/artefactual-mark.svg" alt="" width="18" height="18" />
            </span>
            <span>Artefactual</span>
          </a>
        </nav>
      </footer>
    </div>
  </section>
</template>

<style scoped lang="scss">
.signin-page {
  display: grid;
  min-height: 100vh;
  min-height: 100svh;
  overflow: hidden;
  color: var(--bs-emphasis-color);
  background:
    radial-gradient(
      circle at 100% 0%,
      rgba(var(--bs-primary-rgb), 0.09),
      transparent 38%
    ),
    var(--bs-body-bg);
}

.signin-visual {
  --signin-ambient-x: 0px;
  --signin-ambient-y: 0px;
  --signin-gradient-angle: 145deg;
  --signin-gradient-secondary-x: 26%;
  --signin-gradient-secondary-y: 76%;
  --signin-gradient-x: 62%;
  --signin-gradient-y: 38%;
  --signin-grid-x: 0px;
  --signin-grid-y: 0px;
  --signin-watermark-rotate: -12deg;
  --signin-watermark-x: 0px;
  --signin-watermark-y: 0px;

  position: relative;
  min-width: 0;
  padding: clamp(2rem, 4vw, 4.5rem);
  overflow: hidden;
  color: var(--bs-white);
  isolation: isolate;
  background:
    radial-gradient(
      circle at var(--signin-gradient-x) var(--signin-gradient-y),
      color-mix(in srgb, var(--bs-orange) 48%, transparent),
      color-mix(in srgb, var(--bs-orange) 10%, transparent) 29%,
      transparent 54%
    ),
    radial-gradient(
      circle at var(--signin-gradient-secondary-x)
        var(--signin-gradient-secondary-y),
      color-mix(
        in srgb,
        color-mix(in srgb, var(--bs-primary) 60%, var(--bs-pink)) 38%,
        transparent
      ),
      transparent 46%
    ),
    linear-gradient(
      var(--signin-gradient-angle),
      color-mix(in srgb, var(--bs-primary) 80%, var(--bs-white)),
      transparent 45%
    ),
    linear-gradient(
      155deg,
      var(--bs-primary) 0%,
      color-mix(in srgb, var(--bs-primary) 50%, var(--bs-black)) 78%
    );
  animation: signin-gradient-rotation 5s ease-out both;

  &::before {
    position: absolute;
    z-index: 0;
    inset: -1rem;
    content: "";
    pointer-events: none;
    background-image:
      linear-gradient(rgba(var(--bs-white-rgb), 0.07) 1px, transparent 1px),
      linear-gradient(
        90deg,
        rgba(var(--bs-white-rgb), 0.07) 1px,
        transparent 1px
      );
    background-size: 3rem 3rem;
    mask-image: linear-gradient(
      135deg,
      transparent 5%,
      #000 32%,
      #000 74%,
      transparent 100%
    );
    transform: translate3d(var(--signin-grid-x), var(--signin-grid-y), 0);
    will-change: transform;
  }

  &::after {
    position: absolute;
    z-index: 0;
    top: 12%;
    right: -15%;
    width: min(42rem, 72vw);
    aspect-ratio: 1;
    content: "";
    pointer-events: none;
    background: radial-gradient(
      circle,
      color-mix(in srgb, var(--bs-orange) 34%, transparent) 0%,
      color-mix(
          in srgb,
          color-mix(in srgb, var(--bs-primary) 40%, var(--bs-pink)) 20%,
          transparent
        )
        33%,
      transparent 70%
    );
    border-radius: 50%;
    filter: blur(10px);
    animation: signin-ambient-shift 5s ease-out both;
    will-change: transform, opacity;
  }
}

.signin-watermark {
  position: absolute;
  z-index: 1;
  top: 11%;
  left: -17%;
  width: min(48rem, 76vw);
  opacity: 0.075;
  pointer-events: none;
  filter: grayscale(1) brightness(3);
  transform: translate3d(
      var(--signin-watermark-x),
      var(--signin-watermark-y),
      0
    )
    rotate(var(--signin-watermark-rotate));
  will-change: transform;
}

.signin-visual-brand,
.signin-visual-copy {
  z-index: 2;
  position: relative;
}

.signin-visual-logo {
  display: inline-flex;
  width: 3rem;
  height: 3rem;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 0.8rem;
  box-shadow: 0 0.75rem 2rem
    color-mix(
      in srgb,
      color-mix(in srgb, var(--bs-primary) 50%, var(--bs-black)) 24%,
      transparent
    );

  img {
    width: 100%;
    height: 100%;
  }
}

.signin-visual-copy {
  max-width: 37rem;
}

.signin-visual-title {
  margin-bottom: 1rem;
  font-size: clamp(2.8rem, 5vw, 5rem);
  font-weight: 650;
  letter-spacing: -0.018em;
  line-height: 1.02;
}

.signin-visual-description {
  max-width: 32rem;
  color: rgba(var(--bs-white-rgb), 0.76);
  font-size: clamp(1rem, 1.4vw, 1.2rem);
}

.signin-panel {
  position: relative;
  min-width: 0;
}

.signin-theme-toggle {
  position: absolute;
  z-index: 3;
  top: 1rem;
  right: 1rem;
}

.signin-card {
  max-width: 29rem;
  margin-block: auto;
}

.signin-brand img {
  border-radius: var(--bs-border-radius-lg);
  box-shadow: 0 0.75rem 2rem rgba(var(--bs-primary-rgb), 0.16);
}

.signin-eyebrow {
  font-size: 0.78rem;
  font-weight: 750;
  letter-spacing: 0.11em;
  text-transform: uppercase;
}

.signin-title {
  color: var(--bs-emphasis-color);
  font-size: clamp(2.25rem, 6vw, 3.25rem);
  font-weight: 650;
  letter-spacing: -0.025em;
  line-height: 1.04;
}

.signin-description {
  max-width: 27rem;
  color: var(--bs-secondary-color);
  font-size: 1.05rem;
  line-height: 1.65;
}

.signin-button {
  min-height: 3.5rem;
  padding: 0.9rem 1.25rem;
  font-size: 1rem;
  font-weight: 650;
  box-shadow: 0 0.9rem 2rem rgba(var(--bs-primary-rgb), 0.2);
  transition:
    $btn-transition,
    transform 150ms ease;

  &:hover:not(:disabled) {
    box-shadow: 0 1rem 2.4rem rgba(var(--bs-primary-rgb), 0.28);
    transform: translateY(-1px);
  }
}

.signin-security-note {
  color: var(--bs-secondary-color);
  font-size: 0.86rem;
  line-height: 1.5;
}

.signin-footer {
  max-width: 29rem;
  border-top: var(--bs-border-width) solid var(--bs-border-color);
}

.signin-footer-links {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.55rem;
}

.signin-footer-link {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.55rem;
  color: var(--bs-secondary-color);
  font-size: 0.8rem;
  font-weight: 500;
  background: transparent;
  border: var(--bs-border-width) solid transparent;
  border-radius: var(--bs-border-radius-pill);
  transition:
    color 150ms ease,
    background-color 150ms ease,
    border-color 150ms ease;

  &:hover,
  &:focus-visible {
    color: var(--bs-secondary-color);
    background: var(--bs-tertiary-bg);
    border-color: var(--bs-border-color);
  }

  &:focus-visible {
    outline: var(--bs-focus-ring-width) solid var(--bs-focus-ring-color);
    outline-offset: 2px;
  }
}

.signin-footer-icon {
  display: inline-flex;
  width: 1.15rem;
  height: 1.15rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  opacity: 0.72;
  transition: opacity 150ms ease;

  svg {
    width: 0.95rem;
    height: 0.95rem;
  }

  img {
    filter: grayscale(1);
    transition: filter 150ms ease;
  }
}

.signin-footer-link:hover .signin-footer-icon,
.signin-footer-link:focus-visible .signin-footer-icon {
  opacity: 1;
}

.signin-footer-link:hover .signin-footer-icon img,
.signin-footer-link:focus-visible .signin-footer-icon img {
  filter: none;
}

.signin-footer-separator {
  color: var(--bs-border-color);
  font-size: 0.9rem;
}

@property --signin-gradient-angle {
  syntax: "<angle>";
  inherits: false;
  initial-value: 145deg;
}

@keyframes signin-gradient-rotation {
  0% {
    --signin-gradient-angle: 132deg;
  }

  100% {
    --signin-gradient-angle: 178deg;
  }
}

@keyframes signin-ambient-shift {
  0% {
    opacity: 0.72;
    transform: translate3d(
        calc(var(--signin-ambient-x) - 1.5rem),
        calc(var(--signin-ambient-y) - 1rem),
        0
      )
      scale(0.9);
  }

  100% {
    opacity: 1;
    transform: translate3d(
        calc(var(--signin-ambient-x) + 2.5rem),
        calc(var(--signin-ambient-y) + 1.75rem),
        0
      )
      scale(1.1);
  }
}

@include media-breakpoint-up(lg) {
  .signin-page {
    grid-template-columns: minmax(0, 1.15fr) minmax(28rem, 0.85fr);
  }

  .signin-panel {
    background: rgba(var(--bs-body-bg-rgb), 0.96);
    box-shadow: -2rem 0 5rem rgba(var(--bs-primary-rgb), 0.18);
  }
}

@include media-breakpoint-down(sm) {
  .signin-panel {
    padding-inline: 1.25rem !important;
  }

  .signin-footer-links {
    gap: 0.35rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .signin-visual {
    animation: none;
  }

  .signin-visual::before {
    transform: none;
  }

  .signin-visual::after {
    animation: none;
    transform: none;
  }

  .signin-watermark {
    transform: rotate(-12deg);
  }

  .signin-button {
    transition: none;

    &:hover:not(:disabled) {
      transform: none;
    }
  }

  .spinner-border {
    animation: none;
  }

  .signin-footer-link,
  .signin-footer-icon,
  .signin-footer-icon img {
    transition: none;
  }
}

@media (forced-colors: active) {
  .signin-button {
    border: 1px solid ButtonText;
    box-shadow: none;
  }
}
</style>
