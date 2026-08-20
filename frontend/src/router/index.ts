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
          meta: { title: 'Лиды', sectionName: 'leads' },
        },
        {
          path: 'chats',
          name: 'chats',
          component: () => import('@/views/ChatsView.vue'),
          meta: { title: 'Чаты', sectionName: 'chats' },
        },
        {
          path: 'deals',
          name: 'deals',
          component: () => import('@/views/DealsView.vue'),
          meta: { title: 'Сделки', sectionName: 'deals' },
        },
        {
          path: 'tasks',
          name: 'tasks',
          component: () => import('@/views/TasksView.vue'),
          meta: { title: 'Задачи', sectionName: 'tasks' },
        },
        {
          path: 'products-catalog',
          redirect: '/trade/products',
        },
        {
          path: 'production-calendar',
          name: 'production-calendar',
          component: () => import('@/views/ProductionCalendarView.vue'),
          meta: { title: 'Календарь производства', sectionName: 'production-calendar' },
        },
        {
          path: 'trade',
          component: () => import('@/layouts/TradeLayout.vue'),
          meta: { title: 'Торговля', sectionName: 'trade' },
          redirect: { name: 'trade-outgoing-invoices' },
          children: [
            {
              path: 'outgoing-invoices',
              name: 'trade-outgoing-invoices',
              component: () => import('@/views/trade/OutgoingInvoicesView.vue'),
              meta: { title: 'Расходные накладные', sectionName: 'trade' },
            },
            {
              path: 'incoming-invoices',
              name: 'trade-incoming-invoices',
              component: () => import('@/views/trade/IncomingInvoicesView.vue'),
              meta: { title: 'Приходные накладные', sectionName: 'trade' },
            },
            {
              path: 'stock-balances',
              name: 'trade-stock-balances',
              component: () => import('@/views/trade/StockBalancesView.vue'),
              meta: { title: 'Остатки на складе', sectionName: 'trade' },
            },
            {
              path: 'suppliers',
              name: 'trade-suppliers',
              component: () => import('@/views/trade/SuppliersView.vue'),
              meta: { title: 'Поставщики', sectionName: 'trade' },
            },
            {
              path: 'products',
              name: 'products-catalog',
              component: () => import('@/views/ProductsCatalogView.vue'),
              meta: { title: 'Товары и услуги', sectionName: 'trade' },
            },
          ],
        },
        {
          path: 'salary',
          component: () => import('@/layouts/SalaryLayout.vue'),
          meta: { title: 'Зарплата', sectionName: 'salary' },
          redirect: { name: 'salary-add-deal' },
          children: [
            {
              path: 'add-deal',
              name: 'salary-add-deal',
              component: () => import('@/views/salary/SalaryAddDealView.vue'),
              meta: { title: 'Добавить сделку', sectionName: 'salary' },
            },
            {
              path: 'report',
              name: 'salary-report',
              component: () => import('@/views/salary/SalaryReportView.vue'),
              meta: { title: 'Отчёт', sectionName: 'salary' },
            },
          ],
        },
        {
          path: 'payment-calendar',
          name: 'payment-calendar',
          component: () => import('@/views/PaymentCalendarView.vue'),
          meta: { title: 'Календарь оплат', sectionName: 'payment-calendar' },
        },
        {
          path: 'analytics',
          component: () => import('@/layouts/AnalyticsLayout.vue'),
          meta: { title: 'Аналитика', requiresAdmin: true },
          redirect: { name: 'analytics' },
          children: [
            {
              path: '',
              name: 'analytics',
              component: () => import('@/views/analytics/AnalyticsView.vue'),
              meta: { title: 'Аналитика', requiresAdmin: true },
            },
          ],
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
  const { isAuthenticated, isAdmin, canAccessSection, getDefaultRouteName, hydrateFromStorage } =
    useAuth()

  hydrateFromStorage()

  if (to.meta.public) {
    if (to.name === 'login' && isAuthenticated.value) {
      return { name: getDefaultRouteName() }
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
    return { name: getDefaultRouteName() }
  }

  const sectionName = typeof to.meta.sectionName === 'string' ? to.meta.sectionName : ''
  if (sectionName && !canAccessSection(sectionName)) {
    return { name: getDefaultRouteName() }
  }

  return true
})

export default router
