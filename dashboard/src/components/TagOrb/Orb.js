import React from "react";
import PropTypes from "prop-types";
import { emphasize } from "@material-ui/core/styles/colorManipulator";
import { withStyles } from "@material-ui/core/styles";

const styles = () => ({
  root: {
    borderRadius: "100%",
    border: "1px solid",
    boxSizing: "border-box",
  },
});

class Oval extends React.Component {
  static displayName = "TagOrb.Oval";

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
    const inlineStyle = {
      backgroundColor: colour,
      borderColor: emphasize(colour, 0.15),
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

export default withStyles(styles)(Oval);
