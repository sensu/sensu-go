import React from "react";
import PropTypes from "prop-types";

import { emphasize } from "material-ui/styles/colorManipulator";
import { withStyles } from "material-ui/styles";

import DonutSmallIcon from "material-ui-icons/DonutSmall";
import ExploreIcon from "material-ui-icons/Explore";
import VisibilityIcon from "material-ui-icons/Visibility";

const icons = {
  DonutSmall: DonutSmallIcon,
  Explore: ExploreIcon,
  Visibility: VisibilityIcon,
};

const styles = theme => ({
  iconContainer: { display: "flex" },
  circle: {
    display: "inline-flex",
    backgroundColor: theme.palette.primary.contrastText,
    color: theme.palette.primary.dark,
  },
  smallCircle: {
    display: "inline-flex",
    margin: "16px 0 0 -8px",
  },
});

class OrganizationIcon extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    icon: PropTypes.string.isRequired,
    iconColor: PropTypes.string,
    iconSize: PropTypes.string,
  };

  static defaultProps = { iconColor: "", iconSize: "24" };

  render() {
    const { classes, icon, iconColor, iconSize } = this.props;

    const mainIcon = {
      margin: iconSize * 0.08,
      height: iconSize * 0.83,
      width: iconSize * 0.83,
    };

    const circle = {
      width: iconSize,
      height: iconSize,
      borderRadius: "100%",
    };

    const smallCircle = {
      backgroundColor: iconColor,
      border: "1px solid",
      borderColor: emphasize(iconColor, 0.15),
      alignSelf: "flex-end",
      width: iconSize / 3.1,
      height: iconSize / 3.1,
      borderRadius: iconSize / 3.1,
    };

    const DisplayIcon = icons[icon];

    return (
      <div className={classes.iconContainer}>
        <div className={classes.circle} style={circle}>
          <DisplayIcon style={mainIcon} />
        </div>
        <div className={classes.smallCircle} style={smallCircle} />
      </div>
    );
  }
}

export default withStyles(styles)(OrganizationIcon);
