import React from "react";
import compose from "lodash/fp/compose";
import PropTypes from "prop-types";
import { withRouter, routerShape } from "found";

import AppBar from "material-ui/AppBar";
import Avatar from "material-ui/Avatar";
import MaterialToolbar from "material-ui/Toolbar";
import Typography from "material-ui/Typography";
import IconButton from "material-ui/IconButton";
import Menu, { MenuItem } from "material-ui/Menu";
import { withStyles } from "material-ui/styles";
import MenuIcon from "material-ui-icons/Menu";
import AppSearch from "./AppSearch";
import { logout } from "../utils/authentication";

import logo from "../assets/wordmark-white.svg";
import sampleAvatar from "../assets/sample-avatar.png";

const styles = theme => ({
  appBar: {
    transition: theme.transitions.create("width"),
  },
  toolbar: {
    // paddingLeft: 12,
    // paddingRight: 12,
  },
  title: {
    marginLeft: 20,
    flex: "0 1 auto",
  },
  grow: {
    flex: "1 1 auto",
  },
  logo: {
    height: 16,
    marginRight: theme.spacing.unit * 1,
    verticalAlign: "baseline",
  },
  avatar: {
    height: 32,
    width: 32,
    borderColor: "#fff",
    borderWidth: 1,
    marginRight: -12,
  },
  search: {
    marginRight: theme.spacing.unit,
  },
  drawerButton: {
    marginLeft: -12,
  },
});

class Toolbar extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
    toggleToolbar: PropTypes.func.isRequired,
    router: routerShape.isRequired,
  };

  state = {
    menuAnchorEl: null,
    menuOpen: false,
  };

  //
  // Handlers

  handleMenuButtonClick = event => {
    this.setState({
      menuOpen: !this.state.menuOpen,
      menuAnchorEl: event.currentTarget,
    });
  };

  handleMenuRequestClose = () => {
    this.setState({ menuOpen: false });
  };

  handleLogout = async () => {
    await logout();
    this.props.router.push("/login");
  };

  //
  // Render

  render() {
    const { toggleToolbar, classes } = this.props;

    return (
      <AppBar className={classes.appBar}>
        <MaterialToolbar className={classes.toolbar}>
          <IconButton
            className={classes.drawerButton}
            onClick={toggleToolbar}
            aria-label="Menu"
            color="inherit"
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
          </Typography>
          <div className={classes.grow} />
          <AppSearch className={classes.search} />
          <IconButton
            aria-owns={this.state.menuOpen ? "profile-dropdown-menu" : null}
            aria-haspopup="true"
            onClick={this.handleMenuButtonClick}
          >
            <Avatar className={classes.avatar} src={sampleAvatar} />
          </IconButton>
          <Menu
            id="profile-dropdown-menu"
            anchorEl={this.state.menuAnchorEl}
            open={this.state.menuOpen}
            onClose={this.handleMenuRequestClose}
          >
            <MenuItem>Profile</MenuItem>
            <MenuItem>My account</MenuItem>
            <MenuItem onClick={this.handleLogout}>Logout</MenuItem>
          </Menu>
        </MaterialToolbar>
      </AppBar>
    );
  }
}

export default compose(withStyles(styles), withRouter)(Toolbar);
