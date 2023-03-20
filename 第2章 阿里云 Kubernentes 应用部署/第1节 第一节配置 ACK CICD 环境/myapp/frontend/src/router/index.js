import Vue from 'vue'
import VueRouter from 'vue-router'
import LoginForm from '../components/LoginForm.vue'
import GamePage from '../components/GamePage.vue'
import auth from '../services/auth'

Vue.use(VueRouter)

const routes = [
  { path: '/', component: LoginForm },
  { path: '/game', component: GamePage, meta: { requiresAuth: true } }
]

const router = new VueRouter({
  routes
})

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !auth.isAuthenticated()) {
    next('/')
  } else {
    next()
  }
})

export default router
