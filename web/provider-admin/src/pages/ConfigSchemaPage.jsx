import { useState, useEffect, useCallback } from 'react';
import apiClient from '../components/ApiClient';

const FIELD_TYPES = ['string', 'int', 'float', 'bool', 'json', 'list'];
const EMPTY_FORM = { module_key: '', field_key: '', field_type: 'string', default_value: '', validator: '', is_required: false, is_secret: false, description: '', client_scope: '*', sort_order: 0 };

function ConfigSchemaPage() {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [moduleFilter, setModuleFilter] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({ ...EMPTY_FORM });
  const [editId, setEditId] = useState(null);
  const [editForm, setEditForm] = useState({});

  const fetchList = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await apiClient.configSchema.list();
      setItems(Array.isArray(data) ? data : data?.data || []);
    } catch (e) {
      setError(e.message || '加载失败');
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchList(); }, [fetchList]);

  const handleCreate = async () => {
    if (!form.module_key || !form.field_key) return;
    try {
      await apiClient.configSchema.create(form);
      setShowCreate(false);
      setForm({ ...EMPTY_FORM });
      fetchList();
    } catch (e) {
      alert('创建失败: ' + e.message);
    }
  };

  const handleUpdate = async (id) => {
    try {
      await apiClient.configSchema.update(id, editForm);
      setEditId(null);
      setEditForm({});
      fetchList();
    } catch (e) {
      alert('更新失败: ' + e.message);
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('确认删除该 Schema？')) return;
    try {
      await apiClient.configSchema.delete(id);
      fetchList();
    } catch (e) {
      alert('删除失败: ' + e.message);
    }
  };

  const filtered = moduleFilter ? items.filter((i) => i.module_key === moduleFilter) : items;
  const modules = [...new Set(items.map((i) => i.module_key))];

  const tableStyle = { width: '100%', borderCollapse: 'collapse', backgroundColor: '#fff', fontSize: '13px' };
  const thStyle = { padding: '10px 12px', textAlign: 'left', borderBottom: '2px solid #ddd', backgroundColor: '#fafafa', fontWeight: 600 };
  const tdStyle = { padding: '8px 12px', borderBottom: '1px solid #eee' };

  return (
    <div>
      <h2 style={{ margin: '0 0 16px' }}>配置 Schema 管理</h2>
      <div style={{ display: 'flex', gap: '12px', marginBottom: '16px', alignItems: 'center' }}>
        <select value={moduleFilter} onChange={(e) => setModuleFilter(e.target.value)} style={{ padding: '6px 10px', borderRadius: '4px', border: '1px solid #ccc' }}>
          <option value="">全部模块</option>
          {modules.map((m) => <option key={m} value={m}>{m}</option>)}
        </select>
        <button onClick={() => setShowCreate(true)} style={{ padding: '6px 16px', backgroundColor: '#e94560', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer' }}>新增</button>
        <button onClick={fetchList} style={{ padding: '6px 12px', border: '1px solid #ccc', borderRadius: '4px', cursor: 'pointer', background: '#fff' }}>刷新</button>
      </div>

      {error && <div style={{ color: '#d00', marginBottom: '12px' }}>{error}</div>}
      {loading && <div>加载中...</div>}

      {!loading && (
        <table style={tableStyle}>
          <thead><tr>
            <th style={thStyle}>ID</th><th style={thStyle}>模块</th><th style={thStyle}>字段</th><th style={thStyle}>类型</th>
            <th style={thStyle}>必填</th><th style={thStyle}>作用域</th><th style={thStyle}>启用</th><th style={thStyle}>操作</th>
          </tr></thead>
          <tbody>
            {filtered.length === 0 && <tr><td colSpan="8" style={tdStyle}>暂无数据</td></tr>}
            {filtered.map((row) => (
              <tr key={row.id}>
                <td style={tdStyle}>{row.id}</td>
                <td style={tdStyle}>{row.module_key}</td>
                <td style={tdStyle}>{row.field_key}</td>
                <td style={tdStyle}>{row.field_type}</td>
                <td style={tdStyle}>{row.is_required ? '是' : '否'}</td>
                <td style={tdStyle}>{row.client_scope}</td>
                <td style={tdStyle}>{row.is_enabled !== false ? '是' : '否'}</td>
                <td style={tdStyle}>
                  {editId === row.id ? (
                    <>
                      <button onClick={() => handleUpdate(row.id)} style={{ marginRight: '4px', padding: '3px 8px', cursor: 'pointer' }}>保存</button>
                      <button onClick={() => { setEditId(null); setEditForm({}); }} style={{ padding: '3px 8px', cursor: 'pointer' }}>取消</button>
                    </>
                  ) : (
                    <>
                      <button onClick={() => { setEditId(row.id); setEditForm({ ...row }); }} style={{ marginRight: '8px', padding: '3px 8px', cursor: 'pointer' }}>编辑</button>
                      <button onClick={() => handleDelete(row.id)} style={{ color: '#d00', padding: '3px 8px', cursor: 'pointer', border: 'none', background: 'none' }}>删除</button>
                    </>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {showCreate && (
        <Modal title="新增 Schema" onClose={() => { setShowCreate(false); setForm({ ...EMPTY_FORM }); }}>
          <SchemaForm form={form} onChange={setForm} onSubmit={handleCreate} />
        </Modal>
      )}

      {editId && (
        <Modal title={`编辑 Schema (ID: ${editId})`} onClose={() => { setEditId(null); setEditForm({}); }}>
          <SchemaForm form={editForm} onChange={setEditForm} onSubmit={() => handleUpdate(editId)} />
        </Modal>
      )}
    </div>
  );
}

function Modal({ title, children, onClose }) {
  return (
    <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(0,0,0,0.45)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000 }}>
      <div style={{ backgroundColor: '#fff', borderRadius: '8px', padding: '24px', minWidth: '480px', maxWidth: '600px', maxHeight: '80vh', overflowY: 'auto' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
          <h3 style={{ margin: 0 }}>{title}</h3>
          <button onClick={onClose} style={{ border: 'none', background: 'none', fontSize: '20px', cursor: 'pointer' }}>×</button>
        </div>
        {children}
      </div>
    </div>
  );
}

function SchemaForm({ form, onChange, onSubmit }) {
  const set = (k, v) => onChange({ ...form, [k]: v });
  const labelStyle = { display: 'block', marginBottom: '4px', fontSize: '13px', fontWeight: 500 };
  const inputStyle = { width: '100%', padding: '6px 8px', borderRadius: '4px', border: '1px solid #ccc', boxSizing: 'border-box', marginBottom: '12px' };

  return (
    <form onSubmit={(e) => { e.preventDefault(); onSubmit(); }}>
      <label style={labelStyle}>模块标识 <input style={inputStyle} value={form.module_key} onChange={(e) => set('module_key', e.target.value)} /></label>
      <label style={labelStyle}>字段标识 <input style={inputStyle} value={form.field_key} onChange={(e) => set('field_key', e.target.value)} /></label>
      <label style={labelStyle}>字段类型
        <select style={inputStyle} value={form.field_type} onChange={(e) => set('field_type', e.target.value)}>
          {FIELD_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
        </select>
      </label>
      <label style={labelStyle}>默认值 <input style={inputStyle} value={form.default_value || ''} onChange={(e) => set('default_value', e.target.value)} /></label>
      <label style={labelStyle}>校验规则 <input style={inputStyle} value={form.validator || ''} onChange={(e) => set('validator', e.target.value)} /></label>
      <label style={labelStyle}><input type="checkbox" checked={form.is_required} onChange={(e) => set('is_required', e.target.checked)} /> 必填</label>
      <label style={labelStyle}><input type="checkbox" checked={form.is_secret} onChange={(e) => set('is_secret', e.target.checked)} /> 敏感字段</label>
      <label style={labelStyle}>描述 <textarea style={{ ...inputStyle, height: '50px' }} value={form.description || ''} onChange={(e) => set('description', e.target.value)} /></label>
      <label style={labelStyle}>作用域 <input style={inputStyle} value={form.client_scope || '*'} onChange={(e) => set('client_scope', e.target.value)} /></label>
      <label style={labelStyle}>排序 <input type="number" style={inputStyle} value={form.sort_order ?? 0} onChange={(e) => set('sort_order', Number(e.target.value))} /></label>
      <button type="submit" style={{ padding: '8px 24px', backgroundColor: '#e94560', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer' }}>提交</button>
    </form>
  );
}

export default ConfigSchemaPage;
