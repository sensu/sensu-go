import path from "path";
import fs from "fs";

import historyFallback from "connect-history-api-fallback";
import http from "http";
import killable from "killable";
import express from "express";
import compression from "compression";
import proxy from "http-proxy-middleware";

import "./util/exceptionHandler";

const root = fs.realpathSync(process.cwd());
const proxyPaths = ["/auth", "/graphql"];
const port = parseInt(process.env.PORT, 10) || 3001;

const staticAssets = express.Router();

staticAssets.use(express.static(path.join(root, "build/vendor/public")));
staticAssets.use(express.static(path.join(root, "build/lib/public")));
staticAssets.use(express.static(path.join(root, "build/app/public")));

const app = express();
app.use(compression());

app.use(
  proxy(proxyPaths, {
    target: "http://localhost:8080",
    logLevel: "silent",
  }),
);

app.use(staticAssets);

app.use(
  historyFallback({
    verbose: true,
    disableDotRule: true,
  }),
);

app.use(staticAssets);

const server = killable(http.createServer(app));

server.on("error", error => {
  throw error;
});

server.on("listening", () => {
  console.log("listening on", server.address().port);
});

server.listen(port);

["SIGINT", "SIGTERM"].forEach(sig => {
  process.on(sig, () => {
    console.info(`Process Ended via ${sig}`);
    server.kill();
  });
});
