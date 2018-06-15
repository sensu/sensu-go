import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";

const styles = theme => ({
  root: {
    backgroundColor: theme.palette.background.paper,
    borderStyle: "solid",
    borderColor: theme.palette.divider,
    borderTopWidth: 0,
    borderLeftWidth: 0,
    borderRightWidth: 0,
    borderBottomWidth: 1,

    // Keep content from bleeding out of container
    overflow: "hidden",
  },
});

export class TableList extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    children: PropTypes.node.isRequired,
  };

  static defaultProps = {
    className: "",
  };

  render() {
    const { classes, className: classNameProp, children } = this.props;
    const className = `${classes.root} ${classNameProp}`;
    return <div className={className}>{children}</div>;
  }
}

export default withStyles(styles)(TableList);
