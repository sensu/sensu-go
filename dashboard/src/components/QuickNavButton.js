import React from "react";
import PropTypes from "prop-types";
import compose from "lodash/fp/compose";
import { withRouter, routerShape, Link } from "found";

import { withStyles } from "material-ui/styles";

import Typography from "material-ui/Typography";
import IconButton from "material-ui/IconButton";

const styles = theme => ({
  menuText: {
    color: "inherit",
    padding: "4px 0 0",
    fontSize: "0.6875rem",
    // TODO come back to reassess typography
    fontFamily: "SF Pro Text",
  },
  active: {
    color: `${theme.palette.secondary.main} !important`,
    opacity: "1 !important",
  },
  inactive: { color: theme.palette.secondary.contrastText, opacity: 0.71 },
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
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    Icon: PropTypes.func.isRequired,
    caption: PropTypes.string.isRequired,
    router: routerShape.isRequired,
    to: PropTypes.string.isRequired,
  };

  render() {
    const { classes, Icon, router, caption, ...props } = this.props;
    return (
      <Link
        Component={IconButton}
        classes={{
          root: classes.button,
          label: classes.label,
        }}
        to={props.to}
        role="button"
        tabIndex={0}
        activeClassName={classes.active}
        className={classes.inactive}
        {...props}
      >
        <Icon />
        <Typography classes={{ root: classes.menuText }}>{caption}</Typography>
      </Link>
    );
  }
}

export default compose(withStyles(styles), withRouter)(QuickNavButton);
