import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

import DonutSmall from "material-ui-icons/DonutSmall";
import Explore from "material-ui-icons/Explore";
import Visibility from "material-ui-icons/Visibility";
import Emoticon from "material-ui-icons/InsertEmoticon";
import Briefcase from "../icons/Briefcase";
import Heart from "../icons/Heart";
import HalfHeart from "../icons/HalfHeart";
import HeartMug from "../icons/HeartMug";
import Espresso from "../icons/Espresso";
import Poly from "../icons/Poly";

import EnvironmentIcon from "./EnvironmentIcon";

const icons = {
  BRIEFCASE: Briefcase,
  DONUT: DonutSmall,
  EMOTICON: Emoticon,
  EXPLORE: Explore,
  HEART: Heart,
  HALFHEART: HalfHeart,
  MUG: HeartMug,
  ESPRESSO: Espresso,
  POLYGON: Poly,
  VISIBILITY: Visibility,
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
    alignSelf: "flex-end",
    bottom: 0,
    right: 0,
  },
});

class OrganizationIcon extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    icon: PropTypes.string.isRequired,
    iconColour: PropTypes.string.isRequired,
    size: PropTypes.number,
    disableEnvironmentIdicator: PropTypes.bool,
  };

  static defaultProps = {
    className: null,
    size: 24.0,
    disableEnvironmentIdicator: false,
  };

  render() {
    const {
      classes,
      className,
      icon,
      iconColour,
      size,
      disableEnvironmentIdicator,
    } = this.props;

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

    const DisplayIcon = icons[icon];

    return (
      <div className={`${className} ${classes.circle}`} style={circle}>
        <DisplayIcon style={mainIcon} />
        {!disableEnvironmentIdicator && (
          <EnvironmentIcon
            className={classes.smallCircle}
            colour={iconColour}
            size={size / 3.0}
          />
        )}
      </div>
    );
  }
}

export default withStyles(styles)(OrganizationIcon);
