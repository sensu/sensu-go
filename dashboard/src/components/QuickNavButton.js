import React from "react";
import PropTypes from "prop-types";
import compose from "lodash/fp/compose";
import { withRouter, routerShape } from "found";

import { withStyles } from "material-ui/styles";

import Typography from "material-ui/Typography";
import IconButton from "material-ui/IconButton";

const styles = {
  menuIcon: {
    // TODO theme colour for this
    color: "rgba(21, 25,40, .71)",
  },
  menuText: {
    padding: "4px 0 0",
    fontSize: "0.6875rem",
    // TODO come back to reassess typography
    fontFamily: "SF Pro Text",
    color: "theme.palette.primary.contrastTest",
  },
  // TODO theme colour for this
  active: { color: "rgba(151, 198, 115, 1)" },
  label: {
    flexDirection: "column",
  },
  buttonSize: {
    width: 72,
    height: 72,
  },
};

class QuickNavButton extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    Icon: PropTypes.func.isRequired,
    primary: PropTypes.string.isRequired,
    router: routerShape.isRequired,
    href: PropTypes.string,
    active: PropTypes.bool,
    onClick: PropTypes.func,
  };

  static defaultProps = {
    onClick: null,
    href: "",
    active: false,
  };

  render() {
    const { classes, Icon, router, primary, onClick, ...props } = this.props;
    const handleClick = () => this.props.router.push(this.props.href);
    return (
      <IconButton
        classes={{
          root: classes.buttonSize,
          label: classes.label,
        }}
        to={props.href}
        role="button"
        tabIndex={0}
        onClick={onClick || handleClick}
      >
        <Icon
          classes={{ root: classes.menuIcon }}
          className={props.active ? classes.active : null}
        />
        <Typography
          classes={{ root: classes.menuText }}
          className={props.active ? classes.active : null}
        >
          {primary}
        </Typography>
      </IconButton>
    );
  }
}

export default compose(withStyles(styles), withRouter)(QuickNavButton);
