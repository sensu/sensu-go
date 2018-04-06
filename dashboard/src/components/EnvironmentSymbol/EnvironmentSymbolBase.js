import React from "react";
import PropTypes from "prop-types";
import TagOrb from "../TagOrb";

const colours = {
  BLUE: "#8AB8D0",
  GRAY: "#9A9EA5",
  GREEN: "#8AD1AF",
  ORANGE: "#F4AD5F",
  PINK: "#FA8072",
  PURPLE: "#AD8AD1",
  YELLOW: "#FAD66B",
};

class EnvironmentSymbol extends React.Component {
  static propTypes = {
    colour: PropTypes.oneOf(Object.keys(colours)).isRequired,
  };

  render() {
    const { colour, ...props } = this.props;
    const effectiveColour = colours[colour];
    return <TagOrb colour={effectiveColour} {...props} />;
  }
}

export default EnvironmentSymbol;
