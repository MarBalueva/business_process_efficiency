<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script>
export default {
  name: "App"
};
</script>

<style>
body, html, #app {
  margin: 0;
  padding: 0;
  height: 100%;
  font-family: 'Inter', sans-serif;
  background-color: #f5f7fa;
}

#app {
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>