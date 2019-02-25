import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

import DonutSmall from "@material-ui/icons/DonutSmall";
import Explore from "@material-ui/icons/Explore";
import Visibility from "@material-ui/icons/Visibility";
import Emoticon from "@material-ui/icons/InsertEmoticon";
import Hot from "/lib/component/icon/Hot";
import Donut from "/lib/component/icon/Donut";
import Briefcase from "/lib/component/icon/Briefcase";
import Heart from "/lib/component/icon/Heart";
import HalfHeart from "/lib/component/icon/HalfHeart";
import HeartMug from "/lib/component/icon/HeartMug";
import Espresso from "/lib/component/icon/Espresso";
import Poly from "/lib/component/icon/Poly";

const icons = {
  BRIEFCASE: Briefcase,
  DONUTSM: DonutSmall,
  DONUT: Donut,
  EMOTICON: Emoticon,
  EXPLORE: Explore,
  FIRE: Hot,
  HEART: Heart,
  HALFHEART: HalfHeart,
  MUG: HeartMug,
  ESPRESSO: Espresso,
  POLYGON: Poly,
  VISIBILITY: Visibility,
};

const styles = theme => ({
  root: {
    display: "inline-flex",
    position: "relative",
    backgroundColor: theme.palette.primary.contrastText,
    color: theme.palette.primary.dark,
  },
  child: {
    position: "absolute",
    display: "inline-flex",
    alignSelf: "flex-end",
    bottom: 0,
    right: 0,
  },
});

class Icon extends React.PureComponent {
  static propTypes = {
    children: PropTypes.node,
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    icon: PropTypes.string.isRequired,
    size: PropTypes.number,
  };

  static defaultProps = {
    children: null,
    className: null,
    size: 24.0,
  };

  render() {
    const {
      children,
      classes,
      className: classNameProp,
      icon,
      size,
    } = this.props;

    // Classes
    const className = classnames(classNameProp, classes.root);
    const DisplayIcon = icons[icon];

    // Inline styles
    const iconStyles = {
      margin: `calc(${size}px * (1/12))`,
      height: `calc(${size}px * (5/6))`,
      width: `calc(${size}px * (5/6))`,
    };
    const containerStyles = {
      width: size,
      height: size,
      borderRadius: "100%",
    };

    return (
      <div className={className} style={containerStyles}>
        <DisplayIcon style={iconStyles} />
        {children && React.cloneElement(children, { className: classes.child })}
      </div>
    );
  }
}

export default withStyles(styles)(Icon);
