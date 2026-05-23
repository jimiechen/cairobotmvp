import { useState, useEffect, useCallback } from 'react';
import apiClient from '../components/ApiClient';

const LANG_LABELS = { 'zh-CN': '简体中文', en: 'English', ja: '日本語', ko: '한국어' };

function I18nPackPage() {
  const [packs, setPacks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const [publishTarget, setPublishTarget] = useState(null);
  const [publishForm, setPublishForm] = useState({ version: '', description: '' });
  const [publishing, setPublishing] = useState(false);

  const fetchPacks = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await apiClient.i18n.listPacks();
      setPacks(Array.isArray(data) ? data : data?.data || []);
    } catch (e) {
      setError(e.message || '加载失败');
      setPacks([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchPacks(); }, [fetchPacks]);

  const openPublish = (pack) => {
    setPublishTarget(pack);
    setPublishForm({ version: (pack.version ? String(Number(pack.version) + 1) : '1'), description: '' });
  };

  const handlePublish = async () => {
    if (!publishForm.version) { alert('请输入版本号'); return; }
    setPublishing(true);
    try {
      await apiClient.i18n.publishPack(publishTarget.lang_code, publishForm);
      setMessage(`${LANG_LABELS[publishTarget.lang_code] || publishTarget.lang_code} 发布成功`);
      setTimeout(() => setMessage(''), 3000);
      setPublishTarget(null);
      fetchPacks();
    } catch (e) {
      alert('发布失败: ' + e.message);
    } finally {
      setPublishing(false);
    }
  };

  const gridStyle = { display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: '16px' };
  const cardStyle = { backgroundColor: '#fff', borderRadius: '8px', padding: '20px', border: '1px solid #e0e0e0', boxShadow: '0 1px 3px rgba(0,0,0,0.08)' };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
        <h2 style={{ margin: 0 }}>语言包发布管理</h2>
        <button onClick={fetchPacks} style={{ padding: '6px 12px', border: '1px solid #ccc', borderRadius: '4px', cursor: 'pointer', background: '#fff', fontSize: '13px' }}>刷新</button>
      </div>

      {error && <div style={{ color: '#d00', marginBottom: '12px', padding: '8px', backgroundColor: '#fff0f0', borderRadius: '4px' }}>{error}</div>}
      {message && <div style={{ color: '#0a0', marginBottom: '12px', padding: '8px', backgroundColor: '#f0fff0', borderRadius: '4px' }}>{message}</div>}
      {loading && <div>加载中...</div>}

      {!loading && packs.length === 0 && (
        <div style={{ textAlign: 'center', padding: '40px', color: '#888', backgroundColor: '#fff', borderRadius: '8px', border: '1px solid #e0e0e0' }}>
          暂无语言包数据，请确认后端服务已启动
        </div>
      )}

      {!loading && packs.length > 0 && (
        <div style={gridStyle}>
          {packs.map((pack) => (
            <div key={pack.lang_code} style={cardStyle}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
                <div>
                  <h3 style={{ margin: '0 0 4px', fontSize: '18px' }}>
                    {LANG_LABELS[pack.lang_code] || pack.lang_code}
                    <span style={{ marginLeft: '8px', fontSize: '12px', color: '#888', fontWeight: 400 }}>({pack.lang_code})</span>
                  </h3>
                  <div style={{ fontSize: '13px', color: '#666' }}>版本 v{pack.version ?? '-'}</div>
                </div>
                <span style={{
                  display: 'inline-block', padding: '3px 10px', borderRadius: '12px', fontSize: '12px', fontWeight: 500,
                  backgroundColor: pack.is_published ? '#d1fae5' : '#fef3c7',
                  color: pack.is_published ? '#065f46' : '#92400e',
                }}>
                  {pack.is_published ? '已发布' : '未发布'}
                </span>
              </div>

              {pack.description && (
                <p style={{ margin: '0 0 12px', fontSize: '14px', color: '#444', lineHeight: '1.5' }}>{pack.description}</p>
              )}

              {pack.published_at && (
                <div style={{ fontSize: '12px', color: '#999', marginBottom: '12px' }}>
                  发布时间：{new Date(pack.published_at).toLocaleString('zh-CN')}
                </div>
              )}

              <button onClick={() => openPublish(pack)} style={{
                width: '100%', padding: '8px', backgroundColor: '#e94560', color: '#fff',
                border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '14px',
              }}>
                发布新版本
              </button>
            </div>
          ))}
        </div>
      )}

      {publishTarget && (
        <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(0,0,0,0.45)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000 }}>
          <div style={{ backgroundColor: '#fff', borderRadius: '8px', padding: '24px', minWidth: '400px', maxWidth: '480px' }}>
            <h3 style={{ margin: '0 0 16px' }}>
              发布新版本 — {LANG_LABELS[publishTarget.lang_code] || publishTarget.lang_code}
            </h3>
            <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', fontWeight: 500 }}>版本号</label>
            <input
              value={publishForm.version}
              onChange={(e) => setPublishForm({ ...publishForm, version: e.target.value })}
              placeholder="如: 2"
              style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc', boxSizing: 'border-box', marginBottom: '12px' }}
            />
            <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', fontWeight: 500 }}>版本描述</label>
            <textarea
              value={publishForm.description}
              onChange={(e) => setPublishForm({ ...publishForm, description: e.target.value })}
              placeholder="本次更新的说明..."
              style={{ width: '100%', height: '80px', padding: '8px', borderRadius: '4px', border: '1px solid #ccc', boxSizing: 'border-box', marginBottom: '16px', resize: 'vertical', fontFamily: 'inherit' }}
            />
            <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
              <button onClick={() => setPublishTarget(null)} style={{ padding: '8px 18px', border: '1px solid #ccc', borderRadius: '4px', cursor: 'pointer', background: '#fff' }}>取消</button>
              <button onClick={handlePublish} disabled={publishing} style={{ padding: '8px 18px', backgroundColor: '#e94560', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer' }}>
                {publishing ? '发布中...' : '确认发布'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default I18nPackPage;
