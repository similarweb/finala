const path = require('path')
const webpack = require('webpack')
const ExtractTextPlugin = require("extract-text-webpack-plugin");
const FaviconsWebpackPlugin = require('favicons-webpack-plugin')

module.exports = (options) => ({
  mode: options.mode,
  entry: options.entry,
  devtool: options.devtool,
  output: Object.assign(
    {
      path: path.resolve(process.cwd(), 'build'),
      publicPath: '/'
    },
    options.output
  ), 
  module: {
    rules: options.module.rules.concat([
      {
        test: /\.js$/, 
        exclude: /node_modules/,
        use: {
          loader: 'babel-loader',
        }
      },
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: ['babel-loader', 'eslint-loader']
      },
      {
        test: /\.(css|sass|scss)$/,
        use: ExtractTextPlugin.extract({ fallback: 'style-loader', use: 'css-loader!sass-loader' })
      },
      {
        test: /\.(png|jp(e*)g)$/,  
        use: [{
            loader: 'url-loader',
            options: { 
                limit: 8000, 
                name: 'images/[hash]-[name].[ext]'
            } 
        }]
      },
      {
        test: /\.svg$/,
        loader: 'svg-inline-loader'
      }
    ]),
  },
  plugins: options.plugins.concat([    
    new ExtractTextPlugin("app.[hash].css"),
    new FaviconsWebpackPlugin('./src/styles/icons/logo.png')


  ]),
  resolveLoader: {
    modules: [
      'node_modules',
    ],
  },
});
