import request from '@/utils/request'

export function publishValue(data) {
  return request({
    url: '/api/admin/v1/config/value/publish',
    method: 'post',
    data
  })
}

export function getValueVersions(query) {
  return request({
    url: '/api/admin/v1/config/value/versions',
    method: 'get',
    params: query
  })
}
