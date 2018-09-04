import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

const styles = theme => ({
  root: {
    display: "flex",
    alignItems: "center",
  },
  bottomMargin: {
    marginBottom: theme.spacing.unit * 2,
  },
});

class Content extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    children: PropTypes.node.isRequired,
    bottomMargin: PropTypes.bool,
  };

  static defaultProps = {
    bottomMargin: false,
    className: "",
    container: false,
  };

  render() {
    const {
      children,
      classes,
      className: classNameProp,
      bottomMargin,
    } = this.props;

    const className = classnames(classes.root, classNameProp, {
      [classes.bottomMargin]: bottomMargin,
    });

    return <div className={className}>{children}</div>;
  }
}

export default withStyles(styles)(Content);
