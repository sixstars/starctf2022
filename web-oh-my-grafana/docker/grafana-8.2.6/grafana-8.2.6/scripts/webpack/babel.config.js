module.exports = function getBabelConfig(options = {}) {
  const babelOptions = {
    cacheDirectory: true,
    cacheCompression: false,
    babelrc: false,
    // Note: order is bottom-to-top and/or right-to-left
    presets: [
      [
        '@babel/preset-env',
        {
          bugfixes: true,
          browserslistEnv: process.env.BABEL_ENV || options.BABEL_ENV || 'production',
          useBuiltIns: 'entry',
          corejs: '3.10',
        },
      ],
      [
        '@babel/preset-typescript',
        {
          allowNamespaces: true,
          allowDeclareFields: true,
        },
      ],
      [
        '@babel/preset-react',
        {
          runtime: 'automatic',
        },
      ],
    ],
    plugins: [
      [
        '@babel/plugin-transform-typescript',
        {
          allowNamespaces: true,
          allowDeclareFields: true,
        },
      ],
      ['@babel/plugin-proposal-object-rest-spread', { loose: true }],
      '@babel/plugin-transform-react-constant-elements',
      '@babel/plugin-proposal-nullish-coalescing-operator',
      '@babel/plugin-proposal-optional-chaining',
      '@babel/plugin-syntax-dynamic-import', // needed for `() => import()` in routes.ts
      'angularjs-annotate',
    ],
  };

  if (options.REACT_REFRESH) {
    babelOptions.plugins.push('react-refresh/babel');
  }

  return babelOptions;
};
