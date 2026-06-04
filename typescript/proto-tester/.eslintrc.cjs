const path = require('path');

// 尝试引用 admin-web 配置，不存在则回退到基础配置
const adminWebEslintPath = path.resolve(__dirname, '../admin-web/.eslintrc.js');
let adminWebConfig = {};
try {
  adminWebConfig = require(adminWebEslintPath);
} catch (e) {
  // admin-web eslint 配置不存在，使用默认配置
}

module.exports = {
  ...adminWebConfig,
  root: true,
  env: {
    browser: true,
    es2020: true,
    node: true,
  },
  extends: ['eslint:recommended', 'plugin:react/recommended', 'plugin:react-hooks/recommended'],
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module',
    ecmaFeatures: { jsx: true },
  },
  settings: {
    react: { version: 'detect' },
  },
  rules: {
    'react/react-in-jsx-scope': 'off',
    'react/prop-types': 'off',
  },
};
