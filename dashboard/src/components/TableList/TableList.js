import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

const styles = theme => ({
  root: {
    backgroundColor: theme.palette.background.paper,
    border: 1,
    borderStyle: "solid",
    borderColor: theme.palette.divider,
    borderRadius: 2,

    // Keep content from bleeding out of container
    overflow: "hidden",

    // Shadow
    boxShadow: theme.shadows[2],
    [theme.breakpoints.up("lg")]: {
      boxShadow: theme.shadows[0],
    },
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
