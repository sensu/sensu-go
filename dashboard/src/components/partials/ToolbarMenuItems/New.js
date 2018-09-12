import React from "react";
import { compose, setDisplayName, defaultProps } from "recompose";

import NewIcon from "/icons/New";
import MenuItem from "./MenuItem";

const enhance = compose(
  setDisplayName("ToolbarMenuItems.New"),
  defaultProps({
    title: "New…",
    titleCondensed: "New",
    icon: <NewIcon />,
  }),
);
export default enhance(MenuItem);
