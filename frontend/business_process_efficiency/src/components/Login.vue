<template>
  <div class="login-container">
    <div class="login-card">
      <h1 class="title">Вход</h1>
      <form @submit.prevent="submitForm">
        <div class="form-group">
          <label for="login">Логин</label>
          <input type="text" id="login" v-model="login" placeholder="Введите логин" />
        </div>

        <div class="form-group">
          <label for="password">Пароль</label>
          <input type="password" id="password" v-model="password" placeholder="Введите пароль" />
        </div>

        <div v-if="error" class="error-message">{{ error }}</div>

        <button type="submit">Войти</button>
      </form>
    </div>
  </div>
</template>

<script>
import api from '../api/axios'

export default {
  name: "Login",
  data() {
    return {
      login: "",
      password: "",
      error: ""
    };
  },
  methods: {
    async submitForm() {
      this.error = ""

      if (!this.login.trim()) {
        this.error = "Введите логин"
        return
      }
      if (!this.password.trim()) {
        this.error = "Введите пароль"
        return
      }

      try {
        const response = await api.post('/login', {
          login: this.login,
          password: this.password
        })
        const { token } = response.data

        localStorage.setItem('jwt', token)

        this.$router.push('/profile')
      } catch (err) {
        console.error(err)
        if (err.response && err.response.data && err.response.data.error) {
          this.error = err.response.data.error
        } else {
          this.error = "Произошла ошибка при входе"
        }
      }
    }
  }
};
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: #f5f7fa;
}

.login-card {
  background: #ffffff;
  padding: 2rem 2.5rem;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.05);
  width: 360px;
  text-align: center;
}

.title {
  margin-bottom: 1.5rem;
  font-size: 1.8rem;
  color: #333;
}

.form-group {
  margin-bottom: 1rem;
  text-align: left;
}

label {
  display: block;
  margin-bottom: 0.3rem;
  font-size: 0.9rem;
  color: #555;
}


input[type="text"],
input[type="password"] {
  width: 100%;
  padding: 0.65rem 0.75rem; 
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 1rem;
  outline: none;
  box-sizing: border-box;
  transition: 0.2s;
}

input[type="text"]::placeholder,
input[type="password"]::placeholder {
  color: #a1a1aa;
  text-align: left;
}

input[type="text"]:focus,
input[type="password"]:focus {
  border-color: #a5b4fc;
  box-shadow: 0 0 0 2px rgba(165,180,252,0.2);
}

button {
  margin-top: 1.5rem;
  width: 100%;
  padding: 0.75rem;
  border: none;
  border-radius: 8px;
  background: #4f46e5;
  color: #fff;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: 0.2s;
}

button:hover {
  background: #4338ca;
}

.error-message {
  color: #dc2626;
  font-size: 0.9rem;
  margin-top: 0.5rem;
  text-align: center;
}
</style>