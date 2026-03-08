import { createRouter, createWebHistory } from 'vue-router'
import Login from '../components/Login.vue'
import Employees from '../components/Employees.vue'
import EmployeeCard from '../components/EmployeeCard.vue'
import DictionariesPage from '../components/DictionariesPage.vue'
import ProfilePage from '../components/ProfilePage.vue'
import ProcessRegistry from '../components/ProcessRegistry.vue'
import ProcessCard from '../components/ProcessCard.vue'

const routes = [
  { path: '/login', name: 'Login', component: Login },
  { path: '/employees', name: 'Employees', component: Employees, meta: { requiresAuth: true } },
  { path: '/employees/:id', name: 'EmployeeCard', component: EmployeeCard, meta: { requiresAuth: true } },
  { path: '/dictionaries', name: 'DictionariesPage', component: DictionariesPage, meta: { requiresAuth: true } },
  { path: '/profile', name: 'ProfilePage', component: ProfilePage, meta: { requiresAuth: true } },
  { path: '/processes', name: 'ProcessRegistry', component: ProcessRegistry, meta: { requiresAuth: true } },
  { path: '/processes/:id', name: 'ProcessCard', component: ProcessCard, meta: { requiresAuth: true } },
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
    next({ path: '/profile' })
  } else {
    next()
  }
})

export default router