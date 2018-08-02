import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

const styles = () => ({
  root: {
    verticalAlign: "bottom",
  },
  leftAligned: {
    marginLeft: -8,
    paddingRight: 4,
  },
  rightAligned: {
    marginRight: -8,
    paddingLeft: 4,
  },
  icon: {
    fontSize: 20,
  },
});

class ButtonIcon extends React.PureComponent {
  static propTypes = {
    alignment: PropTypes.oneOf(["left", "right"]),
    classes: PropTypes.object.isRequired,
    component: PropTypes.oneOfType([PropTypes.func, PropTypes.string]),
    children: PropTypes.node.isRequired,
  };

  static defaultProps = {
    alignment: "left",
    component: "span",
  };

  render() {
    const {
      alignment,
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
      [classes.leftAligned]: alignment === "left",
      [classes.rightAligned]: alignment === "right",
    });

    return (
      <Component className={className} {...props}>
        {children}
      </Component>
    );
  }
}

export default withStyles(styles)(ButtonIcon);
