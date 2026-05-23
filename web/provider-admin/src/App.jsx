import { Routes, Route, NavLink } from 'react-router-dom';
import ConfigSchemaPage from './pages/ConfigSchemaPage';
import ConfigValuePage from './pages/ConfigValuePage';
import I18nStringPage from './pages/I18nStringPage';
import I18nPackPage from './pages/I18nPackPage';

const NAV_ITEMS = [
  { path: '/admin/config/schema', label: '配置 Schema' },
  { path: '/admin/config/value', label: '配置值' },
  { path: '/admin/i18n/strings', label: '多语言 Key' },
  { path: '/admin/i18n/packs', label: '语言包发布' },
];

const navStyle = {
  display: 'flex',
  gap: '4px',
  padding: '12px 20px',
  backgroundColor: '#1a1a2e',
  marginBottom: '20px',
};

const linkStyle = (isActive) => ({
  padding: '8px 16px',
  borderRadius: '4px',
  textDecoration: 'none',
  color: isActive ? '#fff' : '#aaa',
  backgroundColor: isActive ? '#e94560' : 'transparent',
  fontSize: '14px',
});

function App() {
  return (
    <div style={{ fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif', minHeight: '100vh', backgroundColor: '#f5f5f5' }}>
      <nav style={navStyle}>
        {NAV_ITEMS.map((item) => (
          <NavLink key={item.path} to={item.path} style={({ isActive }) => linkStyle(isActive)} end={item.path === '/admin/config/schema'}>
            {item.label}
          </NavLink>
        ))}
      </nav>
      <div style={{ padding: '0 20px' }}>
        <Routes>
          <Route path="/admin/config/schema" element={<ConfigSchemaPage />} />
          <Route path="/admin/config/value" element={<ConfigValuePage />} />
          <Route path="/admin/i18n/strings" element={<I18nStringPage />} />
          <Route path="/admin/i18n/packs" element={<I18nPackPage />} />
        </Routes>
      </div>
    </div>
  );
}

export default App;
