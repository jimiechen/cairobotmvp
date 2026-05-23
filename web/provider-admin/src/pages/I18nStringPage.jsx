import { useState, useEffect, useCallback } from 'react';
import apiClient from '../components/ApiClient';

const TEMPLATE_TYPES = [
  { value: 'plain', label: '纯文本', color: '#3b82f6' },
  { value: 'named', label: '命名参数', color: '#10b981' },
  { value: 'icu', label: 'ICU 格式', color: '#f59e0b' },
];
const GROUPS = ['common', 'app', 'error'];
const EMPTY_FORM = { string_key: '', string_value: '', template_type: 'plain', params: [], group_name: 'common' };
const EMPTY_PARAM = { name: '', type: 'string', required: false };

function I18nStringPage() {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [lang, setLang] = useState('zh-CN');
  const [groupFilter, setGroupFilter] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [editId, setEditId] = useState(null);
  const [form, setForm] = useState({ ...EMPTY_FORM });
  const [previewKey, setPreviewKey] = useState(null);
  const [previewParams, setPreviewParams] = useState({});
  const [previewResult, setPreviewResult] = useState('');

  const fetchList = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const params = { lang };
      if (groupFilter) params.group_name = groupFilter;
      const data = await apiClient.i18n.listStrings(params);
      setItems(Array.isArray(data) ? data : data?.data || []);
    } catch (e) {
      setError(e.message || '加载失败');
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, [lang, groupFilter]);

  useEffect(() => { fetchList(); }, [fetchList]);

  const openCreate = () => { setEditId(null); setForm({ ...EMPTY_FORM }); setShowModal(true); };
  const openEdit = (row) => { setEditId(row.id); setForm({ string_key: row.string_key, string_value: row.string_value, template_type: row.template_type || 'plain', params: row.params || [], group_name: row.group_name || 'common' }); setShowModal(true); };

  const handleSubmit = async () => {
    if (!form.string_key) return;
    try {
      if (editId) await apiClient.i18n.updateString(editId, form);
      else await apiClient.i18n.createString(form);
      setShowModal(false);
      fetchList();
    } catch (e) {
      alert('操作失败: ' + e.message);
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('确认删除？')) return;
    try { await apiClient.i18n.deleteString(id); fetchList(); } catch (e) { alert('删除失败: ' + e.message); }
  };

  const handlePreview = async (row) => {
    setPreviewKey(row);
    setPreviewParams({});
    setPreviewResult('');
    if (row.template_type === 'named' && row.params?.length > 0) return;
    try {
      const res = await apiClient.i18n.preview({ string_key: row.string_key, lang, params: {} });
      setPreviewResult(res?.result || JSON.stringify(res));
    } catch (e) { setPreviewResult('预览失败: ' + e.message); }
  };

  const doPreview = async () => {
    try {
      const res = await apiClient.i18n.preview({ string_key: previewKey.string_key, lang, params: previewParams });
      setPreviewResult(res?.result || JSON.stringify(res));
    } catch (e) { setPreviewResult('预览失败: ' + e.message); }
  };

  const filtered = groupFilter ? items.filter((i) => i.group_name === groupFilter) : items;
  const ttMap = Object.fromEntries(TEMPLATE_TYPES.map((t) => [t.value, t]));
  const tableStyle = { width: '100%', borderCollapse: 'collapse', backgroundColor: '#fff', fontSize: '13px' };
  const thStyle = { padding: '10px 12px', textAlign: 'left', borderBottom: '2px solid #ddd', backgroundColor: '#fafafa', fontWeight: 600 };
  const tdStyle = { padding: '8px 12px', borderBottom: '1px solid #eee' };

  return (
    <div>
      <h2 style={{ margin: '0 0 16px' }}>多语言 Key 管理</h2>
      <div style={{ display: 'flex', gap: '12px', marginBottom: '16px', alignItems: 'center', flexWrap: 'wrap' }}>
        <select value={lang} onChange={(e) => setLang(e.target.value)} style={selStyle}>
          <option value="zh-CN">简体中文</option><option value="en">English</option>
        </select>
        <select value={groupFilter} onChange={(e) => setGroupFilter(e.target.value)} style={selStyle}>
          <option value="">全部分组</option>{GROUPS.map((g) => <option key={g} value={g}>{g}</option>)}
        </select>
        <button onClick={openCreate} style={btnPrimary}>新增</button>
        <button onClick={fetchList} style={btnDefault}>刷新</button>
      </div>

      {error && <div style={{ color: '#d00', marginBottom: '12px' }}>{error}</div>}
      {loading && <div>加载中...</div>}
      {!loading && (
        <table style={tableStyle}>
          <thead><tr><th style={thStyle}>Key</th><th style={thStyle}>值（截断）</th><th style={thStyle}>模板类型</th><th style={thStyle}>分组</th><th style={thStyle}>版本</th><th style={thStyle}>操作</th></tr></thead>
          <tbody>
            {filtered.length === 0 && <tr><td colSpan="6" style={tdStyle}>暂无数据</td></tr>}
            {filtered.map((row) => (
              <tr key={row.id}>
                <td style={tdStyle}><code>{row.string_key}</code></td>
                <td style={tdStyle}>{(row.string_value || '').substring(0, 50)}{(row.string_value || '').length > 50 ? '...' : ''}</td>
                <td style={tdStyle}>{(() => { const t = ttMap[row.template_type] || ttMap.plain; return <span style={{ display: 'inline-block', padding: '2px 8px', borderRadius: '4px', color: '#fff', fontSize: '12px', backgroundColor: t.color }}>{t.label}</span>; })()}</td>
                <td style={tdStyle}>{row.group_name}</td>
                <td style={tdStyle}>{row.version ?? '-'}</td>
                <td style={tdStyle}>
                  <button onClick={() => openEdit(row)} style={linkBtn}>编辑</button>
                  <button onClick={() => handlePreview(row)} style={{ ...linkBtn, marginLeft: '8px', color: '#7c3aed' }}>预览</button>
                  <button onClick={() => handleDelete(row.id)} style={{ ...linkBtn, marginLeft: '8px', color: '#d00' }}>删除</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {showModal && (<I18nModal title={editId ? '编辑多语言 Key' : '新增多语言 Key'} form={form} setForm={setForm} onClose={() => setShowModal(false)} onSubmit={handleSubmit} />)}
      {previewKey && (<PreviewModal item={previewKey} params={previewParams} setParams={setPreviewParams} result={previewResult} onPreview={doPreview} onClose={() => setPreviewKey(null)} />)}
    </div>
  );
}

const selStyle = { padding: '6px 10px', borderRadius: '4px', border: '1px solid #ccc' };
const btnPrimary = { padding: '6px 16px', backgroundColor: '#e94560', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer' };
const btnDefault = { padding: '6px 12px', border: '1px solid #ccc', borderRadius: '4px', cursor: 'pointer', background: '#fff' };
const linkBtn = { padding: '3px 8px', cursor: 'pointer', border: 'none', background: 'none', textDecoration: 'underline', color: '#3b82f6' };

function I18nModal({ title, form, setForm, onClose, onSubmit }) {
  const set = (k, v) => setForm({ ...form, [k]: v });
  const addParam = () => set('params', [...(form.params || []), { ...EMPTY_PARAM }]);
  const removeParam = (i) => set('params', form.params.filter((_, idx) => idx !== i));
  const updateParam = (i, k, v) => { const p = [...form.params]; p[i] = { ...p[i], [k]: v }; set('params', p); };

  return (
    <div style={overlayStyle}>
      <div style={modalBodyStyle}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
          <h3 style={{ margin: 0 }}>{title}</h3><button onClick={onClose} style={{ border: 'none', background: 'none', fontSize: '20px', cursor: 'pointer' }}>×</button>
        </div>
        <label style={lblStyle}>Key <input style={inpStyle} value={form.string_key} onChange={(e) => set('string_key', e.target.value)} /></label>
        <label style={lblStyle}>值 <textarea style={{ ...inpStyle, height: '60px' }} value={form.string_value} onChange={(e) => set('string_value', e.target.value)} /></label>
        <label style={lblStyle}>模板类型
          <select style={inpStyle} value={form.template_type} onChange={(e) => set('template_type', e.target.value)}>
            {TEMPLATE_TYPES.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
          </select>
        </label>
        {(form.template_type === 'named' || form.template_type === 'icu') && (
          <div style={{ marginBottom: '12px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
              <span style={{ fontWeight: 500, fontSize: '13px' }}>参数列表</span>
              <button onClick={addParam} style={{ ...btnDefault, padding: '2px 10px', fontSize: '12px' }}>+ 添加参数</button>
            </div>
            {(form.params || []).map((p, i) => (
              <div key={i} style={{ display: 'flex', gap: '8px', alignItems: 'center', marginBottom: '4px' }}>
                <input placeholder="名称" value={p.name} onChange={(e) => updateParam(i, 'name', e.target.value)} style={{ ...inpStyle, flex: 1 }} />
                <select value={p.type} onChange={(e) => updateParam(i, 'type', e.target.value)} style={{ ...inpStyle, width: '80px' }}>
                  <option value="string">string</option><option value="int">int</option><option value="float">float</option>
                </select>
                <label><input type="checkbox" checked={p.required} onChange={(e) => updateParam(i, 'required', e.target.checked)} /> 必填</label>
                <button onClick={() => removeParam(i)} style={{ ...linkBtn, color: '#d00' }}>移除</button>
              </div>
            ))}
          </div>
        )}
        <label style={lblStyle}>分组
          <select style={inpStyle} value={form.group_name} onChange={(e) => set('group_name', e.target.value)}>
            {GROUPS.map((g) => <option key={g} value={g}>{g}</option>)}
          </select>
        </label>
        <button onClick={onSubmit} style={btnPrimary}>提交</button>
      </div>
    </div>
  );
}

function PreviewModal({ item, params, setParams, result, onPreview, onClose }) {
  const paramInputs = (item.params || []).map((p) => (
    <div key={p.name} style={{ marginBottom: '8px' }}>
      <label style={{ fontSize: '13px', marginRight: '8px' }}>{p.name} ({p.type}){p.required ? '*' : ''}:</label>
      <input value={params[p.name] || ''} onChange={(e) => setParams({ ...params, [p.name]: e.target.value })} style={inpStyle} />
    </div>
  ));

  return (
    <div style={overlayStyle}>
      <div style={{ ...modalBodyStyle, maxWidth: '500px' }}>
        <h3 style={{ margin: '0 0 12px' }}>模板预览 — {item.string_key}</h3>
        <div style={{ marginBottom: '12px', fontSize: '13px', color: '#666' }}>原始值: {item.string_value}</div>
        {paramInputs.length > 0 ? <div style={{ marginBottom: '12px' }}>{paramInputs}<button onClick={onPreview} style={btnPrimary}>执行预览</button></div> : null}
        {result && <div style={{ marginTop: '12px', padding: '12px', backgroundColor: '#f0fff0', borderRadius: '4px', whiteSpace: 'pre-wrap' }}>{result}</div>}
        <button onClick={onClose} style={{ ...btnDefault, marginTop: '12px' }}>关闭</button>
      </div>
    </div>
  );
}

const overlayStyle = { position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(0,0,0,0.45)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000 };
const modalBodyStyle = { backgroundColor: '#fff', borderRadius: '8px', padding: '24px', minWidth: '520px', maxWidth: '640px', maxHeight: '80vh', overflowY: 'auto' };
const lblStyle = { display: 'block', marginBottom: '4px', fontSize: '13px', fontWeight: 500 };
const inpStyle = { width: '100%', padding: '6px 8px', borderRadius: '4px', border: '1px solid #ccc', boxSizing: 'border-box', marginBottom: '12px' };

export default I18nStringPage;
