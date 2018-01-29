import React from "react";
import PropTypes from "prop-types";

import { withRouter, routerShape } from "found";
import { withStyles } from "material-ui/styles";

import Typography from "material-ui/Typography";
import IconButton from "material-ui/IconButton";

const styles = {
  menuicon: {
    padding: "",
    width: 24,
    color: "rgba(21, 25,40, .71)",
  },
  menutext: {
    padding: "4px 0 0",
    fontSize: "0.6875rem",
    color: "#2D3555",
    fontFamily: "SF Pro Text",
  },
  label: {
    flexDirection: "column",
  },
  root: { width: 72, height: 72 },
};

class QuickNavButton extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    Icon: PropTypes.func.isRequired,
    primary: PropTypes.string.isRequired,
    router: routerShape.isRequired,
    href: PropTypes.string,
    onClick: PropTypes.func,
  };

  static defaultProps = {
    onClick: null,
    href: "",
  };

  render() {
    const { classes, Icon, router, primary, onClick, ...props } = this.props;
    const handleClick = () => this.props.router.push(this.props.href);

    return (
      <IconButton
        classes={{ root: classes.root, label: classes.label }}
        to={props.href}
        role="button"
        tabIndex={0}
        onClick={onClick || handleClick}
      >
        <Icon className={classes.menuicon} />
        <Typography className={classes.menutext} {...props.children}>
          {primary}
        </Typography>
      </IconButton>
    );
  }
}

export default withRouter(withStyles(styles)(QuickNavButton));
