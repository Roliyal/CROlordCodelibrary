const { merge } = require('webpack-merge');
const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const { VueLoaderPlugin } = require('vue-loader');
const TerserPlugin = require('terser-webpack-plugin');

const commonConfig = {
    entry: './src/main.js',
    resolve: {
        alias: {
            '@': path.resolve(__dirname, 'src'),
        },
        extensions: ['.js', '.vue', '.json'],
    },
    module: {
        rules: [
            {
                test: /\.vue$/,
                loader: 'vue-loader',
            },
            {
                test: /\.js$/,
                loader: 'babel-loader',
                exclude: /node_modules/,
            },
            {
                test: /\.css$/,
                use: ['style-loader', 'css-loader'],
            },
            {
                test: /\.(png|svg|jpg|jpeg|gif)$/i,
                type: 'asset/resource',
            },
        ],
    },
    plugins: [
        new VueLoaderPlugin(),
    ],
};

const productionConfig = (env) => ({
    mode: 'production',
    output: {
        path: path.resolve(__dirname, 'dist'),
        filename: env.gray ? 'js/[name].gray.[contenthash].js' : 'js/[name].[contenthash].js',
        publicPath: '/',
        clean: true,
    },
    optimization: {
        minimize: true,
        minimizer: [new TerserPlugin()],
    },
    devtool: false,
    plugins: [
        new HtmlWebpackPlugin({
            template: './public/index.html',
            title: env.gray ? 'Vue Go Guess Number Gray' : 'Vue Go Guess Number',
            templateParameters: {
                BASE_URL: '/',
                SCRIPT_NAME: env.gray ? 'main.gray' : 'main',
            },
            filename: 'index.html',
            inject: 'body',
        }),
    ],
});

const developmentConfig = {
    mode: 'development',
    output: {
        path: path.resolve(__dirname, 'dist'),
        filename: 'js/[name].[contenthash].js',
        publicPath: '/',
        clean: true,
    },
    devServer: {
        static: path.join(__dirname, 'public'),
        port: 8080,
        proxy: {
            '/api': {
                target: 'http://app-go-backend-service.cicd.svc.cluster.local:8081',
                changeOrigin: true,
                pathRewrite: { '^/api': '' },
            },
        },
        hot: true,
    },
    devtool: 'source-map',
    plugins: [
        new HtmlWebpackPlugin({
            template: './public/index.html',
            title: 'Vue Go Guess Number',
            templateParameters: {
                BASE_URL: '/',
                SCRIPT_NAME: 'main',
            },
            filename: 'index.html',
            inject: 'body',
        }),
    ],
};

module.exports = (env) => {
    if (env.production || env.gray) {
        return merge(commonConfig, productionConfig(env));
    }
    return merge(commonConfig, developmentConfig);
};
