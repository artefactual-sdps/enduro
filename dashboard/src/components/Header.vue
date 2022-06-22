<template>
  <nav class="navbar navbar-expand-lg bg-light">
    <div class="container-fluid px-3">
      <router-link class="navbar-brand" :to="{ name: 'index' }"
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
