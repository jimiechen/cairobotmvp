import request from '@/utils/request'

export function publishPack(data) {
  return request({
    url: '/api/admin/v1/i18n/pack/publish',
    method: 'post',
    data
  })
}

export function rollbackPack(data) {
  return request({
    url: '/api/admin/v1/i18n/pack/rollback',
    method: 'post',
    data
  })
}
