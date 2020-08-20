const path = require("path");
const webpack = require("webpack");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const FaviconsWebpackPlugin = require("favicons-webpack-plugin");

module.exports = (options) => ({
  mode: options.mode,
  entry: options.entry,
  devtool: options.devtool,
  output: Object.assign(
    {
      path: path.resolve(process.cwd(), "build", "static"),
      publicPath: "/static",
    },
    options.output
  ),
  module: {
    rules: options.module.rules.concat([
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: "babel-loader",
        },
      },
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: ["babel-loader", "eslint-loader"],
      },
      {
        test: /\.(css|sass|scss)$/,
        use: [
          {
            loader: MiniCssExtractPlugin.loader,
            options: {
              hmr: process.env.NODE_ENV === "development",
            },
          },
          "css-loader",
          "sass-loader",
        ],
      },
      {
        test: /\.(png|jp(e*)g)$/,
        use: [
          {
            loader: "url-loader",
            options: {
              limit: 8000,
              name: "images/[hash]-[name].[ext]",
            },
          },
        ],
      },
      {
        test: /\.svg$/,
        loader: "svg-inline-loader",
      },
      {
        test: /\.inline.svg$/,
        loader: "svg-react-loader",
      },
    ]),
  },
  plugins: options.plugins.concat([
    new MiniCssExtractPlugin(), // "app.[hash].css"
    new FaviconsWebpackPlugin("./src/styles/icons/icon.png"),
  ]),
  resolveLoader: {
    modules: ["node_modules"],
  },
});
