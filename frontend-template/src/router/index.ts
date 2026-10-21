import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'root',
      // Keep the application mountable without adding a placeholder page.
      component: { render: () => null },
    },
  ],
})

export default router
