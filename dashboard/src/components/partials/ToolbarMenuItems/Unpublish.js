import React from "react";
import { compose, setDisplayName, defaultProps } from "recompose";

import UnpublishIcon from "/icons/Unpublish";
import MenuItem from "./MenuItem";

const enhance = compose(
  setDisplayName("ToolbarMenuItems.Unpublish"),
  defaultProps({
    title: "Unpublish",
    description: "Unpublish target item(s).",
    icon: <UnpublishIcon />,
  }),
);
export default enhance(MenuItem);
