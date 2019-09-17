const path = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
var ManifestPlugin = require('webpack-manifest-plugin');

module.exports = {
    mode: 'production',

    entry: './js/main.js',
    output: {
        filename: 'bundle.[contenthash].js',
        path: path.resolve(__dirname, 'assets/public')
    },

    module: {
        rules: [
            {
                test: /\.js$/,
                use: [
                    {
                        loader: 'babel-loader',
                        options: {
                            presets: [
                                '@babel/preset-env'
                            ]
                        }
                    }
                ],
                exclude: /node_modules/,
            },
            {
                test: /\.css/,
                use: [
                    MiniCssExtractPlugin.loader,
                    {
                        loader: 'css-loader',
                        options: {
                            url: false
                        }
                    }
                ]
            }
        ]
    },
    plugins: [
        new MiniCssExtractPlugin({
            filename: 'bundle.[contenthash].css'
        }),
        new ManifestPlugin()
    ]
};
