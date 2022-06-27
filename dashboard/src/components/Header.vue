<template>
  <nav class="navbar navbar-dark navbar-expand-lg">
    <div class="container-fluid">
      <router-link class="navbar-brand mb-0 h1" :to="{ name: 'index' }"
        >Enduro</router-link
      >

      <div class="collapse navbar-collapse" id="navbarSupportedContent">
        <ul class="navbar-nav me-auto">
          <li class="nav-item">
            <router-link
              class="nav-link"
              active-class="active"
              :to="{ name: 'packages' }"
              >Packages</router-link
            >
          </li>
        </ul>
      </div>
    </div>
  </nav>
</template>

<style lang="scss" scoped>
.navbar-brand {
  user-select: none;
}

.navbar {
  background-color: #5e2750;
}

@media (min-width: 576px) {
  .nav-link.active {
    color: white;
    background-color: #fff2;
  }
}

.nav-link.active {
  border-radius: var(--bs-border-radius-sm) !important;
}
</style>
