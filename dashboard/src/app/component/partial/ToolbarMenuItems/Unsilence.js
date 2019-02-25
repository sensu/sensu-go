import React from "react";
import { compose, setDisplayName, defaultProps } from "recompose";

import UnsilenceIcon from "/lib/component/icon/Unsilence";
import MenuItem from "./MenuItem";

const enhance = compose(
  setDisplayName("ToolbarMenuItems.Silence"),
  defaultProps({
    autoClose: false,
    title: "Clear silence",
    titleCondensed: "Clear silence",
    description: "Clear silences for target item(s).",
    icon: <UnsilenceIcon />,
  }),
);
export default enhance(MenuItem);
