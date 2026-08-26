import { createRouter, createWebHashHistory } from 'vue-router'

/* Layout */
import Layout from '@/layout'

/**
 * Route metadata consumed by the sidebar, breadcrumb and page cache.
 *
 * hidden: true                   hide the route from the sidebar
 * alwaysShow: true               keep the root menu visible for a single child
 * redirect: noRedirect           render the breadcrumb item without navigation
 * name: 'router-name'            stable key used by <keep-alive>
 * meta : {
    roles: ['admin','editor']     roles allowed to access the route
    title: 'title'                sidebar and breadcrumb label
    icon: 'svg-name'              sidebar icon
    noCache: true                 disable page caching
    affix: true                   pin the route in tags-view
    breadcrumb: false             hide the route from breadcrumbs
    activeMenu: '/example/list'   sidebar path highlighted for this route
  }
 */

/**
 * Base routes available before the dynamic permission menu is loaded.
 */
export const constantRoutes = [
  {
    path: '/redirect',
    component: Layout,
    hidden: true,
    children: [
      {
        path: '/redirect/:path*',
        component: () => import('@/views/redirect/index')
      }
    ]
  },
  {
    path: '/login',
    component: () => import('@/views/login/index'),
    hidden: true
  },
  {
    path: '/auth-redirect',
    component: () => import('@/views/login/auth-redirect'),
    hidden: true
  },
  {
    path: '/404',
    component: () => import('@/views/error-page/404'),
    hidden: true
  },
  {
    path: '/401',
    component: () => import('@/views/error-page/401'),
    hidden: true
  },
  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        component: () => import('@/views/dashboard/index'),
        name: 'Dashboard',
        meta: { title: '首页', icon: 'dashboard', affix: true }
      }
    ]
  },
  {
    path: '/profile',
    component: Layout,
    redirect: '/profile/index',
    hidden: true,
    children: [
      {
        path: 'index',
        component: () => import('@/views/profile/index'),
        name: 'Profile',
        meta: { title: '个人中心', icon: 'user', noCache: true }
      }
    ]
  }
]

/**
 * asyncRoutes
 * the routes that need to be dynamically loaded based on user roles
 */
export const asyncRoutes = [

]

const router = createRouter({
  history: createWebHashHistory(), // 使用 hash 模式
  scrollBehavior: () => ({ top: 0 }),
  routes: constantRoutes
})

// Detail see: https://github.com/vuejs/vue-router/issues/1234#issuecomment-357941465
export function resetRouter() {
  const newRouter = createRouter({
    history: createWebHashHistory(),
    scrollBehavior: () => ({ top: 0 }),
    routes: constantRoutes
  })
  router.clearRoutes()
  newRouter.getRoutes().forEach(route => {
    router.addRoute(route)
  })
}

export default router
