import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue')
      },
      {
        path: 'configs',
        name: 'Configs',
        component: () => import('@/views/configs/ConfigList.vue')
      },
      {
        path: 'configs/:key',
        name: 'ConfigDetail',
        component: () => import('@/views/configs/ConfigDetail.vue')
      },
      {
        path: 'stats',
        name: 'Stats',
        component: () => import('@/views/stats/StatsOverview.vue')
      },
      {
        path: 'frequency',
        name: 'Frequency',
        component: () => import('@/views/frequency/FrequencyControl.vue')
      },
      {
        path: 'budget',
        name: 'Budget',
        component: () => import('@/views/budget/BudgetManagement.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router 