import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

const styles = theme => {
  const padding = theme.spacing.unit / 2;

  return {
    root: {
      verticalAlign: "bottom",
    },
    leftAligned: {
      marginLeft: -padding,
      paddingRight: padding,
    },
    rightAligned: {
      marginRight: -padding,
      paddingLeft: padding,
    },
    icon: {
      fontSize: theme.spacing.unit * 2.5,
    },
  };
};

class ButtonIcon extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    component: PropTypes.oneOfType([PropTypes.func, PropTypes.string]),
    children: PropTypes.node.isRequired,
    alignRight: PropTypes.bool,
  };

  static defaultProps = {
    component: "span",
    alignRight: false,
  };

  render() {
    const {
      alignRight,
      classes,
      component: Component,
      children: childrenProp,
      ...props
    } = this.props;

    const children = React.cloneElement(childrenProp, {
      classes: {
        root: classes.icon,
      },
    });

    const className = classnames(classes.root, {
      [classes.leftAligned]: !alignRight,
      [classes.rightAligned]: alignRight,
    });

    return (
      <Component className={className} {...props}>
        {children}
      </Component>
    );
  }
}

export default withStyles(styles)(ButtonIcon);
