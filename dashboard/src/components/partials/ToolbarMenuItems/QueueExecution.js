import React from "react";
import { compose, setDisplayName, defaultProps } from "recompose";

import QueueIcon from "@material-ui/icons/Queue";
import MenuItem from "./MenuItem";

const enhance = compose(
  setDisplayName("ToolbarMenuItems.QueueExecution"),
  defaultProps({
    title: "Execute",
    description: "Queue up ad-hoc execution for check(s).",
    icon: <QueueIcon />,
  }),
);
export default enhance(MenuItem);
