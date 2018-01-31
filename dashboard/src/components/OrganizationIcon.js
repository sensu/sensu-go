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
  mainIcon: {
    margin: 2,
    height: 20,
    width: 20,
  },
  circle: {
    display: "inline-flex",
    width: 24,
    height: 24,
    borderRadius: 24,
    backgroundColor: theme.palette.primary.contrastText,
    color: theme.palette.primary.dark,
  },
  smallCircle: {
    display: "inline-flex",
    margin: "16px 0 0 -8px",
    width: 8,
    height: 8,
    borderRadius: 8,
  },
});

class OrganizationIcon extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    icon: PropTypes.string.isRequired,
    iconColor: PropTypes.string,
  };

  static defaultProps = { iconColor: "" };

  render() {
    const { classes, icon, iconColor } = this.props;
    const background = {
      backgroundColor: iconColor,
      border: "1px solid",
      borderColor: emphasize(iconColor, 0.15),
    };
    const DisplayIcon = icons[icon];

    return (
      <div className={classes.iconContainer}>
        <div className={classes.circle}>
          <DisplayIcon className={classes.mainIcon} />
        </div>
        <div className={classes.smallCircle} style={background} />
      </div>
    );
  }
}

export default withStyles(styles)(OrganizationIcon);
