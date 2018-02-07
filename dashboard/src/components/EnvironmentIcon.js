import React from "react";
import PropTypes from "prop-types";

import { emphasize } from "material-ui/styles/colorManipulator";
import { withStyles } from "material-ui/styles";

const styles = () => ({
  root: {
    borderRadius: "100%",
    border: "1px solid",
    boxSizing: "border-box",
  },
});

class EnvironmentIcon extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    color: PropTypes.string,
    size: PropTypes.number,
  };

  static defaultProps = {
    className: null,
    color: "#8AB8D0",
    size: 8.0,
  };

  render() {
    const { classes, className, color, size, ...props } = this.props;
    const borderWidth = Math.floor(size * (1 / 8));
    const inlineStyle = {
      backgroundColor: color,
      borderColor: emphasize(color, 0.15),
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
