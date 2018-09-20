import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import ResizeObserver from "react-resize-observer";

const styles = () => ({
  root: {
    width: "100%",
    textAlign: "right",
  },
});

class Autosizer extends React.Component {
  static displayName = "ToolbarMenu.Autosizer";

  static propTypes = {
    children: PropTypes.func.isRequired,
    // provided by withStyles prop
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    component: PropTypes.oneOfType([PropTypes.func, PropTypes.string]),
    position: PropTypes.oneOf(["relative", "fixed"]),
    style: PropTypes.object,
  };

  static defaultProps = {
    className: null,
    component: "div",
    position: "relative",
    style: {},
  };

  state = {
    width: null,
  };

  handleResize = rect => {
    this.setState(state => {
      if (state.width === rect.width) {
        return null;
      }

      return { width: rect.width };
    });
  };

  render() {
    const { width } = this.state;
    const {
      component: Component,
      children,
      classes,
      className: classNameProp,
      style,
      position,
    } = this.props;

    const className = classnames(classNameProp, classes.root);

    return (
      <Component style={{ ...style, position }} className={className}>
        <ResizeObserver onResize={this.handleResize} />
        {width && children({ width })}
      </Component>
    );
  }
}

export default withStyles(styles)(Autosizer);
