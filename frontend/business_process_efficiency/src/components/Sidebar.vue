<template>
  <aside class="sidebar">
    <div class="menu">

      <div class="logo">
        BPE System
      </div>

      <router-link to="/profile" class="menu-item">Профиль</router-link>
      <router-link to="/processes" class="menu-item">Реестр процессов</router-link>
      <router-link to="/dashboards" class="menu-item">Дашборды</router-link>
      <router-link to="/dictionaries" class="menu-item">Справочники</router-link>
      <router-link to="/employees" class="menu-item">Сотрудники</router-link>

    </div>

    <button class="logout" @click="logout">
      Выйти
    </button>
  </aside>
</template>

<script>
export default {
  name: "Sidebar",
  methods: {
    logout() {
      localStorage.removeItem("jwt")
      this.$router.push("/login")
    }
  }
}
</script>

<style scoped>

.sidebar {
  position: fixed;
  left: 0;
  top: 0;

  width: 240px;
  height: 100vh;

  background: #ffffff;
  border-right: 1px solid var(--color-muted-bg);

  display: flex;
  flex-direction: column;

  padding: 24px 18px;
  box-sizing: border-box;
}

.logo {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: 30px;
}

.menu {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.menu-item {
  text-decoration: none;
  color: #374151;
  font-size: 15px;
  padding: 10px 12px;
  border-radius: 8px;
  transition: all 0.2s;
}

.menu-item:hover {
  background: var(--color-soft-bg);
}

.router-link-active {
  background: #eef2ff;
  color: var(--color-primary);
  font-weight: 500;
}

.logout {
  margin-top: auto;  

  border: none;
  background: var(--color-soft-bg);

  padding: 10px;
  border-radius: 8px;

  font-size: 14px;
  cursor: pointer;
  transition: 0.2s;
}

.logout:hover {
  background: var(--color-muted-bg);
}

</style>
