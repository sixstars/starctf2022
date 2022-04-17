'use strict';

const path = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');

module.exports = function (options) {
  return {
    test: /\.scss$/,
    exclude: /node_modules/,
    use: [
      MiniCssExtractPlugin.loader,
      {
        loader: 'css-loader',
        options: {
          importLoaders: 2,
          url: options.preserveUrl,
          sourceMap: options.sourceMap,
        },
      },
      {
        loader: 'postcss-loader',
        options: {
          sourceMap: options.sourceMap,
          postcssOptions: {
            config: path.resolve(__dirname),
          },
        },
      },
      {
        loader: 'sass-loader',
        options: {
          sourceMap: options.sourceMap,
        },
      },
    ],
  };
};
