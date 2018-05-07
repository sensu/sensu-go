import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";

const styles = theme => ({
  [theme.breakpoints.up("sm")]: {
    root: {
      // Shadow
      borderRadius: 2,
      boxShadow: theme.shadows[2],
      "& > :first-child": {
        borderTopLeftRadius: 2,
        borderTopRightRadius: 2,
        overflow: "hidden",
      },
      "& > :last-child": {
        borderBottomLeftRadius: 2,
        borderBottomRightRadius: 2,
        overflow: "hidden",
      },
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

    const className = classnames(classes.root, classNameProp);

    return <div className={className}>{children}</div>;
  }
}

export default withStyles(styles)(TableList);
