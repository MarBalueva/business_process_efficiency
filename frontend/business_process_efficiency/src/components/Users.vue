<template>
  <div class="users-page">
    <h1>Пользователи</h1>
    <ul>
      <li v-for="user in users" :key="user.ID">{{ user.Login }}</li>
    </ul>
  </div>
</template>

<script>
import api from '../api/axios'

export default {
  name: 'Users',
  data() {
    return {
      users: []
    }
  },
  async created() {
    try {
      const response = await api.get('/users')
      this.users = response.data
    } catch (err) {
      console.error(err)
      this.$router.push('/login')
    }
  }
}
</script>

<style scoped>
.users-page {
  padding: 2rem;
  font-family: 'Inter', sans-serif;
}
</style>