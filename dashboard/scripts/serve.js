import chalk from "chalk";
import compress from "koa-compress";
import connect from "koa-connect";
import historyFallback from "connect-history-api-fallback";
import http from "http";
import killable from "killable";
import Koa from "koa";
import koaWebpack from "koa-webpack";
import proxy from "http-proxy-middleware";

import "./exceptionHandler";
import assertEnv from "./assertEnv";
import { loading } from "./log";
import createCompiler from "./createCompiler";
import getConfig from "../config/webpack.config";

process.env.NODE_ENV = process.env.NODE_ENV || "development";
assertEnv();

const proxyPaths = ["/auth", "/graphql"];
const port = parseInt(process.env.PORT, 10) || 3001;
const hotPort = parseInt(process.env.HOT_PORT, 10) || port + 1;

const config = getConfig();

const compiler = createCompiler(config);

let compiled = false;
compiler.hooks.done.tap("serve", stats => {
  const { errors } = stats.toJson({});

  if (!compiled && !errors.length) {
    compiled = true;
    console.log();
    console.log(
      `You can now visit the app in your browser`,
      chalk.gray(`visit http://localhost:${port}`),
    );
  }
});

const app = new Koa();

const webpackMiddleware = koaWebpack({
  compiler,
  dev: {
    publicPath: config.output.publicPath,
    logLevel: "silent",
  },
  hot: process.env.NODE_ENV === "development" && {
    port: hotPort,
    logLevel: "silent",
  },
});

app.use(compress());
app.use(
  connect(
    proxy(proxyPaths, {
      target: "http://localhost:8080",
      logLevel: "silent",
    }),
  ),
);
app.use(connect(historyFallback()));
app.use(webpackMiddleware);

const server = killable(http.createServer(app.callback()));

server.on("error", error => {
  throw error;
});

server.listen(port);

["SIGINT", "SIGTERM"].forEach(sig => {
  process.on(sig, () => {
    loading.stop(true);
    console.info(`Process Ended via ${sig}`);
    server.kill();
    webpackMiddleware.close(() => process.exit(0));
  });
});
