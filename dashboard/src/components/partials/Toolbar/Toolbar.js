import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import createStyledComponent from "/components/createStyledComponent";

const Fill = createStyledComponent({
  name: "Fill",
  component: "div",
  styles: () => ({
    flex: "1 1 auto",
  }),
});

const styles = theme => ({
  root: {
    display: "flex",
    flex: "1 1 100%",
    alignItems: "center",
    color: theme.palette.text.secondary,
  },
});

class Toolbar extends React.PureComponent {
  static propTypes = {
    // provided by withStyles
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    component: PropTypes.oneOfType([PropTypes.string, PropTypes.func]),
    left: PropTypes.node,
    middle: PropTypes.node,
    right: PropTypes.node,
  };

  static defaultProps = {
    component: "div",
    className: null,
    left: null,
    middle: null,
    right: null,
  };

  render() {
    const {
      classes,
      className: classNameProp,
      component: Component,
      left,
      middle,
      right,
      ...props
    } = this.props;
    const className = classnames(classNameProp, classes.root);

    return (
      <Component className={className} {...props}>
        {left}
        <Fill />
        {middle}
        <Fill />
        {right}
      </Component>
    );
  }
}

export default withStyles(styles)(Toolbar);
