import React from "react";
import PropTypes from "prop-types";

import AppBar from "material-ui/AppBar";
import Avatar from "material-ui/Avatar";
import MaterialToolbar from "material-ui/Toolbar";
import Typography from "material-ui/Typography";
import IconButton from "material-ui/IconButton";
import MenuIcon from "material-ui-icons/Menu";
import { withStyles } from "material-ui/styles";

import AppSearch from "./AppSearch";

import logo from "../assets/sensu-logo-white.png";
import sampleAvatar from "../assets/sample-avatar.png";

const styles = theme => ({
  appBar: {
    transition: theme.transitions.create("width"),
  },
  title: {
    marginLeft: 24,
    flex: "0 1 auto",
  },
  grow: {
    flex: "1 1 auto",
  },
  logo: {
    height: 15,
    marginRight: theme.spacing.unit * 1,
    verticalAlign: "baseline",
  },
  avatar: {
    height: 32,
    width: 32,
    borderColor: "#fff",
    borderWidth: 1,
  },
  search: {
    marginRight: theme.spacing.unit,
  },
  hamburgerButton: {
    // [theme.breakpoints.up("sm")]: {
    //   display: "none",
    // },
  },
});

class Toolbar extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    toggleToolbar: PropTypes.func.isRequired,
  };

  render() {
    const { toggleToolbar, classes } = this.props;

    return (
      <AppBar className={classes.appBar}>
        <MaterialToolbar>
          <IconButton
            className={classes.hamburgerButton}
            onClick={toggleToolbar}
            aria-label="Menu"
            color="contrast"
          >
            <MenuIcon />
          </IconButton>
          <Typography
            className={classes.title}
            type="title"
            color="inherit"
            noWrap
          >
            <img alt="sensu logo" src={logo} className={classes.logo} />
            Sensu
          </Typography>
          <div className={classes.grow} />
          <AppSearch className={classes.search} />
          <IconButton>
            <Avatar className={classes.avatar} src={sampleAvatar} />
          </IconButton>
        </MaterialToolbar>
      </AppBar>
    );
  }
}

export default withStyles(styles)(Toolbar);
