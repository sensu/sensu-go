import React from "react";
import { compose, setDisplayName, defaultProps } from "recompose";

import DeleteIcon from "/icons/Delete";
import MenuItem from "./MenuItem";

const enhance = compose(
  setDisplayName("ToolbarMenuItems.Delete"),
  defaultProps({
    autoClose: false,
    description: "Permenantly delete resource.",
    icon: <DeleteIcon />,
    title: "Delete…",
    titleCondensed: "Delete",
  }),
);
export default enhance(MenuItem);
