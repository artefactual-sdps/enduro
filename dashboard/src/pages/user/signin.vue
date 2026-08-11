<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref } from "vue";

import IconDocumentation from "~icons/clarity/book-line";
import IconLogin from "~icons/clarity/login-line";
import IconShield from "~icons/clarity/shield-check-line";

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
      <div class="signin-card w-100">
        <div class="signin-brand d-flex align-items-center gap-3 mb-5">
          <img src="/logo.png" alt="" width="52" height="52" />
          <span class="fs-4 fw-semibold text-primary">Enduro</span>
        </div>

        <p class="signin-eyebrow mb-2">Secure workspace</p>
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
          <IconShield aria-hidden="true" class="flex-shrink-0 fs-5" />
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
            class="signin-footer-link"
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
            class="signin-footer-link"
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
  --signin-ink: #291426;
  --signin-muted: #6f6270;
  --signin-purple: #5e2750;
  --signin-purple-dark: #301128;
  --signin-purple-light: #8b4779;
  --signin-orange: #f36f2a;
  --signin-focus: #b44712;
  --signin-surface: #fbf9fb;

  display: grid;
  min-height: 100vh;
  min-height: 100svh;
  overflow: hidden;
  color: var(--signin-ink);
  background:
    radial-gradient(circle at 100% 0%, rgb(94 39 80 / 9%), transparent 38%),
    var(--signin-surface);
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
  color: #fff;
  isolation: isolate;
  background:
    radial-gradient(
      circle at var(--signin-gradient-x) var(--signin-gradient-y),
      rgb(243 111 42 / 48%),
      rgb(243 111 42 / 10%) 29%,
      transparent 54%
    ),
    radial-gradient(
      circle at var(--signin-gradient-secondary-x)
        var(--signin-gradient-secondary-y),
      rgb(167 74 132 / 38%),
      transparent 46%
    ),
    linear-gradient(
      var(--signin-gradient-angle),
      var(--signin-purple-light),
      transparent 45%
    ),
    linear-gradient(
      155deg,
      var(--signin-purple) 0%,
      var(--signin-purple-dark) 78%
    );
  animation: signin-gradient-rotation 5s ease-out both;

  &::before {
    position: absolute;
    z-index: 0;
    inset: -1rem;
    content: "";
    pointer-events: none;
    background-image:
      linear-gradient(rgb(255 255 255 / 7%) 1px, transparent 1px),
      linear-gradient(90deg, rgb(255 255 255 / 7%) 1px, transparent 1px);
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
      rgb(243 111 42 / 34%) 0%,
      rgb(179 78 115 / 20%) 33%,
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
  background: #fff;
  border-radius: 0.8rem;
  box-shadow: 0 0.75rem 2rem rgb(32 8 26 / 24%);

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
  color: rgb(255 255 255 / 76%);
  font-size: clamp(1rem, 1.4vw, 1.2rem);
}

.signin-panel {
  position: relative;
  min-width: 0;
}

.signin-card {
  max-width: 29rem;
  margin-block: auto;
}

.signin-brand img {
  border-radius: 0.85rem;
  box-shadow: 0 0.75rem 2rem rgb(94 39 80 / 16%);
}

.signin-eyebrow {
  color: var(--signin-purple);
  font-size: 0.78rem;
  font-weight: 750;
  letter-spacing: 0.11em;
  text-transform: uppercase;
}

.signin-title {
  color: var(--signin-ink);
  font-size: clamp(2.25rem, 6vw, 3.25rem);
  font-weight: 650;
  letter-spacing: -0.025em;
  line-height: 1.04;
}

.signin-description {
  max-width: 27rem;
  color: var(--signin-muted);
  font-size: 1.05rem;
  line-height: 1.65;
}

.signin-button {
  min-height: 3.5rem;
  padding: 0.9rem 1.25rem;
  font-size: 1rem;
  font-weight: 650;
  border-color: var(--signin-purple);
  background: var(--signin-purple);
  box-shadow: 0 0.9rem 2rem rgb(94 39 80 / 20%);
  transition:
    transform 150ms ease,
    box-shadow 150ms ease,
    background-color 150ms ease;

  &:hover:not(:disabled) {
    border-color: #4f1e43;
    background: #4f1e43;
    box-shadow: 0 1rem 2.4rem rgb(94 39 80 / 28%);
    transform: translateY(-1px);
  }

  &:focus-visible {
    outline: 3px solid var(--signin-focus);
    outline-offset: 3px;
    box-shadow: 0 0 0 0.25rem rgb(94 39 80 / 22%);
  }

  &:disabled {
    border-color: var(--signin-purple);
    background: var(--signin-purple);
  }
}

.signin-security-note {
  color: var(--signin-muted);
  font-size: 0.86rem;
  line-height: 1.5;

  svg {
    color: var(--signin-purple);
  }
}

.signin-footer {
  max-width: 29rem;
  border-top: 1px solid #eee8ec;
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
  color: #786d76;
  font-size: 0.8rem;
  font-weight: 500;
  text-decoration: none !important;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 2rem;
  transition:
    color 150ms ease,
    background-color 150ms ease,
    border-color 150ms ease;

  &:hover,
  &:focus-visible {
    color: var(--signin-muted);
    background: rgb(94 39 80 / 4%);
    border-color: rgb(94 39 80 / 9%);
  }

  &:focus-visible {
    outline: 3px solid var(--signin-focus);
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
  color: var(--signin-purple);
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
  color: #c7bec5;
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
    background: rgb(251 249 251 / 96%);
    box-shadow: -2rem 0 5rem rgb(34 12 29 / 18%);
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
