import js from '@eslint/js'
import { defineConfig, globalIgnores } from 'eslint/config'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import globals from 'globals'
import tseslint from 'typescript-eslint'

export default defineConfig([
  globalIgnores(['dist', 'node_modules']),
  js.configs.recommended,
  react.configs.flat.recommended,
  tseslint.configs.recommended,
  {
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: { ...globals.browser, ...globals.node },
    },
    plugins: { 'react-hooks': reactHooks },
    settings: { react: { version: 'detect' } },
    rules: {
      // react-hooks 7 の recommended は React Compiler 向けの rule も有効にするため、v4 と同じ 2 rule だけを有効にする
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'error',
      'react/react-in-jsx-scope': 'off',
      '@typescript-eslint/no-explicit-any': 'error',
      'no-unused-vars': 'off',
      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
    },
  },
  {
    files: ['**/*.jsx'],
    rules: { '@typescript-eslint/no-unused-vars': 'off' },
  },
  {
    files: ['src/__tests__/**/*'],
    languageOptions: { globals: { ...globals.vitest } },
    rules: { '@typescript-eslint/no-explicit-any': 'off' },
  },
])
