/* eslint-disable import/no-webpack-loader-syntax */

import appIcon from "!!file-loader?name=[name].[ext]!./app-icon.png";
import appIconSmall from "!!file-loader?name=[name].[ext]!./app-icon-sm.png";

export default ({ emitFile }) =>
  emitFile(
    "manifest.json",
    JSON.stringify({
      short_name: "Sensu",
      name: "Sensu UI",
      icons: [
        {
          src: appIcon,
          sizes: "1024x1024",
          type: "image/png",
        },
        {
          src: appIconSmall,
          sizes: "64x64",
          type: "image/png",
        },
      ],
      start_url: "/",
      display: "standalone",
      theme_color: "#000000",
      background_color: "#ffffff",
    }),
  );
