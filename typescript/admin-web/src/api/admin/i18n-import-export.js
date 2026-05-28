import request from '@/utils/request'

function buildFormData(file, packID) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('pack_id', String(packID))
  return formData
}

export function importCSV(file, packID) {
  return request({
    url: '/api/admin/v1/i18n/import/csv?pack_id=' + packID,
    method: 'post',
    data: buildFormData(file, packID),
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 30000
  })
}

export function exportCSV(packID) {
  return request({
    url: '/api/admin/v1/i18n/export/csv?pack_id=' + packID,
    method: 'get',
    responseType: 'blob'
  })
}
