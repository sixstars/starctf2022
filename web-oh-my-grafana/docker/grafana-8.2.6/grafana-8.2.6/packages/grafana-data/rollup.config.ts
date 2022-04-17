import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import sourceMaps from 'rollup-plugin-sourcemaps';
import json from '@rollup/plugin-json';
import { terser } from 'rollup-plugin-terser';

const pkg = require('./package.json');

const libraryName = pkg.name;

const buildCjsPackage = ({ env }) => {
  return {
    input: `compiled/index.js`,
    output: [
      {
        file: `dist/index.${env}.js`,
        name: libraryName,
        format: 'cjs',
        sourcemap: true,
        exports: 'named',
        globals: {},
      },
    ],
    external: [
      'lodash',
      'rxjs',
      '@grafana/schema', // Load from host
    ],
    plugins: [
      json({
        include: ['../../node_modules/moment-timezone/data/packed/latest.json'],
      }),
      commonjs({
        include: /node_modules/,
      }),
      resolve(),
      sourceMaps(),
      env === 'production' && terser(),
    ],
  };
};
export default [buildCjsPackage({ env: 'development' }), buildCjsPackage({ env: 'production' })];
