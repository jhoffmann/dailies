const path = require('path');

module.exports = {
  entry: './web/static/js/main.js',
  output: {
    path: path.resolve(__dirname, 'web/static/js'),
    filename: 'bundle.js',
  },
  resolve: {
    extensions: ['.js']
  }
};