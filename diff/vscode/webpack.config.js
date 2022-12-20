const path = require("path");
const fs = require("fs");

module.exports = {
  entry: {
    diff: {
      import: "./diff.ts",
      filename: "diff.js",
    },
    diffCmd: {
      import: "./cmd.ts",
      filename: "cmd.js",
    },
  },
  output: {
    path: path.resolve(__dirname, "gen"),
    libraryTarget: "umd", // for nodejs need this
  },
  module: {
    rules: [
      {
        test: /\.ts$/,
        // exclude: /(node_modules)/,
        use: {
          loader: "ts-loader",
          options: {
            transpileOnly: true,
          },
        },
      },
      {
        test: /\.(js)$/,
        exclude: /(node_modules)/,
        resolve: {
          extensions: [".ts", ".js"],
        },
        use: {
          loader: "babel-loader",
          options: {
            presets: ["@babel/preset-env"],
          },
        },
      },
    ],
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./"),
    },
    extensions: [".ts", ".js"],
  },
  target: "node",
  plugins: [],
  // devtool: "source-map",
};
