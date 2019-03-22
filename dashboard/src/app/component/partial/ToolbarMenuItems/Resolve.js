import React from "react";
import { compose, setDisplayName, defaultProps } from "recompose";

import SmallCheckIcon from "/lib/component/icon/SmallCheck";
import MenuItem from "./MenuItem";

const enhance = compose(
  setDisplayName("ToolbarMenuItems.Resolve"),
  defaultProps({
    title: "Resolve",
    description: "Set status of event(s) to resolved state.",
    icon: <SmallCheckIcon />,
  }),
);
export default enhance(MenuItem);
