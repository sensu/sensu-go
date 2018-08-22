import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import MUIAppBar from "@material-ui/core/AppBar";
import MaterialToolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";
import IconButton from "@material-ui/core/IconButton";
import { withStyles } from "@material-ui/core/styles";
import MenuIcon from "@material-ui/icons/Menu";
import EnvironmentLabel from "/components/EnvironmentLabel";
import Wordmark from "/icons/SensuWordmark";
import Drawer from "/components/Drawer";

class AppBar extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    viewer: PropTypes.object,
    environment: PropTypes.object,
    loading: PropTypes.bool.isRequired,
  };

  static defaultProps = { environment: null, viewer: null };

  static fragments = {
    viewer: gql`
      fragment AppBar_viewer on Viewer {
        ...Drawer_viewer
      }
      ${Drawer.fragments.viewer}
    `,

    environment: gql`
      fragment AppBar_environment on Environment {
        ...EnvironmentLabel_environment
        ...Drawer_environment
      }

      ${EnvironmentLabel.fragments.environment}
      ${Drawer.fragments.environment}
    `,
  };

  static styles = theme => ({
    container: {
      paddingTop: "env(safe-area-inset-top)",
      backgroundColor: theme.palette.primary.dark,
    },
    toolbar: {
      marginLeft: -12, // Account for button padding to match style guide.
      // marginRight: -12, // Label is not a button at this time.
      backgroundColor: theme.palette.primary.main,
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

  state = {
    drawerOpen: false,
  };

  handleToggleDrawer = () => {
    this.setState(state => ({ drawerOpen: !state.drawerOpen }));
  };

  render() {
    const { environment, viewer, loading, classes } = this.props;

    return (
      <React.Fragment>
        <MUIAppBar className={classes.appBar} position="static">
          <div className={classes.container}>
            <MaterialToolbar className={classes.toolbar}>
              <IconButton
                onClick={this.handleToggleDrawer}
                aria-label="Menu"
                color="inherit"
              >
                <MenuIcon />
              </IconButton>
              <Typography
                className={classes.title}
                variant="title"
                color="inherit"
                noWrap
              >
                <Wordmark alt="sensu logo" className={classes.logo} />
              </Typography>
              <div className={classes.grow} />
              {environment && <EnvironmentLabel environment={environment} />}
            </MaterialToolbar>
          </div>
        </MUIAppBar>
        <Drawer
          loading={loading}
          viewer={viewer}
          open={this.state.drawerOpen}
          onToggle={this.handleToggleDrawer}
          environment={environment}
          className={classes.drawer}
        />
      </React.Fragment>
    );
  }
}

export default withStyles(AppBar.styles)(AppBar);
