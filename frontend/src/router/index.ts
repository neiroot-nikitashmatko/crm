import { createRouter, createWebHistory } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import { useAuth } from '@/composables/useAuth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true, title: 'Авторизация' },
    },
    {
      path: '/',
      component: AppLayout,
      redirect: '/leads',
      children: [
        {
          path: 'leads',
          name: 'leads',
          component: () => import('@/views/LeadsView.vue'),
          meta: { title: 'Лиды' },
        },
        {
          path: 'deals',
          name: 'deals',
          component: () => import('@/views/DealsView.vue'),
          meta: { title: 'Сделки' },
        },
        {
          path: 'tasks',
          name: 'tasks',
          component: () => import('@/views/TasksView.vue'),
          meta: { title: 'Задачи' },
        },
        {
          path: 'products-catalog',
          name: 'products-catalog',
          component: () => import('@/views/ProductsCatalogView.vue'),
          meta: { title: 'Каталог товаров' },
        },
        {
          path: 'production-calendar',
          name: 'production-calendar',
          component: () => import('@/views/ProductionCalendarView.vue'),
          meta: { title: 'Календарь производства' },
        },
        {
          path: 'employees',
          component: () => import('@/layouts/EmployeesLayout.vue'),
          meta: { title: 'Сотрудники', requiresAdmin: true },
          redirect: { name: 'employees-list' },
          children: [
            {
              path: 'list',
              name: 'employees-list',
              component: () => import('@/views/employees/EmployeesListView.vue'),
              meta: { title: 'Список сотрудников', requiresAdmin: true },
            },
            {
              path: 'new',
              name: 'employees-new',
              component: () => import('@/views/employees/AddEmployeeView.vue'),
              meta: { title: 'Добавить сотрудника', requiresAdmin: true },
            },
          ],
        },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const { isAuthenticated, isAdmin } = useAuth()

  if (to.meta.public) {
    if (to.name === 'login' && isAuthenticated.value) {
      return { path: '/leads' }
    }

    return true
  }

  if (!isAuthenticated.value) {
    return {
      name: 'login',
      query: { redirect: to.fullPath },
    }
  }

  if (to.meta.requiresAdmin && !isAdmin.value) {
    return { path: '/leads' }
  }

  return true
})

export default router
