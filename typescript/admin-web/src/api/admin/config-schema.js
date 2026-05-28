import request from '@/utils/request'

export function listSchema(query) {
  return request({
    url: '/api/admin/v1/config/schema',
    method: 'get',
    params: query
  })
}

export function createSchema(data) {
  return request({
    url: '/api/admin/v1/config/schema',
    method: 'post',
    data
  })
}

export function updateSchema(data) {
  return request({
    url: '/api/admin/v1/config/schema',
    method: 'put',
    data
  })
}

export function deleteSchema(id) {
  return request({
    url: '/api/admin/v1/config/schema',
    method: 'delete',
    params: { id }
  })
}
