import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "@material-ui/core/styles";
import { NavLink } from "react-router-dom";

import Typography from "@material-ui/core/Typography";
import IconButton from "@material-ui/core/IconButton";

const styles = theme => ({
  menuText: {
    color: "inherit",
    padding: "4px 0 0",
    fontSize: "0.6875rem",
  },
  active: {
    color: `${theme.palette.secondary.main} !important`,
    opacity: "1 !important",
  },
  link: {
    color: theme.typography.caption.color,
    fontFamily: "SF Pro Text", // TODO come back to reassess typography
  },
  label: {
    flexDirection: "column",
  },
  button: {
    width: 72,
    height: 72,
  },
});

class QuickNavButton extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    Icon: PropTypes.func.isRequired,
    caption: PropTypes.string.isRequired,
    to: PropTypes.string.isRequired,
    organization: PropTypes.string.isRequired,
    environment: PropTypes.string.isRequired,
    exact: PropTypes.bool,
  };

  static defaultProps = {
    exact: NavLink.defaultProps.exact,
  };

  render() {
    const {
      classes,
      Icon,
      caption,
      to,
      organization,
      environment,
      exact,
    } = this.props;

    return (
      <IconButton
        classes={{
          root: classes.button,
          label: classes.label,
        }}
        className={classes.link}
        component={NavLink}
        to={`/${organization}/${environment}/${to}`}
        activeClassName={classes.active}
        exact={exact}
      >
        <Icon />
        <Typography variant="caption" classes={{ root: classes.menuText }}>
          {caption}
        </Typography>
      </IconButton>
    );
  }
}

export default withStyles(styles)(QuickNavButton);
