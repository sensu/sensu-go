import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";

const styles = theme => ({
  root: {
    marginLeft: theme.spacing.unit,
    marginRight: theme.spacing.unit,

    [theme.breakpoints.up("md")]: {
      margin: 0,
    },
  },
});

class Content extends React.Component {
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

export default withStyles(styles)(Content);
