import React from "react";
import PropTypes from "prop-types";

import { emphasize } from "material-ui/styles/colorManipulator";
import { withStyles } from "material-ui/styles";

import DonutSmallIcon from "material-ui-icons/DonutSmall";
import ExploreIcon from "material-ui-icons/Explore";
import VisibilityIcon from "material-ui-icons/Visibility";
import Heart from "../icons/Heart";
import HalfHeart from "../icons/HalfHeart";
import HeartMug from "../icons/HeartMug";
import Espresso from "../icons/Espresso";
import Poly from "../icons/Poly";

const icons = {
  DonutSmall: DonutSmallIcon,
  Explore: ExploreIcon,
  Visibility: VisibilityIcon,
  Heart,
  HalfHeart,
  HeartMug,
  Espresso,
  Poly,
};

const styles = theme => ({
  circle: {
    display: "inline-flex",
    position: "relative",
    backgroundColor: theme.palette.primary.contrastText,
    color: theme.palette.primary.dark,
  },
  smallCircle: {
    position: "absolute",
    display: "inline-flex",
    bottom: 0,
    right: 0,
  },
});

class OrganizationIcon extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    icon: PropTypes.string.isRequired,
    iconColor: PropTypes.string,
    size: PropTypes.number,
  };

  static defaultProps = { iconColor: "", size: 24 };

  render() {
    const { classes, icon, iconColor, size } = this.props;

    const mainIcon = {
      margin: `calc(${size}px * (1/12)`,
      height: `calc(${size}px * (5/6)`,
      width: `calc(${size}px * (5/6)`,
    };

    const circle = {
      width: size,
      height: size,
      borderRadius: "100%",
    };

    const smallCircle = {
      backgroundColor: iconColor,
      border: "1px solid",
      borderColor: emphasize(iconColor, 0.15),
      alignSelf: "flex-end",
      width: size / 3.0,
      height: size / 3.0,
      borderRadius: "100%",
    };

    const DisplayIcon = icons[icon];

    return (
      <div className={classes.circle} style={circle}>
        <DisplayIcon style={mainIcon} />
        <div className={classes.smallCircle} style={smallCircle} />
      </div>
    );
  }
}

export default withStyles(styles)(OrganizationIcon);
