import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import TableCell from "@material-ui/core/TableCell";

const styles = theme => {
  const bgColor = theme.palette.background.paper;

  return {
    root: {
      position: "relative",
    },
    container: {
      top: 0,
      right: 0,
      display: "flex",
      position: "absolute",
      paddingLeft: theme.spacing.unit * 1.5,
      paddingRight: theme.spacing.unit * 1.5,
      backgroundColor: bgColor,
      boxShadow: `
        ${-theme.spacing.unit * 4}px
        0px
        ${theme.spacing.unit * 2}px
        ${-theme.spacing.unit}px
        ${bgColor}
      `,
      overflow: "hidden",
    },
  };
};

class TableToolbarCell extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    classes: PropTypes.object.isRequired,
    disabled: PropTypes.bool,
  };

  static defaultProps = {
    disabled: false,
  };

  render() {
    const { children, classes, disabled } = this.props;

    return (
      <TableCell padding="none" className={classes.root}>
        {!disabled && <div className={classes.container}>{children()}</div>}
      </TableCell>
    );
  }
}

export default withStyles(styles)(TableToolbarCell);
