'use strict';

const { merge } = require('webpack-merge');
const common = require('./webpack.common.js');
const path = require('path');
const { DefinePlugin } = require('webpack');
const ReactRefreshWebpackPlugin = require('@pmmmwh/react-refresh-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const getBabelConfig = require('./babel.config');

module.exports = merge(common, {
  devtool: 'inline-source-map',
  mode: 'development',

  entry: {
    app: ['./public/app/index.ts'],
    dark: './public/sass/grafana.dark.scss',
    light: './public/sass/grafana.light.scss',
  },

  watchOptions: {
    ignored: /node_modules/,
  },

  devServer: {
    devMiddleware: {
      writeToDisk: true,
    },
    historyApiFallback: true,
    hot: true,
    open: false,
    port: 3333,
    proxy: {
      '!/public/build': 'http://localhost:3000',
    },
    static: {
      publicPath: '/public/build/',
    },
  },

  module: {
    // Note: order is bottom-to-top and/or right-to-left
    rules: [
      {
        test: /\.tsx?$/,
        use: {
          loader: 'babel-loader',
          options: getBabelConfig({ BABEL_ENV: 'dev', REACT_REFRESH: true }),
        },
        exclude: /node_modules/,
        include: [path.resolve(__dirname, '../../public/'), path.resolve(__dirname, '../../packages/')],
      },
      require('./sass.rule.js')({
        sourceMap: false,
        preserveUrl: false,
      }),
    ],
  },

  // https://webpack.js.org/guides/build-performance/#output-without-path-info
  output: {
    filename: '[name].js',
    pathinfo: false,
  },

  // https://webpack.js.org/guides/build-performance/#avoid-extra-optimization-steps
  optimization: {
    runtimeChunk: true,
    removeAvailableModules: false,
    removeEmptyChunks: false,
    splitChunks: false,
  },

  plugins: [
    new MiniCssExtractPlugin({
      filename: 'grafana.[name].[fullhash].css',
    }),
    new HtmlWebpackPlugin({
      filename: path.resolve(__dirname, '../../public/views/index.html'),
      template: path.resolve(__dirname, '../../public/views/index-template.html'),
      hash: true,
      inject: false,
      chunksSortMode: 'none',
      excludeChunks: ['dark', 'light'],
    }),
    new ReactRefreshWebpackPlugin(),
    new DefinePlugin({
      'process.env': {
        NODE_ENV: JSON.stringify('development'),
      },
    }),
  ],
});
