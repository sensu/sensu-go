import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";

const styles = theme => ({
  root: {
    paddingRight: theme.spacing.unit / 2,
    verticalAlign: "bottom",
  },
  icon: {
    fontSize: theme.spacing.unit * 2.5,
  },
});

class ButtonIcon extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    component: PropTypes.oneOfType([PropTypes.func, PropTypes.string]),
    children: PropTypes.node.isRequired,
  };

  static defaultProps = {
    component: "span",
  };

  render() {
    const {
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

    return (
      <Component className={classes.root} {...props}>
        {children}
      </Component>
    );
  }
}

export default withStyles(styles)(ButtonIcon);
