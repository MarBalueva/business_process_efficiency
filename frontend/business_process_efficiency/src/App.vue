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
body, html, #app {
  margin: 0;
  padding: 0;
  height: 100%;
  font-family: 'Inter', sans-serif;
  background-color: #f5f7fa;
}

.layout {
  display: flex;
  height: 100vh;
  font-family: 'Inter', sans-serif;
}

.content {
  flex: 1;
  padding: 40px;
  background-color: #f5f7fa;
  overflow-y: auto;
}

</style>