import React from "react";
import { compose, setDisplayName, defaultProps } from "recompose";

import SilenceIcon from "/icons/Silence";
import MenuItem from "./MenuItem";

const enhance = compose(
  setDisplayName("ToolbarMenuItems.Silence"),
  defaultProps({
    autoClose: false,
    title: "Silence",
    description: "Create a silence for target item(s).",
    icon: <SilenceIcon />,
  }),
);
export default enhance(MenuItem);
