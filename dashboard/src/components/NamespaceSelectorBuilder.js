import React from "react";
import PropTypes from "prop-types";

import classNames from "classnames";
import Typography from "material-ui/Typography";
import { withStyles } from "material-ui/styles";

import arrowIcon from "material-ui-icons/ArrowDropDown";

const styles = theme => ({
  default: {
    // TODO come back to reassess typography
    fontFamily: "SF Pro Text",
    color: theme.palette.primary.contrastText,
    opacity: 0.9,
    display: "block",
  },
  org: {
    fontWeight: "lighter",
    fontSize: "0.75rem",
  },
  envContainer: {
    margin: "-5px 0 0",
    display: "flex",
    justifyContent: "space-between",
  },
  env: { fontSize: "1.25rem" },
  arrow: { color: theme.palette.primary.contrastText },
});

class NamespaceSelectorBuilder extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    DropdownArrow: PropTypes.func.isRequired,
    org: PropTypes.string.isRequired,
    env: PropTypes.string.isRequired,
  };

  static defaultProps = { DropdownArrow: arrowIcon };

  render() {
    const { classes, org, env, DropdownArrow } = this.props;

    return (
      <div className={classes.selectorContainer}>
        <Typography className={classNames(classes.default, classes.org)}>
          {org}
        </Typography>
        <div className={classes.envContainer}>
          <Typography className={classNames(classes.default, classes.env)}>
            {env}
          </Typography>
          <span className={classes.arrow}>
            <DropdownArrow />
          </span>
        </div>
      </div>
    );
  }
}

export default withStyles(styles)(NamespaceSelectorBuilder);
