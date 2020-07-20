const path = require("path");
const webpack = require("webpack");
const webpackDevMiddleware = require("webpack-dev-middleware");
const webpackHotMiddleware = require("webpack-hot-middleware");

module.exports = function middleware(app, config) {
  const compiler = webpack(config);

  const middleware = webpackDevMiddleware(compiler, {
    publicPath: config.output.publicPath,
    historyApiFallback: true,
  });

  app.use(middleware);
  app.use(webpackHotMiddleware(compiler));

  const fs = middleware.fileSystem;
  app.get("/api/v1/settings", (req, res) => {
    fs.readFile(path.join(compiler.outputPath, "index.html"), (err, file) => {
      res.send({ api_endpoint: "http://127.0.0.1:8081" });
    });
  });
  app.get("*", (req, res) => {
    fs.readFile(path.join(compiler.outputPath, "index.html"), (err, file) => {
      if (err) {
        res.sendStatus(404);
      } else {
        res.send(file.toString());
      }
    });
  });
};
