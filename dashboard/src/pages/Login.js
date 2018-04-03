import React from "react";
import PropTypes from "prop-types";
import compose from "lodash/fp/compose";
import { withRouter, routerShape } from "found";
import { withStyles } from "material-ui/styles";

import Paper from "material-ui/Paper";
import Button from "material-ui/Button";
import TextField from "material-ui/TextField";
import Typography from "material-ui/Typography";
import AppRoot from "../components/AppRoot";
import Logo from "../icons/SensuLogoGraphic";
import Wordmark from "../icons/SensuWordmark";
import { authenticate } from "../utils/authentication";

// defaultRoute describe the location where the user will land after a
// successful login.
const defaultRoute = "/";

class Login extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    router: routerShape.isRequired,
  };

  static styles = theme => ({
    "@global html": {
      background: theme.palette.primary.main,
    },
    loginCard: {
      margin: "0 auto",
      padding: 48,
      maxWidth: 450,
      textAlign: "left",
      alignSelf: "center",
      height: "auto",
      minHeight: 500,
    },
    textField: {
      marginBottom: theme.spacing.unit,
    },
    actionsContainer: {
      marginTop: theme.spacing.unit * 5,
      textAlign: "right",
    },
    button: {
      textTransform: "uppercase",
    },
    icon: {
      color: theme.palette.secondary.main,
      marginBottom: theme.spacing.unit,
    },
    wordmark: {
      color: theme.palette.secondary.main,
      fontSize: theme.spacing.unit * 1.5,
      marginBottom: theme.spacing.unit,
      marginLeft: theme.spacing.unit / 2,
    },
    headline: {
      marginTop: theme.spacing.unit * 2,
    },
    form: {
      marginTop: theme.spacing.unit * 3,
    },
  });

  state = {
    username: null,
    password: null,
  };

  handleSubmit = ev => {
    const { username, password } = this.state;
    const { router } = this.props;

    const handleSuccess = () => router.replace(defaultRoute);
    const handleFailure = () => {
      this.setState({
        disabled: false,
        authError: "Bad username or password.",
      });
    };

    // Stop default form behaviour
    ev.preventDefault();

    // Disable form
    this.setState({ disabled: true });

    // Authenticate user
    const authPromise = authenticate(username, password);
    authPromise.then(handleSuccess).catch(handleFailure);
  };

  render() {
    const { classes } = this.props;
    const { authError, disabled } = this.state;

    const changeField = (name, val) =>
      this.setState({
        authError: null,
        [name]: val,
      });
    const handleUsername = ev => changeField("username", ev.target.value);
    const handlePassword = ev => this.setState({ password: ev.target.value });

    return (
      <AppRoot>
        <Paper className={classes.loginCard}>
          <Logo className={classes.icon} />
          <Wordmark className={classes.wordmark} />
          <div className={classes.headline}>
            <Typography variant="headline">Sign in</Typography>
            <Typography variant="subheading">
              with your Sensu Account
            </Typography>
          </div>
          <form className={classes.form} onSubmit={this.handleSubmit}>
            <TextField
              name="username"
              label="Username"
              aria-label="Username"
              autoComplete="username"
              autoCorrect="false"
              autoCapitalize="none"
              spellCheck={false}
              className={classes.textField}
              fullWidth
              margin="normal"
              onChange={handleUsername}
              disabled={disabled}
              error={!!authError}
            />
            <TextField
              type="password"
              name="password"
              label="Password"
              aria-label="Password"
              autoComplete="current-password"
              className={classes.textField}
              fullWidth
              onChange={handlePassword}
              disabled={disabled}
              error={!!authError}
              helperText={authError}
            />
            <div className={classes.actionsContainer}>
              <Button
                type="submit"
                color="primary"
                variant="raised"
                disabled={disabled}
                className={classes.button}
              >
                Log in
              </Button>
            </div>
          </form>
        </Paper>
      </AppRoot>
    );
  }
}

export default compose(withStyles(Login.styles), withRouter)(Login);
