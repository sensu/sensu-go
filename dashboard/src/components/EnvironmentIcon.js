import React from "react";
import PropTypes from "prop-types";

import { emphasize } from "material-ui/styles/colorManipulator";
import { withStyles } from "material-ui/styles";

const colours = {
  BLUE: "#8AB8D0",
  GRAY: "#9A9EA5",
  GREEN: "#8AD1AF",
  ORANGE: "#F4AD5F",
  PINK: "#FA8072",
  PURPLE: "#AD8AD1",
  YELLOW: "#FAD66B",
};

const styles = () => ({
  root: {
    borderRadius: "100%",
    border: "1px solid",
    boxSizing: "border-box",
  },
});

class EnvironmentIcon extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    colour: PropTypes.string.isRequired,
    size: PropTypes.number,
  };

  static defaultProps = {
    className: null,
    size: 8.0,
  };

  render() {
    const { classes, className, colour, size, ...props } = this.props;
    const borderWidth = Math.floor(size * (1 / 8));
    const effectiveColour = colours[colour];
    const inlineStyle = {
      backgroundColor: effectiveColour,
      borderColor: emphasize(effectiveColour, 0.15),
      borderWidth,
      width: size,
      height: size,
    };

    return (
      <div
        style={inlineStyle}
        className={`${classes.root} ${className}`}
        {...props}
      />
    );
  }
}

export default withStyles(styles)(EnvironmentIcon);
