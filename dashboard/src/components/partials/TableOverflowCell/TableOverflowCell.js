import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";

import TableCell from "@material-ui/core/TableCell";

const styles = () => ({
  root: {
    width: "100%",
    maxWidth: 0,

    "&, & *": {
      whiteSpace: "nowrap",
      textOverflow: "ellipsis",
      overflow: "hidden",
    },
  },
});

class TableOverflowCell extends React.PureComponent {
  static propTypes = { classes: PropTypes.object.isRequired };

  render() {
    const { classes, ...props } = this.props;
    return <TableCell {...props} className={classes.root} padding="none" />;
  }
}

export default withStyles(styles)(TableOverflowCell);
