const path = require('path');

module.exports = {
  entry: './js/src/index.ts',
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        use: 'ts-loader',
        exclude: /node_modules/,
      },
    ],
  },
  resolve: {
    extensions: ['.tsx', '.ts', '.js'],
  },
  output: {
    filename: 'liv-format.js',
    path: path.resolve(__dirname, 'js/dist'),
    library: 'LIVFormat',
    libraryTarget: 'umd',
  },
  experiments: {
    asyncWebAssembly: true,
  },
  devServer: {
    static: {
      directory: path.join(__dirname, 'js/dist'),
    },
    compress: true,
    port: 9000,
  },
};