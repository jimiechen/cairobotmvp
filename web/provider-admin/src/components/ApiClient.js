const BASE_URL = '/api/v1';

async function request(url, options = {}) {
  const res = await fetch(`${BASE_URL}${url}`, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
    body: options.body ? JSON.stringify(options.body) : undefined,
  });
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(`请求失败 [${res.status}]: ${text || res.statusText}`);
  }
  const contentType = res.headers.get('content-type') || '';
  if (contentType.includes('json')) return res.json();
  return res.text();
}

export const apiClient = {
  get: (url) => request(url, { method: 'GET' }),
  post: (url, body) => request(url, { method: 'POST', body }),
  put: (url, body) => request(url, { method: 'PUT', body }),
  delete: (url) => request(url, { method: 'DELETE' }),

  configSchema: {
    list: () => apiClient.get('/config/schema'),
    create: (data) => apiClient.post('/config/schema', data),
    update: (id, data) => apiClient.put(`/config/schema/${id}`, data),
    delete: (id) => apiClient.delete(`/config/schema/${id}`),
  },

  configValue: {
    get: (env, moduleKey) => apiClient.get(`/config/value/${env}?module_key=${moduleKey}`),
    save: (env, data) => apiClient.put(`/config/value/${env}`, data),
    publish: (env, moduleKey) => apiClient.post('/config/value/publish', { env, module_key: moduleKey }),
  },

  i18n: {
    listStrings: (params) => {
      const qs = new URLSearchParams(params).toString();
      return apiClient.get(`/i18n/strings?${qs}`);
    },
    createString: (data) => apiClient.post('/i18n/strings', data),
    updateString: (id, data) => apiClient.put(`/i18n/strings/${id}`, data),
    deleteString: (id) => apiClient.delete(`/i18n/strings/${id}`),
    preview: (data) => apiClient.post('/i18n/preview', data),
    listPacks: () => apiClient.get('/i18n/packs'),
    publishPack: (langCode, data) => apiClient.post(`/i18n/packs/${langCode}/publish`, data),
  },
};

export default apiClient;
