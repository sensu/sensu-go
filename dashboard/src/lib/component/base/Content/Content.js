import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

const styles = theme => ({
  root: {
    display: "flex",
    alignItems: "center",
  },
  marginBottom: {
    marginBottom: theme.spacing.unit * 2,
  },
});

class Content extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    children: PropTypes.node.isRequired,
    marginBottom: PropTypes.bool,
  };

  static defaultProps = {
    className: "",
    container: false,
    marginBottom: false,
  };

  render() {
    const {
      children,
      classes,
      className: classNameProp,
      marginBottom,
    } = this.props;

    const className = classnames(classes.root, classNameProp, {
      [classes.marginBottom]: marginBottom,
    });

    return <div className={className}>{children}</div>;
  }
}

export default withStyles(styles)(Content);
