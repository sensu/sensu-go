import React from "react";
import { compose, setDisplayName, defaultProps } from "recompose";

import UnsilenceIcon from "/icons/Unsilence";
import MenuItem from "./MenuItem";

const enhance = compose(
  setDisplayName("ToolbarMenuItems.Silence"),
  defaultProps({
    title: "Clear Silences",
    titleCondensed: "Unsilence",
    description: "Clear silences for target item(s).",
    icon: <UnsilenceIcon />,
  }),
);
export default enhance(MenuItem);
