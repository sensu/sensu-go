import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

import Typography from "material-ui/Typography";
import IconButton from "material-ui/IconButton";
import NamespaceLink from "./NamespaceLink";

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
  };

  render() {
    const { classes, Icon, caption, ...props } = this.props;
    return (
      <NamespaceLink
        Component={IconButton}
        classes={{
          root: classes.button,
          label: classes.label,
        }}
        role="button"
        tabIndex={0}
        activeClassName={classes.active}
        className={classes.link}
        {...props}
      >
        <Icon />
        <Typography variant="caption" classes={{ root: classes.menuText }}>
          {caption}
        </Typography>
      </NamespaceLink>
    );
  }
}

export default withStyles(styles)(QuickNavButton);
