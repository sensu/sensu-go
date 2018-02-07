import React from "react";
import PropTypes from "prop-types";

import classNames from "classnames";
import Typography from "material-ui/Typography";
import { withStyles } from "material-ui/styles";

import OrganizationIcon from "./OrganizationIcon";

const styles = theme => ({
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
    fontWeight: "lighter",
  },
  env: { margin: "0 4px", fontWeight: "bold" },
});

class NamespaceLabelBuilder extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    org: PropTypes.string.isRequired,
    env: PropTypes.string.isRequired,
    icon: PropTypes.string.isRequired,
    iconColor: PropTypes.string,
  };

  static defaultProps = { iconColor: "" };

  render() {
    const { classes, icon, iconColor, org, env } = this.props;

    return (
      <div className={classes.orgEnvContainer}>
        <Typography className={classNames(classes.default, classes.org)}>
          {org}
        </Typography>
        <Typography className={classNames(classes.default, classes.env)}>
          {env}
        </Typography>
        <OrganizationIcon icon={icon} iconColor={iconColor} />
      </div>
    );
  }
}

export default withStyles(styles)(NamespaceLabelBuilder);
