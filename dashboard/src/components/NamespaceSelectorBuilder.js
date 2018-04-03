import React from "react";
import PropTypes from "prop-types";
import classNames from "classnames";
import { withStyles } from "material-ui/styles";

import Typography from "material-ui/Typography";
import ArrowIcon from "material-ui-icons/ArrowDropDown";

const styles = theme => ({
  label: {
    color: theme.palette.primary.contrastText,
    opacity: 0.9,
    display: "block",
  },
  org: {
    fontWeight: "lighter",
    fontSize: "0.75rem",
  },
  envContainer: {
    margin: "-6px 0 0",
    display: "flex",
    justifyContent: "space-between",
  },
  env: { fontSize: "1.25rem" },
  arrow: { color: theme.palette.primary.contrastText },
});

class NamespaceSelectorBuilder extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    org: PropTypes.string.isRequired,
    env: PropTypes.string.isRequired,
  };

  render() {
    const { classes, org, env } = this.props;

    return (
      <div className={classes.selectorContainer}>
        <Typography className={classNames(classes.label, classes.org)}>
          {org}
        </Typography>
        <div className={classes.envContainer}>
          <Typography className={classNames(classes.label, classes.env)}>
            {env}
          </Typography>
          <span className={classes.arrow}>
            <ArrowIcon />
          </span>
        </div>
      </div>
    );
  }
}

export default withStyles(styles)(NamespaceSelectorBuilder);
