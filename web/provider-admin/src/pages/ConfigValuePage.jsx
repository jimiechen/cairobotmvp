import { useState, useEffect, useCallback } from 'react';
import apiClient from '../components/ApiClient';

const ENV_OPTIONS = [
  { value: 'dev', label: '开发环境 (dev)' },
  { value: 'test', label: '测试环境 (test)' },
  { value: 'prod', label: '生产环境 (prod)' },
];

const DEFAULT_MODULES = ['ai', 'firmware', 'app', 'gateway'];

function ConfigValuePage() {
  const [env, setEnv] = useState('dev');
  const [moduleKey, setModuleKey] = useState('ai');
  const [configJson, setConfigJson] = useState('{\n  \n}');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  const fetchConfig = useCallback(async () => {
    if (!moduleKey) return;
    setLoading(true);
    setError('');
    setMessage('');
    try {
      const data = await apiClient.configValue.get(env, moduleKey);
      const jsonStr = typeof data === 'object' ? JSON.stringify(data, null, 2) : data || '{\n}';
      setConfigJson(jsonStr);
    } catch (e) {
      setError('加载配置失败: ' + e.message);
    } finally {
      setLoading(false);
    }
  }, [env, moduleKey]);

  useEffect(() => { fetchConfig(); }, [fetchConfig]);

  const handleSave = async () => {
    try {
      JSON.parse(configJson);
    } catch (e) {
      alert('JSON 格式错误: ' + e.message);
      return;
    }
    setSaving(true);
    setError('');
    try {
      await apiClient.configValue.save(env, { module_key: moduleKey, config_json: JSON.parse(configJson) });
      setMessage('保存成功');
      setTimeout(() => setMessage(''), 3000);
    } catch (e) {
      setError('保存失败: ' + e.message);
    } finally {
      setSaving(false);
    }
  };

  const handlePublish = async () => {
    if (!confirm(`确认发布配置到 ${env} 环境？`)) return;
    setPublishing(true);
    setError('');
    try {
      await apiClient.configValue.publish(env, moduleKey);
      setMessage('发布成功');
      setTimeout(() => setMessage(''), 3000);
    } catch (e) {
      setError('发布失败: ' + e.message);
    } finally {
      setPublishing(false);
    }
  };

  const toolbarStyle = { display: 'flex', gap: '12px', marginBottom: '16px', alignItems: 'center', flexWrap: 'wrap' };
  const selectStyle = { padding: '6px 10px', borderRadius: '4px', border: '1px solid #ccc', fontSize: '14px' };
  const btnPrimary = { padding: '6px 18px', backgroundColor: '#e94560', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '14px' };
  const btnWarn = { padding: '6px 18px', backgroundColor: '#f0a030', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '14px' };

  return (
    <div>
      <h2 style={{ margin: '0 0 16px' }}>配置值管理</h2>
      <div style={toolbarStyle}>
        <label>环境:
          <select value={env} onChange={(e) => setEnv(e.target.value)} style={selectStyle}>
            {ENV_OPTIONS.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
        </label>
        <label>模块:
          <select value={moduleKey} onChange={(e) => setModuleKey(e.target.value)} style={selectStyle}>
            {DEFAULT_MODULES.map((m) => <option key={m} value={m}>{m}</option>)}
          </select>
        </label>
        <button onClick={handleSave} disabled={saving} style={btnPrimary}>{saving ? '保存中...' : '保存'}</button>
        <button onClick={handlePublish} disabled={publishing} style={btnWarn}>{publishing ? '发布中...' : '发布'}</button>
      </div>

      {error && <div style={{ color: '#d00', marginBottom: '12px', padding: '8px', backgroundColor: '#fff0f0', borderRadius: '4px' }}>{error}</div>}
      {message && <div style={{ color: '#0a0', marginBottom: '12px', padding: '8px', backgroundColor: '#f0fff0', borderRadius: '4px' }}>{message}</div>}
      {loading && <div style={{ color: '#666' }}>加载中...</div>}

      <div style={{ backgroundColor: '#fff', borderRadius: '8px', padding: '16px', border: '1px solid #ddd' }}>
        <label style={{ display: 'block', marginBottom: '8px', fontWeight: 500, fontSize: '14px' }}>
          配置 JSON ({env}/{moduleKey})
        </label>
        <textarea
          value={configJson}
          onChange={(e) => setConfigJson(e.target.value)}
          style={{
            width: '100%', minHeight: '400px', fontFamily: '"SF Mono", "Menlo", "Consolas", monospace',
            fontSize: '13px', lineHeight: '1.5', padding: '12px',
            border: '1px solid #ccc', borderRadius: '4px', boxSizing: 'border-box',
            resize: 'vertical', backgroundColor: '#fafafa',
          }}
          spellCheck="false"
        />
        <div style={{ marginTop: '8px', fontSize: '12px', color: '#888' }}>
          提示：请输入合法的 JSON 格式，保存前会进行格式校验
        </div>
      </div>
    </div>
  );
}

export default ConfigValuePage;
