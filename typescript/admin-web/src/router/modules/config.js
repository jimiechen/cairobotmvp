import Layout from '@/layout'

const configRouter = {
  path: '/config',
  component: Layout,
  redirect: '/config/schema',
  name: 'ConfigManage',
  meta: { title: '配置管理', icon: 'form' },
  children: [
    {
      path: 'schema',
      component: () => import('@/views/config/schema-list'),
      name: 'ConfigSchema',
      meta: { title: 'Schema 管理', icon: 'list', roles: ['admin'] }
    },
    {
      path: 'value',
      component: () => import('@/views/config/value-publish'),
      name: 'ConfigValue',
      meta: { title: '配置发布', icon: 'edit', roles: ['admin'] }
    }
  ]
}

export default configRouter
