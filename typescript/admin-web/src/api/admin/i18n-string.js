import request from '@/utils/request'

export function listStrings(packID) {
  return request({
    url: '/api/admin/v1/i18n/string',
    method: 'get',
    params: { pack_id: packID }
  })
}

export function createString(data) {
  return request({
    url: '/api/admin/v1/i18n/string',
    method: 'post',
    data
  })
}

export function updateString(data) {
  return request({
    url: '/api/admin/v1/i18n/string',
    method: 'put',
    data
  })
}

export function deleteString(id) {
  return request({
    url: '/api/admin/v1/i18n/string',
    method: 'delete',
    params: { id }
  })
}
