import Layout from '@/layout'

const i18nRouter = {
  path: '/i18n',
  component: Layout,
  redirect: '/i18n/string',
  name: 'I18nManage',
  meta: { title: '国际化管理', icon: 'language' },
  children: [
    {
      path: 'string',
      component: () => import('@/views/i18n/string-list'),
      name: 'I18nString',
      meta: { title: '字符串管理', icon: 'document', roles: ['admin'] }
    },
    {
      path: 'pack',
      component: () => import('@/views/i18n/pack-manage'),
      name: 'I18nPack',
      meta: { title: '语言包管理', icon: 's-grid', roles: ['admin'] }
    },
    {
      path: 'import-export',
      component: () => import('@/views/i18n/import-export'),
      name: 'I18nImportExport',
      meta: { title: 'CSV 导入导出', icon: 'download', roles: ['admin'] }
    }
  ]
}

export default i18nRouter
