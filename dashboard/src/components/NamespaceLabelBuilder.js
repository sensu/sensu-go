import React from "react";
import PropTypes from "prop-types";

import classNames from "classnames";
import Typography from "material-ui/Typography";
import { withStyles } from "material-ui/styles";

import Icon from "material-ui/Icon";
import DonutSmallIcon from "material-ui-icons/DonutSmall";
import ExploreIcon from "material-ui-icons/Explore";

import theme from "./Theme/Default";

const icons = { DonutSmall: DonutSmallIcon, Explore: ExploreIcon };

const styles = {
  orgEnvContainer: {
    display: "flex",
    alignSelf: "center",
  },
  default: {
    // TODO come back to reassess typography
    fontFamily: "SF Pro Text",
    color: theme.palette.primary.contrastText,
    opacity: 0.9,
    display: "inline",
    fontSize: "1rem",
  },
  org: {
    margin: "0 4px 0 0",
    fontWeight: "lighter",
  },
  env: { fontWeight: "bold" },
  icon: { margin: "0 0 0 4px" },
};

class OrganizationEnvironment extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    icon: PropTypes.string.isRequired,
    org: PropTypes.string.isRequired,
    env: PropTypes.string.isRequired,
  };

  render() {
    const { classes, icon, org, env } = this.props;
    const DisplayIcon = icons[icon];

    return (
      <div className={classes.orgEnvContainer}>
        <Typography className={classNames(classes.default, classes.org)}>
          {org}
        </Typography>
        <Typography className={classNames(classes.default, classes.env)}>
          {env}
        </Typography>
        <Icon className={classes.icon}>
          <DisplayIcon />
        </Icon>
      </div>
    );
  }
}

export default withStyles(styles)(OrganizationEnvironment);
