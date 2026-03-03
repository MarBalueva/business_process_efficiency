import { createRouter, createWebHistory } from 'vue-router'
import Login from '../components/Login.vue'
import Users from '../components/Users.vue'

const routes = [
  { path: '/login', name: 'Login', component: Login },
  { 
    path: '/users', 
    name: 'Users', 
    component: Users,
    meta: { requiresAuth: true }
  },
  { path: '/', redirect: '/login' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('jwt')
  if (to.meta.requiresAuth && !token) {
    next({ path: '/login' })
  } else if (to.path === '/login' && token) {
    next({ path: '/users' })
  } else {
    next()
  }
})

export default router