import React from "react";
import { withStyles } from "@material-ui/core/styles";

import reset from "./reset.css";

class ResetStyles extends React.PureComponent {
  render() {
    return null;
  }
}

export default withStyles({ "@global": reset })(ResetStyles);
