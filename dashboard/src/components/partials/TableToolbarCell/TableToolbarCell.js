import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import TableCell from "@material-ui/core/TableCell";

const styles = theme => {
  const bgColor = theme.palette.background.paper;

  return {
    root: {
      position: "relative",
    },
    container: {
      display: "flex",
      paddingLeft: theme.spacing.unit * 1.5,
      paddingRight: theme.spacing.unit * 1.5,
    },
    floating: {
      top: 0,
      right: 0,
      position: "absolute",
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
    floating: PropTypes.bool,
  };

  static defaultProps = {
    disabled: false,
    floating: false,
  };

  renderCell = () => {
    const children = this.props.children();
    const containerClasses = classnames(this.props.classes.container, {
      [this.props.classes.floating]: this.props.floating,
    });

    return <div className={containerClasses}>{children}</div>;
  };

  render() {
    const { classes, disabled } = this.props;

    return (
      <TableCell padding="none" className={classes.root}>
        {!disabled && this.renderCell()}
      </TableCell>
    );
  }
}

export default withStyles(styles)(TableToolbarCell);
