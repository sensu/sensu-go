import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import MUIAppBar from "material-ui/AppBar";
import MaterialToolbar from "material-ui/Toolbar";
import Typography from "material-ui/Typography";
import IconButton from "material-ui/IconButton";
import { withStyles } from "material-ui/styles";
import MenuIcon from "material-ui-icons/Menu";
import EnvironmentLabel from "./EnvironmentLabel";
import Wordmark from "../icons/SensuWordmark";

class AppBar extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    environment: PropTypes.object,
    toggleToolbar: PropTypes.func.isRequired,
  };

  static defaultProps = { environment: null };

  static fragments = {
    environment: gql`
      fragment AppBar_environment on Environment {
        ...EnvironmentLabel_environment
      }

      ${EnvironmentLabel.fragments.environment}
    `,
  };

  static styles = theme => ({
    container: {
      paddingTop: "env(safe-area-inset-top)",
      backgroundColor: theme.palette.primary.dark,
    },
    appBar: {
      transition: theme.transitions.create("width"),
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

  render() {
    const { environment, toggleToolbar, classes } = this.props;

    return (
      <MUIAppBar className={classes.appBar}>
        <div className={classes.container}>
          <MaterialToolbar className={classes.toolbar}>
            <IconButton
              onClick={toggleToolbar}
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
    );
  }
}

export default withStyles(AppBar.styles)(AppBar);
