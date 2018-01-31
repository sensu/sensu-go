import React from "react";
import PropTypes from "prop-types";

import AppBar from "material-ui/AppBar";
import MaterialToolbar from "material-ui/Toolbar";
import Typography from "material-ui/Typography";
import IconButton from "material-ui/IconButton";
import { withStyles } from "material-ui/styles";

import MenuIcon from "material-ui-icons/Menu";

import logo from "../assets/logo/wordmark/white.svg";
import NamespaceLabel from "./NamespaceLabel";

const styles = theme => ({
  appBar: {
    transition: theme.transitions.create("width"),
  },
  toolbar: {
    marginLeft: -12, // Account for button padding to match style guide.
    marginRight: -12,
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
        <MaterialToolbar className={classes.toolbar}>
          <IconButton onClick={toggleToolbar} aria-label="Menu" color="inherit">
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
          {/* TODO should use some environment variables */}
          <NamespaceLabel className={classes.OrgEnv} />
        </MaterialToolbar>
      </AppBar>
    );
  }
}

export default withStyles(styles)(Toolbar);
