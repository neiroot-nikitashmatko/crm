import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { migrateAuthStorage, setUnauthorizedHandler } from '@/api/session'
import { useAuth } from '@/composables/useAuth'
import './styles/global.css'

migrateAuthStorage()
useAuth().hydrateFromStorage()

const app = createApp(App)

setUnauthorizedHandler(() => {
  const { logout } = useAuth()
  logout()
  if (router.currentRoute.value.name !== 'login') {
    void router.push({
      name: 'login',
      query: { redirect: router.currentRoute.value.fullPath },
    })
  }
})

app.use(router)
app.mount('#app')
