
module.exports = (app, options) => {
  const webpackConfig = require('../config/webpack.config.development.js');
    const middleware = require('./middleware.development.js');
    middleware(app, webpackConfig);
};
  