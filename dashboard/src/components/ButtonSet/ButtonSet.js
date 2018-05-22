import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

const styles = theme => ({
  root: {
    "& button": {
      marginRight: theme.spacing.unit,
      "&:last-child": {
        marginRight: "inherit",
      },
    },
  },
});

// Provides to spec spacing between buttons in a set.
// https://material.io/guidelines/components/buttons.html#buttons-style
class ButtonSet extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    component: PropTypes.oneOfType([PropTypes.func, PropTypes.string]),
  };

  static defaultProps = {
    className: null,
    component: "div",
  };

  render() {
    const {
      className: classNameProp,
      classes,
      component: Component,
      ...props
    } = this.props;

    const className = classnames(classes.root, classNameProp);
    return <Component className={className} {...props} />;
  }
}

export default withStyles(styles)(ButtonSet);
