module.exports = {
  root: true,
  env: { browser: true, es2020: true },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:react-hooks/recommended',
    'plugin:boundaries/recommended',
  ],
  ignorePatterns: ['dist', '.eslintrc.cjs'],
  parser: '@typescript-eslint/parser',
  plugins: ['react-refresh', 'boundaries'],
  settings: {
    'import/resolver': {
      typescript: { project: './tsconfig.json' },
    },
    'boundaries/elements': [
      { type: 'app',      pattern: 'src/app/**' },
      { type: 'routes',   pattern: 'src/routes/**' },
      { type: 'features', pattern: 'src/features/*', mode: 'folder', capture: ['feature'] },
      { type: 'entities', pattern: 'src/entities/*', mode: 'folder', capture: ['entity'] },
      { type: 'shared',   pattern: 'src/shared/**' },
    ],
    'boundaries/include': ['src/**/*'],
    'boundaries/ignore': ['src/main.tsx', 'src/vite-env.d.ts'],
  },
  rules: {
    'react-refresh/only-export-components': [
      'warn',
      { allowConstantExport: true },
    ],
    'boundaries/dependencies': [
      'warn',
      {
        default: 'disallow',
        rules: [
          { from: ['app'],      allow: ['app', 'routes', 'features', 'entities', 'shared'] },
          { from: ['routes'],   allow: ['app', 'routes', 'features', 'entities', 'shared'] },
          // features are allowed to read URL constants from routes/paths and app-level
          // providers (AuthContext, useAuth) — both are framework glue, not feature code.
          { from: ['features'], allow: ['app', 'routes', 'features', 'entities', 'shared'] },
          { from: ['entities'], allow: ['entities', 'shared'] },
          { from: ['shared'],   allow: ['shared'] },
        ],
      },
    ],
    'boundaries/no-private': 'off',
  },
}
