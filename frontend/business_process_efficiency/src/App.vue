<template>
  <div class="layout">

    <Sidebar v-if="$route.path !== '/login'" />

    <main class="content">
      <router-view />
    </main>

  </div>
</template>

<script>
import Sidebar from "./components/Sidebar.vue"

export default {
  components: {
    Sidebar
  }
}
</script>

<style>
.layout {
  display: flex;
  height: 100vh;
}

.content {
  flex: 1;
  padding: 40px;
  background-color: var(--color-bg);
  overflow-y: auto;
}

</style>
